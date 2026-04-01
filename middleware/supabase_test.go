package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mustafaaksoy/gosupabase/auth"
)

const testSecret = "super-secret-jwt-token-for-testing"

func makeToken(claims map[string]interface{}, secret string) string {
	header := base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, _ := json.Marshal(claims)
	payloadEnc := base64URLEncode(payload)

	sigInput := header + "." + payloadEnc
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	sig := base64URLEncode(mac.Sum(nil))

	return sigInput + "." + sig
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func fixed32(b []byte) []byte {
	if len(b) >= 32 {
		return b[len(b)-32:]
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

func makeES256Token(t *testing.T, claims map[string]interface{}, kid string, priv *ecdsa.PrivateKey) string {
	t.Helper()
	headerJSON := fmt.Sprintf(`{"alg":"ES256","kid":"%s","typ":"JWT"}`, kid)
	header := base64URLEncode([]byte(headerJSON))
	payload, _ := json.Marshal(claims)
	payloadEnc := base64URLEncode(payload)
	sigInput := header + "." + payloadEnc
	sum := sha256.Sum256([]byte(sigInput))
	r, s, err := ecdsa.Sign(rand.Reader, priv, sum[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	sig := append(fixed32(r.Bytes()), fixed32(s.Bytes())...)
	return sigInput + "." + base64URLEncode(sig)
}

func resetJWKSCache() {
	jwksCacheMu.Lock()
	defer jwksCacheMu.Unlock()
	jwksCache = map[string]cachedJWKS{}
}

func TestSupabaseAuthHappyPath(t *testing.T) {
	claims := map[string]interface{}{
		"sub":   "user-123",
		"role":  "authenticated",
		"email": "test@example.com",
		"aud":   "authenticated",
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
	}
	token := makeToken(claims, testSecret)

	handler := SupabaseAuth(testSecret, "", "auto")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := auth.GetClaims(r.Context())
		if c == nil {
			t.Fatal("expected claims in context")
		}
		if c.Subject != "user-123" {
			t.Errorf("sub = %q, want %q", c.Subject, "user-123")
		}
		if c.Role != "authenticated" {
			t.Errorf("role = %q, want %q", c.Role, "authenticated")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSupabaseAuthMissingHeader(t *testing.T) {
	handler := SupabaseAuth(testSecret, "", "auto")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSupabaseAuthInvalidToken(t *testing.T) {
	handler := SupabaseAuth(testSecret, "", "auto")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSupabaseAuthWrongSecret(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token := makeToken(claims, "wrong-secret")

	handler := SupabaseAuth(testSecret, "", "auto")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSupabaseAuthExpiredToken(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
		"exp": float64(time.Now().Add(-time.Hour).Unix()),
	}
	token := makeToken(claims, testSecret)

	handler := SupabaseAuth(testSecret, "", "auto")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] != "token expired" {
		t.Errorf("error = %q, want %q", body["error"], "token expired")
	}
}

func TestRequireRolesAllowed(t *testing.T) {
	handler := RequireRoles("admin", "authenticated")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithClaims(req.Context(), &auth.Claims{Role: "admin"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireRolesDenied(t *testing.T) {
	handler := RequireRoles("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithClaims(req.Context(), &auth.Claims{Role: "authenticated"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestRequireRolesNoClaims(t *testing.T) {
	handler := RequireRoles("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRequireRolesAllowedFromRolesArray(t *testing.T) {
	handler := RequireRoles("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithClaims(req.Context(), &auth.Claims{Roles: []string{"member", "admin"}})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSupabaseAuthExtractsAppMetadataRoles(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"app_metadata": map[string]interface{}{
			"roles": []interface{}{"editor", "admin"},
		},
	}
	token := makeToken(claims, testSecret)

	authMw := SupabaseAuth(testSecret, "", "auto")
	roleMw := RequireRoles("admin")
	final := authMw(roleMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	final.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSupabaseAuthBadBearerPrefix(t *testing.T) {
	handler := SupabaseAuth(testSecret, "", "auto")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSupabaseAuthNonHS256Algorithm(t *testing.T) {
	header := base64URLEncode([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64URLEncode([]byte(`{"sub":"x"}`))
	sig := base64URLEncode([]byte("fake"))
	token := fmt.Sprintf("%s.%s.%s", header, payload, sig)

	handler := SupabaseAuth(testSecret, "", "auto")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSupabaseAuthModeJWKSRejectsHS256(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token := makeToken(claims, testSecret)
	handler := SupabaseAuth(testSecret, "", "jwks")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSupabaseAuthModeHS256AllowsHS256(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token := makeToken(claims, testSecret)
	handler := SupabaseAuth(testSecret, "", "hs256")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSupabaseAuthExtractsUserMetadataRolesString(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"user_metadata": map[string]interface{}{
			"roles": "admin",
		},
	}
	token := makeToken(claims, testSecret)
	authMw := SupabaseAuth(testSecret, "", "auto")
	roleMw := RequireRoles("admin")
	final := authMw(roleMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	final.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireRolesCaseInsensitive(t *testing.T) {
	handler := RequireRoles("Admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	ctx := auth.WithClaims(req.Context(), &auth.Claims{Role: "admin"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSupabaseAuthExtractsTopLevelRolesArray(t *testing.T) {
	claims := map[string]interface{}{
		"sub":   "user-123",
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"roles": []interface{}{"admin"},
	}
	token := makeToken(claims, testSecret)
	authMw := SupabaseAuth(testSecret, "", "auto")
	roleMw := RequireRoles("admin")
	final := authMw(roleMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	final.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestExtractRolesFromMultipleLocations(t *testing.T) {
	payload := []byte(`{
		"role":"authenticated",
		"roles":["editor"],
		"app_metadata":{"roles":["admin"]},
		"user_metadata":{"roles":"owner"}
	}`)
	got := extractRoles(payload, "authenticated", nil)
	joined := strings.Join(got, ",")
	for _, expected := range []string{"authenticated", "editor", "admin", "owner"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected role %q in %v", expected, got)
		}
	}
}

func TestAddInterfaceRolesIgnoresUnsupportedType(t *testing.T) {
	var got []string
	add := func(v string) { got = append(got, v) }
	addInterfaceRoles(123, add)
	if len(got) != 0 {
		t.Fatalf("expected no roles added, got %v", got)
	}
}

func TestSupabaseAuthES256WithJWKS(t *testing.T) {
	resetJWKSCache()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	kid := "test-kid-1"
	x := base64URLEncode(priv.PublicKey.X.Bytes())
	y := base64URLEncode(priv.PublicKey.Y.Bytes())

	var mu sync.Mutex
	jwksHits := 0
	supabase := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/.well-known/jwks.json" {
			http.NotFound(w, r)
			return
		}
		mu.Lock()
		jwksHits++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"keys":[{"kty":"EC","kid":"%s","crv":"P-256","x":"%s","y":"%s"}]}`, kid, x, y)
	}))
	defer supabase.Close()

	claims := map[string]interface{}{
		"sub": "es-user",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token := makeES256Token(t, claims, kid, priv)
	authMw := SupabaseAuth("", supabase.URL, "auto")
	next := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	rr1 := httptest.NewRecorder()
	next.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	rr2 := httptest.NewRecorder()
	next.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("second request status = %d, want 200", rr2.Code)
	}

	mu.Lock()
	defer mu.Unlock()
	if jwksHits != 1 {
		t.Fatalf("expected 1 JWKS fetch due to cache, got %d", jwksHits)
	}
}

func TestVerifyES256PartsInvalidSignature(t *testing.T) {
	resetJWKSCache()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	otherPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	kid := "bad-sig-kid"
	x := base64URLEncode(priv.PublicKey.X.Bytes())
	y := base64URLEncode(priv.PublicKey.Y.Bytes())
	supabase := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"keys":[{"kty":"EC","kid":"%s","crv":"P-256","x":"%s","y":"%s"}]}`, kid, x, y)
	}))
	defer supabase.Close()

	claims := map[string]interface{}{
		"sub": "user",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	// Sign with a different key than the key published in JWKS.
	token := makeES256Token(t, claims, kid, otherPriv)
	parts := strings.Split(token, ".")
	if _, err := verifyES256Parts(parts, kid, supabase.URL); err == nil {
		t.Fatal("expected invalid signature error")
	}
}

func TestFetchJWKSNon200(t *testing.T) {
	resetJWKSCache()
	supabase := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer supabase.Close()
	if _, err := fetchJWKS(supabase.URL); err == nil {
		t.Fatal("expected fetchJWKS to fail on non-200")
	}
}

func TestVerifyES256PartsInvalidLengthSignature(t *testing.T) {
	resetJWKSCache()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	kid := "len-kid"
	x := base64URLEncode(priv.PublicKey.X.Bytes())
	y := base64URLEncode(priv.PublicKey.Y.Bytes())
	supabase := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"keys":[{"kty":"EC","kid":"%s","crv":"P-256","x":"%s","y":"%s"}]}`, kid, x, y)
	}))
	defer supabase.Close()

	header := base64URLEncode([]byte(fmt.Sprintf(`{"alg":"ES256","kid":"%s","typ":"JWT"}`, kid)))
	payload := base64URLEncode([]byte(`{"sub":"x","exp":9999999999}`))
	invalidSig := base64URLEncode([]byte{1, 2, 3}) // not 64 bytes
	parts := []string{header, payload, invalidSig}
	if _, err := verifyES256Parts(parts, kid, supabase.URL); err == nil {
		t.Fatal("expected invalid signature length error")
	}
}

func TestFixed32PadsOrTrims(t *testing.T) {
	short := []byte{1, 2, 3}
	padded := fixed32(short)
	if len(padded) != 32 {
		t.Fatalf("len(padded) = %d, want 32", len(padded))
	}
	if new(big.Int).SetBytes(padded).Cmp(new(big.Int).SetBytes(short)) != 0 {
		t.Fatal("padded value changed numeric content")
	}
}

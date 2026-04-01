package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

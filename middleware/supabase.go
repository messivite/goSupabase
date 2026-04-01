package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mustafaaksoy/gosupabase/auth"
)

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type cachedJWKS struct {
	keysByKid map[string]*ecdsa.PublicKey
	expiresAt time.Time
}

var (
	jwksCacheMu sync.Mutex
	jwksCache   = map[string]cachedJWKS{}
)

// SupabaseAuth validates Supabase JWTs.
// - HS256: validates with SUPABASE_JWT_SECRET
// - ES256: validates with JWKS from SUPABASE_URL/.well-known/jwks.json
func SupabaseAuth(jwtSecret, supabaseURL, validationMode string) func(http.Handler) http.Handler {
	secretBytes := []byte(jwtSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "missing or invalid authorization header",
				})
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")

			claims, err := verifyToken(token, secretBytes, supabaseURL, validationMode)
			if err != nil {
				msg := "invalid token"
				if err.Error() == "token expired" {
					msg = "token expired"
				}
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": msg})
				return
			}

			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func verifyToken(token string, secret []byte, supabaseURL, validationMode string) (*auth.Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalid
	}

	headerJSON, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, errInvalid
	}
	var hdr jwtHeader
	if json.Unmarshal(headerJSON, &hdr) != nil {
		return nil, errInvalid
	}

	mode := strings.ToLower(strings.TrimSpace(validationMode))
	if mode == "" {
		mode = "auto"
	}

	switch hdr.Alg {
	case "HS256":
		if mode == "jwks" {
			return nil, errInvalid
		}
		return verifyHS256Parts(parts, secret)
	case "ES256":
		if mode == "hs256" {
			return nil, errInvalid
		}
		return verifyES256Parts(parts, hdr.Kid, supabaseURL)
	default:
		return nil, errInvalid
	}
}

// RequireRoles returns middleware that checks the JWT role claim against allowed roles.
func RequireRoles(allowed ...string) func(http.Handler) http.Handler {
	set := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		set[strings.ToLower(strings.TrimSpace(r))] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := auth.GetClaims(r.Context())
			if c == nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			if !hasAnyAllowedRole(c, set) {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasAnyAllowedRole(c *auth.Claims, allowed map[string]bool) bool {
	for _, role := range c.EffectiveRoles() {
		if allowed[strings.ToLower(strings.TrimSpace(role))] {
			return true
		}
	}
	return false
}

func verifyHS256Parts(parts []string, secret []byte) (*auth.Claims, error) {
	sig, err := base64URLDecode(parts[2])
	if err != nil {
		return nil, errInvalid
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, errInvalid
	}

	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, errInvalid
	}
	var claims auth.Claims
	if json.Unmarshal(payloadJSON, &claims) != nil {
		return nil, errInvalid
	}
	claims.Roles = extractRoles(payloadJSON, claims.Role, claims.Roles)

	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errExpired
	}

	return &claims, nil
}

func verifyES256Parts(parts []string, kid, supabaseURL string) (*auth.Claims, error) {
	if kid == "" || strings.TrimSpace(supabaseURL) == "" {
		return nil, errInvalid
	}

	pub, err := getECDSAPublicKey(supabaseURL, kid)
	if err != nil {
		return nil, errInvalid
	}

	sig, err := base64URLDecode(parts[2])
	if err != nil || len(sig) != 64 {
		return nil, errInvalid
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])

	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if !ecdsa.Verify(pub, sum[:], r, s) {
		return nil, errInvalid
	}

	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, errInvalid
	}
	var claims auth.Claims
	if json.Unmarshal(payloadJSON, &claims) != nil {
		return nil, errInvalid
	}
	claims.Roles = extractRoles(payloadJSON, claims.Role, claims.Roles)
	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errExpired
	}
	return &claims, nil
}

func getECDSAPublicKey(supabaseURL, kid string) (*ecdsa.PublicKey, error) {
	jwksCacheMu.Lock()
	cached, ok := jwksCache[supabaseURL]
	if ok && time.Now().Before(cached.expiresAt) {
		if key := cached.keysByKid[kid]; key != nil {
			jwksCacheMu.Unlock()
			return key, nil
		}
	}
	jwksCacheMu.Unlock()

	keysByKid, err := fetchJWKS(supabaseURL)
	if err != nil {
		return nil, err
	}

	jwksCacheMu.Lock()
	jwksCache[supabaseURL] = cachedJWKS{
		keysByKid: keysByKid,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	key := keysByKid[kid]
	jwksCacheMu.Unlock()
	if key == nil {
		return nil, errInvalid
	}
	return key, nil
}

func fetchJWKS(supabaseURL string) (map[string]*ecdsa.PublicKey, error) {
	url := strings.TrimRight(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errInvalid
	}

	var payload jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	keys := make(map[string]*ecdsa.PublicKey)
	for _, k := range payload.Keys {
		if k.Kid == "" || k.Kty != "EC" || k.Crv != "P-256" {
			continue
		}
		xBytes, errX := base64URLDecode(k.X)
		yBytes, errY := base64URLDecode(k.Y)
		if errX != nil || errY != nil {
			continue
		}
		pub := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
		keys[k.Kid] = pub
	}
	return keys, nil
}

func base64URLDecode(s string) ([]byte, error) {
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}

func extractRoles(payloadJSON []byte, role string, roles []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, 8)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			return
		}
		seen[v] = true
		out = append(out, v)
	}
	add(role)
	for _, r := range roles {
		add(r)
	}

	var raw map[string]interface{}
	if json.Unmarshal(payloadJSON, &raw) != nil {
		return out
	}
	addInterfaceRoles(raw["roles"], add)
	if appMeta, ok := raw["app_metadata"].(map[string]interface{}); ok {
		addInterfaceRoles(appMeta["roles"], add)
	}
	if userMeta, ok := raw["user_metadata"].(map[string]interface{}); ok {
		addInterfaceRoles(userMeta["roles"], add)
	}
	return out
}

func addInterfaceRoles(v interface{}, add func(string)) {
	switch x := v.(type) {
	case string:
		add(x)
	case []interface{}:
		for _, item := range x {
			if s, ok := item.(string); ok {
				add(s)
			}
		}
	}
}

type tokenError string

func (e tokenError) Error() string { return string(e) }

var (
	errInvalid = tokenError("invalid token")
	errExpired = tokenError("token expired")
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

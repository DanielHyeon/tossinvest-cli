package httpapi

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/networkboundary"
)

const (
	CapabilityVersion = "tossos-httpapi-capability/v1"
	CapabilityTTL     = 60 * time.Second
	maxCapabilitySize = 8 << 10
)

var identityComponent = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:@/-]{0,127}$`)

var ErrCapabilityInvalid = errors.New("httpapi: capability is invalid")

// CapabilityClaims is the complete signed mutation authority. It intentionally
// grants only one exact request and carries no route family or wildcard.
type CapabilityClaims struct {
	Version        string    `json:"version"`
	Nonce          string    `json:"nonce"`
	Actor          string    `json:"actor"`
	Client         string    `json:"client"`
	Method         string    `json:"method"`
	Resource       string    `json:"resource"`
	BodyDigest     string    `json:"body_digest"`
	IdempotencyKey string    `json:"idempotency_key"`
	IfMatch        string    `json:"if_match"`
	Audience       string    `json:"audience"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type CapabilityBinding struct {
	Method         string
	Resource       string
	BodyDigest     string
	IdempotencyKey string
	IfMatch        string
	Audience       string
}

type CapabilityVerifier struct {
	publicKey ed25519.PublicKey
	now       func() time.Time
}

func NewCapabilityVerifier(publicKey ed25519.PublicKey, now func() time.Time) (*CapabilityVerifier, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%w: Ed25519 public key has the wrong size", ErrCapabilityInvalid)
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &CapabilityVerifier{publicKey: append(ed25519.PublicKey(nil), publicKey...), now: now}, nil
}

func SignCapability(privateKey ed25519.PrivateKey, claims CapabilityClaims) (string, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("%w: Ed25519 private key has the wrong size", ErrCapabilityInvalid)
	}
	if err := validateCapabilityClaims(claims); err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("%w: encoding claims: %v", ErrCapabilityInvalid, err)
	}
	signature := ed25519.Sign(privateKey, payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (v *CapabilityVerifier) Verify(token string, binding CapabilityBinding) (CapabilityClaims, error) {
	var zero CapabilityClaims
	if v == nil || len(token) == 0 || len(token) > maxCapabilitySize || strings.Count(token, ".") != 1 {
		return zero, ErrCapabilityInvalid
	}
	parts := strings.SplitN(token, ".", 2)
	if strings.Contains(parts[0], "=") || strings.Contains(parts[1], "=") {
		return zero, ErrCapabilityInvalid
	}
	payload, err := base64.RawURLEncoding.Strict().DecodeString(parts[0])
	if err != nil {
		return zero, ErrCapabilityInvalid
	}
	signature, err := base64.RawURLEncoding.Strict().DecodeString(parts[1])
	if err != nil || len(signature) != ed25519.SignatureSize || !ed25519.Verify(v.publicKey, payload, signature) {
		return zero, ErrCapabilityInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var claims CapabilityClaims
	if err := decoder.Decode(&claims); err != nil {
		return zero, ErrCapabilityInvalid
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return zero, ErrCapabilityInvalid
	}
	canonical, err := json.Marshal(claims)
	if err != nil || !bytes.Equal(canonical, payload) {
		return zero, ErrCapabilityInvalid
	}
	if err := validateCapabilityClaims(claims); err != nil {
		return zero, err
	}
	now := v.now().UTC()
	if now.Before(claims.IssuedAt.UTC()) || !now.Before(claims.ExpiresAt.UTC()) {
		return zero, ErrCapabilityInvalid
	}
	if claims.Method != binding.Method || claims.Resource != binding.Resource ||
		claims.BodyDigest != binding.BodyDigest || claims.IdempotencyKey != binding.IdempotencyKey ||
		claims.IfMatch != binding.IfMatch || claims.Audience != binding.Audience {
		return zero, ErrCapabilityInvalid
	}
	return claims, nil
}

func validateCapabilityClaims(claims CapabilityClaims) error {
	if claims.Version != CapabilityVersion || !identityComponent.MatchString(claims.Nonce) ||
		!identityComponent.MatchString(claims.Actor) || !identityComponent.MatchString(claims.Client) ||
		claims.Method != "POST" || !validResource(claims.Resource) ||
		!validDigest(claims.BodyDigest) || !validIdempotencyKey(claims.IdempotencyKey) || !validIfMatch(claims.IfMatch) {
		return ErrCapabilityInvalid
	}
	audience, err := networkboundary.ParseOrigin(claims.Audience)
	if err != nil || audience.Scheme != "https" || audience.String() != claims.Audience {
		return ErrCapabilityInvalid
	}
	issued := claims.IssuedAt.UTC()
	expires := claims.ExpiresAt.UTC()
	if issued.IsZero() || expires.IsZero() || !expires.After(issued) || expires.Sub(issued) > CapabilityTTL {
		return ErrCapabilityInvalid
	}
	return nil
}

func validResource(resource string) bool {
	return strings.HasPrefix(resource, "/api/v1/") && !strings.ContainsAny(resource, "?#\\\x00\r\n") &&
		!strings.Contains(resource, "//") && !strings.Contains(resource, "..") && len(resource) <= 256
}

func validDigest(digest string) bool {
	if len(digest) != 64 || strings.ToLower(digest) != digest {
		return false
	}
	decoded, err := hex.DecodeString(digest)
	return err == nil && len(decoded) == 32
}

func validIdempotencyKey(key string) bool {
	return len(key) >= 16 && len(key) <= 128 && identityComponent.MatchString(key)
}

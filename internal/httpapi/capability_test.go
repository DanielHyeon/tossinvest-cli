package httpapi

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestCapabilityIsBoundToEveryMutationIdentityComponent(t *testing.T) {
	t.Parallel()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC)
	bodyDigest := sha256.Sum256([]byte(`{"preset":"safe"}`))
	claims := CapabilityClaims{
		Version: CapabilityVersion, Nonce: "nonce-0123456789abcdef",
		Actor: "operator:local", Client: "ios:device-a", Method: "POST",
		Resource: "/api/v1/optimization/previews", BodyDigest: hex.EncodeToString(bodyDigest[:]),
		IdempotencyKey: "idem-0123456789abcdef", IfMatch: `"7"`, Audience: "https://localhost:443",
		IssuedAt: now, ExpiresAt: now.Add(CapabilityTTL),
	}
	token, err := SignCapability(privateKey, claims)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewCapabilityVerifier(publicKey, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	want := CapabilityBinding{
		Method: claims.Method, Resource: claims.Resource, BodyDigest: claims.BodyDigest,
		IdempotencyKey: claims.IdempotencyKey, IfMatch: claims.IfMatch, Audience: claims.Audience,
	}
	verified, err := verifier.Verify(token, want)
	if err != nil {
		t.Fatalf("valid capability: %v", err)
	}
	if verified.Actor != claims.Actor || verified.Client != claims.Client || verified.Nonce != claims.Nonce {
		t.Fatalf("verified claims drift: %+v", verified)
	}

	mutations := []func(*CapabilityBinding){
		func(v *CapabilityBinding) { v.Method = "PATCH" },
		func(v *CapabilityBinding) { v.Resource += "/other" },
		func(v *CapabilityBinding) { v.BodyDigest = string(make([]byte, 64)) },
		func(v *CapabilityBinding) { v.IdempotencyKey += "-other" },
		func(v *CapabilityBinding) { v.IfMatch = `"8"` },
		func(v *CapabilityBinding) { v.Audience = "https://other.vpn.test:443" },
	}
	for i, mutate := range mutations {
		changed := want
		mutate(&changed)
		if _, err := verifier.Verify(token, changed); err == nil {
			t.Fatalf("binding mutation %d was accepted", i)
		}
	}
}

func TestCapabilityRefusesExpiryOverSixtySecondsAndBoundaryReuse(t *testing.T) {
	t.Parallel()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC)
	base := CapabilityClaims{
		Version: CapabilityVersion, Nonce: "nonce-0123456789abcdef", Actor: "operator:local", Client: "ios:a",
		Method: "POST", Resource: "/api/v1/optimization/previews", BodyDigest: strings.Repeat("0", 64),
		IdempotencyKey: "idem-0123456789abcdef", IfMatch: `"7"`, Audience: "https://localhost:443",
		IssuedAt: now, ExpiresAt: now.Add(CapabilityTTL),
	}
	if _, err := SignCapability(privateKey, func() CapabilityClaims { c := base; c.ExpiresAt = now.Add(CapabilityTTL + time.Nanosecond); return c }()); err == nil {
		t.Fatal("capability wider than 60 seconds was signed")
	}
	token, err := SignCapability(privateKey, base)
	if err != nil {
		t.Fatal(err)
	}
	binding := CapabilityBinding{Method: base.Method, Resource: base.Resource, BodyDigest: base.BodyDigest,
		IdempotencyKey: base.IdempotencyKey, IfMatch: base.IfMatch, Audience: base.Audience}
	verifier, _ := NewCapabilityVerifier(publicKey, func() time.Time { return now.Add(CapabilityTTL) })
	if _, err := verifier.Verify(token, binding); err == nil {
		t.Fatal("capability was accepted at its exclusive expiry boundary")
	}
	verifier, _ = NewCapabilityVerifier(publicKey, func() time.Time { return now.Add(-time.Nanosecond) })
	if _, err := verifier.Verify(token, binding); err == nil {
		t.Fatal("capability was accepted before issuance")
	}
}

func TestCapabilityRejectsMalleableOrRepeatedTokenEncoding(t *testing.T) {
	t.Parallel()
	publicKey, _, _ := ed25519.GenerateKey(rand.Reader)
	verifier, _ := NewCapabilityVerifier(publicKey, time.Now)
	for _, token := range []string{"", "a", "a.b.c", "a=.b", "a..b"} {
		if _, err := verifier.Verify(token, CapabilityBinding{}); err == nil {
			t.Fatalf("malformed token %q was accepted", token)
		}
	}
}

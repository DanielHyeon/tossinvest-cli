package attest

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// testProtectionSigner is deliberately test-only. Production code contains
// verification primitives and no private-key generation, storage, or writer.
type testProtectionSigner struct {
	TrustKey ProtectionTrustKey
	Private  ed25519.PrivateKey
}

func newTestProtectionSigner(t *testing.T, keyID string) testProtectionSigner {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return testProtectionSigner{
		TrustKey: ProtectionTrustKey{
			KeyID:     keyID,
			Role:      ProtectionSignerRole,
			Algorithm: ProtectionSignatureAlgorithm,
			PublicKey: base64.RawURLEncoding.EncodeToString(public),
			NotBefore: protectionNow.Add(-24 * time.Hour),
			NotAfter:  protectionNow.Add(7 * 24 * time.Hour),
			Status:    ProtectionKeyActive,
		},
		Private: private,
	}
}

func signedTestEnvelope(t *testing.T, m ProtectionCapabilityMatrix, signer testProtectionSigner) protectionSignedEnvelope {
	t.Helper()
	payload, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	envelope := protectionSignedEnvelope{
		EnvelopeVersion: ProtectionEnvelopeVersion,
		Domain:          ProtectionSignatureDomain,
		Algorithm:       ProtectionSignatureAlgorithm,
		KeyID:           signer.TrustKey.KeyID,
		Payload:         base64.RawURLEncoding.EncodeToString(payload),
	}
	envelope.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(signer.Private, protectionSigningMessage(envelope)))
	return envelope
}

func writeSignedProtectionMatrix(t *testing.T, m ProtectionCapabilityMatrix, signer testProtectionSigner, root ProtectionTrustRoot) string {
	t.Helper()
	base := t.TempDir()
	runtimeDir := filepath.Join(base, "runtime")
	trustDir := filepath.Join(base, "trust")
	policyDir := filepath.Join(base, "policy")
	if err := os.Mkdir(runtimeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(trustDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(policyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if root.Generation == 0 {
		root.Generation = 1
	}
	sort.Slice(root.Keys, func(i, j int) bool { return root.Keys[i].KeyID < root.Keys[j].KeyID })
	rootPath := filepath.Join(trustDir, ProtectionTrustRootFileName)
	attestationPath := filepath.Join(runtimeDir, ProtectionFileName)
	writeCanonicalJSONFile(t, rootPath, root, 0o444)
	writeCanonicalJSONFile(t, attestationPath, signedTestEnvelope(t, m, signer), 0o600)
	rootData, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	owner, ok := currentProtectionOwnerUID()
	if !ok {
		t.Skip("protection ownership metadata unavailable")
	}
	writeCanonicalJSONFile(t, filepath.Join(policyDir, ProtectionVerifierPolicyFileName), protectionVerifierPolicy{
		FormatVersion: ProtectionVerifierPolicyVersion, Generation: root.Generation,
		AttestationPath: attestationPath, TrustRootPath: rootPath, TrustRootOwner: owner,
		TrustRootDigest: protectionDigest(rootData),
	}, 0o444)
	return attestationPath
}

func newTestProtectionVerifier(path string, now time.Time) *ProtectionVerifier {
	owner, _ := currentProtectionOwnerUID()
	return &ProtectionVerifier{source: protectionPolicySource{path: testProtectionPolicyPath(path), owner: fileIdentity{UID: owner}}, clock: func() time.Time { return now }}
}

func parseTestProtectionCapability(path string) (parsedProtectionCapability, error) {
	return newTestProtectionVerifier(path, protectionNow).parse()
}

func VerifyProtectionCapability(parsed parsedProtectionCapability, now time.Time, scope ProtectionScope, evidence map[string][]byte) (VerifiedProtectionCapability, error) {
	parsed.verifier.clock = func() time.Time { return now }
	return parsed.verifier.verifyParsed(parsed, scope, evidence, nil)
}

func ParseProtectionCapability(path string, policy protectionVerifierPolicy) (parsedProtectionCapability, error) {
	policy.AttestationPath = path
	if err := rewriteTestVerifierPolicyFromTrust(path, policy); err != nil {
		return parsedProtectionCapability{}, err
	}
	return newTestProtectionVerifier(path, protectionNow).parse()
}

func LoadProtectionCapability(path string, policy protectionVerifierPolicy, now time.Time, scope ProtectionScope, evidence map[string][]byte) (VerifiedProtectionCapability, error) {
	parsed, err := ParseProtectionCapability(path, policy)
	if err != nil {
		return VerifiedProtectionCapability{}, err
	}
	return VerifyProtectionCapability(parsed, now, scope, evidence)
}

func parseProtectionCapability(path string, owner fileIdentity, policy protectionVerifierPolicy) (parsedProtectionCapability, error) {
	if err := rewriteTestVerifierPolicyFromTrust(path, policy); err != nil {
		return parsedProtectionCapability{}, err
	}
	verifier := newTestProtectionVerifier(path, protectionNow)
	loaded, err := loadProtectionVerifierPolicy(verifier.source, nil)
	if err != nil {
		return parsedProtectionCapability{}, err
	}
	return parseSignedProtectionCapability(verifier, loaded, owner)
}

func testProtectionTrustRootPath(path string) string {
	return filepath.Join(filepath.Dir(filepath.Dir(path)), "trust", ProtectionTrustRootFileName)
}

func testProtectionPolicyPath(path string) string {
	return filepath.Join(filepath.Dir(filepath.Dir(path)), "policy", ProtectionVerifierPolicyFileName)
}

func testProtectionTrustPolicy(path string) protectionVerifierPolicy {
	data, err := os.ReadFile(testProtectionPolicyPath(path))
	if err == nil {
		var policy protectionVerifierPolicy
		if json.Unmarshal(data, &policy) == nil {
			return policy
		}
	}
	owner, _ := currentProtectionOwnerUID()
	return protectionVerifierPolicy{FormatVersion: ProtectionVerifierPolicyVersion, Generation: 1, AttestationPath: path, TrustRootPath: testProtectionTrustRootPath(path), TrustRootOwner: owner}
}

func rewriteTestVerifierPolicyFromTrust(path string, policy protectionVerifierPolicy) error {
	policy.FormatVersion = ProtectionVerifierPolicyVersion
	if policy.Generation == 0 {
		policy.Generation = 1
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	policyPath := testProtectionPolicyPath(path)
	if err := os.Chmod(policyPath, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(policyPath, data, 0o644); err != nil {
		return err
	}
	return os.Chmod(policyPath, 0o444)
}

func rewriteTestRootAndPolicy(t *testing.T, path string, root ProtectionTrustRoot) {
	t.Helper()
	rootPath := testProtectionTrustRootPath(path)
	if err := os.Chmod(rootPath, 0o644); err != nil {
		t.Fatal(err)
	}
	writeCanonicalJSONFile(t, rootPath, root, 0o644)
	if err := os.Chmod(rootPath, 0o444); err != nil {
		t.Fatal(err)
	}
	rootData, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	policy := testProtectionTrustPolicy(path)
	policy.Generation = root.Generation
	policy.TrustRootDigest = protectionDigest(rootData)
	if err := rewriteTestVerifierPolicyFromTrust(path, policy); err != nil {
		t.Fatal(err)
	}
}

func writeCanonicalJSONFile(t *testing.T, path string, value any, mode os.FileMode) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}
}

func readTestEnvelope(t *testing.T, path string) protectionSignedEnvelope {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var envelope protectionSignedEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope
}

func rewriteTestEnvelope(t *testing.T, path string, envelope protectionSignedEnvelope) {
	t.Helper()
	writeCanonicalJSONFile(t, path, envelope, 0o600)
}

func TestProtectionSignedEnvelopeVerifies(t *testing.T) {
	path := writeProtectionMatrix(t, validProtectionMatrix(t))
	policy := testProtectionTrustPolicy(path)
	parsed, err := parseTestProtectionCapability(path)
	if err != nil {
		t.Fatalf("ParseProtectionCapability: %v", err)
	}
	if _, err := VerifyProtectionCapability(parsed, protectionNow, validProtectionScope(), cloneEvidence()); err != nil {
		t.Fatalf("VerifyProtectionCapability: %v", err)
	}
	if _, err := LoadProtectionCapability(path, policy, protectionNow, validProtectionScope(), cloneEvidence()); err != nil {
		t.Fatalf("LoadProtectionCapability: %v", err)
	}
}

func TestProtectionVerifierIsSealedAndRechecksHardRevocation(t *testing.T) {
	if _, err := new(ProtectionVerifier).Verify(validProtectionScope(), cloneEvidence()); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("zero verifier = %v, want ErrProtectionTrust", err)
	}
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 1, Keys: []ProtectionTrustKey{signer.TrustKey}}
	path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
	verifier := newTestProtectionVerifier(path, protectionNow)
	parsed, err := verifier.parse()
	if err != nil {
		t.Fatal(err)
	}
	revoked := signer.TrustKey
	revokedAt := protectionNow.Add(-time.Minute)
	revoked.Status, revoked.RevokedAt, revoked.RevocationReason = ProtectionKeyRevoked, &revokedAt, "hard revoke during authorization"
	rewriteTestRootAndPolicy(t, path, ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 2, Keys: []ProtectionTrustKey{revoked}})
	if _, err := verifier.verifyParsed(parsed, validProtectionScope(), cloneEvidence(), nil); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("parse-before-revoke authorization = %v, want ErrProtectionTrust", err)
	}
}

func TestProtectionVerifierRechecksRevocationAfterEvidenceValidation(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 1, Keys: []ProtectionTrustKey{signer.TrustKey}}
	path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
	verifier := newTestProtectionVerifier(path, protectionNow)
	parsed, err := verifier.parse()
	if err != nil {
		t.Fatal(err)
	}
	revoked := signer.TrustKey
	revokedAt := protectionNow.Add(-time.Minute)
	revoked.Status, revoked.RevokedAt, revoked.RevocationReason = ProtectionKeyRevoked, &revokedAt, "revoke while evidence is validated"
	if _, err := verifier.verifyParsed(parsed, validProtectionScope(), cloneEvidence(), func() {
		rewriteTestRootAndPolicy(t, path, ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 2, Keys: []ProtectionTrustKey{revoked}})
	}); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("revocation after evidence validation = %v, want ErrProtectionTrust", err)
	}
}

func TestProtectionVerifierRejectsPolicyGenerationRollbackAndReuse(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 2, Keys: []ProtectionTrustKey{signer.TrustKey}}
	path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
	verifier := newTestProtectionVerifier(path, protectionNow)
	parsed, err := verifier.parse()
	if err != nil {
		t.Fatal(err)
	}
	policy := testProtectionTrustPolicy(path)
	policy.Generation = 1
	if err := rewriteTestVerifierPolicyFromTrust(path, policy); err != nil {
		t.Fatal(err)
	}
	if _, err := verifier.verifyParsed(parsed, validProtectionScope(), cloneEvidence(), nil); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("generation rollback = %v", err)
	}
	policy.Generation = 2
	policy.TrustRootDigest = digestBytes([]byte("same generation replacement"))
	if err := rewriteTestVerifierPolicyFromTrust(path, policy); err != nil {
		t.Fatal(err)
	}
	if _, err := verifier.verifyParsed(parsed, validProtectionScope(), cloneEvidence(), nil); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("generation reuse = %v", err)
	}
}

func TestProtectionVerifierRejectsRollbackAcrossDirectVerifyCalls(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 2, Keys: []ProtectionTrustKey{signer.TrustKey}}
	path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
	verifier := newTestProtectionVerifier(path, protectionNow)
	if _, err := verifier.Verify(validProtectionScope(), cloneEvidence()); err != nil {
		t.Fatalf("initial verify: %v", err)
	}

	root.Generation = 1
	rewriteTestRootAndPolicy(t, path, root)
	if _, err := verifier.Verify(validProtectionScope(), cloneEvidence()); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("cross-call generation rollback = %v, want ErrProtectionTrust", err)
	}

	root.Generation = 2
	secondary := newTestProtectionSigner(t, "secondary")
	root.Keys = append(root.Keys, secondary.TrustKey)
	rewriteTestRootAndPolicy(t, path, root)
	if _, err := verifier.Verify(validProtectionScope(), cloneEvidence()); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("cross-call generation reuse = %v, want ErrProtectionTrust", err)
	}
}

func TestProtectionVerifierDoesNotLatchUnverifiedHigherGeneration(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 1, Keys: []ProtectionTrustKey{signer.TrustKey}}
	path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
	verifier := newTestProtectionVerifier(path, protectionNow)
	if _, err := verifier.Verify(validProtectionScope(), cloneEvidence()); err != nil {
		t.Fatalf("initial verify: %v", err)
	}

	rewriteTestRootAndPolicy(t, path, ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 2})
	if _, err := verifier.Verify(validProtectionScope(), cloneEvidence()); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("unverified higher generation = %v, want ErrProtectionTrust", err)
	}
	rewriteTestRootAndPolicy(t, path, root)
	if _, err := verifier.Verify(validProtectionScope(), cloneEvidence()); err != nil {
		t.Fatalf("valid generation poisoned by invalid higher policy: %v", err)
	}
}

func TestVerifiedProtectionCapabilityContainsOnlyMatchedAuthority(t *testing.T) {
	matrix := validProtectionMatrix(t)
	extra := matrix.Capabilities[0]
	extra.AccountRef = "99999999"
	matrix.Capabilities = append(matrix.Capabilities, extra)
	redigestMatrix(t, &matrix)
	path := writeProtectionMatrix(t, matrix)
	verified, err := newTestProtectionVerifier(path, protectionNow).Verify(validProtectionScope(), cloneEvidence())
	if err != nil {
		t.Fatal(err)
	}
	if got := verified.Capability(); got.AccountRef != "12345678" || got.Replace.Mode != ReplaceAtomic {
		t.Fatalf("matched capability = %+v", got)
	}
	scope := verified.Scope()
	scope.Tools[ToolVerifyExecutionCapability] = ToolBuild{}
	if verified.Scope().Tools[ToolVerifyExecutionCapability] == (ToolBuild{}) {
		t.Fatal("verified scope exposed mutable tool authority")
	}
}

func TestProtectionSignedEnvelopeRequiresTrustRootAndSignature(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, ProtectionTrustRoot{
		FormatVersion: ProtectionTrustRootFormatVersion,
		Keys:          []ProtectionTrustKey{signer.TrustKey},
	})

	policy := testProtectionTrustPolicy(path)
	if err := os.Remove(policy.TrustRootPath); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseProtectionCapability(path, policy); !errors.Is(err, ErrProtectionTrust) {
		t.Fatalf("missing trust root error = %v, want ErrProtectionTrust", err)
	}

	path = writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, ProtectionTrustRoot{
		FormatVersion: ProtectionTrustRootFormatVersion,
		Keys:          []ProtectionTrustKey{signer.TrustKey},
	})
	envelope := readTestEnvelope(t, path)
	signature, _ := base64.RawURLEncoding.DecodeString(envelope.Signature)
	signature[0] ^= 0x80
	envelope.Signature = base64.RawURLEncoding.EncodeToString(signature)
	rewriteTestEnvelope(t, path, envelope)
	if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionSignature) {
		t.Fatalf("forged signature error = %v, want ErrProtectionSignature", err)
	}
}

func TestProtectionParser_DistinctNonRootDirectories(t *testing.T) {
	path := writeProtectionMatrix(t, validProtectionMatrix(t))
	policy := testProtectionTrustPolicy(path)
	if policy.TrustRootOwner == 0 {
		t.Skip("runner is root; non-root layout is covered on non-root CI")
	}
	if _, err := ParseProtectionCapability(path, policy); err != nil {
		t.Fatalf("separated rootless layout: %v", err)
	}
}

func TestProtectionTrustPolicyRejectsSameParentAndRuntimeUIDReplacement(t *testing.T) {
	t.Run("same parent", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix(t))
		rootData, err := os.ReadFile(testProtectionTrustRootPath(path))
		if err != nil {
			t.Fatal(err)
		}
		sameParentRoot := filepath.Join(filepath.Dir(path), ProtectionTrustRootFileName)
		if err := os.WriteFile(sameParentRoot, rootData, 0o444); err != nil {
			t.Fatal(err)
		}
		owner, _ := currentProtectionOwnerUID()
		policy := protectionVerifierPolicy{FormatVersion: ProtectionVerifierPolicyVersion, Generation: 1, AttestationPath: path, TrustRootPath: sameParentRoot, TrustRootOwner: owner, TrustRootDigest: protectionDigest(rootData)}
		if _, err := ParseProtectionCapability(path, policy); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("same-parent trust error = %v, want ErrProtectionTrust", err)
		}
	})

	t.Run("runtime uid replacement", func(t *testing.T) {
		originalSigner := newTestProtectionSigner(t, "original")
		originalRoot := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{originalSigner.TrustKey}}
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), originalSigner, originalRoot)
		policy := testProtectionTrustPolicy(path)

		attacker := newTestProtectionSigner(t, "attacker")
		attackerRoot := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{attacker.TrustKey}}
		if err := os.Chmod(policy.TrustRootPath, 0o644); err != nil {
			t.Fatal(err)
		}
		writeCanonicalJSONFile(t, policy.TrustRootPath, attackerRoot, 0o644)
		if err := os.Chmod(policy.TrustRootPath, 0o444); err != nil {
			t.Fatal(err)
		}
		rewriteTestEnvelope(t, path, signedTestEnvelope(t, validProtectionMatrix(t), attacker))
		if _, err := ParseProtectionCapability(path, policy); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("replacement error = %v, want pinned digest ErrProtectionTrust", err)
		}
	})
}

func TestProtectionTrustPolicyMustBeAbsoluteCanonicalAndPinned(t *testing.T) {
	path := writeProtectionMatrix(t, validProtectionMatrix(t))
	validPolicy := testProtectionTrustPolicy(path)
	for _, tc := range []struct {
		name   string
		mutate func(*protectionVerifierPolicy)
	}{
		{"relative path", func(p *protectionVerifierPolicy) { p.TrustRootPath = ProtectionTrustRootFileName }},
		{"wrong basename", func(p *protectionVerifierPolicy) {
			p.TrustRootPath = filepath.Join(filepath.Dir(p.TrustRootPath), "other.json")
		}},
		{"missing digest", func(p *protectionVerifierPolicy) { p.TrustRootDigest = "" }},
		{"wrong digest", func(p *protectionVerifierPolicy) { p.TrustRootDigest = digestBytes([]byte("other root")) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := validPolicy
			tc.mutate(&policy)
			if _, err := ParseProtectionCapability(path, policy); !errors.Is(err, ErrProtectionTrust) {
				t.Fatalf("error = %v, want ErrProtectionTrust", err)
			}
		})
	}
}

func TestProtectionSignedEnvelopeBindsHeaderPayloadAndEvidence(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{signer.TrustKey}}

	for _, tc := range []struct {
		name   string
		mutate func(*protectionSignedEnvelope)
	}{
		{"key id swap", func(e *protectionSignedEnvelope) { e.KeyID = "secondary" }},
		{"wrong domain", func(e *protectionSignedEnvelope) { e.Domain = "TossOS/other/v1" }},
		{"algorithm swap", func(e *protectionSignedEnvelope) { e.Algorithm = "Ed448" }},
		{"payload swap", func(e *protectionSignedEnvelope) {
			payload, _ := base64.RawURLEncoding.DecodeString(e.Payload)
			e.Payload = base64.RawURLEncoding.EncodeToString([]byte(strings.Replace(string(payload), `"profile":"prod-kr"`, `"profile":"prod-us"`, 1)))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
			envelope := readTestEnvelope(t, path)
			tc.mutate(&envelope)
			rewriteTestEnvelope(t, path, envelope)
			if _, err := parseTestProtectionCapability(path); err == nil {
				t.Fatal("swapped signed field was accepted")
			}
		})
	}

	path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
	parsed, err := parseTestProtectionCapability(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*ProtectionScope, map[string][]byte)
		want   error
	}{
		{"cross account replay", func(s *ProtectionScope, _ map[string][]byte) { s.AccountRef = "99999999" }, ErrProtectionScope},
		{"cross profile replay", func(s *ProtectionScope, _ map[string][]byte) { s.Profile = "other" }, ErrProtectionScope},
		{"cross market replay", func(s *ProtectionScope, _ map[string][]byte) { s.Market = MarketUS }, ErrProtectionScope},
		{"tool build replay", func(s *ProtectionScope, _ map[string][]byte) {
			s.Tools[ToolVerifyExecutionCapability] = ToolBuild{Version: "1.2.3", Build: digestBytes([]byte("other"))}
		}, ErrProtectionScope},
		{"evidence swap", func(_ *ProtectionScope, e map[string][]byte) {
			e[executionEvidenceName], e[triggerEvidenceName] = e[triggerEvidenceName], e[executionEvidenceName]
		}, ErrProtectionEvidence},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scope, evidence := validProtectionScope(), cloneEvidence()
			tc.mutate(&scope, evidence)
			if _, err := VerifyProtectionCapability(parsed, protectionNow, scope, evidence); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestProtectionTrustKeyLifecycle(t *testing.T) {
	t.Run("unknown key", func(t *testing.T) {
		signer := newTestProtectionSigner(t, "signer")
		other := newTestProtectionSigner(t, "other")
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, ProtectionTrustRoot{
			FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{other.TrustKey},
		})
		if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})

	for _, tc := range []struct {
		name   string
		mutate func(*ProtectionTrustKey, *ProtectionCapabilityMatrix)
	}{
		{"revoked", func(k *ProtectionTrustKey, _ *ProtectionCapabilityMatrix) {
			at := protectionNow.Add(-time.Minute)
			k.Status, k.RevokedAt, k.RevocationReason = ProtectionKeyRevoked, &at, "operator compromise response"
		}},
		{"not yet valid", func(k *ProtectionTrustKey, _ *ProtectionCapabilityMatrix) {
			k.NotBefore = protectionNow.Add(time.Second)
		}},
		{"expired key", func(k *ProtectionTrustKey, _ *ProtectionCapabilityMatrix) { k.NotAfter = protectionNow }},
		{"matrix issued before key", func(k *ProtectionTrustKey, m *ProtectionCapabilityMatrix) { k.NotBefore = m.IssuedAt.Add(time.Minute) }},
		{"matrix expires after key", func(k *ProtectionTrustKey, m *ProtectionCapabilityMatrix) { k.NotAfter = m.ExpiresAt.Add(-time.Minute) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			signer := newTestProtectionSigner(t, "primary")
			m := validProtectionMatrix(t)
			tc.mutate(&signer.TrustKey, &m)
			path := writeSignedProtectionMatrix(t, m, signer, ProtectionTrustRoot{
				FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{signer.TrustKey},
			})
			parsed, err := parseTestProtectionCapability(path)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if _, err := VerifyProtectionCapability(parsed, protectionNow, validProtectionScope(), cloneEvidence()); !errors.Is(err, ErrProtectionTrust) {
				t.Fatalf("error = %v, want ErrProtectionTrust", err)
			}
		})
	}
}

func TestProtectionTrustRootRotationOverlap(t *testing.T) {
	oldKey := newTestProtectionSigner(t, "old")
	newKey := newTestProtectionSigner(t, "new")
	oldKey.TrustKey.NotBefore = protectionNow.Add(-48 * time.Hour)
	oldKey.TrustKey.NotAfter = protectionNow.Add(48 * time.Hour)
	newKey.TrustKey.NotBefore = protectionNow.Add(-2 * time.Hour)
	newKey.TrustKey.NotAfter = protectionNow.Add(7 * 24 * time.Hour)
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{oldKey.TrustKey, newKey.TrustKey}}
	for _, signer := range []testProtectionSigner{oldKey, newKey} {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		parsed, err := parseTestProtectionCapability(path)
		if err != nil {
			t.Fatalf("parse %s: %v", signer.TrustKey.KeyID, err)
		}
		if _, err := VerifyProtectionCapability(parsed, protectionNow, validProtectionScope(), cloneEvidence()); err != nil {
			t.Fatalf("verify %s during overlap: %v", signer.TrustKey.KeyID, err)
		}
	}
}

func TestProtectionTrustRootRejectsAmbiguousOrMalformedKeys(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*ProtectionTrustRoot)
	}{
		{"duplicate id", func(r *ProtectionTrustRoot) { r.Keys = append(r.Keys, r.Keys[0]) }},
		{"duplicate public key", func(r *ProtectionTrustRoot) { k := r.Keys[0]; k.KeyID = "alias"; r.Keys = append(r.Keys, k) }},
		{"wrong role", func(r *ProtectionTrustRoot) { r.Keys[0].Role = "EXECUTION_SIGNER" }},
		{"wrong algorithm", func(r *ProtectionTrustRoot) { r.Keys[0].Algorithm = "Ed448" }},
		{"short key", func(r *ProtectionTrustRoot) {
			r.Keys[0].PublicKey = base64.RawURLEncoding.EncodeToString(make([]byte, 31))
		}},
		{"padded key", func(r *ProtectionTrustRoot) { r.Keys[0].PublicKey += "=" }},
		{"non UTC key time", func(r *ProtectionTrustRoot) {
			r.Keys[0].NotBefore = r.Keys[0].NotBefore.In(time.FixedZone("KST", 9*60*60))
		}},
		{"active with revocation metadata", func(r *ProtectionTrustRoot) {
			at := protectionNow
			r.Keys[0].RevokedAt = &at
			r.Keys[0].RevocationReason = "ambiguous"
		}},
		{"revoked without reason", func(r *ProtectionTrustRoot) {
			at := protectionNow
			r.Keys[0].Status = ProtectionKeyRevoked
			r.Keys[0].RevokedAt = &at
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			signer := newTestProtectionSigner(t, "primary")
			root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{signer.TrustKey}}
			tc.mutate(&root)
			path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
			if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionTrust) {
				t.Fatalf("error = %v, want ErrProtectionTrust", err)
			}
		})
	}
}

func TestProtectionTrustRootJSONMustBeStrictAndCanonical(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{signer.TrustKey}}
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"whitespace", func(data []byte) []byte { return append(data, '\n') }},
		{"unknown field", func(data []byte) []byte { return []byte(strings.TrimSuffix(string(data), "}") + `,"unknown":true}`) }},
		{"duplicate field", func(data []byte) []byte {
			return []byte(strings.Replace(string(data), `"format_version":1`, `"format_version":1,"format_version":1`, 1))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
			rootPath := testProtectionTrustRootPath(path)
			data, err := os.ReadFile(rootPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(rootPath, 0o644); err != nil {
				t.Fatal(err)
			}
			mutated := tc.mutate(data)
			if err := os.WriteFile(rootPath, mutated, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(rootPath, 0o444); err != nil {
				t.Fatal(err)
			}
			policy := testProtectionTrustPolicy(path)
			policy.TrustRootDigest = protectionDigest(mutated)
			if err := rewriteTestVerifierPolicyFromTrust(path, policy); err != nil {
				t.Fatal(err)
			}
			if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionTrust) {
				t.Fatalf("error = %v, want ErrProtectionTrust", err)
			}
		})
	}
}

func TestProtectionSignedEnvelopeRejectsNonCanonicalEncoding(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{signer.TrustKey}}

	for _, tc := range []struct {
		name   string
		mutate func(string, protectionSignedEnvelope) []byte
	}{
		{"envelope whitespace", func(_ string, e protectionSignedEnvelope) []byte {
			data, _ := json.MarshalIndent(e, "", "  ")
			return data
		}},
		{"envelope unknown field", func(_ string, e protectionSignedEnvelope) []byte {
			data, _ := json.Marshal(e)
			return []byte(strings.TrimSuffix(string(data), "}") + `,"unknown":true}`)
		}},
		{"envelope duplicate field", func(_ string, e protectionSignedEnvelope) []byte {
			data, _ := json.Marshal(e)
			return []byte(strings.Replace(string(data), `"key_id":"primary"`, `"key_id":"primary","key_id":"primary"`, 1))
		}},
		{"padded payload", func(_ string, e protectionSignedEnvelope) []byte {
			e.Payload += "="
			data, _ := json.Marshal(e)
			return data
		}},
		{"padded signature", func(_ string, e protectionSignedEnvelope) []byte {
			e.Signature += "="
			data, _ := json.Marshal(e)
			return data
		}},
		{"payload unknown field", func(_ string, e protectionSignedEnvelope) []byte {
			payload, _ := base64.RawURLEncoding.DecodeString(e.Payload)
			payload = []byte(strings.Replace(string(payload), `"format_version":1`, `"format_version":1,"unknown":true`, 1))
			e.Payload = base64.RawURLEncoding.EncodeToString(payload)
			e.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(signer.Private, protectionSigningMessage(e)))
			data, _ := json.Marshal(e)
			return data
		}},
		{"payload whitespace", func(_ string, e protectionSignedEnvelope) []byte {
			payload, _ := base64.RawURLEncoding.DecodeString(e.Payload)
			e.Payload = base64.RawURLEncoding.EncodeToString(append(payload, '\n'))
			e.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(signer.Private, protectionSigningMessage(e)))
			data, _ := json.Marshal(e)
			return data
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
			envelope := readTestEnvelope(t, path)
			if err := os.WriteFile(path, tc.mutate(path, envelope), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := parseTestProtectionCapability(path); err == nil {
				t.Fatal("noncanonical encoding accepted")
			}
		})
	}
}

func TestProtectionSignedMatrixRequiresOneSemanticCanonicalForm(t *testing.T) {
	t.Run("evidence order", func(t *testing.T) {
		matrix := validProtectionMatrix(t)
		matrix.Evidence[0], matrix.Evidence[1] = matrix.Evidence[1], matrix.Evidence[0]
		redigestMatrix(t, &matrix)
		if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, matrix)); !errors.Is(err, ErrProtectionInvalid) {
			t.Fatalf("reordered evidence = %v", err)
		}
	})
	t.Run("capability order", func(t *testing.T) {
		matrix := validProtectionMatrix(t)
		extra := matrix.Capabilities[0]
		extra.AccountRef = "99999999"
		matrix.Capabilities = append(matrix.Capabilities, extra)
		matrix.Capabilities[0], matrix.Capabilities[1] = matrix.Capabilities[1], matrix.Capabilities[0]
		redigestMatrix(t, &matrix)
		if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, matrix)); !errors.Is(err, ErrProtectionInvalid) {
			t.Fatalf("reordered capabilities = %v", err)
		}
	})
	t.Run("non UTC timestamp", func(t *testing.T) {
		matrix := validProtectionMatrix(t)
		matrix.IssuedAt = matrix.IssuedAt.In(time.FixedZone("KST", 9*60*60))
		matrix.ExpiresAt = matrix.ExpiresAt.In(time.FixedZone("KST", 9*60*60))
		redigestMatrix(t, &matrix)
		if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, matrix)); !errors.Is(err, ErrProtectionInvalid) {
			t.Fatalf("non-UTC timestamps = %v", err)
		}
	})
	t.Run("trust key order", func(t *testing.T) {
		alpha := newTestProtectionSigner(t, "alpha")
		zulu := newTestProtectionSigner(t, "zulu")
		root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Generation: 1, Keys: []ProtectionTrustKey{zulu.TrustKey, alpha.TrustKey}}
		if err := root.validate(); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("unsorted trust keys = %v", err)
		}
	})
}

func TestProtectionTrustRootRejectsUnsafeFileAndTOCTOU(t *testing.T) {
	signer := newTestProtectionSigner(t, "primary")
	root := ProtectionTrustRoot{FormatVersion: ProtectionTrustRootFormatVersion, Keys: []ProtectionTrustKey{signer.TrustKey}}

	t.Run("mode", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		rootPath := testProtectionTrustRootPath(path)
		if err := os.Chmod(rootPath, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("symlink", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		rootPath := testProtectionTrustRootPath(path)
		target := filepath.Join(secureTempDir(t), ProtectionTrustRootFileName)
		writeCanonicalJSONFile(t, target, root, 0o444)
		if err := os.Remove(rootPath); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, rootPath); err != nil {
			t.Fatal(err)
		}
		if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("hardlink", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		rootPath := testProtectionTrustRootPath(path)
		if err := os.Link(rootPath, filepath.Join(filepath.Dir(path), "root-alias")); err != nil {
			t.Fatal(err)
		}
		if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("owner", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		policy := testProtectionTrustPolicy(path)
		policy.TrustRootOwner++
		if _, err := loadProtectionTrustRoot(policy, nil); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("parent symlink", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		policy := testProtectionTrustPolicy(path)
		linkDir := filepath.Join(filepath.Dir(filepath.Dir(policy.TrustRootPath)), "trust-link")
		if err := os.Symlink(filepath.Dir(policy.TrustRootPath), linkDir); err != nil {
			t.Fatal(err)
		}
		policy.TrustRootPath = filepath.Join(linkDir, ProtectionTrustRootFileName)
		if _, err := ParseProtectionCapability(path, policy); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("writable parent", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		policy := testProtectionTrustPolicy(path)
		if err := os.Chmod(filepath.Dir(policy.TrustRootPath), 0o757); err != nil {
			t.Fatal(err)
		}
		if _, err := ParseProtectionCapability(path, policy); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("non-traversable parent", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		policy := testProtectionTrustPolicy(path)
		if err := os.Chmod(filepath.Dir(policy.TrustRootPath), 0o750); err != nil {
			t.Fatal(err)
		}
		if _, err := ParseProtectionCapability(path, policy); !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("post-read replacement", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		policy := testProtectionTrustPolicy(path)
		rootPath := policy.TrustRootPath
		_, err := loadProtectionTrustRoot(policy, func() {
			replacement := filepath.Join(filepath.Dir(rootPath), "replacement-root")
			writeCanonicalJSONFile(t, replacement, root, 0o444)
			if renameErr := os.Rename(replacement, rootPath); renameErr != nil {
				t.Fatal(renameErr)
			}
		})
		if !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("error = %v, want ErrProtectionTrust", err)
		}
	})
	t.Run("post-read trust parent mode", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		policy := testProtectionTrustPolicy(path)
		parent := filepath.Dir(policy.TrustRootPath)
		_, err := loadProtectionTrustRoot(policy, func() {
			if chmodErr := os.Chmod(parent, 0o757); chmodErr != nil {
				t.Fatal(chmodErr)
			}
		})
		if !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("trust parent mode race = %v", err)
		}
	})
	t.Run("post-read policy parent mode", func(t *testing.T) {
		path := writeSignedProtectionMatrix(t, validProtectionMatrix(t), signer, root)
		verifier := newTestProtectionVerifier(path, protectionNow)
		parent := filepath.Dir(verifier.source.path)
		_, err := loadProtectionVerifierPolicy(verifier.source, func() {
			if chmodErr := os.Chmod(parent, 0o757); chmodErr != nil {
				t.Fatal(chmodErr)
			}
		})
		if !errors.Is(err, ErrProtectionTrust) {
			t.Fatalf("policy parent mode race = %v", err)
		}
	})
}

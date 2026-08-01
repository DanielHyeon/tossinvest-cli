package attest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var protectionNow = time.Date(2026, 7, 31, 3, 0, 0, 0, time.UTC)

const (
	executionEvidenceName = "verify-execution-capability.evidence.jsonl"
	triggerEvidenceName   = "verify-observes-the-trigger.evidence.jsonl"
)

var protectionEvidenceBytes = map[string][]byte{
	executionEvidenceName: []byte("execution evidence\n"),
	triggerEvidenceName:   []byte("trigger evidence\n"),
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestMatrix(t *testing.T, m ProtectionCapabilityMatrix) string {
	t.Helper()
	data, err := canonicalProtectionMatrix(m)
	if err != nil {
		t.Fatal(err)
	}
	return digestBytes(data)
}

func validProtectionMatrix(t *testing.T) ProtectionCapabilityMatrix {
	t.Helper()
	m := ProtectionCapabilityMatrix{
		FormatVersion: ProtectionFormatVersion,
		IssuedAt:      protectionNow.Add(-time.Hour),
		ExpiresAt:     protectionNow.Add(24 * time.Hour),
		Capabilities: []ConditionalCapability{{
			AccountRef: "12345678", Profile: "prod-kr", Market: MarketKR,
			Session: SessionRegular, ConditionalType: ConditionalSingle,
			OrderType:   OrderMarket,
			Trigger:     TriggerCapability{Source: TriggerLastTrade, Direction: TriggerFallsToOrBelow},
			Quantity:    QuantityCapability{Minimum: 1, Maximum: 999, PartialFill: true},
			Persistence: PersistenceCapability{SurvivesProcessExit: true, SurvivesRestart: true},
			Reservation: ReservationCapability{ReservesSellableQuantity: true},
			Idempotency: IdempotencyCapability{Create: true, ClientOrderID: true},
			Replace:     ReplaceCapability{Mode: ReplaceAtomic, ContinuousCoverage: true, NewIdentifierRecorded: true},
		}},
	}
	m.Evidence = []ProtectionEvidence{
		{Tool: ToolVerifyExecutionCapability, Version: "1.2.3", Build: digestBytes([]byte("execution build")), Source: executionEvidenceName, Digest: digestBytes(protectionEvidenceBytes[executionEvidenceName])},
		{Tool: ToolVerifyObservesTrigger, Version: "1.2.3", Build: digestBytes([]byte("trigger build")), Source: triggerEvidenceName, Digest: digestBytes(protectionEvidenceBytes[triggerEvidenceName])},
	}
	redigestMatrix(t, &m)
	return m
}

func validProtectionScope() ProtectionScope {
	return ProtectionScope{
		AccountRef: "12345678", Profile: "prod-kr", Market: MarketKR,
		Session: SessionRegular, ConditionalType: ConditionalSingle, OrderType: OrderMarket,
		TriggerSource: TriggerLastTrade, Quantity: 1,
		Tools: map[ProtectionTool]ToolBuild{
			ToolVerifyExecutionCapability: {Version: "1.2.3", Build: digestBytes([]byte("execution build"))},
			ToolVerifyObservesTrigger:     {Version: "1.2.3", Build: digestBytes([]byte("trigger build"))},
		},
	}
}

func cloneEvidence() map[string][]byte {
	out := make(map[string][]byte, len(protectionEvidenceBytes))
	for name, data := range protectionEvidenceBytes {
		out[name] = append([]byte(nil), data...)
	}
	return out
}

func writeProtectionMatrix(t *testing.T, m ProtectionCapabilityMatrix) string {
	t.Helper()
	signer := newTestProtectionSigner(t, "test-protection-key")
	return writeSignedProtectionMatrix(t, m, signer, ProtectionTrustRoot{
		FormatVersion: ProtectionTrustRootFormatVersion,
		Keys:          []ProtectionTrustKey{signer.TrustKey},
	})
}

func secureTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func parseAndVerify(t *testing.T, m ProtectionCapabilityMatrix, scope ProtectionScope, evidence map[string][]byte) error {
	t.Helper()
	parsed, err := parseTestProtectionCapability(writeProtectionMatrix(t, m))
	if err != nil {
		return err
	}
	_, err = VerifyProtectionCapability(parsed, protectionNow, scope, evidence)
	return err
}

func TestProtectionMatrixParseAndEvidenceVerificationAreSeparate(t *testing.T) {
	parsed, err := parseTestProtectionCapability(writeProtectionMatrix(t, validProtectionMatrix(t)))
	if err != nil {
		t.Fatalf("ParseProtectionCapability: %v", err)
	}
	if _, err := VerifyProtectionCapability(parsed, protectionNow, validProtectionScope(), nil); !errors.Is(err, ErrProtectionEvidence) {
		t.Fatalf("verification without evidence bytes = %v, want ErrProtectionEvidence", err)
	}
	verified, err := VerifyProtectionCapability(parsed, protectionNow, validProtectionScope(), cloneEvidence())
	if err != nil {
		t.Fatalf("VerifyProtectionCapability: %v", err)
	}
	if got := verified.Capability().Replace.Mode; got != ReplaceAtomic {
		t.Fatalf("replace mode = %s", got)
	}
}

func TestProtectionMatrixRecomputesEveryEvidenceDigest(t *testing.T) {
	parsed, err := parseTestProtectionCapability(writeProtectionMatrix(t, validProtectionMatrix(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(map[string][]byte)
	}{
		{"missing", func(e map[string][]byte) { delete(e, triggerEvidenceName) }},
		{"fake bytes", func(e map[string][]byte) { e[executionEvidenceName] = []byte("forged") }},
		{"extra evidence", func(e map[string][]byte) { e["unclaimed.evidence"] = []byte("extra") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			evidence := cloneEvidence()
			tc.mutate(evidence)
			if _, err := VerifyProtectionCapability(parsed, protectionNow, validProtectionScope(), evidence); !errors.Is(err, ErrProtectionEvidence) {
				t.Fatalf("error = %v, want ErrProtectionEvidence", err)
			}
		})
	}
}

func TestProtectionMatrixDigestBindsCanonicalMatrix(t *testing.T) {
	m := validProtectionMatrix(t)
	m.Capabilities[0].Quantity.Maximum++
	if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, m)); !errors.Is(err, ErrProtectionInvalid) {
		t.Fatalf("mutated matrix = %v, want ErrProtectionInvalid", err)
	}
	m = validProtectionMatrix(t)
	m.Evidence[0].CapabilityDigest = digestBytes([]byte("different matrix"))
	if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, m)); !errors.Is(err, ErrProtectionInvalid) {
		t.Fatalf("unbound evidence = %v, want ErrProtectionInvalid", err)
	}
	m = validProtectionMatrix(t)
	m.ExpiresAt = m.ExpiresAt.Add(time.Hour)
	if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, m)); !errors.Is(err, ErrProtectionInvalid) {
		t.Fatalf("unbound validity window = %v, want ErrProtectionInvalid", err)
	}
	m = validProtectionMatrix(t)
	m.Evidence[0].Version = "1.2.4"
	if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, m)); !errors.Is(err, ErrProtectionInvalid) {
		t.Fatalf("unbound evidence metadata = %v, want ErrProtectionInvalid", err)
	}
}

func TestProtectionMatrixRejectsUnknownLegacyAndTrailingJSON(t *testing.T) {
	cases := map[string]func(*ProtectionCapabilityMatrix){
		"legacy version": func(m *ProtectionCapabilityMatrix) { m.FormatVersion = 0 },
		"newer version":  func(m *ProtectionCapabilityMatrix) { m.FormatVersion = 2 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := validProtectionMatrix(t)
			mutate(&m)
			path := writeProtectionMatrix(t, m)
			if _, err := parseTestProtectionCapability(path); err == nil {
				t.Fatal("invalid matrix was accepted")
			}
		})
	}
}

func TestProtectionMatrixRejectsTimeAndExactScopeMismatch(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ProtectionCapabilityMatrix, *ProtectionScope)
		want   error
	}{
		{"expired", func(m *ProtectionCapabilityMatrix, _ *ProtectionScope) {
			m.ExpiresAt = protectionNow
			redigestMatrix(t, m)
		}, ErrProtectionExpired},
		{"future issue", func(m *ProtectionCapabilityMatrix, _ *ProtectionScope) {
			m.IssuedAt = protectionNow.Add(time.Second)
			m.ExpiresAt = protectionNow.Add(time.Hour)
			redigestMatrix(t, m)
		}, ErrProtectionInvalid},
		{"account", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.AccountRef = "99999999" }, ErrProtectionScope},
		{"profile", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.Profile = "other" }, ErrProtectionScope},
		{"market", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.Market = MarketUS }, ErrProtectionScope},
		{"session", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.Session = SessionExtended }, ErrProtectionScope},
		{"conditional type", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.ConditionalType = ConditionalOCO }, ErrProtectionScope},
		{"order type", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.OrderType = OrderLimit }, ErrProtectionScope},
		{"trigger", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.TriggerSource = TriggerMarkPrice }, ErrProtectionScope},
		{"quantity", func(_ *ProtectionCapabilityMatrix, s *ProtectionScope) { s.Quantity = 1000 }, ErrProtectionScope},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, scope := validProtectionMatrix(t), validProtectionScope()
			tc.mutate(&m, &scope)
			if !errors.Is(parseAndVerify(t, m, scope, cloneEvidence()), tc.want) {
				t.Fatalf("want %v", tc.want)
			}
		})
	}
}

func TestProtectionMatrixRejectsMissingSafetyClaimsAndMetadata(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ProtectionCapabilityMatrix)
	}{
		{"partial fill", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Quantity.PartialFill = false
			redigestMatrix(t, m)
		}},
		{"process persistence", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Persistence.SurvivesProcessExit = false
			redigestMatrix(t, m)
		}},
		{"restart persistence", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Persistence.SurvivesRestart = false
			redigestMatrix(t, m)
		}},
		{"reservation", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Reservation.ReservesSellableQuantity = false
			redigestMatrix(t, m)
		}},
		{"idempotency", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Idempotency.Create = false
			redigestMatrix(t, m)
		}},
		{"client order identity", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Idempotency.ClientOrderID = false
			redigestMatrix(t, m)
		}},
		{"replace mode", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Replace.Mode = "BEST_EFFORT"
			redigestMatrix(t, m)
		}},
		{"replace continuity", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Replace.ContinuousCoverage = false
			redigestMatrix(t, m)
		}},
		{"replace identifier", func(m *ProtectionCapabilityMatrix) {
			m.Capabilities[0].Replace.NewIdentifierRecorded = false
			redigestMatrix(t, m)
		}},
		{"tool missing", func(m *ProtectionCapabilityMatrix) { m.Evidence = m.Evidence[:1] }},
		{"tool version", func(m *ProtectionCapabilityMatrix) { m.Evidence[0].Version = "" }},
		{"tool build", func(m *ProtectionCapabilityMatrix) { m.Evidence[0].Build = "dirty" }},
		{"evidence digest", func(m *ProtectionCapabilityMatrix) { m.Evidence[0].Digest = "sha256:nope" }},
		{"source path", func(m *ProtectionCapabilityMatrix) { m.Evidence[0].Source = "../" + executionEvidenceName }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validProtectionMatrix(t)
			tc.mutate(&m)
			if err := parseAndVerify(t, m, validProtectionScope(), cloneEvidence()); !errors.Is(err, ErrProtectionInvalid) {
				t.Fatalf("error = %v, want ErrProtectionInvalid", err)
			}
		})
	}
}

func bindEvidence(m *ProtectionCapabilityMatrix) {
	for i := range m.Evidence {
		m.Evidence[i].CapabilityDigest = m.CapabilityDigest
	}
}

func redigestMatrix(t *testing.T, m *ProtectionCapabilityMatrix) {
	t.Helper()
	m.CapabilityDigest = digestMatrix(t, *m)
	bindEvidence(m)
}

func TestProtectionAccountFormatIsStrict(t *testing.T) {
	for _, account := range []string{"acct-12345678", "1234-5678", "123-45678", "1234_5678", "-12345678", "12345678-", "123--45678", "1234 5678", "123"} {
		t.Run(account, func(t *testing.T) {
			m := validProtectionMatrix(t)
			m.Capabilities[0].AccountRef = account
			redigestMatrix(t, &m)
			if _, err := parseTestProtectionCapability(writeProtectionMatrix(t, m)); !errors.Is(err, ErrProtectionInvalid) {
				t.Fatalf("account %q accepted: %v", account, err)
			}
		})
	}
	parsed, err := parseTestProtectionCapability(writeProtectionMatrix(t, validProtectionMatrix(t)))
	if err != nil {
		t.Fatal(err)
	}
	scope := validProtectionScope()
	scope.AccountRef = "acct-12345678"
	if _, err := VerifyProtectionCapability(parsed, protectionNow, scope, cloneEvidence()); !errors.Is(err, ErrProtectionScope) {
		t.Fatalf("malformed runtime account = %v", err)
	}
}

func TestProtectionMatrixRejectsUnsafeFileAndDirectParent(t *testing.T) {
	t.Run("wrong basename", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix(t))
		wrong := filepath.Join(filepath.Dir(path), "other.json")
		if err := os.Rename(path, wrong); err != nil {
			t.Fatal(err)
		}
		if _, err := parseTestProtectionCapability(wrong); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("mode", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix(t))
		if err := os.Chmod(path, 0o640); err != nil {
			t.Fatal(err)
		}
		if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("symlink", func(t *testing.T) {
		target := writeProtectionMatrix(t, validProtectionMatrix(t))
		link := filepath.Join(secureTempDir(t), ProtectionFileName)
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		policy := testProtectionTrustPolicy(target)
		policy.AttestationPath = link
		if err := rewriteTestVerifierPolicyFromTrust(target, policy); err != nil {
			t.Fatal(err)
		}
		if _, err := newTestProtectionVerifier(target, protectionNow).parse(); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("hardlink", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix(t))
		if err := os.Link(path, filepath.Join(filepath.Dir(path), "second-link")); err != nil {
			t.Fatal(err)
		}
		if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("parent mode", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix(t))
		if err := os.Chmod(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if _, err := parseTestProtectionCapability(path); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("parent symlink", func(t *testing.T) {
		target := writeProtectionMatrix(t, validProtectionMatrix(t))
		realDir := filepath.Dir(target)
		root := secureTempDir(t)
		linkDir := filepath.Join(root, "profile")
		if err := os.Symlink(realDir, linkDir); err != nil {
			t.Fatal(err)
		}
		policy := testProtectionTrustPolicy(target)
		policy.AttestationPath = filepath.Join(linkDir, ProtectionFileName)
		if err := rewriteTestVerifierPolicyFromTrust(target, policy); err != nil {
			t.Fatal(err)
		}
		if _, err := newTestProtectionVerifier(target, protectionNow).parse(); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("owner", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix(t))
		owner, _ := currentProtectionOwnerUID()
		if _, err := parseProtectionCapability(path, fileIdentity{UID: owner + 1}, testProtectionTrustPolicy(path)); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestProtectionCapabilityIsSeparateFromLegacyAttestation(t *testing.T) {
	legacy := Attestation{FormatVersion: FormatVersion, AccountRef: "123-45678", IssuedAt: protectionNow.Add(-time.Hour), ExpiresAt: protectionNow.Add(time.Hour), Endpoints: []string{"GET /api/v1/accounts"}}
	path := filepath.Join(t.TempDir(), FileName)
	if err := Save(path, legacy); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := loaded.Verify(protectionNow, legacy.AccountRef, legacy.Endpoints); err != nil {
		t.Fatalf("legacy execution attestation changed: %v", err)
	}
}

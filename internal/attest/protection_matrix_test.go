package attest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var protectionNow = time.Date(2026, 7, 31, 3, 0, 0, 0, time.UTC)

func validProtectionMatrix() ProtectionCapabilityMatrix {
	return ProtectionCapabilityMatrix{
		FormatVersion: ProtectionFormatVersion,
		IssuedAt:      protectionNow.Add(-time.Hour),
		ExpiresAt:     protectionNow.Add(24 * time.Hour),
		Evidence: []ProtectionEvidence{
			{Tool: ToolVerifyExecutionCapability, Version: "1.2.3", Build: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
			{Tool: ToolVerifyObservesTrigger, Version: "1.2.3", Build: "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", Digest: "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"},
		},
		Capabilities: []ConditionalCapability{{
			AccountRef: "123-45678", Profile: "prod-kr", Market: MarketKR,
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
}

func validProtectionScope() ProtectionScope {
	return ProtectionScope{
		AccountRef: "12345678", Profile: "prod-kr", Market: MarketKR,
		Session: SessionRegular, ConditionalType: ConditionalSingle, OrderType: OrderMarket,
		TriggerSource: TriggerLastTrade, Quantity: 1,
		Tools: map[ProtectionTool]ToolBuild{
			ToolVerifyExecutionCapability: {Version: "1.2.3", Build: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			ToolVerifyObservesTrigger:     {Version: "1.2.3", Build: "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
		},
	}
}

func writeProtectionMatrix(t *testing.T, m ProtectionCapabilityMatrix) string {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), ProtectionFileName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestProtectionMatrixStrictLoadAndVerify(t *testing.T) {
	path := writeProtectionMatrix(t, validProtectionMatrix())
	got, err := LoadProtectionCapability(path, protectionNow, validProtectionScope())
	if err != nil {
		t.Fatalf("LoadProtectionCapability: %v", err)
	}
	if len(got.Capabilities) != 1 || got.Capabilities[0].Replace.Mode != ReplaceAtomic {
		t.Fatalf("matrix = %+v", got)
	}
}

func TestProtectionMatrixRejectsUnknownLegacyAndTrailingJSON(t *testing.T) {
	data, err := json.Marshal(validProtectionMatrix())
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string][]byte{
		"unknown field":  []byte(strings.Replace(string(data), `"format_version":1`, `"format_version":1,"surprise":true`, 1)),
		"legacy version": []byte(strings.Replace(string(data), `"format_version":1`, `"format_version":0`, 1)),
		"newer version":  []byte(strings.Replace(string(data), `"format_version":1`, `"format_version":2`, 1)),
		"trailing json":  append(append([]byte(nil), data...), []byte(` {}`)...),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ProtectionFileName)
			if err := os.WriteFile(path, body, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadProtectionCapability(path, protectionNow, validProtectionScope()); err == nil {
				t.Fatal("invalid matrix was accepted")
			}
		})
	}
}

func TestProtectionMatrixRejectsExpiredFutureAndIdentityMismatch(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ProtectionCapabilityMatrix, *ProtectionScope)
		want   error
	}{
		{"expired", func(m *ProtectionCapabilityMatrix, _ *ProtectionScope) { m.ExpiresAt = protectionNow }, ErrProtectionExpired},
		{"future issue", func(m *ProtectionCapabilityMatrix, _ *ProtectionScope) { m.IssuedAt = protectionNow.Add(time.Second) }, ErrProtectionInvalid},
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
			m, scope := validProtectionMatrix(), validProtectionScope()
			tc.mutate(&m, &scope)
			_, err := LoadProtectionCapability(writeProtectionMatrix(t, m), protectionNow, scope)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestProtectionMatrixRejectsMissingSafetyClaimsAndBadEvidence(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ProtectionCapabilityMatrix)
	}{
		{"partial fill", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Quantity.PartialFill = false }},
		{"process persistence", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Persistence.SurvivesProcessExit = false }},
		{"restart persistence", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Persistence.SurvivesRestart = false }},
		{"reservation", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Reservation.ReservesSellableQuantity = false }},
		{"create idempotency", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Idempotency.Create = false }},
		{"client order identity", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Idempotency.ClientOrderID = false }},
		{"replace continuity", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Replace.ContinuousCoverage = false }},
		{"replace identity", func(m *ProtectionCapabilityMatrix) { m.Capabilities[0].Replace.NewIdentifierRecorded = false }},
		{"tool missing", func(m *ProtectionCapabilityMatrix) { m.Evidence = m.Evidence[:1] }},
		{"tool version", func(m *ProtectionCapabilityMatrix) { m.Evidence[0].Version = "" }},
		{"tool build", func(m *ProtectionCapabilityMatrix) { m.Evidence[0].Build = "dirty" }},
		{"evidence digest", func(m *ProtectionCapabilityMatrix) { m.Evidence[0].Digest = "sha256:nope" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validProtectionMatrix()
			tc.mutate(&m)
			if _, err := LoadProtectionCapability(writeProtectionMatrix(t, m), protectionNow, validProtectionScope()); !errors.Is(err, ErrProtectionInvalid) {
				t.Fatalf("error = %v, want ErrProtectionInvalid", err)
			}
		})
	}
}

func TestProtectionMatrixRejectsToolBuildMismatch(t *testing.T) {
	scope := validProtectionScope()
	scope.Tools[ToolVerifyExecutionCapability] = ToolBuild{Version: "1.2.4", Build: scope.Tools[ToolVerifyExecutionCapability].Build}
	_, err := LoadProtectionCapability(writeProtectionMatrix(t, validProtectionMatrix()), protectionNow, scope)
	if !errors.Is(err, ErrProtectionScope) {
		t.Fatalf("error = %v, want ErrProtectionScope", err)
	}
}

func TestProtectionMatrixRejectsUnsafeFiles(t *testing.T) {
	t.Run("mode", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix())
		if err := os.Chmod(path, 0o640); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadProtectionCapability(path, protectionNow, validProtectionScope()); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v, want ErrProtectionFile", err)
		}
	})
	t.Run("symlink", func(t *testing.T) {
		target := writeProtectionMatrix(t, validProtectionMatrix())
		link := filepath.Join(t.TempDir(), ProtectionFileName)
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadProtectionCapability(link, protectionNow, validProtectionScope()); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v, want ErrProtectionFile", err)
		}
	})
	t.Run("directory", func(t *testing.T) {
		if _, err := LoadProtectionCapability(t.TempDir(), protectionNow, validProtectionScope()); !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v, want ErrProtectionFile", err)
		}
	})
	t.Run("owner", func(t *testing.T) {
		path := writeProtectionMatrix(t, validProtectionMatrix())
		_, err := loadProtectionCapability(path, protectionNow, validProtectionScope(), fileIdentity{UID: uint32(os.Geteuid() + 1)})
		if !errors.Is(err, ErrProtectionFile) {
			t.Fatalf("error = %v, want ErrProtectionFile", err)
		}
	})
}

func TestProtectionCapabilityIsSeparateFromLegacyAttestation(t *testing.T) {
	legacy := Attestation{
		FormatVersion: FormatVersion,
		AccountRef:    "123-45678",
		IssuedAt:      protectionNow.Add(-time.Hour),
		ExpiresAt:     protectionNow.Add(time.Hour),
		Endpoints:     []string{"GET /api/v1/accounts"},
	}
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

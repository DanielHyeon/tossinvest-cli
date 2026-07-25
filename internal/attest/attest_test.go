package attest_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func valid() attest.Attestation {
	return attest.Attestation{
		FormatVersion: attest.FormatVersion,
		AccountRef:    "123-45678",
		IssuedAt:      now.Add(-24 * time.Hour),
		ExpiresAt:     now.Add(30 * 24 * time.Hour),
		SoakDays:      3,
		Endpoints: []string{
			"GET /api/v1/accounts",
			"GET /api/v1/orders",
			"GET /api/v1/orders/{id}",
			"GET /api/v1/buying-power",
			"GET /api/v1/holdings",
			"POST /api/v1/orders",
			"POST /api/v1/orders/{id}/cancel",
		},
		VerifiedBy: "operator",
	}
}

func writeAttestation(t *testing.T, a attest.Attestation) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), attest.FileName)
	if err := attest.Save(path, a); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return path
}

// TestRoundTrip: what the verification tool writes is what the engine reads.
func TestRoundTrip(t *testing.T) {
	path := writeAttestation(t, valid())

	got, err := attest.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.AccountRef != "123-45678" {
		t.Errorf("account = %q", got.AccountRef)
	}
	if !got.ExpiresAt.Equal(valid().ExpiresAt) {
		t.Errorf("expiry = %v, want %v", got.ExpiresAt, valid().ExpiresAt)
	}
	if got.SoakDays != 3 {
		t.Errorf("soak days = %d", got.SoakDays)
	}
	if err := got.Verify(now, "123-45678", nil); err != nil {
		t.Errorf("a freshly written attestation must verify: %v", err)
	}
}

// TestSaveIsOwnerOnly: the file names an account and its limits.
func TestSaveIsOwnerOnly(t *testing.T) {
	path := writeAttestation(t, valid())

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("mode = %v, want 0600", mode)
	}
}

// TestMissingFileIsItsOwnError so the engine can say "run the verification"
// rather than printing a path that does not exist.
func TestMissingFileIsItsOwnError(t *testing.T) {
	_, err := attest.Load(filepath.Join(t.TempDir(), "nope.json"))
	if !errors.Is(err, attest.ErrMissing) {
		t.Fatalf("err = %v, want ErrMissing", err)
	}
}

func TestMalformedFileIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), attest.FileName)
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := attest.Load(path); !errors.Is(err, attest.ErrMalformed) {
		t.Fatalf("err = %v, want ErrMalformed", err)
	}
}

// TestNewerFormatIsRefused: reading a field we do not know about as absent could
// turn "US orders were never verified" into "nothing to report".
func TestNewerFormatIsRefused(t *testing.T) {
	a := valid()
	a.FormatVersion = attest.FormatVersion + 1
	path := writeAttestation(t, a)

	if _, err := attest.Load(path); !errors.Is(err, attest.ErrFormatTooNew) {
		t.Fatalf("err = %v, want ErrFormatTooNew", err)
	}
}

// TestVerifyRefusals walks every way verification says no.
func TestVerifyRefusals(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(*attest.Attestation)
		account    string
		required   []string
		wantErr    error
		wantDetail string
	}{
		{
			name:    "expired",
			mutate:  func(a *attest.Attestation) { a.ExpiresAt = now.Add(-time.Second) },
			account: "123-45678",
			wantErr: attest.ErrExpired,
			// The scenario in engine-safety requires a re-verification prompt.
			wantDetail: "verify-execution-capability",
		},
		{
			name:    "expires exactly now",
			mutate:  func(a *attest.Attestation) { a.ExpiresAt = now },
			account: "123-45678",
			wantErr: attest.ErrExpired,
		},
		{
			name:    "different account",
			mutate:  func(*attest.Attestation) {},
			account: "999-00000",
			wantErr: attest.ErrAccountMismatch,
		},
		{
			name:     "endpoint never verified",
			mutate:   func(a *attest.Attestation) { a.Endpoints = []string{"GET /api/v1/accounts"} },
			account:  "123-45678",
			required: []string{"POST /api/v1/orders"},
			wantErr:  attest.ErrEndpointNotAttested,
		},
		{
			name:    "no expiry at all",
			mutate:  func(a *attest.Attestation) { a.ExpiresAt = time.Time{} },
			account: "123-45678",
			wantErr: attest.ErrIncomplete,
		},
		{
			name:    "no account",
			mutate:  func(a *attest.Attestation) { a.AccountRef = "" },
			account: "123-45678",
			wantErr: attest.ErrIncomplete,
		},
		{
			name:    "no endpoints",
			mutate:  func(a *attest.Attestation) { a.Endpoints = nil },
			account: "123-45678",
			wantErr: attest.ErrIncomplete,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := valid()
			tc.mutate(&a)

			err := a.Verify(now, tc.account, tc.required)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantDetail != "" && !contains(err.Error(), tc.wantDetail) {
				t.Errorf("message %q does not tell the operator about %q", err, tc.wantDetail)
			}
		})
	}
}

// TestAccountMatchIgnoresFormatting: brokers write the same account with and
// without hyphens depending on the endpoint, and a hyphen must not be why an
// engine refuses to start.
func TestAccountMatchIgnoresFormatting(t *testing.T) {
	a := valid()
	a.AccountRef = "123-45678"

	if err := a.Verify(now, "12345678", nil); err != nil {
		t.Errorf("hyphenless form must match: %v", err)
	}
	if err := a.Verify(now, "123 45678", nil); err != nil {
		t.Errorf("spaced form must match: %v", err)
	}
	if err := a.Verify(now, "12345679", nil); !errors.Is(err, attest.ErrAccountMismatch) {
		t.Errorf("a genuinely different account must not match: %v", err)
	}
}

// TestEndpointMatchIsCaseAndSpaceInsensitive keeps a cosmetic difference in how
// the verification tool wrote an endpoint from becoming a refusal to start.
func TestEndpointMatchIsCaseAndSpaceInsensitive(t *testing.T) {
	a := valid()
	a.Endpoints = []string{"get   /API/v1/orders"}

	if missing := a.MissingEndpoints([]string{"GET /api/v1/orders"}); len(missing) != 0 {
		t.Errorf("missing = %v, want none", missing)
	}
}

// TestExpiresWithin is the input to the critical "credentials expiring soon"
// alert (task 4.3).
func TestExpiresWithin(t *testing.T) {
	a := valid() // expires in 30 days

	if a.ExpiresWithin(now, 24*time.Hour) {
		t.Error("30 days out must not read as expiring within a day")
	}
	if !a.ExpiresWithin(now, 31*24*time.Hour) {
		t.Error("30 days out must read as expiring within 31 days")
	}
	var none attest.Attestation
	if !none.ExpiresWithin(now, time.Second) {
		t.Error("an attestation with no expiry must always read as expiring")
	}
}

// TestMaskHidesEnoughToBeSafeAndShowsEnoughToBeUseful.
func TestMask(t *testing.T) {
	cases := map[string]string{
		"":            "(none)",
		"12":          "**",
		"1234":        "****",
		"123-45678":   "*****5678",
		"12345678901": "*******8901",
	}
	for in, want := range cases {
		if got := attest.Mask(in); got != want {
			t.Errorf("Mask(%q) = %q, want %q", in, got, want)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle ||
		len(needle) == 0 || indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

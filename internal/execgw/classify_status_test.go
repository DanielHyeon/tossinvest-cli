package execgw_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// TestStatusOfReadsWrappedSentinels.
//
// internal/official carries the HTTP status on authentication refusals since
// change a082, so every sentinel this maps back to a status can arrive wrapped.
// The mapping is the input to "definitive refusal or ambiguous", which is the
// input to whether a symbol gets blocked.
func TestStatusOfReadsWrappedSentinels(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		known  bool
	}{
		{"bare auth", official.ErrAuth, http.StatusUnauthorized, true},
		{"auth carrying its status code", fmt.Errorf("%w (HTTP 401)", official.ErrAuth), http.StatusUnauthorized, true},
		{"address refusal carrying its status code", fmt.Errorf("%w (HTTP 403)", official.ErrIPNotAllowed), http.StatusUnauthorized, true},
		{"rate limited maps to 429, which the journal treats as ambiguous", official.ErrRateLimited, http.StatusTooManyRequests, true},
		{"api error keeps its own code", &official.APIError{Code: 422, Body: "nope"}, 422, true},
		{"unrelated error is unknown", errors.New("nope"), 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, known := execgw.StatusOfForTest(tc.err)
			if known != tc.known {
				t.Fatalf("known = %v, want %v for %v", known, tc.known, tc.err)
			}
			if known && status != tc.status {
				t.Errorf("status = %d, want %d for %v", status, tc.status, tc.err)
			}
		})
	}
}

package official

// auth_refusal_report_test.go covers the reporting half of change a082.
//
// The token race was invisible for three days because 401 and 403 arrived as the
// same bare sentinel with the status code thrown away. An operator reading
// "official: authentication failed" could not tell a token another process had
// rotated from a credential the broker will not accept.

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestAnAuthRefusalCarriesItsStatusCode(t *testing.T) {
	for _, tc := range []struct {
		code int
		body string
		want error
	}{
		{401, `{"message":"unauthorized"}`, ErrAuth},
		{403, `{"message":"forbidden"}`, ErrAuth},
		{401, `{"message":"ip not allowed"}`, ErrIPNotAllowed},
		{403, `{"message":"IP blocked"}`, ErrIPNotAllowed},
	} {
		err := classifyStatus(tc.code, []byte(tc.body))
		if !errors.Is(err, tc.want) {
			t.Errorf("HTTP %d %q classified as %v, want %v — this change moves the "+
				"message, never the verdict", tc.code, tc.body, err, tc.want)
		}
		if code := codeIn(t, tc.code); !strings.Contains(err.Error(), code) {
			t.Errorf("HTTP %d reported as %q, which does not name %s; 401 and 403 have "+
				"different causes and one message cannot stand for both",
				tc.code, err, code)
		}
	}
}

func TestAnAuthRefusalDoesNotCarryTheResponseBody(t *testing.T) {
	const secretish = "account-18301045921-session-detail"
	for _, code := range []int{401, 403} {
		err := classifyStatus(code, []byte(`{"message":"`+secretish+`"}`))
		if strings.Contains(err.Error(), secretish) {
			t.Errorf("HTTP %d reported the response body: %q — this error is logged and "+
				"the body can carry account identifiers", code, err)
		}
	}
}

// TestNothingDecidesAnAuthRefusalByReadingItsMessage is the guard the wrap
// depends on. Every consumer in this repository asks errors.Is or errors.As, so
// changing the message text is safe. If somebody switches to matching strings,
// this is where the cost shows up.
func TestNothingDecidesAnAuthRefusalByReadingItsMessage(t *testing.T) {
	wrapped := classifyStatus(401, []byte(`{"message":"unauthorized"}`))
	odd := errors.New("official: authentication failed")

	if !errors.Is(wrapped, ErrAuth) {
		t.Fatal("the wrapped refusal no longer satisfies errors.Is(err, ErrAuth)")
	}
	if errors.Is(odd, ErrAuth) {
		t.Fatal("an unrelated error with the same text satisfies errors.Is(err, ErrAuth); " +
			"identity has stopped being identity")
	}
	if !ShouldFallback(wrapped) {
		t.Error("ShouldFallback stopped recognising a wrapped authentication refusal")
	}
	var apiErr *APIError
	if errors.As(wrapped, &apiErr) {
		t.Error("a wrapped sentinel now also unwraps to *APIError; the passthrough and " +
			"the sentinel paths must stay distinct")
	}
}

// codeIn renders the status a refusal must name.
//
// It fails loudly on anything outside the two it knows, because strings.Contains
// against "" is always true and a default of "" would silently disarm the
// assertion the moment a row is added to the table.
func codeIn(t *testing.T, code int) string {
	t.Helper()
	switch code {
	case 401, 403:
		return strconv.Itoa(code)
	}
	t.Fatalf("codeIn has no rendering for HTTP %d; add one rather than letting the "+
		"assertion pass on an empty needle", code)
	return ""
}

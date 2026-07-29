package verifylive

// endpoints.go answers one question about a finished verification: which broker
// calls did this record actually prove?
//
// # Why the question is asked here
//
// The engine's automation gate is interlocked on a capability attestation, and
// that attestation is written by the read-only soak. The soak proves reads by
// running unattended for days; it contains no mutation transport at all, so it
// can never prove that placing or cancelling an order works. Those two calls are
// in the gate's required set (internal/app/engine.RequiredEndpoints) and the only
// evidence for them in this system is this record — written by runs a person
// approved batch by batch, against a real account.
//
// # Why it reads Call.Endpoint and not the step ids
//
// Call.Endpoint is already spelled the way internal/soak and internal/app/engine
// spell their endpoint sets, on purpose, so the three compare without a
// translation table. Deriving "conditional-register means POST
// /api/v1/conditional-orders" from step ids would build exactly the translation
// table that comment exists to avoid, and would claim a call happened rather than
// observing that it did.
//
// # What this file does not decide
//
// It does not decide what an attestation may carry. It reports what the record
// proves; internal/soak decides which of those a read-only tool is allowed to
// borrow (its LiveOnlyEndpoints set) and refuses the rest. Keeping the policy out
// of here is what stops "this record contains a successful GET" from turning into
// "a supervised run can stand in for the soak".

import (
	"sort"
	"strings"
	"time"
)

// EndpointEvidence is what one verification record proves.
type EndpointEvidence struct {
	// AccountRefs are every account reference the record names, sorted and
	// deduplicated. It is a list rather than a single value because a record
	// naming two accounts is a misconfiguration the caller has to refuse, and it
	// cannot refuse what it cannot see.
	AccountRefs []string
	// Endpoints maps "METHOD /path" to the newest successful call inside the
	// window. Newest, because the question the age check asks is "when was this
	// last known to work".
	//
	// Keyed by the spelling the record used, not by a normalised form. The
	// attestation this feeds is compared against internal/app/engine's required
	// set, and that comparison is what the shared spelling exists for — handing
	// on an upper-cased key would put a fourth spelling into a chain whose whole
	// point is that there are not four. Normalisation is used to decide whether
	// two entries are the same endpoint, never to decide what to write down.
	Endpoints map[string]time.Time
}

// SucceededEndpoints reads the record for calls that worked recently enough.
//
// A call counts when three things hold, and each is a way the evidence could
// otherwise be weaker than it looks:
//
//	no error         a 422 that refused an order is not proof that placing
//	                 orders works; it is the opposite
//	not too old      age >= maxAge is already too old. The boundary is exclusive
//	                 and stated in a test, so nobody has to infer it from a
//	                 comparison operator
//	not in the future a timestamp after now means a clock that cannot be
//	                 trusted, and the safe reading of an untrustworthy clock is
//	                 that it proves nothing
//
// A zero maxAge admits nothing, which is the fail-closed direction for a caller
// that forgot to set a window.
func SucceededEndpoints(entries []Entry, now time.Time, maxAge time.Duration) EndpointEvidence {
	ev := EndpointEvidence{Endpoints: map[string]time.Time{}}
	accounts := map[string]bool{}
	// spelling keeps the first form each endpoint was written in, so two entries
	// that differ only in case land on one key instead of two.
	spelling := map[string]string{}

	for _, e := range entries {
		if ref := strings.TrimSpace(e.AccountRef); ref != "" {
			accounts[ref] = true
		}
		for _, c := range e.Calls {
			written := strings.Join(strings.Fields(strings.TrimSpace(c.Endpoint)), " ")
			if written == "" || strings.TrimSpace(c.Error) != "" {
				continue
			}
			age := now.Sub(c.At)
			if age < 0 || age >= maxAge {
				continue
			}
			key := normaliseEndpoint(written)
			name, seen := spelling[key]
			if !seen {
				name = written
				spelling[key] = name
			}
			if prev, had := ev.Endpoints[name]; !had || c.At.After(prev) {
				ev.Endpoints[name] = c.At
			}
		}
	}

	for ref := range accounts {
		ev.AccountRefs = append(ev.AccountRefs, ref)
	}
	sort.Strings(ev.AccountRefs)
	return ev
}

// normaliseEndpoint spells an endpoint the one way the three packages compare
// them. It mirrors internal/attest's own normalisation rather than importing it,
// because this package must not depend on the attestation format to read its own
// record.
func normaliseEndpoint(e string) string {
	return strings.Join(strings.Fields(strings.ToUpper(strings.TrimSpace(e))), " ")
}

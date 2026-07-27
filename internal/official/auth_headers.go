package official

// auth_headers.go exposes the headers an authenticated request carries, for the
// one caller that has to build a request this package cannot build for it
// (extend-execution-contract task 7.3).
//
// # Why this exists
//
// The idempotent replay of an IN_DOUBT order (execgw.HTTPReplay) resends the
// exact bytes that were stored when the order was first dispatched. Every method
// on this client does the opposite: it *builds* a body from structured fields,
// which is precisely what a replay must not do — a change in the serialisation
// rules would send different bytes under the same idempotency key, and the
// broker's answer to that is `422 idempotency-key-conflict` (openapi), which
// proves nothing about the original order.
//
// So the replay transport lives outside this package and needs only one thing
// from it: the Authorization and account headers, produced by the same token
// manager every other call uses. Duplicating token acquisition on the replay side
// would mean two caches, two refresh policies and two ways to be stale.
//
// # What it deliberately is not
//
// It is a read of the current credentials, not a request builder and not a token
// getter: the bearer value is assembled into a header map here so no caller ends
// up holding a raw token and inventing its own header name. It performs no
// mutation and grants no capability the caller did not already have — anything
// holding this client can already call every endpoint on it.

import (
	"context"
	"strconv"
)

// Header names the authenticated request carries. Exported so a caller building
// its own request cannot get the spelling subtly wrong.
const (
	HeaderAuthorization = "Authorization"
	HeaderAccount       = "X-Tossinvest-Account"
)

// AuthHeaders returns the headers an authenticated, account-scoped request needs.
//
// It acquires (or refreshes) the OAuth token through the shared token manager and
// resolves the account sequence lazily, exactly as the client's own verbs do. A
// caller must treat the result as valid only for the request it is about to make:
// the token expires, and this map does not.
func (c *Client) AuthHeaders(ctx context.Context) (map[string]string, error) {
	tok, err := c.tm.token(ctx)
	if err != nil {
		return nil, err
	}
	seq, err := c.ensureAccountSeq(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		HeaderAuthorization: "Bearer " + tok,
		HeaderAccount:       strconv.Itoa(seq),
	}, nil
}

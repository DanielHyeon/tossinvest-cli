package official

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// apiAccount is the official API shape for one account entry.
// Endpoint: GET /api/v1/accounts
// Schema (openapi.latest.json component "Account"):
//
//	accountNo   string  — human account number (e.g. "123-45")
//	accountSeq  integer — identifier used in subsequent API calls
//	accountType string  — enum: BROKERAGE | ...
type apiAccount struct {
	AccountNo   string `json:"accountNo"`
	AccountSeq  int    `json:"accountSeq"`
	AccountType string `json:"accountType"`
}

// Accounts fetches the authenticated user's account list from the official API
// and adapts it to []domain.Account. A fully decoded response also primes an
// unresolved client from a positive first account sequence; explicit positive
// and invalid negative configured values are preserved.
func (c *Client) Accounts(ctx context.Context) ([]domain.Account, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.accountsLocked(ctx)
}

// accountsLocked fetches and adapts the account list while the caller holds
// c.mu. Both public discovery and lazy account scoping use this one path so a
// public Accounts request that is already in flight can prime the following
// account-scoped request without a second HTTP call.
func (c *Client) accountsLocked(ctx context.Context) ([]domain.Account, error) {
	var raw []apiAccount
	if err := c.get(ctx, "/api/v1/accounts", nil, &raw); err != nil {
		return nil, err
	}
	selected := c.accountSeq.Load()
	if selected == 0 && len(raw) > 0 && raw[0].AccountSeq > 0 {
		c.accountSeq.Store(int64(raw[0].AccountSeq))
	} else if selected > 0 && !c.accountSeqExplicit {
		if len(raw) == 0 || raw[0].AccountSeq <= 0 || int64(raw[0].AccountSeq) != selected {
			return nil, fmt.Errorf(
				"account sequence drift: selected implicit sequence %d no longer matches the first account record",
				selected,
			)
		}
	}
	return adaptAccounts(raw), nil
}

// SelectedAccountSeq reports the account sequence the next account-scoped
// request would use. A false result means the value is unresolved or invalid.
// It performs no discovery and is used by engine startup to prove its journal
// account number and official request header name the same first record.
func (c *Client) SelectedAccountSeq() (int, bool) {
	seq := c.accountSeq.Load()
	return int(seq), seq > 0
}

// adaptAccounts converts a slice of official API account records to the
// canonical domain representation. It is a pure function (no I/O).
//
// Mapping rationale (cross-referenced with internal/client/account.go WTS adapter):
//
//   - accountSeq (int) → ID (string):
//     WTS uses item.Key as the stable account identifier passed to subsequent
//     calls. The official equivalent is accountSeq (an integer key). We
//     convert int→string to match the domain.Account.ID type.
//
//   - accountNo (string) → DisplayName:
//     WTS puts item.DisplayName in domain.Account.DisplayName. The official
//     accountNo is the human-readable account number and is the closest
//     analogue to the displayed identity.
//
//   - accountType (string enum) → Type:
//     WTS copies item.Type verbatim. We do the same with the official enum
//     value (e.g. "BROKERAGE") to avoid an unnecessary translation layer.
//
//   - Name, Markets, Currency, Primary: not present in the official accounts
//     endpoint; left as zero values. Currency and Primary can be enriched by
//     callers if needed.
func adaptAccounts(raw []apiAccount) []domain.Account {
	out := make([]domain.Account, 0, len(raw))
	for _, a := range raw {
		out = append(out, domain.Account{
			ID:          strconv.Itoa(a.AccountSeq), // official accountSeq → stable call-time key
			DisplayName: a.AccountNo,                // official accountNo  → human display value
			Type:        a.AccountType,              // official accountType → raw enum (e.g. BROKERAGE)
			// Name     — not available from official API
			// Markets  — not available from official API
			// Currency — not available from official API
			// Primary  — not available from official API (no primaryKey concept in this endpoint)
		})
	}
	return out
}

package official

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

// ErrAccountIdentityAuthority means an official production account read could
// not prove one exact selected account identity.
var ErrAccountIdentityAuthority = errors.New("official: account identity authority unavailable")

// VerifyAuthoritativeAccountIdentity re-reads the official account list and
// proves that accountID is the unique selected account used by account-scoped
// official calls. It returns no caller-reusable identity material.
func (c *Client) VerifyAuthoritativeAccountIdentity(ctx context.Context, accountID string) error {
	if ctx == nil || !canonicalAccountSequence(accountID) {
		return fmt.Errorf("%w: invalid request", ErrAccountIdentityAuthority)
	}
	if _, ok := c.AuthorityOrigin(); !ok {
		return ErrAuthorityOrigin
	}
	accounts, err := c.Accounts(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAccountIdentityAuthority, err)
	}
	if _, ok := c.AuthorityOrigin(); !ok {
		return ErrAuthorityOrigin
	}
	selected, ok := c.SelectedAccountSeq()
	if !ok || strconv.Itoa(selected) != accountID {
		return fmt.Errorf("%w: selected account mismatch", ErrAccountIdentityAuthority)
	}
	matches := 0
	for _, account := range accounts {
		if account.ID == accountID {
			matches++
		}
	}
	if matches != 1 {
		return fmt.Errorf("%w: expected one exact account, got %d", ErrAccountIdentityAuthority, matches)
	}
	return nil
}

func canonicalAccountSequence(value string) bool {
	if value == "" || value[0] == '0' {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	parsed, err := strconv.ParseUint(value, 10, 63)
	return err == nil && parsed > 0
}

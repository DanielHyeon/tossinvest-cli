package official

import (
	"net/http"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
)

// NewWithAccountBudget constructs an explicitly enforcing client. A failed
// budget or unsupported transport never yields a usable client.
func NewWithAccountBudget(creds Credentials, cacheFile string, budget *accountbudget.Budget, cleanup func(error), opts ...Option) (*Client, error) {
	if budget == nil || cleanup == nil {
		return nil, accountbudget.ErrUnavailable
	}
	c := New(creds, cacheFile, opts...)
	c.configMu.Lock()
	defer c.configMu.Unlock()
	if c.hc == nil {
		return nil, accountbudget.ErrUnavailable
	}
	base, ok := c.hc.Transport.(*http.Transport)
	if !ok {
		return nil, accountbudget.ErrUnavailable
	}
	wrapped, err := accountbudget.NewTransport(budget, base, cleanup)
	if err != nil {
		return nil, err
	}
	hc := *c.hc
	hc.Transport = wrapped
	c.hc = &hc
	c.tm.hc = c.hc
	c.budgetEnforced = true
	c.budget = budget
	c.budgetTransport = wrapped
	c.budgetCleanup = cleanup
	return c, nil
}

// UsesAccountBudget reports the sealed transport and token-client binding.
// It does not grant monetary authority or expose the budget.
func (c *Client) UsesAccountBudget(expected *accountbudget.Budget) bool {
	if c == nil || expected == nil {
		return false
	}
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	return c.configurationSealed && c.budgetEnforced && c.budget == expected &&
		c.budgetTransport != nil && c.hc != nil && c.hc.Transport == c.budgetTransport &&
		c.tm != nil && c.tm.hc == c.hc
}

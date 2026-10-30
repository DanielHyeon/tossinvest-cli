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
	c.budget = budget
	c.budgetTransport = wrapped
	c.budgetCleanup = cleanup
	return c, nil
}

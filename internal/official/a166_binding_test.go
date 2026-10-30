//go:build linux

package official

import (
	"net/http"
	"testing"
)

func TestA166UsesAccountBudgetBinding(t *testing.T) {
	b := a166Budget(t)
	foreign := a166Budget(t)
	var nilClient *Client
	if nilClient.UsesAccountBudget(b) {
		t.Fatal("nil client")
	}
	off := New(Credentials{}, "")
	if off.UsesAccountBudget(b) || off.UsesAccountBudget(nil) {
		t.Fatal("OFF binding")
	}
	cases := []struct {
		name  string
		alter func(*Client)
	}{
		{"none", func(*Client) {}}, {"unsealed", func(c *Client) { c.configurationSealed = false }}, {"not_enforced", func(c *Client) { c.budgetEnforced = false }},
		{"budget_missing", func(c *Client) { c.budget = nil }}, {"wrapper_missing", func(c *Client) { c.budgetTransport = nil }},
		{"wrapper_drift", func(c *Client) { c.hc.Transport = &http.Transport{} }}, {"client_missing", func(c *Client) { c.hc = nil }},
		{"token_missing", func(c *Client) { c.tm = nil }}, {"token_client_drift", func(c *Client) { c.tm.hc = &http.Client{} }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {}, WithBaseURL("https://example.invalid"))
			if e != nil {
				t.Fatal(e)
			}
			if _, ok := c.AuthorityOrigin(); ok {
				t.Fatal("test endpoint gained monetary authority")
			}
			if !c.UsesAccountBudget(b) || c.UsesAccountBudget(nil) || c.UsesAccountBudget(foreign) {
				t.Fatal("binding identity")
			}
			tt.alter(c)
			if got := c.UsesAccountBudget(b); got != (tt.name == "none") {
				t.Fatalf("binding after drift = %v", got)
			}
		})
	}
}

func TestA166UsesAccountBudgetConcurrentOptionReplay(t *testing.T) {
	b := a166Budget(t)
	c, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 1000; i++ {
			WithBaseURL("https://example.invalid")(c)
		}
	}()
	for i := 0; i < 1000; i++ {
		if !c.UsesAccountBudget(b) {
			t.Fatal("binding changed")
		}
	}
	<-done
}

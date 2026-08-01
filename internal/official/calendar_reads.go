package official

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// MarketCalendar fetches trading-hours calendar info for a market.
//
// Endpoint: GET /api/v1/market-calendar/{country} ("KR" | "US")
// Query:    date (optional) "YYYY-MM-DD"; defaults to today on the server.
// Response: {previous,today,next}BusinessDay with per-session open/close times
// (KST). KR and US return structurally different session shapes.
//
// ponytail: the payload is returned as a decoded map rather than typed structs.
// Nothing in the codebase computes over the calendar shape (it is surfaced
// verbatim as JSON to MCP callers), so faithful typing would be pure boilerplate
// with two divergent KR/US schemas. Add typed structs if a consumer ever needs
// to reason about specific fields.
func (c *Client) MarketCalendar(ctx context.Context, country, date string) (map[string]any, error) {
	country = strings.ToUpper(strings.TrimSpace(country))
	if country != "KR" && country != "US" {
		return nil, fmt.Errorf("market calendar country must be KR or US, got %q", country)
	}
	if err := validateMarketCalendarDate(date); err != nil {
		return nil, err
	}
	q := url.Values{}
	if date != "" {
		q.Set("date", date)
	}
	var out map[string]any
	if err := c.get(ctx, "/api/v1/market-calendar/"+country, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func validateMarketCalendarDate(date string) error {
	if date == "" {
		return nil
	}
	if len(date) != len("2006-01-02") || date[4] != '-' || date[7] != '-' {
		return fmt.Errorf("market calendar date must be exact YYYY-MM-DD, got %q", date)
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil || parsed.Format("2006-01-02") != date {
		return fmt.Errorf("market calendar date must be exact YYYY-MM-DD, got %q", date)
	}
	return nil
}

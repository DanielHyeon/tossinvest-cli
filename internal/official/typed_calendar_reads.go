package official

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type MarketCalendarSession struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type MarketCalendarSessions struct {
	PreMarket     *MarketCalendarSession `json:"preMarket"`
	RegularMarket *MarketCalendarSession `json:"regularMarket"`
	AfterMarket   *MarketCalendarSession `json:"afterMarket"`
}

type MarketCalendarDay struct {
	Date          string                  `json:"date"`
	Integrated    *MarketCalendarSessions `json:"integrated"`
	PreMarket     *MarketCalendarSession  `json:"preMarket"`
	RegularMarket *MarketCalendarSession  `json:"regularMarket"`
	AfterMarket   *MarketCalendarSession  `json:"afterMarket"`
}

type MarketCalendarResponse struct {
	PreviousBusinessDay MarketCalendarDay `json:"previousBusinessDay"`
	Today               MarketCalendarDay `json:"today"`
	NextBusinessDay     MarketCalendarDay `json:"nextBusinessDay"`
}

// TypedMarketCalendar is the timezone-safe computation boundary for scheduling.
func (c *Client) TypedMarketCalendar(ctx context.Context, country, date string) (MarketCalendarResponse, error) {
	country = strings.ToUpper(strings.TrimSpace(country))
	if country != "KR" && country != "US" {
		return MarketCalendarResponse{}, fmt.Errorf("market calendar country must be KR or US, got %q", country)
	}
	if err := validateMarketCalendarDate(date); err != nil {
		return MarketCalendarResponse{}, err
	}
	query := url.Values{}
	if date != "" {
		query.Set("date", date)
	}
	var out MarketCalendarResponse
	if err := c.get(ctx, "/api/v1/market-calendar/"+country, query, &out); err != nil {
		return MarketCalendarResponse{}, err
	}
	return out, nil
}

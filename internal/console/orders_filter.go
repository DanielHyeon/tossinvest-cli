package console

// orders_filter.go is the orders screen's filtering (change
// console-orders-screen, D6).
//
// # Links and query parameters, and nothing else
//
// This console has no JavaScript, no CDN and no build step. A filter is therefore
// a set of links, each carrying the whole choice in the query string, and the
// current one is rendered as text rather than as a link to itself.
//
// # The filter never reaches the broker
//
// Every one of these is applied to the reading the cache already holds. Passing
// them as broker parameters would split the cache by filter combination —
// /orders?status=live and ?status=closed would each be a key of their own — and
// the two calls a refresh costs would become four, then eight, out of the same
// §0.4 budget. The reading is one page of each endpoint; the filter is a view of
// it.

import (
	"net/http"
	"net/url"
	"strings"
)

// orderFilterChoice is the filter in effect, as the query string spelled it.
//
// An unrecognised value is dropped rather than matched against nothing: a
// hand-typed ?market=KRW would otherwise empty the table and read as "no orders
// in that market", which is a measurement nobody made.
type orderFilterChoice struct {
	Market string
	Side   string
	State  string
}

// The values each filter accepts. They are the console's own vocabulary rather
// than the broker's: "live" is internal/brokerstate's non-terminal, not the
// OPEN group label of the broker's request parameter, and conflating the two is
// the mistake brokerstate's package doc opens with.
const (
	filterStateLive   = "live"
	filterStateClosed = "closed"
)

// filterChoiceFrom reads the choice out of a request.
func filterChoiceFrom(r *http.Request) orderFilterChoice {
	q := r.URL.Query()
	return orderFilterChoice{
		Market: normaliseFilter(q.Get("market"), "KR", "US"),
		Side:   normaliseFilter(q.Get("side"), "BUY", "SELL"),
		State:  normaliseFilter(q.Get("state"), filterStateLive, filterStateClosed),
	}
}

// normaliseFilter keeps a value only when it is one this screen can mean.
func normaliseFilter(value string, allowed ...string) string {
	trimmed := strings.TrimSpace(value)
	for _, a := range allowed {
		if strings.EqualFold(trimmed, a) {
			return a
		}
	}
	return ""
}

// Applied reports that at least one filter narrows the table.
func (c orderFilterChoice) Applied() bool {
	return c.Market != "" || c.Side != "" || c.State != ""
}

// query renders the choice as a query string, for a link.
func (c orderFilterChoice) query() string {
	q := url.Values{}
	if c.Market != "" {
		q.Set("market", c.Market)
	}
	if c.Side != "" {
		q.Set("side", c.Side)
	}
	if c.State != "" {
		q.Set("state", c.State)
	}
	if len(q) == 0 {
		return "/orders"
	}
	return "/orders?" + q.Encode()
}

// with returns the choice with one axis replaced.
func (c orderFilterChoice) with(axis, value string) orderFilterChoice {
	switch axis {
	case "market":
		c.Market = value
	case "side":
		c.Side = value
	case "state":
		c.State = value
	}
	return c
}

// SummaryText names the filter in effect, for the page's one-line summary.
func (c orderFilterChoice) SummaryText() string {
	var parts []string
	if c.Market != "" {
		parts = append(parts, "시장 "+c.Market)
	}
	if c.Side != "" {
		parts = append(parts, sideLabel(c.Side))
	}
	switch c.State {
	case filterStateLive:
		parts = append(parts, "미체결")
	case filterStateClosed:
		parts = append(parts, "종결")
	}
	if len(parts) == 0 {
		return "전체"
	}
	return strings.Join(parts, " · ")
}

// orderFilterGroup is one axis of the filter, rendered as a row of links.
type orderFilterGroup struct {
	Label   string
	Options []orderFilterOption
}

// orderFilterOption is one link. On is the choice already in effect, which is
// rendered as text: a link to the page you are on is a link that teaches nothing.
type orderFilterOption struct {
	Label string
	Href  string
	On    bool
}

// filterGroups builds the three link rows for the current choice.
func filterGroups(current orderFilterChoice) []orderFilterGroup {
	build := func(label, axis string, options [][2]string) orderFilterGroup {
		g := orderFilterGroup{Label: label}
		for _, o := range options {
			value, text := o[0], o[1]
			next := current.with(axis, value)
			g.Options = append(g.Options, orderFilterOption{
				Label: text,
				Href:  next.query(),
				On:    currentValue(current, axis) == value,
			})
		}
		return g
	}
	return []orderFilterGroup{
		build("시장", "market", [][2]string{{"", "전체"}, {"KR", "KR"}, {"US", "US"}}),
		build("방향", "side", [][2]string{{"", "전체"}, {"BUY", "매수"}, {"SELL", "매도"}}),
		build("상태", "state", [][2]string{
			{"", "전체"}, {filterStateLive, "미체결"}, {filterStateClosed, "종결"},
		}),
	}
}

func currentValue(c orderFilterChoice, axis string) string {
	switch axis {
	case "market":
		return c.Market
	case "side":
		return c.Side
	default:
		return c.State
	}
}

// filterRows applies the choice to the rows already fetched.
//
// A conditional order matches the 미체결 state filter and never the 종결 one: the
// list came from the broker's OPEN group, so every row in it is watching. It
// matches no side filter at all, because the conditional payload carries no side
// — filtering by 매수 must not hide it silently, so a side filter is treated as
// not applying to conditionals rather than as excluding them.
func filterRows(rows []orderRow, c orderFilterChoice) []orderRow {
	if !c.Applied() {
		return rows
	}
	out := make([]orderRow, 0, len(rows))
	for _, r := range rows {
		if c.Market != "" && !strings.EqualFold(r.Market, c.Market) {
			continue
		}
		if c.Side != "" && !r.Conditional && !strings.EqualFold(r.Side, sideLabel(c.Side)) {
			continue
		}
		switch c.State {
		case filterStateLive:
			if !r.Live && !r.Unresolved {
				continue
			}
		case filterStateClosed:
			if r.Live || r.Unresolved {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

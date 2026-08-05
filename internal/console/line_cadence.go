package console

// line_cadence.go is change a080: the protection line is redrawn on the engine's
// cadence instead of the broker cache TTL.
//
// # The two questions that were one constant
//
// Both trading screens used to derive their reload period from holdingsTTL:
//
//	func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }
//
// holdingsTTL answers "how often may this console ask the broker" — a rate-budget
// question, settled by the account's retry matrix. How often the protection line
// should be redrawn is a different question with a different owner: the engine
// judges every engine.DefaultExitObservationInterval, and a screen on a thirty
// second reload shows a moved baseline, a promoted rung or a new pending proposal
// up to thirty seconds after the judgement that produced it.
//
// One constant standing for both meant the two could not be argued separately. It
// also meant a change to either was silently a change to the other.
//
// # Why redrawing faster costs the broker nothing
//
// The comment that used to defend the derivation said a period under the TTL
// "would be a reload that costs broker calls faster than the budget the spec
// fixes". That is not true of this cache. holdingsCache.get refreshes only when
// the last attempt is a TTL old:
//
//	if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
//		c.refreshLocked(ctx, now)
//	}
//
// so the TTL, not the reload period, bounds the call rate. Reloading six times as
// often reaches the cache six times as often and the broker exactly as often. The
// overview does not even do that — it reads through peek and makes no call at all
// (change console-operator-overview D4).
//
// # Why there is no JavaScript here
//
// The obvious way to redraw one column without reloading the page is a script
// that fetches a fragment. This console does not take that route, and the reason
// is older than this change: "JavaScript, CSP nonce, `unsafe-inline` script,
// 외부 CSS/폰트 도입" is a stated Non-Goal of the change that built these screens
// (archive/2026-07-31-streamline-trading-views), the operator-console spec fixes
// it as a CSP regression scenario, and three separate changes hold tests on it.
// The one script in this package renders on a POST-only preview sub-page that is
// deliberately outside the screen list those tests walk.
//
// The full reload is affordable here because the things a reload usually costs
// were already paid off:
//
//   - open folds survive, because a reloading screen's disclosure state lives in
//     the URL rather than in a native <details> (a055 §6)
//   - there is nothing to lose mid-edit, because these two screens carry no form,
//     input or button at all — a057 moved those to /position-management
//   - the broker is not asked again, per the bound above
//
// So the cadence moves and the mechanism does not.

import "time"

// lineRefreshInterval is how often the two protection-line screens redraw.
//
// It is the engine's observation period (engine.DefaultExitObservationInterval),
// duplicated rather than imported: the console runs as its own process and must
// not compile against the engine runtime, which is the same boundary that keeps a
// broker, a journal writer and an engine controller out of this package. If the
// two ever drift, these screens become slower than the engine — never wrong, and
// never a source of extra broker calls.
const lineRefreshInterval = 5 * time.Second

// lineRefreshSeconds is that period as the meta refresh writes it.
func lineRefreshSeconds() int { return int(lineRefreshInterval / time.Second) }

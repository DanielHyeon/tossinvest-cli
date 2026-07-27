package console

// overview.go is the operator overview (change console-operator-overview): one
// screen that answers "what is this account doing right now".
//
// # Why it exists
//
// Nothing here is newly visible. The holdings are on /positions, the round trips
// are on /history, the leftovers are on /verify and the engine's state is on /.
// What was missing is one place that holds them at once — an operator deciding
// whether to start the engine was making that join in their head, from four
// screens, every time. This screen makes the join once.
//
// # What it is not
//
// It is a 관측창. There is no form on it, no POST, no button and therefore no
// confirmation of any kind — no typed string, no second click, no extra approval
// (D11; 사용자 지시 2026-07-27). StockOS's dashboard of the same name mounts an
// order ticket and an auto-order launcher; what is borrowed here is which numbers
// belong on one screen, and not what to let somebody press.
//
// # It makes no broker call
//
// The cache is read with peek and never refreshed (D4). A screen that reloads
// itself every TTL and refreshes on each reload would put the console's
// longest-lived tab permanently on the account's rate budget, and the scarce
// thing that budget protects is a supervised live verification. The cost is
// written down rather than hidden: with /positions never opened, the account
// panel says "아직 읽지 않음" forever, so it carries a link to the screen that
// fills it.
//
// # "We could not read it" is not zero
//
// Every number is a reading: a value, whether it was measured, and — when it was
// not — which enumerated reason applies. They are separate because the operator's
// next move is different for each one: wait, fix the credentials, start the
// engine, wire the seam, open the screen that fills the cache, edit the config
// file, or look at the ledger row that will not parse. A row with no value is
// still rendered with its label, because a row that quietly disappears is not an
// unmeasured marker, it is concealment.

import (
	"context"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// --- the Guardian limits seam ---------------------------------------------------

// GateLimits is the Guardian's five ceilings, as the config file spells them.
//
// It is a console-local value rather than the config type design D8 named, for a
// reason D8 did not have in front of it: engineproc_test.go's
// TestTheConsoleDecidesNothingAboutTheGate bans the identifier AutomationGate
// from every file in this package, and it is right to. The console asks the
// engine process whether it may run and displays the answer verbatim; a console
// holding the gate's own type is one edit away from deciding for itself that the
// gate looks fine. Carrying five numbers and a currency keeps that guard intact
// and hands this package strictly less than the design proposed.
//
// Zero means "not set", which is what the file says and not what the console
// concludes: the engine's startup interlock treats an unset ceiling as a refusal
// rather than as permission, and the screen renders it as 미설정.
type GateLimits struct {
	MaxOrderQuantity   float64
	MaxOrderNotional   float64
	MaxTotalExposure   float64
	MaxDailyLossAmount float64
	MaxDailyLossRatio  float64
	// Currency is what the money ceilings are denominated in ("KRW"/"USD").
	// Empty is its own state and the screen says so rather than guessing.
	Currency string
}

// GateLimitsReader reads those ceilings. One method.
//
// It is a seam of its own rather than a third method on AdoptionSettings (D8).
// That seam is the console's config *write* surface and the operator-console spec
// fixes its shape at exactly Load and Save — a rule a static test enforces. The
// discipline that rule expresses is "a seam carries one job", and the way to keep
// it while gaining a second job is to add a seam, not to widen one: a screen that
// only wants to display a ceiling must not thereby hold the ability to write the
// adoption block.
//
// Reading is all it does. The ceilings themselves are a §0.7 human decision taken
// outside any browser, and the console cannot edit them (operator-console:
// 게이트·Guardian 한도·kill switch는 콘솔에서 편집할 수 없다).
type GateLimitsReader interface {
	GateLimits() (GateLimits, error)
}

// --- the reasons a number is missing (D5) ---------------------------------------

// unmeasuredReason names why a value is absent. Each one is kept apart from the
// others because each asks the operator for a different thing.
//
// The enumeration's job is to write down what the operator has to fix, without
// remainder. A reason that exists only as a free sentence cannot be counted and
// cannot be tested for absence, so the next one gets written as a sentence too
// and at that point the enumeration no longer describes the surface (I-2, Manager
// ruling 2026-07-28). Two reasons have been added that way since the draft, each
// because it asks for something none of the others does: config_unreadable sends
// the operator to a config file, and journal_value_unparsable sends them to a row
// in the ledger.
type unmeasuredReason string

const (
	// reasonVerifySuspended: a live verification holds the rate budget. Wait.
	reasonVerifySuspended unmeasuredReason = "verify_suspended"
	// reasonBrokerReadFailed: 429, a network fault, an expired credential. Fix it.
	reasonBrokerReadFailed unmeasuredReason = "broker_read_failed"
	// reasonJournalUnreadable: the ledger is absent, or its schema disagrees with
	// this binary. Start the engine, or update the console.
	reasonJournalUnreadable unmeasuredReason = "journal_unreadable"
	// reasonSeamUnwired: this build was not handed that capability. Wire it.
	reasonSeamUnwired unmeasuredReason = "seam_unwired"
	// reasonNeverFetched: the cache has never been filled. Open the screen that
	// fills it — nothing else will.
	reasonNeverFetched unmeasuredReason = "never_fetched"
	// reasonConfigUnreadable: the seam IS wired and the config it reads could not
	// be parsed. Edit the file. It is none of the five above — the seam is there,
	// no broker and no ledger is involved, and no cache is cold — and it was the
	// one unmeasured state on this screen an operator could not grep for while it
	// was rendered as a code-less sentence.
	reasonConfigUnreadable unmeasuredReason = "config_unreadable"
	// reasonJournalValueUnparsable: the ledger opened, answered, and one of the
	// values it answered with cannot be interpreted — a closed_at that is not a
	// timestamp, a realized_pnl_after_costs that is not a number. It is separate
	// from journal_unreadable because that one's advice is "start the engine so it
	// migrates", and doing that to a ledger that opened perfectly well changes
	// nothing. The row is what has to be looked at.
	reasonJournalValueUnparsable unmeasuredReason = "journal_value_unparsable"
)

// unmeasuredSentences is what each reason tells the operator to do. A bare "—"
// does not distinguish waiting from fixing from wiring, which is the whole reason
// there are codes rather than one boolean.
var unmeasuredSentences = map[unmeasuredReason]string{
	reasonVerifySuspended: "검증 중 — 갱신 보류. 실계좌 검증이 rate limit을 쓰고 있다. 끝나면 다시 읽힌다.",
	reasonBrokerReadFailed: "브로커 조회 실패 — 인증·rate limit·네트워크 중 하나다. " +
		"포지션 화면에 마지막 실패 사유가 그대로 남아 있다.",
	reasonJournalUnreadable: "원장 미판독 — 엔진을 한 번 기동해 마이그레이션하거나, 콘솔이 원장보다 " +
		"오래됐다면 콘솔을 갱신하라.",
	reasonSeamUnwired:  "seam 미배선 — 이 빌드의 콘솔에 그 조회가 연결되어 있지 않다.",
	reasonNeverFetched: "아직 읽지 않음 — 이 값을 채우는 화면을 한 번 열어야 한다.",
	reasonConfigUnreadable: "설정 판독 불가 — seam은 배선돼 있는데 config 파일을 파싱하지 못했다. " +
		"설정 파일을 고쳐라. 0도 무제한도 아니고, 읽히지 않았다는 뜻이다.",
	reasonJournalValueUnparsable: "원장 값 해석 불가 — 원장은 열렸고 그 안의 값 하나를 숫자나 시각으로 " +
		"읽을 수 없다. 엔진을 다시 기동해도 달라지지 않는다; 해당 행을 봐야 한다.",
}

// reading is one number the screen either has or does not, with the reason.
//
// The zero value is unmeasured with no sentence, which is the one state the
// screen must never render; overview_test.go's
// TestNoUnmeasuredValueOnTheScreenIsAReasonlessDash fails on it.
type reading struct {
	text   string
	known  bool
	reason unmeasuredReason
	// detail carries the specific sentence beside a code — the parse error the
	// config seam returned, the ledger value that would not read — or, for the
	// rare case with no code at all, the whole reason. It is never empty when it
	// is used, so a rendered "미측정" always carries something an operator can act
	// on.
	detail string
}

// measured is a value that was read.
func measured(text string) reading { return reading{text: text, known: true} }

// unmeasuredFor is a value that was not read, for one of the enumerated reasons.
func unmeasuredFor(r unmeasuredReason) reading { return reading{reason: r} }

// unmeasuredWithDetail is one of the enumerated reasons plus the specific thing
// that could not be read.
//
// The code is what an operator greps the page for and what a test can assert the
// absence of; the detail is what tells them which file or which row. Neither
// replaces the other — a code with no detail sends somebody to a config file
// without saying what is wrong with it, and a detail with no code is the state
// the ruling on I-2 removed.
func unmeasuredWithDetail(r unmeasuredReason, detail string) reading {
	out := reading{reason: r}
	if trimmed := strings.TrimSpace(detail); trimmed != "" {
		out.detail = trimmed
	}
	return out
}

// unmeasuredBecause is a value that was not read for a reason outside the
// enumeration — the sentence itself is the reason, and it must not be empty.
func unmeasuredBecause(detail string) reading {
	if strings.TrimSpace(detail) == "" {
		detail = "사유를 특정할 수 없다"
	}
	return reading{detail: detail}
}

// Known reports that the value was measured.
func (r reading) Known() bool { return r.known }

// Value renders the number, or the em dash the reason accompanies.
func (r reading) Value() string {
	if !r.known {
		return "—"
	}
	return r.text
}

// Code is the machine reason, for the operator who wants to search the page.
func (r reading) Code() string {
	if r.known {
		return ""
	}
	return string(r.reason)
}

// Why is the sentence beside an unmeasured value. It is never empty for an
// unmeasured reading — a reason-less dash tells nobody whether to wait, to fix
// something or to wire something.
//
// A reading that has both a code and a detail renders both: the code's sentence
// says what kind of thing is wrong and the detail says which one.
func (r reading) Why() string {
	if r.known {
		return ""
	}
	sentence := unmeasuredSentences[r.reason]
	detail := strings.TrimSpace(r.detail)
	switch {
	case sentence != "" && detail != "":
		return sentence + " (" + detail + ")"
	case sentence != "":
		return sentence
	case detail != "":
		return detail
	default:
		return "사유를 특정할 수 없다"
	}
}

// --- the screen -----------------------------------------------------------------

// overviewRecentEvents bounds the recent-activity panel. It is short on purpose:
// the full stream is the history screen's, and a summary that needs scrolling is
// not a summary.
const overviewRecentEvents = 12

// overviewView is the whole screen.
//
// Every panel is assembled independently and none of them can take another one
// down (task 4.11): a missing journal empties the today and recent panels and
// leaves the engine and safety panels working, which is what an operator needs
// when they open this screen precisely because they do not know what is running.
type overviewView struct {
	Now time.Time
	// Journal is the one verdict the ledger gave this render. The three panels
	// that read it share it, and the page prints the notice once.
	Journal journalView
	Engine  engineView
	Account accountPanel
	Today   todayPanel
	Open    openOrdersPanel
	Safety  safetyPanel
	Recent  recentPanel
	// VerifyRunning reports a supervised live verification in progress, and
	// VerifyNote says which process is running it.
	//
	// It is a note and not a reason. This screen refreshes nothing, so nothing is
	// being withheld from it; what a running verification actually costs the
	// operator is that the way out of a cold cache — opening /positions — is
	// itself suspended until the run ends. That is worth saying, and it is not
	// the same sentence as "this value was not read because a verification is
	// running", which on this screen would be false twice over.
	VerifyRunning bool
	VerifyNote    string
}

// NowText renders the instant the screen was built.
func (v overviewView) NowText() string { return v.Now.UTC().Format("2006-01-02 15:04:05Z") }

// --- account -----------------------------------------------------------------------

// accountPanel is what the account holds, per market.
//
// There is no total. domain.Position carries a MarketType and no currency, so a
// KR row is KRW and a US row is USD, and one number made by adding them means
// nothing at all. Not producing a total is better than producing a wrong one
// (spec: 시장을 가로지르는 합계를 만들지 않는다).
type accountPanel struct {
	// Broker is the cache reading the numbers came from — never refreshed here.
	Broker holdingsSnapshot
	// Journal is the ledger state behind the managed/unmanaged split.
	Journal journalView
	Markets []accountMarketRow
}

// accountMarketRow is one market's holdings. Each cell is its own reading, so
// the ledger failing does not take the broker numbers down with it.
type accountMarketRow struct {
	Market   string
	Currency string
	// Unreadable, when set, is the reason no cell in this row could be filled.
	// The row still renders, with its label and the reason.
	Unreadable reading
	// Blocked reports that Unreadable is the whole row.
	Blocked bool
	// Empty reports a market this account holds nothing in. It is not a value of
	// zero: the row says 보유 없음 rather than asserting an evaluation of 0.
	Empty     bool
	Value     reading
	PnL       reading
	Symbols   reading
	Managed   reading
	Unmanaged reading
	// Gain colours the P&L.
	Gain bool
}

// --- today -------------------------------------------------------------------------

// todayPanel is what closed today, per market.
//
// "Today" is each market's own local midnight (D6). trade_outcomes.closed_at is
// stored in UTC and UTC midnight falls at 09:00 KST — an hour into the Korean
// session — so a UTC day would file that morning's fills under yesterday. The
// screen prints which boundary each row used, because choosing a boundary and
// hiding the choice are different mistakes, and the second one has two people
// reading the same screen as two different days.
type todayPanel struct {
	Journal journalView
	Markets []todayMarketRow
}

type todayMarketRow struct {
	Market string
	// Boundary is the local midnight this row counted from, spelled out.
	Boundary string
	Currency string
	// NoTrades reports that nothing closed today in this market. It renders as
	// 거래 없음 rather than as a realised P&L of zero.
	NoTrades bool
	// Unreadable, when set, is why this row has no cells at all.
	Unreadable reading
	Blocked    bool
	Trips      reading
	PnL        reading
	Wins       reading
	Losses     reading
	Gain       bool
}

// --- open orders --------------------------------------------------------------------

// openOrdersPanel is how many live orders the account is carrying.
//
// It is unmeasured in this change and the reason is seam_unwired, not zero. The
// order-reading seam lands in console-orders-screen, where the data it needs is
// being made readable; until then a 0 here would read as "미체결 없음", which is
// the exact failure this change's own rule exists to prevent. This is the first
// place that rule is applied, and it is applied to this change.
type openOrdersPanel struct {
	Count reading
}

// --- safety -----------------------------------------------------------------------

// safetyPanel is the verification state, the leftovers and the Guardian limits.
//
// It reads BOTH markets. c.snapshot() is the KR reading and the two markets keep
// separate evidence files, so a KR-only leftover panel hides US leftovers — and a
// hidden leftover is the thing this panel exists for.
type safetyPanel struct {
	Markets  []safetyMarketRow
	Guardian guardianPanel
}

type safetyMarketRow struct {
	Market  string
	Record  string
	Present bool
	Error   string
	Done    int
	Total   int
	// Outstanding is what the last run left live on the account.
	Outstanding []verifylive.Artifact
}

// Leftovers reports that this market is carrying something.
func (r safetyMarketRow) Leftovers() bool { return len(r.Outstanding) > 0 }

// guardianPanel is the Guardian's ceilings, read and never written.
//
// Whether the gate is open is deliberately not on this panel. That is the engine
// process's judgement, made against its own startup interlock, and the console's
// whole relationship with it is to display the refusal the engine printed — a
// screen that also drew its own ON/OFF badge would be inviting the reader to
// believe the console had checked.
type guardianPanel struct {
	// Wired reports that a limits seam was injected at all.
	Wired bool
	// Read reports that the seam answered. It is separate from Wired because the
	// screen may only describe the config file when something actually read it:
	// a wired-but-failed seam knows nothing about the ceilings and nothing about
	// the currency either, and "한도 통화 (미지정)" from that branch is a statement
	// about a file nobody opened.
	Read bool
	Err  string
	// Currency is what the money ceilings are denominated in, as the config
	// spells it. Empty is its own state and the screen says so — but only when
	// Read, because unset and unread are different things.
	Currency string
	Limits   []guardianLimit
	Axes     []guardianAxis
}

type guardianLimit struct {
	Label string
	Value reading
}

// guardianAxis is one dimension of the gate, and how much of it today has used.
//
// Only one axis is computable: today's realised P&L against max_daily_loss_amount,
// because the ledger froze the realised figure. The other three — open exposure,
// order quantity, order notional — are computed in the engine's process and
// written to no table anywhere, so they are structurally unmeasured and the screen
// says exactly that. Deriving something plausible from the holdings would render
// as a measured limit consumption, and a screen that makes somebody trust a
// protection that is not there is worse than a screen with no limits on it.
type guardianAxis struct {
	Label string
	// Structural is the sentence for an axis no record exists for, "" otherwise.
	Structural string
	Rows       []guardianAxisRow
}

type guardianAxisRow struct {
	Market   string
	Consumed reading
	// Note explains a comparison that was not made — a limit denominated in a
	// currency this market does not trade in, or a limit nobody has set.
	Note string
}

// --- recent -------------------------------------------------------------------------

// recentPanel is the newest exit judgements.
//
// Fills are deliberately absent. StockOS shows a recent-fills strip; journal.
// ReadOnly has no fills accessor (AccountRefs, LivePositionExits,
// AccountExitEvents and AccountTradeTrips are the whole surface), and adding one
// is a change to the ledger's read surface that belongs to console-orders-screen.
type recentPanel struct {
	Journal journalView
	Events  []eventRow
	// Unreadable is why there are no events to show, when the answer is not
	// "there were none".
	Unreadable reading
	Blocked    bool
}

// --- assembly -----------------------------------------------------------------------

// overview builds the screen. Each panel is independent, and none of them
// returns an error: every failure here is a state the page has words for.
func (c *Console) overview(ctx context.Context) overviewView {
	now := c.now()
	v := overviewView{Now: now}

	installed, _ := c.opts.Binary()
	v.Engine = c.readEngine(now, installed)

	live, trips, events, jv := c.overviewLedger(ctx)
	v.Journal = jv
	v.Account = c.accountPanelFrom(now, live, jv)
	v.Today = todayPanelFrom(now, trips, jv)
	v.Recent = recentPanelFrom(events, jv)

	v.Open = openOrdersPanel{Count: unmeasuredFor(reasonSeamUnwired)}
	v.Safety = c.safetyPanelFrom(v.Today)
	v.VerifyRunning, v.VerifyNote = c.verifyHold(now)
	return v
}

// overviewLedger reads everything this screen needs out of the ledger, through
// ONE handle and ONE journalView.
//
// It was two opens and two views: the account panel called livePositions and the
// today and recent panels called a second reader, each opening the file, each
// deciding for itself whether the ledger was readable. Two readings of one file
// can disagree — the engine writes to it while the page renders — and the screen
// then showed one panel's holdings split against another panel's "원장 미판독",
// with the ledger-state notice printed twice underneath. One open, one verdict,
// one notice: the panels can still be empty for different reasons, but they
// cannot disagree about whether the file answered.
//
// A partial read keeps whatever it got and names the failure, like every other
// ledger read on this console.
func (c *Console) overviewLedger(ctx context.Context) ([]journal.PositionExit,
	[]journal.TradeTrip, []eventRow, journalView) {
	ro, jv := c.openJournal(ctx)
	if ro == nil {
		return nil, nil, nil, jv
	}
	defer ro.Close()

	var (
		live   []journal.PositionExit
		trips  []journal.TradeTrip
		events []eventRow
	)

	refs, err := ro.AccountRefs(ctx)
	if err != nil {
		jv.State, jv.Detail = journalFailed, err.Error()
		return nil, nil, nil, jv
	}

	for _, ref := range refs {
		open, err := ro.LivePositionExits(ctx, ref)
		if err != nil {
			jv.State, jv.Detail = journalFailed, err.Error()
			return live, trips, events, jv
		}
		live = append(live, open...)

		got, err := ro.AccountTradeTrips(ctx, ref)
		if err != nil {
			jv.State, jv.Detail = journalFailed, err.Error()
			return live, trips, events, jv
		}
		trips = append(trips, got...)

		seen, err := ro.AccountExitEvents(ctx, ref, overviewRecentEvents)
		if err != nil {
			jv.State, jv.Detail = journalFailed, err.Error()
			return live, trips, events, jv
		}
		for _, e := range seen {
			events = append(events, eventRow{
				Symbol:    e.Symbol,
				At:        e.Event.CreatedAt,
				Observed:  dashIfEmpty(e.Event.ObservedPrice),
				HighWater: dashIfEmpty(e.Event.HighWater),
				Baseline:  dashIfEmpty(e.Event.BaselineAfter),
				Level:     dashIfEmpty(e.Event.LevelAfter),
				Action:    e.Event.Action,
				IntentID:  e.Event.ProposedIntentID,
			})
		}
	}
	return live, trips, events, jv
}

// overviewMarkets is the two markets, in the order the screen lists them. Both
// rows always render: a market with nothing in it says so, and a market that
// silently vanished would be indistinguishable from one carrying no risk.
var overviewMarkets = []struct {
	Market   clock.Market
	Label    string
	Currency string
}{
	{clock.MarketKR, verifylive.MarketKR, "KRW"},
	{clock.MarketUS, verifylive.MarketUS, "USD"},
}

// overviewOtherMarket labels the row that holds everything the two markets above
// did not match.
//
// A holding or a round trip whose market is neither KR nor US used to be dropped
// on the floor: no row, no marker, and /positions still listing it. Two screens
// in one console then disagreed about what the account holds, and the rule this
// change spends most of its words on — a value that vanishes is not an unmeasured
// marker, it is concealment — was being broken on a symbol instead of on a cell.
// The row is deliberately unmeasured rather than summed: nothing here knows what
// currency those rows are in, and D6 forbids inventing one.
const overviewOtherMarket = "기타/미분류"

// overviewMatchesMarket reports a market string belonging to one of the two
// markets the screen splits by.
func overviewMatchesMarket(market string) bool {
	for _, m := range overviewMarkets {
		if strings.EqualFold(strings.TrimSpace(market), m.Label) {
			return true
		}
	}
	return false
}

// overviewOtherSymbols is the sorted, de-duplicated "market:symbol" list an
// 기타/미분류 row names, so an operator can find the rows on /positions.
func overviewOtherSymbols(pairs []string) string {
	seen := map[string]bool{}
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}

// accountPanelFrom builds the holdings panel from the cache and the ledger.
func (c *Console) accountPanelFrom(now time.Time, live []journal.PositionExit,
	jv journalView) accountPanel {
	snap := c.holdings.peek(now)
	panel := accountPanel{Broker: snap, Journal: jv}

	rows := joinPositions(snap.Rows, live, jv.Readable())
	usable := brokerReadable(snap)

	for _, m := range overviewMarkets {
		row := accountMarketRow{Market: m.Label, Currency: m.Currency}
		if !usable.Known() {
			row.Unreadable, row.Blocked = usable, true
			panel.Markets = append(panel.Markets, row)
			continue
		}

		var (
			value, pnl               float64
			symbols, managed, others int
			unknown                  int
		)
		for _, r := range rows {
			if !r.InBroker || !strings.EqualFold(r.Market, m.Label) {
				continue
			}
			symbols++
			value += r.MarketValue
			pnl += r.UnrealizedPnL
			switch {
			case r.Unknown():
				unknown++
			case r.Managed():
				managed++
			default:
				others++
			}
		}

		if symbols == 0 {
			row.Empty = true
			panel.Markets = append(panel.Markets, row)
			continue
		}
		row.Value = measured(decimalText(value))
		row.PnL = measured(decimalText(pnl))
		row.Gain = pnl >= 0
		row.Symbols = measured(strconv.Itoa(symbols))
		if unknown > 0 {
			row.Managed = unmeasuredFor(reasonJournalUnreadable)
			row.Unmanaged = unmeasuredFor(reasonJournalUnreadable)
		} else {
			row.Managed = measured(strconv.Itoa(managed))
			row.Unmanaged = measured(strconv.Itoa(others))
		}
		panel.Markets = append(panel.Markets, row)
	}

	// Everything the two markets did not match. Without this row the holding is
	// simply not on the screen, while /positions goes on listing it.
	if usable.Known() {
		var pairs []string
		for _, r := range rows {
			if !r.InBroker || overviewMatchesMarket(r.Market) {
				continue
			}
			pairs = append(pairs, strings.ToUpper(r.Market)+":"+r.Symbol)
		}
		if len(pairs) > 0 {
			panel.Markets = append(panel.Markets, accountMarketRow{
				Market:   overviewOtherMarket,
				Currency: "통화 불명",
				Blocked:  true,
				Unreadable: unmeasuredBecause("KR·US 어느 쪽도 아닌 시장의 보유가 " +
					strconv.Itoa(len(pairs)) + "건 있다 (" + overviewOtherSymbols(pairs) + "). " +
					"이 화면은 시장별 통화로만 값을 내므로 통화를 모르는 보유의 평가액은 만들지 않는다 — " +
					"수량과 단가는 포지션 화면에 그대로 있다."),
			})
		}
	}
	return panel
}

// brokerReadable reports that the cache holds a usable reading, or names why it
// does not.
//
// A cache that has never been filled reads as never_fetched even while a
// verification is running. The code used to say verify_suspended there, and on
// this screen that sentence — "검증 중 — 갱신 보류 … 끝나면 다시 읽힌다" — is false
// in both directions: nothing is being withheld, because the overview makes no
// broker call at all by contract (D4), and the value will NOT be read when the
// verification ends, because only opening /positions ever fills this cache.
// It told the operator to wait for something that was never going to happen.
//
// What a running verification does cost them is that the way out — opening
// /positions — is itself suspended for the duration. That is a note beside the
// panel (overviewView.VerifyNote) rather than a reason on the value, because it
// is a fact about the other screen and not about this reading.
func brokerReadable(snap holdingsSnapshot) reading {
	switch {
	case !snap.Wired:
		return unmeasuredFor(reasonSeamUnwired)
	case snap.Present:
		return measured("")
	case snap.Error != "":
		return unmeasuredFor(reasonBrokerReadFailed)
	default:
		return unmeasuredFor(reasonNeverFetched)
	}
}

// todayPanelFrom counts what closed since each market's own local midnight.
func todayPanelFrom(now time.Time, trips []journal.TradeTrip, jv journalView) todayPanel {
	panel := todayPanel{Journal: jv}

	for _, m := range overviewMarkets {
		row := todayMarketRow{Market: m.Label, Currency: m.Currency}

		start, label, err := marketDayStart(m.Market, now)
		if err != nil {
			row.Boundary = "(시간대를 읽을 수 없다)"
			row.Unreadable = unmeasuredBecause("시장의 시간대를 읽을 수 없다: " + err.Error())
			row.Blocked = true
			panel.Markets = append(panel.Markets, row)
			continue
		}
		row.Boundary = label

		if !jv.Readable() {
			row.Unreadable, row.Blocked = unmeasuredFor(reasonJournalUnreadable), true
			panel.Markets = append(panel.Markets, row)
			continue
		}

		total := new(big.Rat)
		count, wins, losses := 0, 0, 0
		broken := ""
		for _, trip := range trips {
			if !strings.EqualFold(trip.Market, m.Label) {
				continue
			}
			closed, perr := time.Parse(time.RFC3339, strings.TrimSpace(trip.Outcome.ClosedAt))
			if perr != nil {
				broken = "원장의 청산 시각 " + trip.Outcome.ClosedAt +
					" 을(를) 해석할 수 없어 어느 날짜인지 정할 수 없다"
				break
			}
			if closed.Before(start) {
				continue
			}
			amount, ok := new(big.Rat).SetString(strings.TrimSpace(trip.Outcome.RealizedPnLAfterCosts))
			if !ok {
				broken = "원장의 실현손익 " + trip.Outcome.RealizedPnLAfterCosts +
					" 을(를) 숫자로 읽을 수 없다"
				break
			}
			total.Add(total, amount)
			count++
			if amount.Sign() < 0 {
				losses++
			} else {
				wins++
			}
		}

		switch {
		case broken != "":
			// journal_value_unparsable and not journal_unreadable: the ledger
			// opened and answered. "엔진을 한 번 기동해 마이그레이션하라" would send
			// the operator to do something that changes nothing.
			row.Unreadable = unmeasuredWithDetail(reasonJournalValueUnparsable,
				broken+" — 이 시장의 오늘은 합산하지 않았다.")
			row.Blocked = true
		case count == 0:
			row.NoTrades = true
		default:
			row.Trips = measured(strconv.Itoa(count))
			row.PnL = measured(ratText(total))
			row.Wins = measured(strconv.Itoa(wins))
			row.Losses = measured(strconv.Itoa(losses))
			row.Gain = total.Sign() >= 0
		}
		panel.Markets = append(panel.Markets, row)
	}

	// Round trips in a market neither row counted. Same rule as the account
	// panel's: a trip that matches nothing is named rather than dropped, because
	// a realised P&L that is on /history and not here is two screens disagreeing.
	if jv.Readable() {
		var pairs []string
		for _, trip := range trips {
			if overviewMatchesMarket(trip.Market) {
				continue
			}
			pairs = append(pairs, strings.ToUpper(strings.TrimSpace(trip.Market))+":"+trip.Symbol)
		}
		if len(pairs) > 0 {
			panel.Markets = append(panel.Markets, todayMarketRow{
				Market:   overviewOtherMarket,
				Currency: "통화 불명",
				Boundary: "(경계 없음 — 이 시장의 현지 자정을 모른다)",
				Blocked:  true,
				Unreadable: unmeasuredBecause("KR·US 어느 쪽도 아닌 시장의 왕복이 " +
					strconv.Itoa(len(pairs)) + "건 있다 (" + overviewOtherSymbols(pairs) + "). " +
					"현지 자정도 통화도 모르므로 오늘인지 판정하지 않고 합산하지도 않는다 — " +
					"전체는 거래 이력 화면에 있다."),
			})
		}
	}
	return panel
}

// marketDayStart is the market-local midnight the day containing now began at,
// and the sentence the screen prints to say which boundary that was.
func marketDayStart(m clock.Market, now time.Time) (time.Time, string, error) {
	loc, err := m.Location()
	if err != nil {
		return time.Time{}, "", err
	}
	local := now.In(loc)
	year, month, day := local.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, loc)
	return start, start.Format("2006-01-02 15:04 MST") + " (" + loc.String() + ") 이후", nil
}

// recentPanelFrom takes the newest judgements, newest first.
func recentPanelFrom(events []eventRow, jv journalView) recentPanel {
	panel := recentPanel{Journal: jv}
	if !jv.Readable() {
		panel.Unreadable, panel.Blocked = unmeasuredFor(reasonJournalUnreadable), true
		return panel
	}
	panel.Unreadable = measured("")
	for i := len(events) - 1; i >= 0 && len(panel.Events) < overviewRecentEvents; i-- {
		panel.Events = append(panel.Events, events[i])
	}
	return panel
}

// safetyPanelFrom reads both markets' evidence records and the gate's limits.
func (c *Console) safetyPanelFrom(today todayPanel) safetyPanel {
	var panel safetyPanel
	for _, m := range overviewMarkets {
		view := c.readVerify(m.Label)
		panel.Markets = append(panel.Markets, safetyMarketRow{
			Market:      m.Label,
			Record:      view.Record,
			Present:     view.Present,
			Error:       view.Error,
			Done:        view.Done,
			Total:       view.Total,
			Outstanding: view.Outstanding,
		})
	}
	panel.Guardian = c.guardianPanelFrom(today)
	return panel
}

// guardianPanelFrom reads the limits and computes the one axis that has a record.
func (c *Console) guardianPanelFrom(today todayPanel) guardianPanel {
	panel := guardianPanel{}
	if c.opts.GateLimits == nil {
		absent := unmeasuredFor(reasonSeamUnwired)
		panel.Limits = sharedLimits(absent)
		panel.Axes = append([]guardianAxis{dailyLossAxis(today, GateLimits{}, absent)},
			structuralAxes()...)
		return panel
	}

	panel.Wired = true
	gate, err := c.opts.GateLimits.GateLimits()
	if err != nil {
		panel.Err = err.Error()
		absent := unmeasuredWithDetail(reasonConfigUnreadable, "설정에서 한도를 읽지 못했다: "+err.Error())
		panel.Limits = sharedLimits(absent)
		panel.Axes = append([]guardianAxis{dailyLossAxis(today, GateLimits{}, absent)},
			structuralAxes()...)
		// Read stays false, and the screen therefore says nothing about the limit
		// currency. It used to render 한도 통화 (미지정) here, which is a claim about
		// the file's contents made by a branch that never read the file — the same
		// collapse of "unread" into "unset" that GateLimits.Currency's own doc says
		// this screen does not make.
		return panel
	}

	panel.Read = true
	panel.Currency = strings.TrimSpace(gate.Currency)
	panel.Limits = []guardianLimit{
		{Label: "1회 주문 수량 상한", Value: limitReading(gate.MaxOrderQuantity)},
		{Label: "1회 주문 금액 상한", Value: limitReading(gate.MaxOrderNotional)},
		{Label: "계좌 개방 노출 상한", Value: limitReading(gate.MaxTotalExposure)},
		{Label: "일일 손실 한도(금액)", Value: limitReading(gate.MaxDailyLossAmount)},
		{Label: "일일 손실 한도(비율)", Value: limitReading(gate.MaxDailyLossRatio)},
	}
	panel.Axes = append([]guardianAxis{dailyLossAxis(today, gate, measured(""))}, structuralAxes()...)
	return panel
}

// limitReading renders one configured ceiling. Zero is "not set" — which the
// engine's startup interlock treats as a refusal rather than as permission — and
// it is a measured fact about the file, not a missing reading.
func limitReading(v float64) reading {
	if v <= 0 {
		return measured("미설정")
	}
	return measured(decimalText(v))
}

// sharedLimits renders the five ceilings with one shared reason.
func sharedLimits(why reading) []guardianLimit {
	return []guardianLimit{
		{Label: "1회 주문 수량 상한", Value: why},
		{Label: "1회 주문 금액 상한", Value: why},
		{Label: "계좌 개방 노출 상한", Value: why},
		{Label: "일일 손실 한도(금액)", Value: why},
		{Label: "일일 손실 한도(비율)", Value: why},
	}
}

// structuralAxes are the three dimensions no record exists for anywhere.
//
// OpenExposure and the per-order figures are computed inside the engine process
// and handed to the Guardian; the ledger's tables hold none of them. Waiting does
// not make them appear, which is why they are named 구조적으로 미측정 rather than
// given one of the enumerated reasons — not one of them would be true here, and
// each of them tells the operator to go and do something that would not help.
func structuralAxes() []guardianAxis {
	const why = "구조적으로 미측정 — 이 소진분은 엔진 프로세스가 메모리에서 계산해 Guardian에 넘기는 값이고, " +
		"원장의 어떤 테이블에도 기록이 없다. 콘솔이 보유에서 무언가를 계산해 여기 적으면 그것은 측정된 " +
		"한도 소진분처럼 보이지만 아무도 관측한 적이 없는 숫자다."
	return []guardianAxis{
		{Label: "계좌 개방 노출", Structural: why},
		{Label: "1회 주문 수량", Structural: why},
		{Label: "1회 주문 금액", Structural: why},
	}
}

// dailyLossAxis is the one axis with a frozen record behind it.
//
// The comparison is per market and it is only made when the limit's currency is
// the market's own: max_daily_loss_amount is one number in one currency, today's
// realised P&L is a number per market, and dividing one by the other across a
// currency boundary produces a percentage of nothing.
//
// The currency test comes BEFORE the no-trades test, and the order is the whole
// point. It used to come after, so a US market that had closed nothing rendered
// "0 / 1,000,000" — a KRW ceiling, a USD row, no note, and the pair displayed as
// a measured consumption. The quiet days are exactly the days nobody looks
// closely, and D6's rule does not take the day off because the market did.
func dailyLossAxis(today todayPanel, gate GateLimits, limit reading) guardianAxis {
	axis := guardianAxis{Label: "오늘 실현손익 대 일일 손실 한도(금액)"}
	currency := strings.TrimSpace(gate.Currency)

	for _, m := range overviewMarkets {
		row := guardianAxisRow{Market: m.Label}
		var day todayMarketRow
		for _, t := range today.Markets {
			if t.Market == m.Label {
				day = t
			}
		}

		switch {
		case !limit.Known():
			row.Consumed = limit
		case day.Blocked:
			row.Consumed = day.Unreadable
		case gate.MaxDailyLossAmount <= 0:
			row.Consumed = measured(axisTodayText(day) + " / 미설정")
			row.Note = "한도가 설정되어 있지 않다 — 게이트를 켤 때 엔진이 기동을 거부하는 조건이다."
		case currency == "" || !strings.EqualFold(currency, m.Currency):
			row.Consumed = measured(axisTodayText(day) + " " + m.Currency + " / " +
				limitText(gate.MaxDailyLossAmount) + " " + currencyText(currency))
			row.Note = "한도 통화와 이 시장의 통화가 같지 않아 비율은 계산하지 않았다 — " +
				"통화를 가로지르는 비율은 아무 뜻도 없다."
		case day.NoTrades:
			row.Consumed = measured("거래 없음 / " + limitText(gate.MaxDailyLossAmount) + " " + currency)
			row.Note = "오늘 청산된 왕복이 없다 — 실현손익 0이 아니라 대상이 없다는 뜻이다."
		case !day.PnL.Known():
			row.Consumed = day.PnL
		default:
			row.Consumed = measured(day.PnL.Value() + " / " + limitText(gate.MaxDailyLossAmount) +
				" " + currency + " (" + consumedPercent(day.PnL.text, gate.MaxDailyLossAmount) + ")")
		}
		axis.Rows = append(axis.Rows, row)
	}
	return axis
}

// axisTodayText renders today's realised figure for an axis row: the number, or
// the fact that there was nothing to realise. It never renders a zero for an
// absent measurement — that substitution is what put a KRW ceiling opposite an
// invented USD zero.
func axisTodayText(day todayMarketRow) string {
	switch {
	case day.NoTrades:
		return "거래 없음"
	case day.PnL.Known():
		return day.PnL.Value()
	default:
		return "미측정"
	}
}

// consumedPercent renders how much of the ceiling today's loss has used. A
// profit consumes none of it, which is what a loss ceiling means.
func consumedPercent(pnl string, limit float64) string {
	amount, ok := new(big.Rat).SetString(strings.ReplaceAll(strings.TrimSpace(pnl), ",", ""))
	if !ok || limit <= 0 {
		return "비율 미산출"
	}
	if amount.Sign() >= 0 {
		return "소진 0%"
	}
	ratio := new(big.Rat).Quo(new(big.Rat).Neg(amount), new(big.Rat).SetFloat64(limit))
	ratio.Mul(ratio, big.NewRat(100, 1))
	f, _ := ratio.Float64()
	return "소진 " + strconv.FormatFloat(f, 'f', 1, 64) + "%"
}

func limitText(v float64) string {
	if v <= 0 {
		return "미설정"
	}
	return decimalText(v)
}

func currencyText(c string) string {
	if strings.TrimSpace(c) == "" {
		return "(통화 미지정)"
	}
	return c
}

// ratText renders an exact ledger sum without turning it into a float first.
func ratText(r *big.Rat) string {
	if r.IsInt() {
		return groupDecimalText(r.Num().String())
	}
	text := strings.TrimRight(r.FloatString(4), "0")
	text = strings.TrimSuffix(text, ".")
	return groupDecimalText(text)
}

// --- the page -----------------------------------------------------------------------

type overviewPage struct {
	Nav  string
	Snap overviewView
}

// Refresh and RefreshSeconds: the screen reloads itself at the holdings cache
// TTL, derived from the constant rather than written again (positionsPage's
// precedent). Here every reload costs zero broker calls, so the TTL is a
// freshness bound rather than a budget one — but a second period to keep in step
// with the first is worth avoiding either way.
func (overviewPage) Refresh() bool       { return true }
func (overviewPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }

// registerOverview puts the overview screen on the mux.
//
// It is registered from this file rather than from console.go's routes() so that
// the route table's static guards have to be able to see a route declared
// anywhere in the package. Until this change they read one hardcoded file, and a
// route registered here would have passed every one of them in silence.
func (c *Console) registerOverview(mux *http.ServeMux) {
	mux.HandleFunc("/dashboard", c.session0(c.handleOverview))
}

// handleOverview renders the screen. GET, inside the session gate and outside
// the CSRF gate — there is nothing here to submit.
func (c *Console) handleOverview(w http.ResponseWriter, r *http.Request) {
	c.render(w, "overview", overviewPage{Nav: "overview", Snap: c.overview(r.Context())})
}

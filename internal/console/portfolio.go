package console

// portfolio.go assembles the two dashboard screens' data (change
// add-operator-dashboard, tasks 2.1 and 2.2).
//
// # Two sources, and neither may take the other down
//
// The positions screen is a join between what the broker says the account holds
// and what the engine's journal says it is managing. Both halves are read fresh
// per request and both are allowed to be absent:
//
//	no broker wiring   the journal half renders on its own.
//	no journal file    the broker half renders on its own, and the screen says
//	                   the engine has not run rather than showing an empty table.
//	schema mismatch    named in whichever direction it is, because "update the
//	                   console" and "start the engine" are different instructions
//	                   and neither of them is an empty screen.
//
// A join failure is not a screen failure (spec: 조인 실패가 화면 실패가 되어서는
// 안 된다). Every row therefore carries what it has and says what it does not.
//
// # Eligibility
//
// A position is an exit-policy target when a record justifies its baseline, and
// there are two such records: the entry decision the engine opened it under, or
// the adoption record the engine took it into management with
// (adopt-external-positions). The screen asks journal.Position.ExitEligible,
// which is position.ExitEligible, and never spells the columns itself — a second
// copy of "which columns count" is how a screen comes to report a protected
// position as bare.
//
// The two are distinguished in the *display* rather than in the verdict, because
// they answer different operator questions. "Is this protected" is the label;
// "why is it protected" is 자격 근거, and an operator who finds the engine
// managing a holding they bought by hand needs to be able to see that it was
// adopted rather than entered.
//
// A holding with no journal position at all, and a journal position with neither
// record, are both 관리 외(미편입) — one label, as the spec fixes it — and each
// carries its own reason. Folding an external holding *into* management is the
// engine's reconciliation loop's job and this screen deliberately cannot do it:
// every route is a GET.
//
// There is a third answer and it is not that label: when the journal could not be
// read, a holding is 관리 여부 불명. "Unmanaged" is a finding, and a console that
// has not opened the ledger has not found it — reporting one as the other is how
// an operator would come to believe a protected position was bare.
//
// # Nothing is recomputed
//
// The history screen prints columns. trade_outcomes froze realised P&L, realised
// R, the initial quantity and the exit stage at close; positions supplies the
// symbol and exit_states the t0 entry price. There is no exit price in the schema
// and so there is none on the screen: deriving one from the fills would be
// exactly the recomputation the freeze exists to prevent, and it would disagree
// with the frozen R beside it.

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// exitEventWindow bounds the judgement stream on the history screen. The record
// is append-only and a long-running engine writes one row per observation per
// position, so a page that rendered all of them would eventually render a
// megabyte of table nobody scrolls to the bottom of.
const exitEventWindow = 200

// --- the journal half ------------------------------------------------------------

// journalView is what the journal side of a screen managed to read, or why it
// did not.
//
// The states are mutually exclusive and every one of them has its own sentence
// on the page. "Empty" is deliberately not among them: an empty account is
// State ok with no rows, which reads very differently from a journal nobody
// could open.
type journalView struct {
	Path string
	// State is one of the journalState constants below.
	State journalState
	// Detail is the underlying error, for the operator who wants it.
	Detail string
	// Version is the schema version found on disk, 0 when unknown.
	Version int
	// Build is the schema version this binary understands.
	Build int
}

type journalState string

const (
	journalUnwired journalState = "unwired"
	journalMissing journalState = "missing"
	journalTooNew  journalState = "toonew"
	journalTooOld  journalState = "tooold"
	journalFailed  journalState = "failed"
	journalOK      journalState = "ok"
)

// The template asks these rather than comparing strings, so a renamed state is a
// compile error instead of a branch that silently stops matching.
func (v journalView) Unwired() bool { return v.State == journalUnwired }
func (v journalView) Missing() bool { return v.State == journalMissing }
func (v journalView) TooNew() bool  { return v.State == journalTooNew }
func (v journalView) TooOld() bool  { return v.State == journalTooOld }
func (v journalView) Failed() bool  { return v.State == journalFailed }
func (v journalView) OK() bool      { return v.State == journalOK }

// Readable reports that the journal answered. It is what the page uses to decide
// whether "no rows" means "nothing to manage" or "we could not look".
func (v journalView) Readable() bool { return v.State == journalOK }

// openJournal opens the engine's journal read-only for one request.
//
// Per request, not once at startup: the engine is frequently started after the
// console, and a handle taken at boot would keep reporting "no journal" for the
// rest of the day. Opening a SQLite file is cheap and the alternative is a
// dashboard that has to be restarted to become true.
//
// The returned handle is nil whenever the view is not OK, and the caller closes
// it otherwise.
func (c *Console) openJournal(ctx context.Context) (*journal.ReadOnly, journalView) {
	v := journalView{Path: strings.TrimSpace(c.opts.JournalPath), Build: journal.SchemaVersion}
	if v.Path == "" {
		v.State = journalUnwired
		return nil, v
	}

	ro, err := journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: v.Path})
	switch {
	case err == nil:
		v.State, v.Version = journalOK, ro.SchemaVersion()
		return ro, v
	case errors.Is(err, journal.ErrJournalMissing):
		v.State = journalMissing
	case errors.Is(err, journal.ErrSchemaTooNew):
		v.State, v.Detail = journalTooNew, err.Error()
	case errors.Is(err, journal.ErrSchemaTooOld):
		v.State, v.Detail = journalTooOld, err.Error()
	default:
		v.State, v.Detail = journalFailed, err.Error()
	}
	return nil, v
}

// --- the positions screen ---------------------------------------------------------

// positionsView is the whole positions screen.
type positionsView struct {
	Journal  journalView
	Holdings holdingsSnapshot
	Rows     []positionRow
	// Accounts are the masked account references the journal rows came from.
	// Rendered only when there is more than one, because a single-account
	// journal needs no column repeating the same value.
	Accounts []string
}

// Multi reports that the journal holds more than one account.
func (v positionsView) Multi() bool { return len(v.Accounts) > 1 }

// NoManaged reports a journal that was read and holds no live position at all.
//
// It is a separate sentence from "엔진 미가동" because it is a different fact: the
// ledger answered, and the answer is that the engine is managing nothing. The
// spec pairs the two ("엔진 미가동/관리 포지션 없음") and an operator needs to know
// which of them they are looking at.
func (v positionsView) NoManaged() bool {
	if !v.Journal.Readable() {
		return false
	}
	for _, r := range v.Rows {
		if r.InJournal {
			return false
		}
	}
	return true
}

// positionRow is one symbol, from whichever side knows about it.
//
// The two halves are separate on purpose. A symbol the broker reports and the
// journal does not is an unmanaged holding; a symbol the journal holds and the
// broker does not is a projection that disagrees with the account, which is
// worth seeing rather than hiding.
type positionRow struct {
	Symbol string
	Market string
	Name   string

	// The broker half.
	InBroker      bool
	Quantity      float64
	AvgPrice      float64
	LastPrice     float64
	MarketValue   float64
	UnrealizedPnL float64
	ProfitRate    float64

	// JournalReadable reports that the journal answered at all. It is what keeps
	// "관리 외(미편입)" an observation rather than a guess: when the journal could
	// not be opened, a holding is not known to be unmanaged — it is unknown, and
	// the row says so.
	JournalReadable bool

	// The journal half.
	InJournal       bool
	AccountRef      string
	PositionID      string
	State           string
	JournalQuantity string
	JournalAvgPrice string
	Eligible        bool
	// Adopted reports which record makes the row eligible. It is display only:
	// the verdict is Eligible, and this says why.
	Adopted bool
	HasExit bool
	Exit    journal.ExitState
}

// Basis names the record that justifies the exit baseline, for the operator who
// wants to know why the engine is managing a holding they bought themselves.
func (r positionRow) Basis() string {
	switch {
	case !r.Managed():
		return "—"
	case r.Adopted:
		return "편입 기록"
	default:
		return "진입 결정"
	}
}

// Managed reports that the engine's exit policy owns this position.
func (r positionRow) Managed() bool { return r.InJournal && r.Eligible }

// Unknown reports a holding whose management could not be determined, because
// the journal did not answer. It is deliberately not folded into "unmanaged": a
// console that could not read the ledger has not established that a position is
// unprotected, and saying so would be the screen asserting something it did not
// observe.
func (r positionRow) Unknown() bool { return !r.JournalReadable && !r.InJournal }

// Label is the verdict in the status column. The unmanaged label is fixed by the
// spec to this exact string; there is not a second spelling of it anywhere in
// this package.
func (r positionRow) Label() string {
	switch {
	case r.Unknown():
		return "관리 여부 불명"
	case !r.Managed():
		return "관리 외(미편입)"
	case r.HasExit && r.Exit.Completed:
		return "관리 종료"
	case r.HasExit:
		return "엔진 관리"
	default:
		return "엔진 관리(대기)"
	}
}

// Reason explains an absent exit line. It is empty when there is one to show.
func (r positionRow) Reason() string {
	switch {
	case r.Unknown():
		return "엔진 원장을 읽지 못했으므로 이 보유가 엔진 관리 대상인지 알 수 없다 — 위의 원장 안내를 보라. " +
			"관리 중이 아니라고 단정하지 않는다."
	case !r.InJournal:
		return "엔진 원장에 이 종목의 포지션이 없다 — 엔진이 진입한 포지션이 아니므로 손절·익절 라인도 없다. " +
			"수동 보유를 엔진 관리로 편입하는 것은 이 화면의 기능이 아니다."
	case !r.Eligible:
		return "진입 결정(entry decision)도 편입 기록(adoption)도 없는 포지션이다. exit 정책은 그중 하나의 " +
			"손절가를 기준선으로 삼는데 둘 다 없으므로 대상이 아니다."
	case !r.HasExit:
		return "자격 근거(" + r.Basis() + ")는 있으나 exit 상태가 아직 열리지 않았다 — 반영 직후이거나, 엔진이 " +
			"아직 관측을 시작하지 않았다."
	default:
		return ""
	}
}

// BrokerMissing reports a projection row the account does not confirm.
func (r positionRow) BrokerMissing() bool { return r.InJournal && !r.InBroker }

// Rung renders the ladder rung, or an em dash under a ratchet policy.
func (r positionRow) Rung() string {
	if !r.HasExit || r.Exit.ActiveRung < 0 {
		return "—"
	}
	return strconv.Itoa(r.Exit.ActiveRung)
}

// Pending renders the outstanding exit proposal.
func (r positionRow) Pending() string {
	if !r.HasExit || !r.Exit.Pending() {
		return "—"
	}
	out := r.Exit.PendingAction
	if r.Exit.PendingLevel != "" {
		out += " / " + r.Exit.PendingLevel
	}
	if r.Exit.PendingIntentID != "" {
		out += " (intent " + r.Exit.PendingIntentID + ")"
	}
	return out
}

// Taken renders the cumulative partial take-profit fraction.
func (r positionRow) Taken() string {
	if !r.HasExit || strings.TrimSpace(r.Exit.TakenRatioTotal) == "" {
		return "—"
	}
	return r.Exit.TakenRatioTotal
}

// Qty, Avg, Last, Value, PnL and Rate render the broker half.
func (r positionRow) Qty() string   { return brokerNumber(r.InBroker, r.Quantity) }
func (r positionRow) Avg() string   { return brokerNumber(r.InBroker, r.AvgPrice) }
func (r positionRow) Last() string  { return brokerNumber(r.InBroker, r.LastPrice) }
func (r positionRow) Value() string { return brokerNumber(r.InBroker, r.MarketValue) }
func (r positionRow) PnL() string   { return brokerNumber(r.InBroker, r.UnrealizedPnL) }

// Rate renders the unrealised return as a percentage.
func (r positionRow) Rate() string {
	if !r.InBroker {
		return "—"
	}
	return strconv.FormatFloat(r.ProfitRate*100, 'f', 2, 64) + "%"
}

// Gain reports a non-negative unrealised P&L, for the colour class.
func (r positionRow) Gain() bool { return r.UnrealizedPnL >= 0 }

func brokerNumber(present bool, v float64) string {
	if !present {
		return "—"
	}
	return decimalText(v)
}

// positions builds the positions screen.
func (c *Console) positions(ctx context.Context) positionsView {
	now := c.now()
	hold, why := c.verifyHold(now)

	v := positionsView{Holdings: c.holdings.get(ctx, now, hold, why)}

	ro, jv := c.openJournal(ctx)
	v.Journal = jv
	var journalRows []journal.PositionExit
	if ro != nil {
		defer ro.Close()
		refs, err := ro.AccountRefs(ctx)
		if err != nil {
			v.Journal.State, v.Journal.Detail = journalFailed, err.Error()
		}
		for _, ref := range refs {
			rows, err := ro.LivePositionExits(ctx, ref)
			if err != nil {
				v.Journal.State, v.Journal.Detail = journalFailed, err.Error()
				break
			}
			if len(rows) > 0 {
				v.Accounts = append(v.Accounts, attest.Mask(ref))
			}
			journalRows = append(journalRows, rows...)
		}
	}

	v.Rows = joinPositions(v.Holdings.Rows, journalRows, v.Journal.Readable())
	return v
}

// joinPositions merges the two sides on (market, symbol).
//
// The key carries the market because "005930" on KR and a US listing of the same
// ticker are different instruments, and the projection normalises both fields
// (position_projection.go) while the broker reports "KR"/"US" — so both sides are
// folded to one spelling here rather than compared as they arrive.
//
// A second journal position claiming a key the first already took becomes its own
// row instead of overwriting: two live instances of one symbol is a state worth
// showing, not one worth silently resolving.
func joinPositions(broker []domain.Position, managed []journal.PositionExit,
	journalReadable bool) []positionRow {
	index := map[string]int{}
	rows := make([]positionRow, 0, len(broker)+len(managed))

	for _, h := range broker {
		key := symbolKey(h.MarketType, h.Symbol)
		index[key] = len(rows)
		rows = append(rows, positionRow{
			Symbol:          strings.ToUpper(strings.TrimSpace(h.Symbol)),
			Market:          strings.ToLower(strings.TrimSpace(h.MarketType)),
			Name:            h.Name,
			JournalReadable: journalReadable,
			InBroker:        true,
			Quantity:        h.Quantity,
			AvgPrice:        h.AveragePrice,
			LastPrice:       h.CurrentPrice,
			MarketValue:     h.MarketValue,
			UnrealizedPnL:   h.UnrealizedPnL,
			ProfitRate:      h.ProfitRate,
		})
	}

	for _, m := range managed {
		key := symbolKey(m.Position.Market, m.Position.Symbol)
		at, ok := index[key]
		if !ok {
			at = len(rows)
			index[key] = at
			rows = append(rows, positionRow{
				Symbol:          strings.ToUpper(strings.TrimSpace(m.Position.Symbol)),
				Market:          strings.ToLower(strings.TrimSpace(m.Position.Market)),
				JournalReadable: journalReadable,
			})
		} else if rows[at].InJournal {
			// Already claimed by another instance: give this one its own row.
			at = len(rows)
			rows = append(rows, positionRow{
				Symbol:          strings.ToUpper(strings.TrimSpace(m.Position.Symbol)),
				Market:          strings.ToLower(strings.TrimSpace(m.Position.Market)),
				JournalReadable: journalReadable,
			})
		}
		row := &rows[at]
		row.InJournal = true
		row.AccountRef = attest.Mask(m.Position.AccountRef)
		row.PositionID = m.Position.ID
		row.State = m.Position.State
		row.JournalQuantity = m.Position.Quantity
		row.JournalAvgPrice = m.Position.AvgPrice
		row.Eligible = m.Position.ExitEligible()
		row.Adopted = m.Position.Adopted()
		row.HasExit = m.HasExit
		row.Exit = m.Exit
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Market != rows[j].Market {
			return rows[i].Market < rows[j].Market
		}
		return rows[i].Symbol < rows[j].Symbol
	})
	return rows
}

func symbolKey(market, symbol string) string {
	return strings.ToLower(strings.TrimSpace(market)) + "|" + strings.ToUpper(strings.TrimSpace(symbol))
}

// --- the history screen ------------------------------------------------------------

// historyView is the whole trade-history screen.
type historyView struct {
	Journal  journalView
	Trips    []tripRow
	Events   []eventRow
	Accounts []string
	// Truncated reports that the judgement stream was windowed.
	Truncated bool
}

// Multi reports that more than one account contributed rows.
func (v historyView) Multi() bool { return len(v.Accounts) > 1 }

// tripRow is one frozen round trip, rendered.
//
// Every field is a column that exists. There is no exit price here because there
// is no exit price in the schema, and the header on the page says so rather than
// leaving the operator to wonder whether it was forgotten.
type tripRow struct {
	AccountRef string
	Symbol     string
	Market     string
	// EntryPrice is exit_states.entry_price, frozen at t0. Empty renders "—".
	EntryPrice string
	PnL        string
	R          string
	Quantity   string
	// Held is the hold time, already rendered: "—" when held_seconds was NULL.
	Held     string
	Stage    string
	ClosedAt string
	// Win is the sign of the realised P&L, for the colour class.
	Win bool
}

// Entry renders the frozen entry price, or an em dash when no exit state ever
// froze one.
func (r tripRow) Entry() string {
	if strings.TrimSpace(r.EntryPrice) == "" {
		return "—"
	}
	return r.EntryPrice
}

// eventRow is one exit judgement, rendered.
type eventRow struct {
	AccountRef string
	Symbol     string
	At         string
	Observed   string
	HighWater  string
	Baseline   string
	Level      string
	Action     string
	IntentID   string
}

// Proposal renders the intent this judgement proposed, if any.
func (r eventRow) Proposal() string {
	if strings.TrimSpace(r.Action) == "" {
		return "—"
	}
	if strings.TrimSpace(r.IntentID) == "" {
		return r.Action
	}
	return r.Action + " (intent " + r.IntentID + ")"
}

// history builds the trade-history screen.
func (c *Console) history(ctx context.Context) historyView {
	var v historyView

	ro, jv := c.openJournal(ctx)
	v.Journal = jv
	if ro == nil {
		return v
	}
	defer ro.Close()

	refs, err := ro.AccountRefs(ctx)
	if err != nil {
		v.Journal.State, v.Journal.Detail = journalFailed, err.Error()
		return v
	}

	for _, ref := range refs {
		masked := attest.Mask(ref)

		trips, err := ro.AccountTradeTrips(ctx, ref)
		if err != nil {
			v.Journal.State, v.Journal.Detail = journalFailed, err.Error()
			return v
		}
		for _, t := range trips {
			v.Trips = append(v.Trips, tripRow{
				AccountRef: masked,
				Symbol:     t.Symbol,
				Market:     t.Market,
				EntryPrice: t.EntryPrice,
				PnL:        t.Outcome.RealizedPnLAfterCosts,
				R:          t.Outcome.RealizedR,
				Quantity:   t.Outcome.InitialQuantity,
				Held:       heldText(t.Outcome.HeldSeconds, t.HeldSecondsKnown),
				Stage:      stageText(t.Outcome.ExitRatchetLevel, t.Outcome.ExitRung),
				ClosedAt:   t.Outcome.ClosedAt,
				Win:        !strings.HasPrefix(strings.TrimSpace(t.Outcome.RealizedPnLAfterCosts), "-"),
			})
		}

		events, err := ro.AccountExitEvents(ctx, ref, exitEventWindow)
		if err != nil {
			v.Journal.State, v.Journal.Detail = journalFailed, err.Error()
			return v
		}
		if len(events) == exitEventWindow {
			v.Truncated = true
		}
		for _, e := range events {
			v.Events = append(v.Events, eventRow{
				AccountRef: masked,
				Symbol:     e.Symbol,
				At:         e.Event.CreatedAt,
				Observed:   dashIfEmpty(e.Event.ObservedPrice),
				HighWater:  dashIfEmpty(e.Event.HighWater),
				Baseline:   dashIfEmpty(e.Event.BaselineAfter),
				Level:      dashIfEmpty(e.Event.LevelAfter),
				Action:     e.Event.Action,
				IntentID:   e.Event.ProposedIntentID,
			})
		}
		if len(trips) > 0 || len(events) > 0 {
			v.Accounts = append(v.Accounts, masked)
		}
	}

	sort.SliceStable(v.Trips, func(i, j int) bool { return v.Trips[i].ClosedAt < v.Trips[j].ClosedAt })
	sort.SliceStable(v.Events, func(i, j int) bool { return v.Events[i].At < v.Events[j].At })
	return v
}

// --- rendering helpers --------------------------------------------------------------

// heldText renders a hold time. A NULL held_seconds is "—" and never "0초": the
// column is nullable in the schema, and printing a zero would state a fact the
// journal does not hold (spec: nullable 필드의 NULL은 "—"로 렌더한다).
func heldText(seconds int64, known bool) string {
	if !known {
		return "—"
	}
	if seconds < 0 {
		seconds = 0
	}
	d := time.Duration(seconds) * time.Second
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%d초", seconds)
	case d < time.Hour:
		return fmt.Sprintf("%d분 %d초", seconds/60, seconds%60)
	case d < 24*time.Hour:
		return fmt.Sprintf("%d시간 %d분", seconds/3600, (seconds%3600)/60)
	default:
		return fmt.Sprintf("%d일 %d시간", seconds/86400, (seconds%86400)/3600)
	}
}

// stageText renders the exit stage the round trip reached. The two are exclusive
// by policy — a ratchet has a level, a ladder has a rung — and the schema keeps
// them in separate nullable columns for exactly that reason.
func stageText(ratchetLevel string, rung int) string {
	switch {
	case strings.TrimSpace(ratchetLevel) != "":
		return ratchetLevel
	case rung >= 0:
		return "rung " + strconv.Itoa(rung)
	default:
		return "—"
	}
}

func dashIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

// decimalText renders a broker number without exponent notation and without
// trailing zeros, then groups the integer part in threes.
//
// The broker's own decimals arrive as float64 through internal/domain, so this is
// presentation only — nothing here is arithmetic on a price, and the journal's
// side of the screen stays in the decimal strings the ledger stores.
func decimalText(v float64) string {
	text := strconv.FormatFloat(v, 'f', -1, 64)
	sign := ""
	if strings.HasPrefix(text, "-") {
		sign, text = "-", text[1:]
	}
	whole, frac, _ := strings.Cut(text, ".")

	var b strings.Builder
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	out := sign + b.String()
	if frac != "" {
		out += "." + frac
	}
	return out
}

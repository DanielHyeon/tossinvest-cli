package console

import (
	"strconv"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// The two screens the shell itself has to link to.
//
// They are constants because three different files used to spell them as string
// literals — a route registration, a restart's return path and a nav href — and
// a path that is written in three places is a path that gets moved in two.
const (
	pathOverview      = "/dashboard"
	pathVerifyConsole = "/verify-console"
)

// chrome is the shell every console screen shares: which nav item is current,
// whether the browser reloads itself, and the four facts the status strip says.
//
// Before this there was no shared base at all. Each page struct declared Nav and
// Refresh for itself and the head template found them by duck typing. That is
// enough while the shell is two fields; it stops being enough the moment the
// shell has to say *the same* four facts on every screen, because "the same" is
// not a property that duck-typed fields can have — eight templates each grew
// their own sentence for the same reading instead.
//
// It is embedded rather than wrapped. A {Page, Chrome} wrapper would have turned
// every {{.X}} in eight templates into {{.Page.X}}, spreading the diff across
// hundreds of lines that have nothing to do with what is being changed.
type chrome struct {
	Nav     string
	Refresh bool
	Status  statusStrip

	// Explain is the URL-driven disclosure state, and only the screens that
	// reload themselves fill it in (change a055 §6). A native <details> on those
	// screens closes on the next tick; everywhere else native is right and this
	// stays zero.
	Explain explainState
}

// RefreshSeconds is the default reload period: none.
//
// A screen that reloads defines its own method and that method wins, because a
// method on the outer type shadows one promoted from an embedded field. The
// shadowing is also the hazard this file's tests exist for: delete a screen's
// RefreshSeconds and this zero takes over in silence — it compiles, it renders,
// and the only thing that changed is that the screen stopped updating. The same
// goes for Refresh, which several screens supply as a method returning true.
func (chrome) RefreshSeconds() int { return 0 }

// --- the status strip ---------------------------------------------------------

// The three things the data cell can mean. They are different facts and one
// word for all three would be a lie on two of them.
const (
	// dataCache is a broker reading with its own recorded instant.
	dataCache = "cache"
	// dataOnRequest is a screen that read files synchronously while rendering,
	// so the render instant *is* the reading instant.
	dataOnRequest = "on-request"
	// dataUnavailable is a read that failed.
	dataUnavailable = "unavailable"
)

// statusStrip is the row every screen renders in the same place.
//
// It holds rendered text rather than the readings themselves: the strip's whole
// point is that the wording does not vary by screen, and a struct of raw values
// would let each template word them again.
//
// The reload cell is deliberately absent. It is read straight off .Refresh and
// .RefreshSeconds in the template, so the strip cannot say "2초마다" while the
// meta tag says nothing — the two would be separate copies of one fact.
type statusStrip struct {
	// EngineState is "unwired", "stopped" or "running". Unwired is not stopped:
	// one is a build without the seam, the other is a machine that is not running
	// the engine, and telling an operator the second when the first is true sends
	// them looking for a process that was never supposed to exist.
	EngineState string
	EngineText  string
	// EngineNote carries the provenance mismatch: a running engine started from a
	// different executable than the one installed now.
	EngineNote string

	// Session is the market-hours advisory, carried whole and read by nothing in
	// Go.
	//
	// It would be shorter to store an "open"/"closed" string, and that is exactly
	// what this must not do: computing it here would be console code branching on
	// .Outside, which static_test.go forbids for a reason worth keeping — the
	// window in which the broker actually accepts an order is unmeasured, and
	// every Go read of this advisory is one more place that rule has to go on
	// being true. The template renders it. Nothing consults it.
	Session verifylive.SessionAdvisory

	// DataMode is one of the three constants above; DataText is the instant and
	// its age; DataTone is "", "ok", "warn" or "stale".
	DataMode string
	DataText string
	DataTone string
	// DataNote is the failure reason, or the known hold reason. When a hold is
	// known DataTone is empty: see freshness.tone.
	DataNote string

	// PendingApproval reports a verification run parked on its approval, and
	// PendingHref links to the screen that can approve it. The strip carries no
	// form — approving happens there, not here.
	PendingApproval bool
	PendingHref     string
}

// freshness is what a screen knows about the age of what it is showing.
//
// The screen owns this because only the screen knows which of the three modes it
// is in. The strip owns the wording.
type freshness struct {
	Mode string
	// At is the reading's own instant, already rendered, and Age is how old it is
	// now. Both are meaningful only in dataCache mode.
	At  string
	Age time.Duration
	// TTL is this screen's own cache period, which is where the tone thresholds
	// come from. No new constant: a threshold that is not the screen's own period
	// would warn about an age the screen itself considers current.
	TTL time.Duration
	// Reason explains a dataUnavailable reading.
	Reason string
	// LastAt is the last successful read, rendered, and empty when there is no
	// such record. The strip never invents one.
	LastAt string

	// Hold names a known and correct reason this reading is not being refreshed —
	// the rate-budget hold a running verification puts on the broker caches, for
	// instance. A hold suppresses the tone: a system doing what the spec tells it
	// to do is not a warning, and a warning that is always on is one nobody reads.
	Hold string
}

// tone grades the age against the screen's own TTL.
//
// Empty means no tone at all, which is the answer for every mode except a cache
// reading with no known hold — a render-time read cannot be stale, and a hold is
// an explanation rather than a fault.
func (f freshness) tone() string {
	switch {
	case f.Mode == dataUnavailable:
		return "stale"
	case f.Mode != dataCache, f.Hold != "", f.TTL <= 0:
		return ""
	case f.Age < f.TTL:
		return "ok"
	case f.Age < 2*f.TTL:
		return "warn"
	default:
		return "stale"
	}
}

// onRequest is the freshness of a screen that reads its files while rendering.
func onRequest() freshness { return freshness{Mode: dataOnRequest} }

// --- building it ----------------------------------------------------------------

// chromeFor assembles the shell for one render.
//
// It reads the engine marker and stats the installed binary; it does not call
// snapshot(). snapshot() also parses the soak record, the attestation and the
// verification record, and the strip uses none of them — putting that on a
// screen that reloads every two seconds would be paying for four readings to
// display two.
func (c *Console) chromeFor(nav, market string, f freshness) chrome {
	now := c.now()

	var installed binstamp.Stamp
	if c.opts.Binary != nil {
		installed, _ = c.opts.Binary()
	}
	engine := c.readEngine(now, installed)

	s := statusStrip{Session: verifylive.SessionAdvisoryFor(market, now)}

	switch {
	case !engine.Wired:
		s.EngineState = "unwired"
		s.EngineText = "미배선"
		s.EngineNote = "이 빌드에 엔진 마커 경로가 없다 — 엔진이 꺼져 있다는 뜻이 아니다"
	case engine.Running:
		s.EngineState = "running"
		s.EngineText = "실행 중"
		if engine.PID > 0 {
			s.EngineText += " · PID " + strconv.Itoa(engine.PID)
		}
		if engine.Stale {
			s.EngineNote = "실행 중인 엔진은 설치된 바이너리와 다른 실행 파일에서 기동됐다"
		}
	default:
		s.EngineState = "stopped"
		s.EngineText = "정지"
	}

	s.DataMode = f.Mode
	s.DataTone = f.tone()
	switch f.Mode {
	case dataCache:
		s.DataText = f.At + " (" + strconv.Itoa(int(f.Age.Round(time.Second)/time.Second)) + "초 전)"
	case dataUnavailable:
		s.DataText = "읽지 못함"
		s.DataNote = f.Reason
		if f.LastAt != "" {
			s.DataNote += " · 마지막 성공 " + f.LastAt
		}
	default:
		// The render instant is the reading instant here, so this is not a
		// manufactured timestamp — it is the one fact the screen actually has.
		s.DataText = renderInstant(now)
		s.DataNote = "요청 시 읽음"
	}
	if f.Hold != "" {
		s.DataNote = f.Hold
	}

	if run := c.currentRun(); run != nil {
		if v := run.snapshot(); v.Awaiting {
			s.PendingApproval = true
			s.PendingHref = pathVerifyConsole
		}
	}

	return chrome{Nav: nav, Status: s}
}

// chromeOnRequest is chromeFor for the screens that read at render time, which
// is most of them.
func (c *Console) chromeOnRequest(nav string) chrome {
	return c.chromeFor(nav, verifylive.MarketKR, onRequest())
}

// --- what the cache-backed screens know ------------------------------------------
//
// These live here rather than beside their types because they are one contract
// in three spellings, and the thing worth checking about them is that they agree.

// freshness reads the account cache's five states into the strip's three.
//
// The order matters. A reading that exists but whose last refresh failed is
// *unavailable with a last-success time*, not a cache reading — the numbers on
// screen are real but they are not an answer to "how is the account now", and
// the failure is the fact the operator needs first.
func (s holdingsSnapshot) freshness() freshness {
	switch {
	case !s.Wired:
		return freshness{Mode: dataUnavailable, Reason: "브로커 배선 없음"}
	case s.Error != "":
		f := freshness{Mode: dataUnavailable, Reason: s.Error}
		if s.Present {
			f.LastAt = s.TakenAt()
		}
		return f
	case !s.Present:
		return freshness{Mode: dataUnavailable, Reason: "아직 읽은 값이 없다"}
	}
	f := freshness{Mode: dataCache, TTL: holdingsTTL, At: s.TakenAt(), Age: s.Age}
	if s.Held {
		f.Hold = "갱신 보류 — " + s.HeldReason
	}
	return f
}

// freshness reads the orders cache the same way, against its own TTL.
func (s ordersSnapshot) freshness() freshness {
	switch {
	case !s.Wired:
		return freshness{Mode: dataUnavailable, Reason: "주문 조회 배선 없음"}
	case s.Error != "":
		f := freshness{Mode: dataUnavailable, Reason: s.Error}
		if s.Present {
			f.LastAt = s.TakenAt()
		}
		return f
	case !s.Present:
		return freshness{Mode: dataUnavailable, Reason: "아직 읽은 값이 없다"}
	}
	f := freshness{Mode: dataCache, TTL: ordersTTL, At: s.TakenAt(), Age: s.Age}
	if s.Held {
		f.Hold = "갱신 보류 — " + s.HeldReason
	}
	return f
}

// freshness reads the discovery store's own instant, and deliberately carries no
// tone.
//
// TTL is left at zero, which makes tone() return nothing. That is the honest
// answer here: this reading advances on `tossctl candidate watch`'s tick and on
// nothing else. The console neither causes that tick nor observes whether one is
// due — no scan overnight is the design working, not the reading rotting — so
// grading the age against the tick interval would put a warning on the screen
// every evening for a system behaving exactly as specified. The reason is
// printed instead, which is what the operator can actually act on.
//
// The tempting version of this reads the market-hours advisory and suppresses
// the tone while the market is closed. It is not written that way on purpose:
// every Go read of that advisory is one more place the "advisory only, never a
// branch" rule has to keep being true (static_test.go), and buying a colour with
// that is a bad trade.
func (v signalsView) freshness() freshness {
	if !v.Read.Known() {
		return freshness{Mode: dataUnavailable, Reason: v.Read.Code()}
	}
	return freshness{
		Mode: dataCache,
		At:   v.TakenAt,
		Age:  time.Duration(v.AgeSeconds) * time.Second,
		Hold: "발굴 스캔 tick(" + candidate.DefaultWatchInterval.String() + ")에만 갱신된다 — " +
			"이 콘솔은 스캔을 일으키지 않는다",
	}
}

package engine

// in_doubt_settle.go 는 a193 의 세션 중 IN_DOUBT 해소 실행자임(openspec order-execution 「IN_DOUBT 해소」 · 「세션 중 해소」).
//
// # 왜 있나
//
// 세션 중에 결과 불명(IN_DOUBT)이 된 발주는 다음 기동의 재시작 복구만 해소했음 — 그동안 그 종목의 모든 새 mutation 이 거절되고
// (손절 포함) 알림도 없었음(2026-10-29 사고: 36.8시간 침묵, 운영 이력 6회 전부 11.9~46.3시간 뒤 park). 이 실행자는 엔진이 도는 동안
// 같은 해소 함수(execgw.Resolver.Resolve — 조회 대조)를 부르고, 그 사실을 critical 로 알림.
//
// # 지키는 것
//
//   - 보조 실행자임 — 손절 판정 goroutine 밖에서 돌고, 정지·패닉이 다른 루프를 내리지 않음(auxiliary.go). 기동 복구 뒤에 시작됨.
//   - 대상은 IN_DOUBT 이면서 발주(PLACE)인 attempt 뿐임. RECORDED · DISPATCH_STARTED 는 진행 중인 전송이라 손대지 않음(재시작 규칙 금지).
//   - 재생(같은 본문 재전송)은 부르지 않음 — 조회만 함. 주문을 내는 경로가 이 파일에 없음.
//   - 엔진 정지로 해소가 취소되면 결과를 버림 — 원장 쓰기도 같은 취소된 ctx 를 받아 실패하므로 attempt 는 IN_DOUBT 로 남음(design D3).
//   - 알림은 전부 알림기의 critical 기록 단일 입구로 창 0 한 번(에피소드 = attempt 또는 연속 id). 취소된 ctx 로는 기록하지 않음
//     (기록 실패가 진입 게이트를 잠그므로).
//
// 브로커 호출은 해소(Resolve) 안에서만 일어나고, IN_DOUBT 발주가 없으면 0 임.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// inDoubtSettleName 은 보조 실행자 이름임 — 감독 루프 이름에 나타나면 안 됨(그 정지가 엔진을 내림).
const inDoubtSettleName = "in-doubt-settle"

// errInDoubtSettleUnwired 는 원장이나 해소기가 없는 실행자임 — 조용히 쉬지 않고 정지로 드러냄(OnStop 이 critical).
var errInDoubtSettleUnwired = errors.New("engine: the in-doubt settler has no journal or resolver")

// inDoubtLedger 는 실행자가 쓰는 원장 면임. *journal.Journal 이 구현함. 원장에 쓰는 것은 결과 불명 취소의 닫기 하나뿐(a198 D10 —
// 대상 주문의 원장 종결 기록으로, 브로커 호출 없음).
type inDoubtLedger interface {
	PendingAttempts(ctx context.Context) ([]journal.AttemptRecord, error)
	LookupIntent(ctx context.Context, id string) (journal.Intent, error)
	ArmedExitProposals(ctx context.Context, accountRef string) ([]journal.ArmedExitProposal, error)
	CloseInDoubtCancelOnTargetClose(ctx context.Context, attemptID, source string) (journal.CancelTargetClose, error)
}

// inDoubtResolver 는 기존 해소 함수임. *execgw.Resolver 가 구현함.
type inDoubtResolver interface {
	Resolve(ctx context.Context, attemptID string) (execgw.Resolution, error)
}

// inDoubtSettleDeps 는 실행자의 배선임.
type inDoubtSettleDeps struct {
	Ledger   inDoubtLedger
	Resolver inDoubtResolver
	Critical CriticalRecorder
	Log      *obs.Logger
	Clock    clock.Clock
	Account  string
	// ResolveBudget 은 해소 하나의 예산(execgw.ResolveConfig.MaxDuration)임 — 10분 상한은 그 두 배로 유도함(새 설정 없음).
	ResolveBudget time.Duration
	// NewID 는 연속(원장 못 읽음)·실행 에피소드의 id 를 만듦. 기본은 128비트 난수.
	NewID func() string
}

type inDoubtSettle struct {
	deps     inDoubtSettleDeps
	interval time.Duration
	bound    time.Duration
	runID    string

	// recorded 는 기록에 성공한 알림 key 임 — 프로세스 안 중복 적재를 막는 래치(재시작 뒤 같은 key 는 입구가 같은 행으로 받음).
	recorded map[string]bool
	// unsent 는 기록에 실패한 알림임 — attempt 가 그사이 종결돼 스윕 대상에서 빠져도 다음 스윕 머리에서 다시 기록함(구현 리뷰 1회차 P1-1).
	unsent map[string]obs.Event
	// unreadableSince · unreadableID 는 원장 목록 읽기 연속 실패의 시작과 에피소드 id 임. 성공하면 비움.
	unreadableSince time.Time
	unreadableID    string
	// lastFailure 는 attempt 별 마지막 해소 오류 문구임 — 같은 문구가 이어지면 로그를 한 번만 남김.
	lastFailure map[string]string
}

// InDoubtSettler 는 세션 중 IN_DOUBT 해소 보조 실행자를 돌려줌(cmd/tossctl engineRuntime 이 보조 집합에 넣음).
//
// 엔진 조립의 원장 · 해소기 · 알림기를 그대로 씀 — 기동 복구와 같은 Resolver 인스턴스(수신자 필드에 쓰지 않으므로 공유 안전).
// nil 포인터를 인터페이스에 넣지 않음(typed nil 이 「배선됨」으로 보이지 않게).
func (c *Context) InDoubtSettler(clk clock.Clock) AuxiliaryExecutor {
	deps := inDoubtSettleDeps{Clock: clk, ResolveBudget: execgw.DefaultResolveConfig().MaxDuration}
	if c != nil {
		if c.Journal != nil {
			deps.Ledger = c.Journal
		}
		if c.Resolver != nil {
			deps.Resolver = c.Resolver
			if c.Resolver.Config.MaxDuration > 0 {
				deps.ResolveBudget = c.Resolver.Config.MaxDuration
			}
		}
		if c.Notifier != nil {
			deps.Critical = c.Notifier
		}
		deps.Log = c.Log
		deps.Account = c.AccountRef
		if deps.Clock == nil {
			deps.Clock = c.clock()
		}
	}
	return newInDoubtSettle(deps).executor()
}

func newInDoubtSettle(deps inDoubtSettleDeps) *inDoubtSettle {
	if deps.Clock == nil {
		deps.Clock = clock.System()
	}
	if deps.ResolveBudget <= 0 {
		deps.ResolveBudget = execgw.DefaultResolveConfig().MaxDuration
	}
	if deps.NewID == nil {
		deps.NewID = randomExitID
	}
	return &inDoubtSettle{
		deps: deps,
		// 스윕 간격 = 해소 폴링 간격(5초) — 새 설정 없음.
		interval:    execgw.DefaultResolveConfig().PollInterval,
		bound:       2 * deps.ResolveBudget,
		runID:       deps.NewID(),
		recorded:    map[string]bool{},
		unsent:      map[string]obs.Event{},
		lastFailure: map[string]string{},
	}
}

func (s *inDoubtSettle) executor() AuxiliaryExecutor {
	return AuxiliaryExecutor{Name: inDoubtSettleName, Run: s.Run, OnStop: s.onStop, StopEvent: obs.EventOrderInDoubt}
}

// Run 은 기동 복구 뒤 시작되어 취소될 때까지 스윕함. 취소는 ctx.Err() 로 돌려줌(런타임이 정상 정지로 읽음).
func (s *inDoubtSettle) Run(ctx context.Context) error {
	if s.deps.Ledger == nil || s.deps.Resolver == nil {
		return errInDoubtSettleUnwired
	}
	for {
		// 첫 스윕 **전에** 취소 확인 — 이미 끝난 런타임에서 원장을 건드리지 않음(영값 원장으로 조립한 시험 포함).
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.sweep(ctx); err != nil {
			return err
		}
		if err := s.deps.Clock.Sleep(ctx, s.interval); err != nil {
			return err
		}
	}
}

// sweep 은 한 번의 순회임: 원장 읽기 → (1단) 미결 전부의 처음 봄 · 10분 알림 → (2단) 직렬 해소.
// 오류는 ctx 취소일 때만 돌려줌 — 나머지 실패는 다음 스윕이 다시 함.
func (s *inDoubtSettle) sweep(ctx context.Context) error {
	// 먼저 못 보낸 알림 — 원장 읽기 실패와 무관하게, 대상이 종결돼 사라졌어도.
	s.retryUnsent(ctx)
	recs, err := s.deps.Ledger.PendingAttempts(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		s.noteUnreadable(ctx, err)
		return nil
	}
	s.unreadableSince, s.unreadableID = time.Time{}, ""

	targets := make([]journal.AttemptRecord, 0, len(recs))
	var cancels []journal.AttemptRecord
	for _, rec := range recs {
		// 대상 거름: 조회 대조 해소는 결과 불명인 발주만. 결과 불명인 취소는 원장 종결 기록으로만 닫음(a198 D10). 정정은 기동이 해소하고,
		// 진행 중 전송(RECORDED · DISPATCH_STARTED)과 ACKED 는 손대지 않음.
		switch {
		case rec.State == journal.StateInDoubt && rec.Kind == journal.KindPlace:
			targets = append(targets, rec)
		case rec.State == journal.StateInDoubt && rec.Kind == journal.KindCancel:
			cancels = append(cancels, rec)
		}
	}
	// 1단: 알림을 해소보다 먼저 — 한 해소가 분 단위로 걸려도 다른 attempt 의 알림이 그 뒤에 밀리지 않게 함(Q-F 처분).
	// 취소 뒤의 기록은 record 가 막음(기록 실패가 진입 게이트를 잠그므로). 취소에는 알림을 내지 않음(길 치우기의 경보가 맡음).
	for _, rec := range targets {
		s.noteFirstSight(ctx, rec)
		s.noteStale(ctx, rec)
	}
	// 취소 닫기를 발주 해소보다 먼저 — 원장만 읽어 빠르고, 분 단위 해소 뒤로 밀리면 그만큼 손절이 늦어짐(Manager Q5).
	for _, rec := range cancels {
		if err := s.closeCancel(ctx, rec); err != nil {
			return err
		}
	}
	// 2단: 직렬 해소 — 동시 해소 상한 1. 취소면 resolve 가 ctx 오류를 돌려줌.
	for _, rec := range targets {
		if err := s.resolve(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

// resolve 는 attempt 하나를 기존 해소 함수로 해소함. 취소면 결과를 버리고 ctx 오류를 돌려줌.
func (s *inDoubtSettle) resolve(ctx context.Context, rec journal.AttemptRecord) error {
	res, err := s.deps.Resolver.Resolve(ctx, rec.ID)
	if err != nil {
		if ctx.Err() != nil {
			// 메모리 결과(UNRESOLVED 일 수 있음)는 원장에 없음 — 같은 취소된 ctx 의 원장 쓰기가 실패했으므로 버림(design D3).
			return ctx.Err()
		}
		// 원문 오류는 싣지 않음 — 원장 오류가 주문 번호를 품을 수 있음(구현 리뷰 1회차 P1-2). 분류만 남김.
		class := resolveFailureClass(err)
		if s.lastFailure[rec.ID] != class {
			s.lastFailure[rec.ID] = class
			s.deps.Log.Warn(obs.EventOrderInDoubt,
				obs.FieldAttemptID, rec.ID, "error_class", class,
				obs.FieldDetail, "the in-doubt settler could not resolve this attempt; it retries on the next sweep")
		}
		return nil
	}
	delete(s.lastFailure, rec.ID)
	// 결과 줄 — 계좌 · 수량 · 가격 · 상세 문구(주문 번호가 섞일 수 있음) 없이 사유 코드만.
	s.deps.Log.Event(obs.EventOrderInDoubt,
		obs.FieldAttemptID, res.AttemptID, obs.FieldSymbol, res.Symbol,
		"settled_state", string(res.State), obs.FieldReason, string(res.Reason),
		obs.FieldDetail, "the in-doubt settler resolved this attempt while the engine runs")
	if res.State == journal.StateUnresolvedInDoubt {
		s.noteParkOutsideExit(ctx, res)
	}
	return nil
}

// noteFirstSight 는 IN_DOUBT 발주를 처음 볼 때 critical 한 번(Q-A 처분 — 정본 「등급화된 알림」의 IN_DOUBT 발생).
func (s *inDoubtSettle) noteFirstSight(ctx context.Context, rec journal.AttemptRecord) {
	key := string(obs.EventOrderInDoubt) + "|attempt:" + rec.ID
	if s.handled(key) {
		return
	}
	symbol := s.symbolOf(ctx, rec)
	s.record(ctx, key, obs.Event{
		Type:  obs.EventOrderInDoubt,
		Key:   key,
		Title: strings.TrimSpace(symbol + " 주문 결과 불명(IN_DOUBT) — 엔진이 해소 중"),
		Body: "발주 attempt " + rec.ID + " 의 결과를 브로커가 확정하지 않았다. 엔진이 조회 대조로 해소를 시작한다. " +
			"그동안 이 종목의 다른 주문(손절 포함)은 막힌다. 해소 결과(접수 확정 · 부재 · 판정 불능 park)는 엔진 로그 in-doubt-settle 줄과 " +
			"이어지는 알림에 남는다.",
		Fields: map[string]any{obs.FieldSymbol: symbol, obs.FieldAttemptID: rec.ID, "intent_id": rec.IntentID},
	})
}

// noteStale 은 전송 시작(없으면 기록) 시각부터 10분 이상 IN_DOUBT 인 발주를 critical 한 번으로 알림 — 해소가 오류로 되풀이되거나
// 직렬 대기열에 밀린 경우만 남(건강한 경로는 그 전에 종결됨).
func (s *inDoubtSettle) noteStale(ctx context.Context, rec journal.AttemptRecord) {
	key := string(obs.EventOrderInDoubt) + "|stale:" + rec.ID
	if s.handled(key) {
		return
	}
	stamp := strings.TrimSpace(rec.DispatchStartedAt)
	if stamp == "" {
		stamp = strings.TrimSpace(rec.RecordedAt)
	}
	since, err := time.Parse(time.RFC3339Nano, stamp)
	if err != nil {
		// 시각을 못 읽으면 경과를 모름 — 알리는 쪽으로(침묵보다 중복이 안전).
		since = time.Time{}
	}
	waited := s.deps.Clock.Now().Sub(since)
	if !since.IsZero() && waited < s.bound {
		return
	}
	symbol := s.symbolOf(ctx, rec)
	minutes := int(waited / time.Minute)
	body := fmt.Sprintf("발주 attempt %s 가 %d분째 결과 불명(IN_DOUBT)으로 해소되지 않았다 — 원 주문이 살아 있을 수 있어 그 종목의 다른 "+
		"주문(손절 포함)이 막혀 있다. 엔진 로그의 in-doubt-settle 줄에서 해소 오류를 확인하라.", rec.ID, minutes)
	if since.IsZero() {
		body = "발주 attempt " + rec.ID + " 의 전송 시각을 읽지 못해 경과를 잴 수 없다 — 결과 불명(IN_DOUBT)이 남아 있다. 엔진 로그의 " +
			"in-doubt-settle 줄을 확인하라."
	}
	s.record(ctx, key, obs.Event{
		Type:   obs.EventOrderInDoubt,
		Key:    key,
		Title:  strings.TrimSpace(symbol + " 주문 결과 불명이 해소되지 않고 남아 있다"),
		Body:   body,
		Fields: map[string]any{obs.FieldSymbol: symbol, obs.FieldAttemptID: rec.ID, "waited_minutes": minutes},
	})
}

// noteParkOutsideExit 는 세션 중 park 된 attempt 가 무장된 exit 발의의 것이 아닐 때만 critical 한 번(Q-E 처분).
// exit 발의의 attempt 는 기존 recordParkCause 가 알리므로 여기서 또 알리지 않음. 발의 목록을 못 읽으면 알림(모르면 통지 쪽).
func (s *inDoubtSettle) noteParkOutsideExit(ctx context.Context, res execgw.Resolution) {
	key := string(obs.EventOrderUnresolved) + "|attempt:" + res.AttemptID
	if s.handled(key) {
		return
	}
	armed, err := s.deps.Ledger.ArmedExitProposals(ctx, s.deps.Account)
	if err == nil {
		for _, proposal := range armed {
			if strings.TrimSpace(proposal.IntentID) == strings.TrimSpace(res.IntentID) {
				return
			}
		}
	}
	s.record(ctx, key, obs.Event{
		Type:  obs.EventOrderUnresolved,
		Key:   key,
		Title: strings.TrimSpace(res.Symbol + " 판정 불능(park) — 운영자 해동 필요"),
		Body: "발주 attempt " + res.AttemptID + " 가 세션 중 해소에서 판정 불능(UNRESOLVED_IN_DOUBT)으로 park 됐다. 원 주문이 살아 있을 수 " +
			"있다. 계정 전체의 신규 진입이 엔진 재시작까지 막힌다(청산은 막히지 않음). 브로커에서 확인한 뒤 해동 명령(tossctl engine " +
			"attempt-resolve)으로 닫아야 한다.",
		Fields: map[string]any{obs.FieldSymbol: res.Symbol, obs.FieldAttemptID: res.AttemptID, "intent_id": res.IntentID,
			obs.FieldReason: string(res.Reason)},
	})
}

// noteUnreadable 은 미종결 attempt 목록을 연속 10분 이상 못 읽으면 critical 한 번(Q-F 처분). 에피소드 = 연속 id(벽시계 key 는 역행에
// 흡수되므로 — a094 D−7.1).
func (s *inDoubtSettle) noteUnreadable(ctx context.Context, err error) {
	now := s.deps.Clock.Now()
	if s.unreadableID == "" {
		s.unreadableSince, s.unreadableID = now, s.deps.NewID()
		s.deps.Log.Warn(obs.EventOrderInDoubt, obs.FieldError, obs.MaskAccount(err, s.deps.Account).Error(),
			obs.FieldDetail, "the in-doubt settler could not list unsettled attempts; it retries on the next sweep")
	}
	if now.Sub(s.unreadableSince) < s.bound {
		return
	}
	key := string(obs.EventOrderInDoubt) + "|ledger-unreadable:" + s.unreadableID
	if s.handled(key) {
		return
	}
	s.record(ctx, key, obs.Event{
		Type:  obs.EventOrderInDoubt,
		Key:   key,
		Title: "세션 중 IN_DOUBT 해소가 원장을 읽지 못하고 있다",
		Body: fmt.Sprintf("원장의 미종결 attempt 목록을 %d분째 읽지 못했다 — 그동안 결과 불명 주문의 해소와 그 알림이 멈췄다. 엔진 로그를 "+
			"확인하라.", int(now.Sub(s.unreadableSince)/time.Minute)),
	})
}

// onStop 은 실행자가 취소가 아닌 사유로 멈췄을 때(패닉 포함) critical 한 번. 이미 끝난 ctx 로는 기록하지 않음(런타임이 내려가는 중).
func (s *inDoubtSettle) onStop(ctx context.Context, err error) {
	key := string(obs.EventOrderInDoubt) + "|settler-stopped:" + s.runID
	s.record(ctx, key, obs.Event{
		Type:  obs.EventOrderInDoubt,
		Key:   key,
		Title: "세션 중 IN_DOUBT 해소 실행자가 멈췄다",
		Body: "세션 중 IN_DOUBT 해소 실행자가 멈췄다 — 엔진 재시작 전까지 새 결과 불명 주문은 해소되지 않는다(손절 판정 루프는 계속 돈다). " +
			"엔진 로그를 확인하라.",
		Fields: map[string]any{obs.FieldReason: describe(obs.MaskAccount(err, s.deps.Account))},
	})
}

// record 는 critical 하나를 단일 입구로 창 0 기록함. 취소된 ctx 면 기록하지 않음(실패가 진입 게이트를 잠그므로). 성공했을 때만
// 래치하고, 실패한 알림은 unsent 에 두어 다음 스윕 머리에서 다시 기록함(취소로 못 한 것도 같음 — 다음 기동이 아니라 같은 프로세스의 다음 스윕).
func (s *inDoubtSettle) record(ctx context.Context, key string, e obs.Event) bool {
	if ctx.Err() != nil {
		s.unsent[key] = e
		return false
	}
	if s.deps.Critical == nil {
		s.deps.Log.Warn(e.Type, obs.FieldDetail, "no critical recorder is wired, so this alert is not durable: "+e.Title)
		return false
	}
	if err := s.deps.Critical.RecordCritical(ctx, e, 0); err != nil {
		s.unsent[key] = e
		s.deps.Log.Warn(e.Type, obs.FieldError, obs.MaskAccount(err, s.deps.Account).Error(),
			obs.FieldDetail, "the critical alert could not be recorded; the next sweep retries")
		return false
	}
	delete(s.unsent, key)
	s.recorded[key] = true
	return true
}

// retryUnsent 는 기록에 실패했던 알림을 다시 기록함. 기록 순서는 정해 두지 않음(key 마다 한 행이라 순서가 결과를 바꾸지 않음).
func (s *inDoubtSettle) retryUnsent(ctx context.Context) {
	for key, e := range s.unsent {
		s.record(ctx, key, e)
	}
}

// handled 는 그 key 가 이미 기록됐거나 재시도 대기 중인가임 — 대기 중인 것은 retryUnsent 가 맡으므로 같은 스윕에서 두 번 시도하지 않음.
func (s *inDoubtSettle) handled(key string) bool {
	if s.recorded[key] {
		return true
	}
	_, waiting := s.unsent[key]
	return waiting
}

// resolveFailureClass 는 해소 오류의 분류임 — 로그에 원문 대신 남김(원장 오류는 주문 번호를 품을 수 있음).
func resolveFailureClass(err error) string {
	switch {
	case errors.Is(err, execgw.ErrNotResolvable):
		return "not_resolvable"
	case errors.Is(err, journal.ErrUnexpectedState):
		return "ledger_state_changed"
	case errors.Is(err, journal.ErrAttemptNotFound):
		return "attempt_not_found"
	case errors.Is(err, execgw.ErrMalformedOrder):
		return "broker_order_unreadable"
	case errors.Is(err, execgw.ErrTooManyPages), errors.Is(err, execgw.ErrCursorLoop):
		return "broker_list_incomplete"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	default:
		return "other"
	}
}

// symbolOf 는 알림 제목용 종목임. 못 읽으면 빈 값 — 알림을 막지 않음.
func (s *inDoubtSettle) symbolOf(ctx context.Context, rec journal.AttemptRecord) string {
	intent, err := s.deps.Ledger.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(intent.Symbol)
}

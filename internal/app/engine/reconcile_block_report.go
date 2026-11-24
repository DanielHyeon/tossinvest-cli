package engine

// reconcile_block_report.go 는 대사 차단 때문에 편입되지 못한 보유와 운영자만 풀 수 있는 차단 자체를 알림(a194).
//
// 왜: 차단 분기(judgeHoldings B10)는 「전이 상태는 스스로 풀린다」는 전제로 침묵했는데, 운영자만 풀 수 있는 차단은 스스로
// 풀리지 않아 그 종목에 생긴 보유가 손절 없이 남았음(2026-11-23 운영). 편입 여부 · 차단 · 해제 · 청산 규칙은 바꾸지 않고 알림만 더함.

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const (
	factBlockedOperatorOnly = "blocked_operator_only" // 운영자만 풀 수 있는 차단 아래 — 유예 없음
	factBlockedStale        = "blocked_stale"         // 스스로 풀리는 차단이 유예를 넘김(또는 덮는 블록을 못 찾음)
	// blockedHoldingGrace 는 스스로 풀리는 차단의 유예임(design D2 — 대사 주기 약 62초 × 10).
	blockedHoldingGrace = 10 * time.Minute
	// blockedReportsPerCycle 은 한 주기의 보유 보고 Notify 상한임 — 알림을 줄이는 규칙이 아니라 동기 전송의 주기당 지연 상한(design D3).
	blockedReportsPerCycle = 3
)

// blockedHolding 은 judgeHoldings 의 차단 분기에서 빠진 보유임.
type blockedHolding struct {
	position       journal.Position
	market, symbol string
}

// dueBlockedReport 는 이번 주기에 보고할 보유 하나임.
type dueBlockedReport struct {
	position journal.Position
	fact     string
	cause    string
	since    time.Time // 차단이 걸린 시각(모르면 영값)
	first    time.Time // 이 보유가 처음 빠진 시각 — 보고 순서의 열쇠
}

// reportBlocked 는 judgeHoldings 의 끝에서 부름 — adopt · 기존 무관리 보고 뒤라 이 보고의 동기 전송이 그 주기의 편입을 밀지 않음.
func (d *ReconcileDriver) reportBlocked(ctx context.Context, held []blockedHolding) {
	var blocks []reconcile.Block
	if d.opts.Tracker != nil {
		blocks = d.opts.Tracker.Blocks()
	}
	d.reportBlockedWith(ctx, held, blocks, d.clk.Now())
}

// reportBlockedWith 는 이 주기의 블록 목록으로 보유 보고와 차단 보고를 냄.
func (d *ReconcileDriver) reportBlockedWith(ctx context.Context, held []blockedHolding, blocks []reconcile.Block,
	now time.Time) {
	d.reportBlockedHoldings(ctx, held, blocks, now)
	d.reportOperatorOnlyBlocks(ctx, blocks, now)
}

// reportBlockedHoldings 는 편입 후보인 차단 보유를 판정해 보고함(design D1 · D2 · D3 · D4).
func (d *ReconcileDriver) reportBlockedHoldings(ctx context.Context, held []blockedHolding, blocks []reconcile.Block,
	now time.Time) {
	if d.blockedSince == nil {
		d.blockedSince = map[string]time.Time{}
	}
	if d.blockedRound == nil {
		d.blockedRound = map[string]bool{}
	}
	seen := map[string]bool{}
	var due []dueBlockedReport
	for _, h := range held {
		// 편입 대상이 아닌 보유(제외 · 꺼짐∧미지정)는 차단이 없어도 편입되지 않음 — 「차단 때문에 보호가 빠졌다」가 거짓(D4).
		if d.opts.Adoption.Excludes(h.symbol) || (!d.opts.Adoption.Enabled && !d.opts.Adoption.Included(h.symbol)) {
			continue
		}
		id := h.position.ID
		seen[id] = true
		first, ok := d.blockedSince[id]
		if !ok {
			first = now
			d.blockedSince[id] = now
		}
		r := dueBlockedReport{position: h.position, first: first}
		if b, ok := earliestCoveringBlock(blocks, h.market, h.symbol, true); ok {
			r.fact, r.cause, r.since = factBlockedOperatorOnly, b.Cause, b.Since
		} else {
			if now.Sub(first) < blockedHoldingGrace {
				continue // 스스로 풀리는 차단의 유예 안 — 전이 상태
			}
			// 덮는 블록을 못 찾은 순간(blocked() 와 Blocks() 의 두 읽기 사이 불일치)도 조용히 건너뛰지 않고 같은 길로(D1).
			r.fact, r.cause = factBlockedStale, "미상"
			if b, ok := earliestCoveringBlock(blocks, h.market, h.symbol, false); ok {
				r.cause, r.since = b.Cause, b.Since
			}
		}
		due = append(due, r)
	}
	// 이번 주기에 빠지지 않은 보유(편입 · 해제 · 청산)의 시계는 지움 — 다시 빠지면 새 유예.
	for id := range d.blockedSince {
		if !seen[id] {
			delete(d.blockedSince, id)
			delete(d.blockedRound, id)
		}
	}

	sort.Slice(due, func(i, j int) bool {
		if !due[i].first.Equal(due[j].first) {
			return due[i].first.Before(due[j].first)
		}
		return due[i].position.ID < due[j].position.ID
	})
	// 주기 상한은 차례(round)로 셈: 이번 차례에 시도한 보유는 차례의 모두를 한 번씩 시도할 때까지 다시 시도하지 않음.
	// 셋 이하면 매 주기 전원 시도(critical 은 매 관측 기록 시도 — 메모리 억제가 아님), 넘으면 남은 보유가 다음 주기에 나감.
	// 차례에 남은 보유가 모두 빠졌으면(편입 · 해제) 이번 주기에 새 차례를 시작함 — 한 주기를 빈손으로 보내지 않게(gstack review).
	left := false
	for _, r := range due {
		if !d.blockedRound[r.position.ID] && (d.opts.NotificationsEnabled || !d.unmanaged[r.position.ID][r.fact]) {
			left = true
			break
		}
	}
	if !left {
		d.blockedRound = map[string]bool{}
	}
	attempted, more := 0, false
	for _, r := range due {
		if d.blockedRound[r.position.ID] {
			continue
		}
		if !d.opts.NotificationsEnabled && d.unmanaged[r.position.ID][r.fact] {
			continue // 일반 등급은 같은 사실을 한 번만(래치)
		}
		if attempted == blockedReportsPerCycle {
			more = true
			break
		}
		d.alertBlockedHolding(ctx, r, now)
		d.blockedRound[r.position.ID] = true
		attempted++
	}
	if !more {
		d.blockedRound = map[string]bool{}
	}
}

// earliestCoveringBlock 은 이 (시장, 종목)을 덮는 블록 중 가장 이른 Since 의 것임. operatorOnly 면 운영자만 풀 수 있는 것만 봄.
func earliestCoveringBlock(blocks []reconcile.Block, market, symbol string, operatorOnly bool) (reconcile.Block, bool) {
	var (
		best  reconcile.Block
		found bool
	)
	for _, b := range blocks {
		if !b.Covers(market, symbol) || (operatorOnly && b.Release != reconcile.ReleaseOperatorOnly) {
			continue
		}
		if !found || b.Since.Before(best.Since) {
			best, found = b, true
		}
	}
	return best, found
}

// alertBlockedHolding 은 보유 보고 하나를 냄. 알림 켜짐이면 critical(래치 없이 매 관측 — 중복은 outbox key 와 재알림 창),
// 꺼짐이면 일반 등급 + (포지션, 사실) 래치 — 꺼진 엔진의 critical 은 미전달이 진입 래치 · 모드 승격을 부르기 때문(D3).
func (d *ReconcileDriver) alertBlockedHolding(ctx context.Context, r dueBlockedReport, now time.Time) {
	p := r.position
	moment := journal.RFC3339(now)
	body := fmt.Sprintf("%s 대사에서 이 종목은 대사 차단(원인 %s, %s) 아래에 있어 편입하지 못했다. "+
		"그 시각 손절·익절이 걸려 있지 않았다.\n", kstText(now), r.cause, blockSinceText(r.since))
	if r.fact == factBlockedOperatorOnly {
		body += operatorReleaseGuide + " 그 전까지 이 종목은 앱에서 직접 관리하라."
	} else {
		// 재시작 뒤 복원된 수량 불일치 블록은 조정 크레딧이 없어 일치만으로 안 풀릴 수 있음 — 단정이 아니라 가능성으로 안내(D6).
		body += "차단이 10분 넘게 풀리지 않았다. 스스로 풀리지 않으면 운영자 해제(`tossctl engine reconcile-resolve`)가 필요할 수 있다 — " +
			"그 명령은 엔진을 멈춰야 실행되고 멈춘 동안에는 모든 포지션의 자동 손절이 돌지 않는다. " +
			"엔진 로그의 대사 상태를 확인하고 그 전까지 이 종목은 앱에서 직접 관리하라."
	}
	body += "\n이 알림이 늦게 도착했다면 그 사이 상태가 바뀌었을 수 있으니 콘솔에서 현재 상태를 확인하라."

	t := obs.EventExitPositionAdoptionFailed
	if !d.opts.NotificationsEnabled {
		t = obs.EventExitPositionUnmanaged
		if d.unmanaged[p.ID] == nil {
			d.unmanaged[p.ID] = map[string]bool{}
		}
		d.unmanaged[p.ID][r.fact] = true
	}
	d.alert(ctx, obs.Event{
		Type:  t,
		Key:   reconcileUnmanagedKey(t, r.fact, p.ID),
		Title: d.label(p.Symbol) + " 대사 차단으로 편입하지 못했다 — 보호가 열리지 않았다",
		Body:  body,
		Fields: map[string]any{
			obs.FieldSymbol: p.Symbol,
			"market":        p.Market,
			"position_id":   p.ID,
			"fact":          r.fact,
			"cause":         r.cause,
			"observed_at":   moment,
		},
	})
}

// reportOperatorOnlyBlocks 는 운영자만 풀 수 있는 차단을 보유가 없어도 하루(KST) 한 번 알림(D5).
// 래치는 기록이 성공한 뒤에만 섬 — 실패는 보낸 것으로 세지 않고 다음 주기가 다시 시도. 이번 주기에 없는 key 는 버려 맵이 자라지 않음.
func (d *ReconcileDriver) reportOperatorOnlyBlocks(ctx context.Context, blocks []reconcile.Block, now time.Time) {
	day := now.In(kst).Format("2006-01-02") // kst 는 attestation_watch.go 의 고정 오프셋
	next := map[string]bool{}
	for _, b := range blocks {
		if b.Release != reconcile.ReleaseOperatorOnly {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(b.Symbol))
		key := strings.Join([]string{string(obs.EventReconcilePermanent), string(b.Scope), symbol, b.Cause, day}, "|")
		if d.blockReported[key] || next[key] || d.recordOperatorOnlyBlock(ctx, b, symbol, key, now) {
			next[key] = true
		}
	}
	d.blockReported = next
}

// recordOperatorOnlyBlock 은 차단 보고 하나를 기록하고 기록됐는지 돌려줌.
func (d *ReconcileDriver) recordOperatorOnlyBlock(ctx context.Context, b reconcile.Block, symbol, key string,
	now time.Time) bool {
	target := "계좌 전체"
	if symbol != "" {
		target = d.label(symbol)
	}
	moment := journal.RFC3339(now)
	// 본문은 관측 시각의 사건으로 씀 — 늦게 배달돼도 참이게(D6).
	body := fmt.Sprintf("%s 대사에서 %s에 운영자만 풀 수 있는 대사 차단(원인 %s, %s)이 활성이었다. "+
		"그 시각 그 범위의 신규 진입은 막혀 있었고, 그 범위에 보유가 있었다면 엔진이 편입하지 않아 손절·익절이 걸려 있지 않았다.\n"+
		operatorReleaseGuide+" 풀리지 않은 동안 하루(KST) 한 번 알린다.\n"+
		"이 알림이 늦게 도착했다면 그 사이 상태가 바뀌었을 수 있으니 콘솔에서 현재 상태를 확인하라.",
		kstText(now), target, b.Cause, blockSinceText(b.Since))
	if !d.opts.NotificationsEnabled {
		// 알림이 꺼진 엔진은 critical 을 만들지 않음 — 로그 한 줄만(D5).
		if d.opts.Log != nil {
			d.opts.Log.Warn(obs.EventReconcilePermanent, obs.FieldSymbol, symbol, "scope", string(b.Scope),
				"cause", b.Cause, obs.FieldDetail, body)
		}
		return true
	}
	if d.opts.Alerts == nil {
		return false // 기록할 곳이 없으면 알린 것으로 세지 않음
	}
	// d.alert 는 오류를 삼키므로 Notify 의 반환값을 직접 봄(r2 R2-1).
	err := d.opts.Alerts.Notify(ctx, obs.Event{
		Type:  obs.EventReconcilePermanent,
		Key:   key,
		Title: target + " 운영자만 풀 수 있는 대사 차단",
		Body:  body,
		Fields: map[string]any{
			obs.FieldSymbol: symbol,
			"scope":         string(b.Scope),
			"cause":         b.Cause,
			"observed_at":   moment,
		},
	})
	if err != nil {
		if d.opts.Log != nil {
			d.opts.Log.Error(obs.EventAlertUndelivered, err, "alert", string(obs.EventReconcilePermanent))
		}
		return false
	}
	return true
}

// operatorReleaseGuide 는 운영자 전용 차단의 할 일 문장임. 해제 명령이 받는 원인 목록은 명령 쪽이 정하므로 여기 복제하지 않고,
// 엔진 정지가 모든 포지션의 자동 손절을 멈춘다는 사실을 숨기지 않음(Manager 결정 2026-11-23).
const operatorReleaseGuide = "이 차단은 스스로 풀리지 않는다. 풀려면 운영자 해제(`tossctl engine reconcile-resolve`)가 필요하다 — " +
	"이 명령은 엔진을 멈춰야 실행되고 멈춘 동안에는 모든 포지션의 자동 손절이 돌지 않으므로, 보유 종목의 시장이 닫힌 시간에 실행하라. " +
	"명령이 이 원인을 받지 않으면 거절하며 그 사유를 말한다."

// blockSinceText 는 차단이 걸린 시각 문구임 — 모르면 그렇다고 말함.
func blockSinceText(since time.Time) string {
	if since.IsZero() {
		return "걸린 시각 미상"
	}
	return kstText(since) + "부터"
}

// kstText 는 운영자가 읽는 본문 시각임(한국 시간). 기계가 읽는 observed_at 필드는 RFC3339 UTC 그대로.
func kstText(t time.Time) string { return t.In(kst).Format("2006-01-02 15:04 KST") }

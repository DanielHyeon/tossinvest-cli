# a201 gstack `review` r1 (tasks 3.3) — 2026-11-24

- 대상: worktree `/home/daniel/tossos-work/a201-impl`, 브랜치 `fix/a201-stale-quantity-comparison`, HEAD `065be75f8`, base `35d28a535`(`base-commit.txt`). diff `35d28a535..065be75f8`(119 파일, +13723/−609 — 그중 `analysis/` 산출물을 뺀 코드 · 시험 · 문서 1605 줄).
- 스킬: gstack `review`(skill-start PROTO 1, REPO_MODE solo, SESSION_KIND interactive). 발견만 — 수리 0.
- 리뷰어: 본 리뷰(체크리스트 Pass 1 · Pass 2) + 전문가 5(testing · maintainability · security · performance · simplification, 각자 새 문맥) + red team 1 + native 적대 1(「Think like an attacker and a chaos engineer」). 전부 같은 하네스(Claude) — 다른 계열 보이스 아님. Codex 외부 패스는 돌리지 않음(아래 「건너뛴 단계」).
- 앞선 처분(R1~R6 · V3R1-* · V3R2-*)과 design V8-1~V8-4 · §4 공시 항목은 다시 올리지 않음.

```text
PR type: mixed (application + docs)
Scope Check: CLEAN
Intent: design v3 — 체결이 조정을 대신한 종목에 체결 credit, 다음 깨끗한 비교에서 FILL_APPLIED 해제 (+ V4b 반박 1 자리)
Plan: openspec/changes/a201-…/design.md · tasks.md
Delivered: RunOnce 2 자리 · Observe 2 자리 · persist 원인 갈래 · ValidReconcileReleaseCause case 1 · 새 파일 2 · 시험 · 증거
Plan items: tasks 2.1~3.2 DONE, 3.3(이 리뷰) · 3.4 · 4.1 미착수(정상)
Diff read: full for production Go + tests + change 문서; analysis/ 산출물(원문 · ast.json · 변이 raw)은 인용 대조만
Undeclared behavior changes: none (V4b 의 조정 credit 경로 변경은 design §2 에 공시됨)
GATE_SUMMARY: listed=8 eligible=3 inspected=3 unread=0 tagged={RH-1:1} unmatched=113
BANNED_TERMS: none
```

## 발견

분류: (1) 런타임에 틀린 결과(blocker) (2) 계약(design/spec)과 코드의 불일치 (3) 시험 · 증거 공백 (4) 취향 · 정리(참고).

| # | 분류 | 한 줄 | 재현 입력 / 근거 | 실행 | 제안 처분 |
|---|---|---|---|---|---|
| G1 | (2) — 문언 밖, V4b 와 같은 원리 | `RunOnce` 가 `Tracker.Refresh` 오류로 Observe **앞에서** 반환하면, 그 주기 비교가 credit 종목을 외부 포지션(분쟁)으로 봐도 반박이 돌지 않아 credit 이 살아남고, 다음 깨끗한 비교에서 `FILL_APPLIED` 로 해제 · 진입 열림 · 편입 | reconcileloop.go:460-465(Refresh 오류 → `return cycle`), 반박 자리는 Observe 안뿐(mismatch.go:668 · :716 · 끝 루프). 아래 탐침 | **예**(사본 worktree, 탐침 4 시험) | Manager 결정: (a) Refresh 오류 반환 직전에 `refuteUsableAdjustmentCredits(symbolsInDispute(diff), diff.AsOf)` 를 잠금 아래 부르는 Tracker 메서드 1(조정 credit 경로도 같이 바뀜 — V4b 와 같은 「보수 방향 기존 동작 변경」, 편집 표면 · 번들 늘어남) 또는 (b) design §4 공시 + 별도 change |
| G2 | (3) | design §3-13 앞 절반(「수량 흔적 없는 a192 결속 행 → 체결 credit 안 섬」) 엔진 시험이 V2 (a) 를 재지 않음 — 주입이 체결 0 취소라 종목 워터마크가 안 움직여 (b) 가 먼저 막음. M01(V2 (a) → 분쟁 종목 전부) 에서 이 시험 PASS | `a201_fill_credit_cycles_test.go:553` `TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit`; `analysis/mutation/raw/M01-a-any-disputed-symbol.txt:121-126`(PASS), 원장 M01 은 단위 시험 1 로만 CAUGHT | 원장 raw 대조(재실행 안 함) | 같은 hook 에서 다른 계좌의 같은 종목 체결(W3-b 수법)을 더해 (a) 만 남기고 M01 이 이 시험으로 CAUGHT 인지 재기 |
| G3 | (3) | §3-5 「조정이 반영된 종목 → credit 없음」 엔진 시험이 V2 (c) 를 재지 않음 — W4 입력은 운영자 전용(ATTRIBUTION_FAILED) 블록만 남겨 (d) · B7 이 먼저 막음. M03(V2 (c) 삭제) 에서 PASS. (c) 가 유일한 가드인 입력(수렴한 종목 + 표식 뒤 워터마크 이동)이면 같은 as-of 로 덮어써 원인이 `FILL_APPLIED` 로 기록됨(이름만 틀림, 해제 자체는 조정이 정당화) | `a201_fill_credit_cycles_test.go:399` `TestA201AConvergedSymbolGetsNoFillCredit`; `raw/M03-c-converged-not-excluded.txt:91-96`(PASS); claim-check-v3.md:64(W4 = ATTRIBUTION_FAILED) | 원장 raw 대조(재실행 안 함) | 엔진 시험 1: `start` → `converge.at=a201BeforePositions` 에서 다른 계좌 같은 종목 체결 → 수렴은 커밋 → 다음 깨끗한 주기 `ADJUSTMENT_APPLIED` 단언, M03 CAUGHT 확인 |
| G4 | (4) | 낡은 주석: 해제는 `ADJUSTMENT_APPLIED` 뿐이라고 말하는 자리가 남음 — diff 안 2(`Outcome.Cleared` 문서 mismatch.go:614-615, `persistResult` 문서 :877-883 「out.Cleared only ever carries blocks whose symbol an adjustment converged」 — 이제 거짓), diff 밖 3(execgw/symbolgate.go:146-148, reconcile/mismatch.go:60-64 파일 머리, journal/reconcile_states.go:60-62) | 위 줄 | 읽기 | 문구 정리(한 커밋) — 코드 동작 무관 |
| G5 | (4) | 문구: `a201_fill_credit.go:5` 파일 머리가 「로컬 읽기 뒤」로 씀 — R1 처분 뒤 정확한 표현은 「로컬 읽기 직전 표식 뒤(V8-4)」. `reconcileloop.go:22-30` 「One cycle」 단계 목록에 표식 읽기 · 체결 credit 부여 단계 없음 | 위 줄 | 읽기 | 문구 정리 |
| G6 | (4) | `FILL_APPLIED` 해제마다 남는 로그 줄 없음 — a192 는 `logOrderTerminalReleases` 가 해제마다 한 줄, `FILL_APPLIED` 는 `logCycle` 의 `released=N` 합계뿐이라 `ADJUSTMENT_APPLIED` 와 구별 불가. V8 「사람이 원인을 보지 않은 해제」라 tasks §5 의 운영 관측이 DB 직접 조회에 의존 | reconcileloop.go:468(a192 만), logCycle | 읽기 | 있으면 좋은 관측(a192 D7 모양, 계좌 데이터 없음) — 결정은 Manager |

blocker((1)) 0. 전문가 security · performance · simplification 은 `NO FINDINGS`.

### G1 탐침(사본 worktree `065be75f8`, 실행 뒤 제거)

시험 파일(커밋하지 않음, 아래 전문) — `a201World` 하네스에 `Tracker.Journal` 을 감싸 첫 `ActiveReconcileStates` 를 한 번 실패시킴.

명령: `rtk proxy go test -count=1 -timeout 30m -run 'TestGstackProbe' -v ./internal/app/engine/` → `ok … 0.593s`, `-run 'TestGstackProbeAdj'` → `ok … 0.683s`.

| 판 | k | k+1(사람 매수 10, 외부 포지션) | k+2(접힌 10 = 계좌 10) | 결과 |
|---|---|---|---|---|
| 체결 credit · Refresh 실패 | W2 — QM 1 active, credit | `err=…injected refresh read failure` folded=1 | released=1 adopted=1 | `QUANTITY_MISMATCH\|released:FILL_APPLIED`, entry 허용 |
| 체결 credit · 대조군 | 같음 | awaiting=1 (반박) | awaiting=1 | `QUANTITY_MISMATCH\|active`, entry 거절 |
| 조정 credit · Refresh 실패 | 외부 전량 매도 → converged=1 credit | 같은 실패 | released=1 adopted=1 | `released:ADJUSTMENT_APPLIED`, entry 허용 |
| 조정 credit · 대조군 | 같음 | awaiting=1 | awaiting=1 | active, entry 거절 |

- 해석: 해제 여부가 무관한 장부 읽기 하나의 실패에 달림. 조정 credit 판이 같은 결과라 **클래스는 a201 이전부터 있음**(HEAD 에서 잼 — 조정 경로 코드는 V4b 외 무편집, base 에서 직접 재지는 않음). a201 은 W2 · W3 모양 credit 을 새로 만들어 이 구멍에 닿는 입력을 넓힘.
- 해제된 순간의 수량은 실제로 일치(계좌 10 = 로컬 10)라 해제 순간 노출 차이는 없음 — 그래서 (1) 이 아니라 (2)로 둠. 정본 :240 문언은 「관측」의 반박을 요구하는데 Refresh 실패 주기는 관측이 아님. 같은 원리(관측된 분쟁이 credit 을 못 지우는 조기 반환)를 V3R1-1 은 P1 로 채택해 V4b 로 막았고, 이것은 V4b 가 닫지 않은 둘째 조기 반환.
- 발생 조건: credit 이 선 다음 주기에 (i) 그 종목이 외부 포지션으로 보이고 (ii) 같은 주기 `ActiveReconcileStates` 읽기가 실패. 운영에서 그 읽기 실패 빈도는 재지 않음.

```go
package engine_test

// gstack r1 탐침(커밋하지 않음): RunOnce 의 Refresh 오류 조기 반환이 그 주기 비교의 분쟁 반박을 건너뛰는지 잼.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

type gstackFailOnce struct {
	reconcile.ReconcileStore
	fail bool
}

func (s *gstackFailOnce) ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error) {
	if s.fail {
		s.fail = false
		return nil, errors.New("probe: injected refresh read failure")
	}
	return s.ReconcileStore.ActiveReconcileStates(ctx)
}

func gstackProbe(t *testing.T, failRefresh bool) {
	const symbol = "SYN201"
	w := newA201World(t)
	fs := &gstackFailOnce{ReconcileStore: w.tracker.Journal}
	w.tracker.Journal = fs
	w.start(symbol)
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("k W2", symbol, c)
	if c.Err != nil || c.Converged != 0 {
		t.Fatalf("k = %+v", c)
	}
	// k+1: 사람이 앱으로 10 주 매수 — 로컬 인스턴스 없음 → 외부 포지션(분쟁).
	w.holds(symbol, "10", "100", 100)
	fs.fail = failRefresh
	c = w.next()
	w.logCycle("k+1 external buy", symbol, c)
	// k+2: 접힌 외부 10 = 계좌 10 — 깨끗한 비교.
	c = w.next()
	w.logCycle("k+2 clean", symbol, c)
	t.Logf("RESULT failRefresh=%v rows=%q entry_refused=%v", failRefresh, w.rows(symbol),
		w.tracker.EntryAllowed("kr", symbol) != nil)
}

// 조정 credit 판: 계좌가 전량을 밖에서 팔아 수렴이 0 으로 조정(credit) → 다음 주기 사람 매수 10(외부 포지션) + Refresh 실패 → 깨끗한 주기.
func gstackProbeAdjustment(t *testing.T, failRefresh bool) {
	const symbol = "SYN201"
	w := newA201World(t)
	fs := &gstackFailOnce{ReconcileStore: w.tracker.Journal}
	w.tracker.Journal = fs
	w.hold(symbol, "31")
	w.holds(symbol, "31", "100", 100)
	w.stopOpenAtBroker(symbol, "31")
	if c := w.next(); c.Err != nil {
		t.Fatalf("start = %+v", c)
	}
	w.holdings.items = nil // 사람이 앱으로 전량 매도(장부 체결 없음), 엔진 매도 주문은 열린 채
	c := w.next()
	w.logCycle("k external sale", symbol, c)
	w.holds(symbol, "10", "100", 100)
	fs.fail = failRefresh
	c = w.next()
	w.logCycle("k+1 external buy", symbol, c)
	c = w.next()
	w.logCycle("k+2 clean", symbol, c)
	t.Logf("RESULT-ADJ failRefresh=%v rows=%q entry_refused=%v", failRefresh, w.rows(symbol),
		w.tracker.EntryAllowed("kr", symbol) != nil)
}

func TestGstackProbeAdjRefreshFailure(t *testing.T) { gstackProbeAdjustment(t, true) }
func TestGstackProbeAdjRefreshOK(t *testing.T)      { gstackProbeAdjustment(t, false) }

func TestGstackProbeRefreshFailureSkipsRefutation(t *testing.T) { gstackProbe(t, true) }
func TestGstackProbeRefreshOKControl(t *testing.T)               { gstackProbe(t, false) }
```

## Gate integrity

`Gate edits: 8 listed, 3 read, 1 finding — 0 kept, 0 restored, 1 open`

- `[06ebb0d91005] RH-1 internal/app/engine/a201_stale_quantity_comparison_repro_test.go` — 1단계 「CurrentBehaviour」 시험 6 삭제. 근거는 diff 안에 보임(design §3-1 「1단계 재현 6개를 v3 계약으로 뒤집는다」, tasks 2.2, 새 파일 `a201_fill_credit_cycles_test.go` 의 `TestA201W2…` · `W3a…` · `AFillJustBefore…` · `W3b…` · `AFillCreditReplaces…` · `TheTrailingSymbol…StaysStuck` 6, 지운 시험 6 은 FLM revision base 번들). 판단: 정당화됨(verified — 같은 입력이 새 계약 단언으로 옮겨짐). 비대화 실행이 아니라 `SESSION_KIND: interactive` 지만 사람에게 묻지 않는 지시라 `open`(Manager 확인)으로 둠.
- 나머지 2 hunk(같은 파일 머리 주석)는 게이트 의미 없음.

## 건너뛴 스킬 단계

- Step 1 `git fetch origin` · merge-base: 하지 않음 — base 를 change 의 `base-commit.txt` sha 로 고정.
- Step 1.5 plan 파일 선택 질문 · Step 3.4 VERSION 큐 · Step 3.5 slop scan(`bun run slop:diff` — 이 저장소에 없음): 건너뜀.
- Step 2.5 리뷰 봇(PR 없음 · push 금지): 건너뜀.
- Step 3.6 교차 프로젝트 학습 질문 · 웹 검색(Aside): 묻지 않고 기존 학습 3 만 사용, 웹 검색 안 함.
- Step 4.7 탐색 QA(브라우저 · 스모크): 해당 없음(UI · CLI 표면 변경 없음) — 대신 G1 탐침 실행.
- Step 4.8 Codex 외부 적대 · 구조 리뷰: **돌리지 않음** — Codex 는 이미 impl r1(`impl-r1-codex.md`)로 이 change 를 리뷰했고, 이번 회차는 「gstack review 1회」 지시라 유료 외부 호출을 더하지 않는 쪽(보수 기본값)을 고름. 외부 커버리지는 이 회차에 없음.
- Step 5 Fix-First 전부(AUTO-FIX · ASK · 시험 작성 · 수리) · Step 5.8 재리뷰 루프 · 커밋 · push · PR · VERSION/CHANGELOG: 지시대로 건너뜀.

## 보지 않은 범위

- base(`35d28a535`)에서 G1 조정 credit 판을 직접 재지 않음(HEAD 에서만).
- G2 · G3 변이 재실행 안 함 — 커밋된 raw 원문으로만 대조.
- 운영 장부에서 `ActiveReconcileStates` 실패 빈도.
- `analysis/` 산출물(ast.json · FLM/BTM · coverage · race · selected-tests 원문)의 내용 정합은 재검증하지 않음.
- `Tracker.Resolve` 본문은 `adjusted = nil` 줄 외 읽지 않음, 콘솔 화면의 원인 표시는 grep 으로 소비자 없음만 확인.

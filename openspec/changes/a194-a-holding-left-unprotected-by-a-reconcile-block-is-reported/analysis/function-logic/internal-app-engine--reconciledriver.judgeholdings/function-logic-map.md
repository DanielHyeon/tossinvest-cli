# Function Logic Map: `ReconcileDriver.judgeHoldings`

> **편집 전 지도(pre-edit) — tasks 1.5, 2026-11-23.** 편집 뒤(3.3)에 AST·이 지도·BTM 을 다시 만든다.

- Source: `internal/app/engine/adoption.go` (107-193), base `064f59fd9`
- AST evidence: `ast.json` — branches 16 / returns 0(`continue` 는 logic-map 이 return 으로 세지 않음) / calls 24,
  source SHA-256 `3d66976f07a50aa7e98447cc17c943d394c67694025ad665765d39562ec0bc9d`.
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/adoption.go --func ReconcileDriver.judgeHoldings`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 키 92 · `sites.json` 116행): `adoption.go`·`reconcileloop.go`·`judgeHoldings`·`ReconcileDriver` 0건(json 파싱 후 문자열 세기).
- a163 판정 핀(`a163_judgement_pin_test.go`): ExitObserver 열 함수만 — 이 함수 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot reconcile.Snapshot` | 이번 주기의 안정 스냅샷 | `RunOnce` call@470 — `Observe` 성공 주기에만(`RunOnce` B7 if@463 → return@467; B1 if@422·B2 if@428·B5 if@451 도 이 호출 전에 반환) | 낡으면 `fresh=false` → B11 에서 후보 판정 전부 침묵 |
| `cycle *ReconcileCycle` | 이번 주기의 셈 | `RunOnce` | `Unmanaged`·`Adopted`·`Deferred` 를 이 함수와 `adopt` 가 더함 |
| `d.opts.Tracker` | 비nil(`NewReconcileDriver` 가 nil 거절, reconcileloop.go:314) | `Context.ReconcileDriver` → `c.Reconcile` | `blocked` 의 nil 갈래(B1 if@197)는 생산에서 닿지 않음 |
| `d.opts.Adoption` | 로드된 `config.Adoption`. 거부된 블록은 0 으로 만들어져 `Enabled=false`·목록 빈 값(config/engine.go:127-131 주석) | `Context.ReconcileDriver` | — |
| `d.opts.SnapshotStaleness` | 0 이면 `DefaultAdoptionStaleness` | 옵션 | B1 if@110 |
| 불변식 | 보유 하나는 「이미 관리」→「RECONCILE」→「스냅샷 낡음」→「제외」→「꺼짐∧미지정」→「후보」 순서로 **첫 갈래 하나**에만 듦(파일 머리 주석의 문 순서) | exit-policy 「외부 취득 포지션의 자동 편입」 | — |
| 불변식 | RECONCILE·낡음 갈래는 `unmanaged`·`candidates` 어느 쪽에도 넣지 않음 → `alertUnmanaged`(B16 range@188)에 닿지 않음 = 무알림 | exit-policy 「전이 상태만 무알림」 | **a194 가 바꾸려는 자리** |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@110) | `SnapshotStaleness <= 0` | `stale = DefaultAdoptionStaleness` | — | 하네스 기본(0) — 대부분의 driver 시험 |
| B2 (range@119) | `snapshot.Holdings` 순회 | — | — | 모든 driver 시험 |
| B3 (if@121) | 정규화한 market 빈 값 | `market = DefaultMarket` | — | 미측정(블록 121.19 hit 0) |
| B4 (if@125) | symbol·market 빈 값 또는 수량 0 | — | continue@126 | 미측정(125.71 hit 0) |
| B5 (if@130) | `CurrentPosition` 오류 — fold 실패·거절 | — | continue@134 | 미측정(130.17 hit 0) |
| B6 (if@136) | 포지션 CLOSED 또는 수량 0 | — | continue@137 | 미측정(136.70 hit 0) |
| B7 (if@140) | `p.ExitEligible()` — 이미 관리 | B8/B9 의 수량 증가 검사 | continue@148 | 편입 뒤 주기·수량 증가 시험(reconcileloop_test.go:844 부근) |
| B8 (if@143) | `p.Adopted()` | `checkExternalIncrease` | — | 위와 같음 |
| B9 (else@145) | 엔진이 연 포지션 | `checkEngineOpenedIncrease` | — | a095 수량 증가 시험(145.10 hit 1) |
| B10 (if@152) | `d.blocked(market, symbol)` = `Tracker.Permanent() ∨ Tracker.EntryAllowed(market,symbol) != nil` | 없음 | continue@153 — 무알림. **a194 편집 지점** | `TestAdoptionIsSilentUnderReconcile`(reconcileloop_test.go:358, QUANTITY_MISMATCH 행) · `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`(:383, IDENTIFIER_CONFLICT 행) · `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`(adoption_include_test.go:109, 계좌 전체 영구) |
| B11 (if@155) | `!fresh` — 스냅샷이 판정 순간 낡음 | 없음 | continue@156 — 무알림 | `TestA095AStaleSnapshotStaysSilent`(a095 …_internal_test.go:683) |
| B12 (if@163) | `Adoption.Excludes(symbol)` | `unmanaged` 에 넣음 | continue@165 | `TestAnExcludedSymbolIsReportedNotAdopted`(reconcileloop_test.go:529) · `TestIncludeDoesNotBypassExclusion`(adoption_include_test.go:145) |
| B13 (if@171) | `!Adoption.Enabled ∧ !Adoption.Included(symbol)` | `unmanaged` 에 넣음 | continue@173 | `TestTheUnmanagedAlertSurvivesAdoptionBeingOff`(reconcileloop_test.go:494) · `TestA095OperatorChosenStatesStayNormal`(a095 :494) |
| (B13 뒤) | 후보 | `candidates` 에 넣음(@175) | — | 편입 시험 |
| B14 (range@179) | 후보 순회 | — | — | 편입 시험 |
| B15 (if@180) | 결과가 `adoptAdopted` 아님(연기·시도 실패) | `unmanaged` 에 넣음 | — | `TestACandidateWithNoQuoteIsDeferred`(reconcileloop_test.go:582) · a095 시도 실패 시험 |
| B16 (range@188) | `unmanaged` 순회 | `cycle.Unmanaged++`, `alertUnmanaged`(call@191) | — | 위 B12·B13·B15 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.opts.Journal.CurrentPosition`(call@129) | 보유의 포지션 인스턴스 | 오류 → B5 continue(보고 없음) | AST |
| `d.blocked`(call@152) | RECONCILE 판정 | bool 만 돌려줌 — 원인(Cause)·범위·시각(Since)·해제 규칙(Release)을 잃음 | AST `../../cited-ast/blocked.ast.json` B1 if@197, calls `Tracker.Permanent`@200·`Tracker.EntryAllowed`@200 |
| `Tracker.EntryAllowed`(blocked 안) | 이 종목을 덮는 첫 블록(정렬: Permanent 먼저) | `*execgw.RejectedError{Reason, Detail}` — **Cause·Since·Release·Permanent 칸 없음**(execgw/reason.go:278-283) | AST `../../cited-ast/EntryAllowed.ast.json` B1 range@1327 → B2 if@1328 → return@1329 |
| `Tracker.Permanent`(blocked 안) | 계좌 전체 **수량 불일치** 영구 승격 표지(`t.permanent`) | `Refresh` 가 `hasPermanentQuantityAccountBlock` 로 다시 계산(mismatch.go:1065) — 다른 원인의 계좌 행은 이 표지 밖(그 행은 `EntryAllowed` 의 계좌 범위 `Covers` 가 잡음) | `../../cited-ast/Permanent.ast.json` 의 AST(갈래 없음, return@1317), `../../cited-ast/Refresh.ast.json` |
| `d.opts.Adoption.Excludes`·`Included`(call@163·@171) | 운영자 목록 | 순수 함수 | AST + config/engine.go:135-142 |
| `d.adopt`(call@178) | 후보 시세 읽기·편입 | 시세 배치 읽기 1회(후보 0이면 0회) | AST |
| `d.alertUnmanaged`(call@191) | 무관리 보고 | 알림 켜짐 ∧ 시도 실패 → `alertAdoptionFailed`(critical, `Notify`, 재알림 창 = Notifier 기본 1시간); 그 밖은 normal 메모리 래치 | AST `../../cited-ast/alertUnmanaged.ast.json` B1 if@451 |
| `d.checkExternalIncrease`·`checkEngineOpenedIncrease`(call@144·@146) | 수량 증가 보고 | 이 change 무관 | AST |

## State mutations and fallbacks

- 지역: `candidates`·`unmanaged`. 드라이버 상태는 직접 쓰지 않는다(래치 `d.unmanaged` 는 `alertUnmanaged`·`adoptOne` 이 씀).
- `cycle.Unmanaged` 는 B16 에서만 증가 — 필드 주석(reconcileloop.go:204-206) 「It excludes transition states」.
- a194 편집 계획(design §2 A): B10 에서 `continue` 전에 보유를 모으고 순회 뒤 새 파일의 함수를 부른다. **B10 의 continue 와 편입 0 은 그대로**(exit-policy 편입 조건 SHALL).

## Safety conclusion

- Safe edit boundary: B10 본문(모으기) + 순회 뒤 새 함수 호출 한 곳. B1~B9·B11~B16 의 조건·순서·continue 불변. `adopt` 앞에 새 브로커 호출 0.
- High-risk impact: yes — 대사·편입·알림. 새 보고가 critical 이면 알림 켜진 엔진에서 전달 실패가 계좌 전체 진입 래치·운영 모드 승격으로 이어진다(정본 「배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」).
- 편집 전 결정 필요(1.4 판명): 새 보고에서 `cycle.Unmanaged` 를 올리면 `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`(adoption_include_test.go:133 `cycle.Unmanaged != 0` → Fatal)이 깨진다 — Manager 결정 대기.

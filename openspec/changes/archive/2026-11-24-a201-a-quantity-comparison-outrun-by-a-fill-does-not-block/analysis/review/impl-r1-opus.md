# a201 구현 적대 리뷰 r1 — Opus(작성자와 다른 문맥)

- 대상: `fix/a201-stale-quantity-comparison` HEAD `d31347ab8`, base `35d28a535`(base-commit.txt). 생산 diff: `reconcileloop.go` RunOnce(+6) · `mismatch.go` Observe/persist(+23/-2) · `reconcile_states.go`(1) · 새 `a201_fill_credit.go` · `a201_fill_applied.go`.
- 위험 등급 High-risk(대사 · 진입 차단 해제). 1회차(상한 2).
- codex 보이스: `impl-r1-codex.md`(원문). 아래 표는 두 보이스를 합쳐 중복을 정리한 것.
- 도구 판본: codegraph 1.6.0(호출자 확인은 rg 로 대신 — 아래 「보지 않은 범위」), go test 는 사본 worktree `a201-review-830200`(d31347ab8, 끝나고 제거).

## 결론

- **P0 · P1 없음.** design V8 · §4 공시 밖에서 「실제 차이가 남아 있는데 풀리는」 입력, 영구 · 계좌 범위 · 다른 원인 블록이 풀리는 입력, 같은 as-of · 이전 비교가 credit 을 쓰는 입력을 찾지 못했다. 모든 해제는 credit 보다 엄격히 나중인 비교가 (전체 비교가 진입을 막지 않고) 그 종목을 분쟁으로 보지 않을 때만 일어나며, 이 조건은 조정 credit 과 같은 무편집 분기(B8 · B10 · B29 · `creditUsableBy`)가 진다.
- **P2 1건(문서 ↔ 코드)**, P3 5건(시험 공백 3 · 증거 인용 1 · 미확인 기존 결함 후보 1).

## 발견

| # | 등급 | 한 줄 | 재현 입력 | 실제로 돌렸나 | 근거 | 제안 처분 |
|---|---|---|---|---|---|---|
| R1 | P2(문서↔코드) | 해제 근거 문구 · design V8 · spec 시나리오는 「로컬 읽기 **뒤** 커밋된 체결」을 말하지만 코드가 증명하는 것은 「로컬 읽기 **직전 표식 뒤** 커밋」 — 로컬 읽기와 겹치거나 그 앞(표식과 로컬 읽기 사이)에 커밋돼 로컬 읽기에 이미 든 체결도 credit 을 받는다 | 주기 k: 표식 100 → (표식과 `LocalStateFromJournal` 사이 또는 그 여러 읽기 도중) 엔진 SELL 체결 id 101 커밋 → 로컬 읽기가 그 체결을 반영 → 비교 S 불일치 → 수렴이 S 를 조정 못함(예: `Positions` 읽기 오류로 판 중단) → Observe 성공 → `101 > 100` 으로 체결 credit. k+1 일치 → `FILL_APPLIED`, 근거에 "a fill committed after the local read" 영속(codex C1) | 아니오(RunOnce 하네스에 표식 ~ 로컬 읽기 사이 훅이 없음). 성질 자체는 기존 `TestA201FillOutranSymbolsNamesOnlyMovedUnconvergedQuantitySymbols` 가 보임 — 로컬 읽기 없이 「표식 뒤 체결」만으로 부여 | reconcileloop.go:435-436 · a201_fill_credit.go:25-26(코드 주석은 「로컬 읽기와 겹치거나 그 뒤」라고 정확히 씀) · :65 · mismatch.go:952 · a201_fill_applied.go:3-4 · design V8(:39) · spec ADDED 첫 Scenario WHEN | 논리 무변. 문구만 코드에 맞춤 — 근거 문자열 · `ReconcileReleaseFillApplied` 주석 · design V8 · spec 시나리오를 「로컬 읽기 직전 표식 뒤(로컬 읽기와 겹치거나 그 뒤)에 커밋된 체결」로. spec Requirement 본문(「그 id 뒤로 커밋」)은 이미 정확. 근거 문자열을 고치면 persist 편집이라 FLM 번들 갱신 필요 |
| R2 | P3(시험 공백) | design §3-4 시험 `TestA201ALaterDisputeRefutesTheFillCredit` 은 「나중 분쟁이 체결 credit 을 반박」을 재지 않는다 — k+2 주기 수렴의 새 조정 credit 이 체결 credit 을 덮어써 반박 없이도 같은 결과 | 사본에서 B29 의 `disputed[symbol] ||` 를 `(disputed[symbol] && false) ||` 로(mismatch.go:853) | **예** — `go test -run TestA201 ./internal/app/engine/ ./internal/reconcile/`: engine 전부 PASS(그 시험 포함 `--- PASS: TestA201ALaterDisputeRefutesTheFillCredit`), reconcile 은 V4b 대조 `TestA201WithoutTheAuthorityErrorTheSameInputIsUnchanged` 만 FAIL(조정 credit 판 — AAPL 이 분쟁 뒤 cleared). 기존 시험 `-run 'Credit\|Refut\|Adjust\|A083' ./internal/reconcile/` 4 FAIL(`TestARefutedCreditIsDiscardedPerSymbol` 외). 원문 scratch `x2-no-dispute-refute.txt` · `x2b-preexisting.txt` | mismatch.go:853 · a201_fill_credit_cycles_test.go:298-327 | 공유 분기라 안전 공백은 아님(기존 조정 credit 시험이 지킴). 그 시험의 k+2 를 「계좌 5 · 로컬 5(외부 편입 유지) · 체결 · 조정 없음 → 해제 없음」으로 바꾸면 체결 credit 판으로 직접 잰다(codex C2 제안). 처분은 Manager |
| R3 | P3(시험 공백) | 「종목 워터마크 하나라도 못 읽으면 아무 종목에도 주지 않음」이 시험으로 고정되지 않음 | 사본에서 a201_fill_credit.go:63 `return nil` → `return out` | **예** — `-run TestA201` reconcile · engine 모두 ok(생존). 원문 scratch `x3-partial-watermark-failure.txt` | a201_fill_credit.go:61-64 · a201_fill_credit_test.go:97-120(단일 종목 · 닫힌 장부만) | 이 변이는 위험 방향이 아님(A 의 credit 은 A 자신의 판독으로 정당) — 계약이 「전부 아님」을 요구하는지부터 Manager 결정. 요구하면 두 종목 중 하나만 실패하는 픽스처(codex C3) |
| R4 | P3(시험 공백) | 「표식은 로컬 읽기 **바로 앞**」이 고정되지 않음 — 구조 시험은 `mark < local` 만 봄 | 사본에서 `MarkBeforeLocalRead` 를 `d.stabilise` 앞으로 이동 | **예** — `-run TestA201 ./internal/app/engine/` ok(구조 시험 포함 생존). 원문 scratch `x1-mark-before-stabilise.txt` | a201_fill_credit_structure_test.go:116-119 | 이 변이면 안정화 창(기본 2 초 + 수집 2 회) 동안 커밋돼 로컬 읽기에 이미 든 체결이 credit 을 받음 — V8 류를 조용히 넓힘. 구조 시험이 「표식 문장 바로 다음이 로컬 읽기」를 보면 닫힘(새 가드 아님 — 기존 구조 시험의 단언 강화). 처분 Manager |
| R5 | P3(증거 인용) | Observe FLM 의 B7 행이 「운영자 전용 · 다른 원인 블록은 체결 credit 으로도 안 풀림」의 근거로 `TestA201AConvergedSymbolGetsNoFillCredit` 를 인용 — 그 시험은 수렴된 종목이라 체결 credit 이 아예 서지 않아 「credit 이 선 채 B7 이 거르는」 입력에 닿지 않음 | — | 아니오(읽기) | function-logic/internal-reconcile--tracker.observe/function-logic-map.md B7 행 · a201_fill_credit_cycles_test.go:402-423 | 인용 정정(그 행은 기존 분기 · 「체결 credit 과 무관한 기존 시험」으로 적거나, 계좌 영구 + 종목 QM 이 함께 선 픽스처를 인용) |
| R6 | P3(미확인 · 기존 결함 후보) | 시장 범위(`scope_market`) QM 행 — 위험 버킷 ORPHAN_FILL(risk_bucket_owner.go:595-600)이 쓰는 행 — 은 Refresh 로 `ReleaseOnAdjustedReconcile` 블록이 되지만(mismatch.go:1430-1432), persist 의 해제 요청은 `ScopeMarket` 없이 나가 `scope_market IS NULL` 행만 찾음(reconcile_states.go:463-465). 그 종목에 쓸 credit 이 서면 해제가 「did not confirm」 오류로 끝나고 credit 은 남아(hasBlockFor 참) 깨끗한 주기마다 Observe 오류가 반복될 수 있음 | 종목 S 에 시장 범위 ORPHAN_FILL QM 행만 활성 + S 에 쓸 수 있는 credit(조정이든 체결이든) + 깨끗한 비교 | 아니오(추론). 트래커 자신의 NULL 범위 행이 함께 서는지에 따라 도달 여부가 갈림 — 확인 안 함 | mismatch.go:954-969 · reconcile_states.go:459-466 | a201 이 만든 결함이 아님(조정 credit 에도 같음). 다만 ORPHAN_FILL 행은 **체결이 만든 행**이라 a201 이후 그 체결 자체가 credit 근거가 될 수 있음 — 별도 확인 후보로만 기록 |

## 초점별 판정(반례를 못 찾은 것 — 근거)

1. **V8 밖 틀린 해제**: 해제 경로는 깨끗한 갈래 B6~B10 하나뿐(Observe FLM, 재번호 equal 29). 체결 credit 은 `t.adjusted` 에 같은 값으로 들어가 같은 분기로만 쓰임(구조 시험 `TestA201TheFillCreditRidesTheAdjustmentCreditsOwnBranches` 가 접근 함수 집합을 고정). 부여 주기의 비교는 그 종목 수량 불일치를 담아 차단 갈래로만 가므로 그 주기에는 해제가 없고, 스탬프가 그 주기 `diff.AsOf` 라 `creditUsableBy`(엄격히 나중)로 같은 비교에 못 씀. 영구 · 계좌 범위 블록은 B7(`Release != ReleaseOnAdjustedReconcile`)로, 부여는 `hasBlockFor`(QM ∧ adjusted)로 거름. 다른 원인 행은 종목당 활성 행 1(enterReconcileScopeInTx)이라 QM 블록과 같은 종목에 공존하지 않음.
2. **credit 수명**: Observe 조기 반환 둘 — 깨끗한 갈래 권위 오류는 V4b 반박 뒤 반환, 차단 갈래 권위 오류(@743)는 @728 반박 뒤. persist 오류는 조기 반환이 아니어서 끝 루프 B29 가 돌고, RunOnce 는 Observe 오류면 부여 전에 반환. Refresh 는 `adjusted`/`filled` 를 안 만짐 — 블록이 다른 길로 사라지면 다음 Observe 의 B29(`!hasBlockFor`)가 지움. Restore · Resolve 는 `adjusted = nil`. 곁 맵 `filled` 는 등식(`filled == adjusted != ""`)으로만 쓰이고 쓰기는 부여 메서드뿐 — 등식이 틀리는 길은 「같은 글자 as-of」(design §4 공시)와, 시계 역행으로 옛 스탬프가 새 스탬프를 덮는 경우(같은 §4 「벽시계 역행」)뿐.
3. **그 주기 Observe 무변**: 부여는 Observe 반환 뒤 · 별도 잠금 구간. V4b 는 `if err != nil` 몸 안 — 오류 없는 경로는 구성상 무변. 건강 계수(`note(cycle)`)는 부여가 `cycle` 을 안 만지므로 무변.
4. **a192 결속 행**: 부여는 그 주기 수량 불일치(V2 (a))가 필요하고, 같은 주기 `beginOrderTerminalObservation`/`updateOrderTerminalBindings` 가 그 종목 결속을 `tainted` 로 만듦 → a192 가 풀 수 있었던 행을 가로채지 않음. credit 이 선 뒤 같은 행에 새 결속이 생기려면 삽입 증명이 필요한데 행이 활성이라 삽입이 없음.
5. **RunOnce 순서**: 표식은 로컬 읽기 바로 앞(R1 · R4 는 문구 · 고정 문제). `cycle.Err` 가 수렴 · 편입 오류로 선 주기에도 부여함 — design V2 (c) 「판 중단 모두 포함」과 일치. 「하나라도 못 읽으면 전부 안 줌」은 보수 방향이라 틀린 해제를 못 만듦(대가는 고착 — FillWatermark 실패는 DB 수준이라 종목 하나만 실패하는 입력은 비현실적).
6. **동시성**: 부여 메서드 · Observe · Refresh · AdjustmentApplied 모두 `t.mu` 아래. Observe 와 부여 사이에 끼는 것은 Refresh(진입 게이트 권위 갱신 · 체결 훅)뿐이고 credit 맵을 안 만짐. `AdjustmentApplied` 생산 호출자는 converge.go:189 하나(RunOnce 안, 순차). `-race` 로 a201 시험 3 패키지 실행 PASS(scratch `r0-baseline-race.txt`) — 단 이 시험들은 단일 고루틴이라 인터리빙은 실행으로 재지 않음.
7. **시험이 재는가**: M14 — 행동 고정 시험은 오류 없는 경로의 「보류 사유 · a192 후보 미진입」을 재며 실제로 갈래 차이를 잡음(codex 도 동의). M15 — §4 공시된 같은 글자 as-of 원인 이름만 잼. 원장의 CAUGHT 가 다른 가드 덕인 사례는 R2(§3-4 시험) 외에 찾지 못함. 엔진 시험은 실제 `RunOnce` 를 씀(reconcile 단위 시험은 함수 단위라고 스스로 밝힘).
8. **문서 ↔ 코드**: R1 · R5. 그 밖에 design §5 편집 표면(함수 넷 · `Restore` 무편집)과 diff 일치.
9. **안전 불변식**: 청산 경로 무관(`ExitAllowed` 무조건 참, 해제는 진입 쪽). 새 로그 0, 근거 문자열에 계좌 · 수량 · 가격 없음. 이 change 에 토글 없음(진입 차단 해제 규칙 자체의 변경) — 토글 OFF 불변식 not-applicable.

## 돌려서 확인한 것 / 추론만 한 것

- 돌림(사본 `a201-review-830200`, 판마다 `git checkout --` 원복 · `git diff --quiet` · hash-object == HEAD 확인):
  - r0 무변이 `-race -run TestA201` reconcile · engine · journal: 셋 다 ok.
  - X1(R4) 생존 · X2(R2) a201 엔진 생존 / reconcile V4b 대조 1 FAIL / 기존 4 FAIL · X3(R3) 생존.
  - 원문: `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/6ba8fcd4-b024-437e-a4c2-ebe621826fe5/scratchpad/{r0-baseline-race,x1-mark-before-stabilise,x2-no-dispute-refute,x2b-preexisting,x3-partial-watermark-failure}.txt`(세션 스크래치 — 지워질 수 있음).
- 추론만: R1(코드 + 기존 시험 구조), R5, R6, 초점 1~6 · 9 의 「반례 없음」 판정.

## 보지 않은 범위

- `codegraph affected --filter '*_test.go'` · codegraph 호출자 질의를 돌리지 않음 — 호출자는 `rg`(`\.(AdjustmentApplied|Observe|Resolve|Restore|Refresh|FillsOutranAdjustment)\(`)로만 셈.
- `tools/logic-map` 재실행 안 함 — 분기 근거는 change 의 post-edit 번들(Observe · RunOnce)을 씀.
- 수렴의 「스냅숏보다 장부가 앞선」 경합(표식 전 체결이 로컬에만 든 경우 수렴이 계좌 값으로 되돌리는 문제)은 a201 이전부터의 성질로 보고 따로 재지 않음.
- 위험 버킷 broker-zero 경로는 a201 시험(`TestA201BrokerZeroRefusesAFillAppliedRelease`)을 읽기만 함.
- 실행 인터리빙(동시성)은 실행으로 재지 않음.

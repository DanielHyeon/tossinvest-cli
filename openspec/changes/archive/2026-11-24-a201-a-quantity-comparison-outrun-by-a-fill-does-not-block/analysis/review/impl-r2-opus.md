# a201 구현 적대 리뷰 r2(마지막 회차) — Opus

- 범위: `d31347ab8..7621c5c2e` 델타만(r1 처분 + gstack r1 처분). 생산 논리 변화는 V4c 하나(`RunOnce` Refresh 오류 갈래 return 직전 `RefuteCreditsDisputedBy(diff)` · 새 메서드 `Tracker.RefuteCreditsDisputedBy`). 나머지는 시험(R2 · R4 · R5 · G2 · G3 · V4c 시험 파일) · 주석 · R1 근거 문자열.
- codex 보이스: `impl-r2-codex.md`(원문). 아래 표는 합친 것.
- 실험 사본: `/home/daniel/tossos-work/a201-review2-1160102`(7621c5c2e, 끝나고 제거). 하네스 `run_mut_r2.sh`(판마다 HEAD sha 단언 · `git checkout --` 원복 · `git diff --quiet` · hash-object == HEAD 단언).

## 결론

- **P0 · P1 · P2 없음(Opus).** V4c 가 지우는 집합은 같은 `diff` 로 Observe 가 지웠을 집합의 부분집합이다(아래 초점 1). V4c 가 닫지 못한 조기 반환은 AST 로 열거해 없음을 확인(초점 2). 고친 시험은 변이로 실제로 가드를 잰다(초점 5).
- codex 도 P0~P2 0(원문 `impl-r2-codex.md`). 합친 P3: S1 · S2(Opus), S3 · S4(codex), 그리고 M21 진단 차이(초점 4 — 두 보이스 같은 결론, 별도 행 없음).

## 발견

| # | 등급 | 한 줄 | 재현 입력 | 돌렸나 | 근거 | 제안 처분 |
|---|---|---|---|---|---|---|
| S1 | P3(문서 낡음) | `tasks.md` 3.1 과 `review.md` 「구현 적대 리뷰 r1 처분」 끝 줄이 「리뷰 2회차: not-applicable — 생산 논리 무변」이라고 적혀 있으나, 그 뒤 gstack G1 처분(V4c)이 생산 논리를 바꿨고 지금 2회차가 돌았음 | — | 아니오(읽기) | tasks.md 3.1 · review.md 「리뷰 2회차: not-applicable」 줄 | 2회차 원문(`impl-r2-*.md`)을 가리키게 갱신 |
| S3 | P3(문서 · 공시 폭, codex) | design §2 · V4b · V4c 는 반박 확대를 「자동 해제가 줄어드는 보수 방향」이라고만 적음. 블록 유지는 진입뿐 아니라 그 종목 편입 차단과 청산의 RECONCILE 확정 하한(하한 조회 실패 시 0 주)도 길게 만듦 — 안전 불변식 4(청산 즉시성)를 읽는 사람이 판단할 수 있게 공시가 좁음. 틀린 결과는 아님(분쟁을 본 비교에서만 유지) | Refresh 실패 주기 k+1 에 S 외부 포지션(분쟁) + S credit → V4c 반박 → 이후 S 블록 유지 동안 S 청산은 확정 하한으로 깎임 | 아니오(읽기) | exitwiring.go:192-200 · exitloop.go:1766-1775 · design §2 | design §2 「둘(V4b · V4c)」 문장에 「블록 유지 = 청산 확정 하한 적용도 유지」 한 구절(선택) |
| S4 | P3(처분 근거, codex) | R3 처분 근거 「둘 다 design V2 안」은 같은 처분이 design V2 에 넣은 「구현은 주기 단위 전부-아니면-0」 문장과 모순 — 그 문장 뒤에는 `return out` 변이가 V2 위반. 코드는 계약대로(`return nil`) | 종목 A 판독 성공 · B 판독 실패 → 변이면 A 에만 credit | 아니오(r1 X3 생존 원문 재인용) | review.md R3 행 · ledger.md M16 · design.md:12 | 처분 근거 문구만 「시험 공백 수용(구성에 seam 필요)」으로 |
| S2 | P3(문구) | R1 문구 정정이 spec ADDED 첫 시나리오의 **제목**에는 안 닿음 — 제목은 여전히 「로컬 읽기 뒤 체결이 앞지른 수량 불일치」, WHEN 본문은 「로컬 읽기 직전 … 기록한 뒤(로컬 읽기 도중 … 포함)」로 고쳐짐. 근거 문자열 · 상수 주석 · design V1/V8 · 파일 머리 주석은 일관 | — | 아니오(rg) | specs/reconciliation/spec.md:7 | 제목만 맞춤(선택) |

## 초점별 판정

1. **V4c 가 지우면 안 되는 credit 을 지우는가 — 아니오.**
   - 지우는 조건은 `disputed[종목] ∧ creditUsableBy(credit, diff.AsOf)`(무편집 `refuteUsableAdjustmentCredits`). 같은 주기 수렴이 준 조정 credit 은 스탬프가 같은 `diff.AsOf` 라 「엄격히 나중」이 아니어서 남음 — **변이 Y4(as-of 무시하고 분쟁 종목 credit 전부 삭제)를 행동 시험 `TestA201ARefreshFailureDoesNotRefuteTheCreditItsOwnConvergenceGave` 가 잡음**(구조 시험 둘도 같이 FAIL). 분쟁 아닌 종목은 안 건드림(대조 시험 `…KeepsACreditItsComparisonDoesNotDispute`). `diff.AsOf` 가 비거나 파싱 불가면 `creditUsableBy` 가 거짓이라 아무것도 안 지움 — Observe 와 같은 동작.
   - Refresh 의 오류 반환은 하나(`ActiveReconcileStates` 실패, mismatch.go:1046-1050)이고 그 앞에 추적기 상태 변경이 없음 — 부분 변경 뒤 실패는 없다.
   - 같은 `diff` 로 Observe 가 돌았다면: 차단 갈래 · 깨끗한 갈래 모두 끝 루프 B29 가 `disputed ∧ usable` credit 을 지우고(persist 오류여도 끝 루프는 돔), 권위 오류 반환 둘은 각각 V4b · @728 반박 뒤. 그래서 V4c 삭제 집합 ⊆ Observe 삭제 집합 — 새로 잃는 credit 은 「Observe 가 돌았어도 잃었을 것」뿐.
   - 귀결: 그 종목 블록이 다음 조정 · 체결 credit 또는 운영자 해제까지 남음. `cycle` 은 이미 `Err` 이고 V4c 는 `cycle` 을 안 만져 그 주기 건강 계수 무변, 그 주기 편입 판정은 편집 전에도 이 갈래에서 건너뜀. **청산과 무관하지 않음(codex 지적, S3 — r1 초점 9 의 「청산 경로 무관」 정정)**: 블록이 남는 동안 그 종목 편입은 막히고, 관리 포지션 청산은 RECONCILE 확정 하한으로 깎이며 하한을 못 구하면 0 주(exitwiring.go:192-200 · exitloop.go:1766-1775). 다만 블록을 남기는 것은 그 비교가 그 종목을 분쟁으로 봤을 때뿐 — Observe 가 돌았어도 같은 결과(위 부분집합)라 틀린 결과는 아님.
2. **V4c 가 지워야 하는데 못 지우는 자리 — 없음.** RunOnce post-edit 번들(ast sha `a47a346fb26f` = HEAD 파일 sha 앞 12자 일치)의 반환 다섯: @435 안정화 실패(비교 전 — `diff` 없음), @444 로컬 읽기 오류(`diff` 는 @446 에서 처음 계산 — 없음), @471 Refresh 오류(V4c), @483 Observe 오류(Observe 안에서 반박 완료 — 아래), @491 정상. Observe 번들(sha `0369ee0e88fc` = HEAD 일치)의 반환 셋: @669(V4b 반박 @668 뒤), @743(같은 if 몸 안 @728 반박 뒤), @867(끝 루프 B29 뒤). 셋째 조기 반환은 없음. 수렴 · 편입 오류는 반환하지 않고 Refresh · Observe 로 이어짐.
3. **동시성 — 문제 없음.** 새 메서드는 `t.mu` 를 잡고 그 안에서 `symbolsInDispute(diff)`(인자 `diff` 만 읽는 순수 함수)와 반박을 함. Refresh 는 오류 반환 전에 잠금을 풂(mismatch.go:1048) — 교착 없음. 사이에 끼는 Refresh(게이트 권위 갱신)는 `adjusted` 를 안 만짐.
4. **M21(호출을 Refresh 앞 무조건으로) — 해제 결과는 같고 갈래만 다름.** 성공 경로에서 앞당긴 반박은 B29 가 같은 주기에 지울 credit 을 먼저 지울 뿐. 다른 점은 깨끗한 갈래에서 credit 있던 분쟁(외부 포지션) 종목이 B10(Awaiting) 대신 B8 → `orderTerminalCandidate` 로 가는 것: 그 판정은 `cmp.disputed` 에서 `symbol_disputed` 로 보류(order_terminal_release.go:245-246)하므로 해제는 없고, 차이는 결속의 보류 사유 기록 · 증명 읽기 1회 · `row_changed` 판정 시 결속 삭제(이미 쓸모없는 결속). M14 와 같은 종류 — 행동 시험으로 고정할 틀린 결과 없음(추론). codex 가 같은 차이를 구체 입력(k 결속 · k+1 외부 매도 수렴 credit · 사이 취소 + 외부 매수 1 · k+2 외부 포지션)으로 적음 — 원장 M21 의 「행동 결과 무변」은 「차단 · 해제 결과 무변」으로 좁혀 읽을 것.
5. **고친 시험이 이제 재는가 — 예(변이로 확인).**
   - R2: Y2(B29 `disputed` 삭제 무력화) → `TestA201ALaterDisputeRefutesTheFillCredit` FAIL(r1 에서는 PASS 였음).
   - R4: Y1(표식을 `stabilise` 앞으로) → `TestA201TheFillCreditIsGrantedAtOneSiteAfterASuccessfulObserve` FAIL.
   - V4c: Y3(호출 삭제) → 체결 판 · 조정 판 행동 시험 둘 + 구조 시험 FAIL. Y4 는 위 1.
   - G2: Y5(V2 (a) 를 missing-order 종목까지 넓힘) → `TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit` FAIL, 실패 원문 `rows for SYN201 = ["QUANTITY_MISMATCH|released:FILL_APPLIED"], want … ORDER_TERMINAL_CONFIRMED` — 고친 G2 가 원래 주장(수량 흔적 없는 결속 행은 종결 증명 경로로만)을 여전히 잼. 단위 시험 `TestA201FillOutranSymbols…` 도 FAIL.
   - R5 · G3 은 원장 M19 · M03b 를 읽기만 함(다시 안 돌림).
6. **문서 ↔ 코드**: design V4c 행 · 대응표 「반박」 · §2 「둘(V4b · V4c)」 · §5 「RunOnce 3자리」, spec 새 Scenario(「엄격히 나중 as-of … Refresh 실패」) · MODIFIED 한 문장(「엄격히 later · disputed · 반환 전 만료」) 모두 코드와 같은 폭. R1 문구는 근거 문자열(mismatch.go:952) · 상수 주석 · design V1/V8 · 파일 머리 · spec WHEN 일관, 예외는 S2(시나리오 제목).
7. **r1 처분 재검토**: R3 수용 — 시험을 안 더하는 결정에는 동의, 근거 문구는 S4. R6 「도달 불가」 — 동의: `releaseRiskBucketOwner`(risk_bucket_owner.go:814, `released_at` 쓰는 유일한 자리 :1017)와 `recordRiskBucketBrokerZeroObservation`(:228) 모두 생산 호출자 0(`rg --glob '!*_test.go'` 로 재확인). 위험 버킷 owner 해제를 생산에 배선하는 change 가 생기면 그 전제로 다시 볼 것 — verdict.md 의 단서 그대로.

## 돌린 것 / 추론만 한 것

- 돌림(사본, `-run TestA201`, engine · reconcile):
  - 무변이 대조군: 둘 다 ok, PASS 40 — `r2-control.txt`.
  - Y1 · Y2 · Y3 · Y4 · Y5 모두 CAUGHT(위 5) — `r2-Y1-mark-before-stabilise.txt` · `r2-Y2-b29-no-disputed.txt` · `r2-Y3-v4c-deleted.txt` · `r2-Y4-v4c-any-asof.txt` · `r2-Y5-a-missing-orders.txt`.
  - 원문 위치: `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/6ba8fcd4-b024-437e-a4c2-ebe621826fe5/scratchpad/`(세션 스크래치 — 지워질 수 있음).
- 추론만: 초점 1 의 부분집합 논증, 초점 3, 초점 4(M21), R6 재확인(호출자 rg), S1 ~ S4.

## 보지 않은 범위

- `-race` 는 이번 회차에 다시 돌리지 않음(팀메이트 원문 `analysis/verify/race-gstack-r1.txt` 를 읽지 않음).
- M21 의 「갈래만 다름」을 엔진 시험으로 재현하지 않음.
- G3 · R5 시험은 변이로 다시 재지 않음(원장 인용).
- 시험 · 주석 외 델타(분석 원문 · verify 파일)는 내용 검증 안 함.

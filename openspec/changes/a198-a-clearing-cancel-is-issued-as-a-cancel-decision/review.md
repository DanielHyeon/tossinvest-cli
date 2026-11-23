# a198-a-clearing-cancel-is-issued-as-a-cancel-decision Review

## 계약 작성 (2026-11-23)
- 작성: Manager(Fable). 발견 경위: a197 1단계 probe(브랜치 `feat/a197-external-opposite-buy-cancel`) — 실제 게이트웨이로 길 치우기의 취소가 원장에서 거절됨을 재현. 분기 주장의 AST 는 그 조사 산출물 인용이며 change 안 사본·HEAD 재확인은 tasks 1.3.
- `openspec validate a198-a-clearing-cancel-is-issued-as-a-cancel-decision --strict`: valid (rc 0, 2026-11-23 HEAD 064f59fd9).
- proposal-freeze 리뷰: 미실시(tasks 1.7~1.8).
- Function Logic Map: 편집 대상 기존 함수의 번들은 tasks 1.6 에서 만든다.

## 1단계 계약과 증거 (2026-11-23, 팀메이트 Opus)
- base `e1eaf720c`(1.1). CodeGraph 1.6.0 hard evidence·CodeGraphContext not-applicable·증거 조정(1.2) — `analysis/code-context/`.
- 주장 대조(1.3) — `analysis/claim-check.md`, AST 36 개 `analysis/cited-ast/`, probe 재실행 `analysis/probe/`(a197 원본 바이트 동일 + (b) 확장, 버린 worktree 미커밋 파일). 결정을 바꾸지 않는 정정 1: proposal 「Why」 표의 「키만 없애서는…」 줄(결정 행에 종류 필드 없음 · 매도 주문은 키만 없으면 방향 비교 통과 · checkIdempotencyKey 는 키가 있을 때만).
- 편집 표면(1.4) `analysis/edit-surface.md`, rate limit(1.5) `analysis/rate-limit.md`, pre-edit FLM/BTM 번들 2(1.6) `analysis/function-logic/`(clearTheSymbol · Context.ExitObserver). `check_analysis.py --change …` rc 0(required 0).
- `openspec validate a198-a-clearing-cancel-is-issued-as-a-cancel-decision --strict`: valid (rc 0, 2026-11-23, 이 브랜치 정정 뒤).
- Manager 결정 필요(문서 미편집): ① D4 `warnEpisode` 는 운영 `opts.Log` nil 로 무기록 ② D2 안 — (가-B) 관측기 내부 `execgw.Issuer` 구성 권고(편집 함수 1~2, 하네스 0), 단 D3·조립 Scenario 공집합화 ③ a087 current 번들이 `Context.ExitObserver` sha 를 고정 — 착지 순서 ④ 취소 사전 GET 과 정본 「브로커 조회를 더하지 않는다」 ⑤ 접수된 매수 취소 뒤 재취소(§2-1).

## proposal-freeze 리뷰 r1 (2026-11-23)
- 보이스: codex-cli 0.165.0 `codex exec -s read-only </dev/null`(다른 계열, 적대적 Eng). 원문 `analysis/review/proposal-freeze-r1-codex.md`. 판정 「freeze 불가 — P1 4, P2 2, P0 미입증」.

| # | 발견 | 대조 | 처분 |
|---|---|---|---|
| P1-1 | D4 대로면 운영에서 원인 기록 0(opts.Log nil); 전체 Log 연결은 계좌 노출 위험 | 1.3 claim-check D4 와 같은 결론, 코드 재확인 | **Manager 결정 대기**(싱크 선택) |
| P1-2 | CANCEL 이 ACKED 에 남으면(MarkAcked dispatch.go:181 → Settle :199 사이 crash·쓰기 실패) 기동 복구(acked_boot.go:51 「CANCEL/AMEND … stays ACKED」)·세션 해소(PLACE 만)·운영자 해소(resolution.go:143 UNRESOLVED_IN_DOUBT 만) 어느 것도 못 닫고 clearTheSymbol B3 이 손절을 막음, 계수 제외(onlyEngineCancelsInFlight B5) | 코드 재확인 — 사실. 1단계 산출물에 없던 발견 | **Manager 결정 대기**(복구 계약 = 새 범위) |
| P1-3 | 전일 주문의 당일 취소: ConfirmedCancelOf 는 주문 거래일, 게이트웨이는 취소 거래일 기록 → 확정 취소 재전송·withPending=false 시 대기 방벽 누락 | claim-check (c) 말미와 같은 사실 | **Manager 결정 대기**(결속 계약 = 새 가드) |
| P1-4 | 취소마다 사전 GET(service.go:218 → precheck.go:98) — 정본 「조회 0건」과 문언 충돌, 예산·지연 계상 부족 | rate-limit.md 와 같은 사실 | **Manager 결정 대기**(정본 해석) |
| P2-5 | D3 조립 실패는 엔진 런타임 전체(무충돌 손절 포함)를 막음 — design §3 「새 거절 없음」에 빠짐 | engine.go:684 반환 확인 | **Manager 결정 대기**(D2 안과 함께) |
| P2-6 | 결정 종류는 원장에 없어 §4 「종류 CANCEL 을 원장으로 단언」 불가(AMEND 변이 못 잡음); 결정은 대상 주문번호를 결속하지 않음 | claim-check P5 와 같은 사실 | **Manager 결정 대기**(시험 전략 변경) |

- 사실 오류 정정 반영: 없음(모든 발견이 결정·범위·시험 전략 사안). 리뷰가 제안한 시험 표(원문 하단)는 미반영 — 새 가드·범위 확장 금지 규칙.
- r2: 미실시 — r1 발견이 전부 Manager 결정 대기라 반영한 수정이 없음(재리뷰할 변경 없음).

## Manager 결정 반영 (2026-11-23)
- r1 처분 갱신(위 표의 「Manager 결정 대기」를 대체): P1-1 → D4 를 a129 운영 로그 경로(`noteExitOrder`→`ExitOrderLog`)로 변경 · P1-2 → 범위 밖·이름 붙인 잔여(proposal·design §6·Story limits 공시, 시험 6 ACKED 변형) · P1-3 → 범위 밖 잔여(design §6) · P1-4 → 정본 범위 안으로 해석, 비용 공시(design §2-4) · P2-5 → D2 (가-B)·D3 삭제·조립 Scenario 삭제 · P2-6 → §4 단언·변이 수정. 매수 재취소 → §2-5·§6 공시, a087 → 기다리지 않음(§6).
- 문서: proposal(What Changes 1·2·3, 하지 않는 것, Impact) · design(D2·D3·D4, §2-4·§2-5, §3, §4-2·4-4~4-7·4-9·변이, §6) · spec delta(첫 Requirement 끝 문장·조립 Scenario 삭제, 셋째 Requirement 「운영 로그에 계좌 정보 없이」) · tasks 3.2·3.3 · Story acceptance 2문장·limits 1문장. tracker 재생성 `--check` rc 0.
- 편집 대상 기존 함수: `ExitObserver.clearTheSymbol` · `exitOrderLine.args`(새 pre-edit 번들). `Context.ExitObserver` 번들 삭제(무편집). `check_analysis.py` 「evidence complete or diff-proven exempt」.
- 선행 확인(claim-check 「D2 선행 확인」·「D4 경로 실현 가능성」): 운영 시계 동일 · TTL 값 60s 동일(상수는 별개) · ID 접두만 다름 · D4 경로 기존 함수 편집 1(`args`).
- `openspec validate a198-a-clearing-cancel-is-issued-as-a-cancel-decision --strict`: valid (rc 0, 2026-11-23, 반영 뒤). ADDED 본문 길이 206·239·259자.

## proposal-freeze 리뷰 r2 (2026-11-23, 마지막 회차)
- 보이스: codex-cli 0.165.0 read-only `</dev/null`. 원문 `analysis/review/proposal-freeze-r2-codex.md`. 판정 「freeze 불가 — P1 2, P2 3, P0 없음」. 반박 없음으로 확인된 처분: (가-B) 시계 동일, TTL 60s, D3 삭제, D4 계좌 비노출 방향, P2-6 단언.

| # | 발견 | 대조 | 처분 |
|---|---|---|---|
| r2-1 P1 | `noteExitOrder` 래치 서명(exit_order_log.go:203 `event·action·level·reason·state`, 키 positionID)에 대상 주문이 없어, 같은 포지션의 두 취소가 같은 사유로 실패하면 같은 주기 안에서 둘째가 버려지고, 사유가 다르면 서로 덮어 주기마다 반복 → 「무기록 금지」와 「연속 한 번」을 동시에 못 지킴. 「args 1개면 충분」 불성립 | 코드 재확인 — 사실 | **Manager 결정 대기**(D4 채택 조건 재판단 — `noteExitOrder` 편집 또는 래치 미사용·주기마다 줄, 어느 쪽이든 결정 변경). design D4 의 「같은 결과가 연속 주기에 이어지면 한 번」 문장은 고치지 않음. 수량·가격 **키 부재**는 args 안의 취소 분기로 가능 — D4 문장에 명시(사실 보강) |
| r2-2 P1 | 사전 조회 2s 는 반환 상한이 아님(client.go:336 → token.go:63 끊기지 않는 `mu.Lock()`), 401 재전송(client.go:359) — `N×17s` 는 최악 지연이 아님 | 코드 재확인 — 사실 | **사실 정정 반영**: design §2-4·proposal Impact 를 「HTTP 시간 예산, 전체 최악 지연 미확정」으로, rate-limit.md 를 논리/물리 호출 구분으로. 토큰 경로 수리는 범위 밖 |
| r2-3 P2 | ACKED 계수 제외는 조건부(onlyEngineCancelsInFlight), 지연 경보는 다음 관측 주기에서만 | 코드 재확인 — 사실 | **사실 정정 반영**: design §6·proposal·Story 에 조건 괄호, §4-6 에 「전체 관측 주기 반복으로 단언」 |
| r2-4 P2 | 전일 결속은 「알려진 정본 미충족 잔여」로 · 문서 간 불일치(rate-limit·edit-surface·review·tasks 5.3 대 Story limits 「운영 반영 제외」) | 사실 | design §6 문구·rate-limit·edit-surface·review(이 절) 반영. **tasks 5.3(사람 승인 운영 반영) 대 Story limits(운영 반영 포함 안 함)는 Manager 작성 문장 간 충돌 — 미편집, Manager 결정 대기** |
| r2-5 P2 | r1 파일의 프롬프트 블록이 a196 리뷰 문안 | 사실 — 공유 스크래치 경로를 병행 작업이 덮어씀 | r1 파일에 「정정」 절로 실제 프롬프트 덧붙임(원본 보존) |

## r2 처분 확정과 freeze (Manager, 2026-11-23)
- r2-1 → **D4 재결정**: a129 대기열·래치·줄 타입을 쓰지 않고 새 파일의 새 함수가 `ExitOrderLog` 에 직접 한 줄, 자기 중복 억제(key = 포지션·대상 주문·상태·사유 코드; 접수 확정 시 그 주문 표시 삭제). 편집 기존 함수 = `clearTheSymbol` 하나. `exitOrderLine.args` 번들 삭제. design D4·§4-4·4-5·4-7·4-9·변이·§6(a087 — exit_order_log.go 충돌 소멸), proposal Impact·What Changes 2, tasks 3.2·3.3·4.1 반영.
- r2-2·r2-3·r2-4·r2-5 사실 정정 반영: 승인.
- tasks 5.3 ↔ Story limits 충돌 → 5.3 을 체크박스에서 빼 「## 6. 종결 뒤 (사람 승인 — 이 change 의 종결 조건 아님)」로 이동(a195 와 같은 처리). Story 무변경.
- **Manager 1단계 검증·freeze 2026-11-23** (커밋 `94dec8ae4` 읽고 strict validate 재확인 뒤). tasks 1.8 [x].

## Pre-Edit 선언 (tasks 2.1, 2026-11-23 — High-risk, 면제 없음)

```text
Pre-Edit Gate:
- change id / task id: a198-a-clearing-cancel-is-issued-as-a-cancel-decision / 2.1 → 3.1·3.2
- 대상 심볼(패키지.함수): engine.ExitObserver.clearTheSymbol(기존, 본문 편집 — 유일). 새 코드: 새 파일의 engine.(*ExitObserver).noteClearCancelFailure, 관측기 구조체 새 필드 1(중복 억제 맵). 무편집: submit·record·NewExitObserver·Context.ExitObserver·a129 함수 4·internal/execgw·internal/journal production.
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (codegraph 1.6.0, base e1eaf720c) — clearTheSymbol 호출자 record 1 자리(exitloop.go:1395); 편집 자리 :1684(발행)·:1696(B9)·:1717(B11).
- CodeGraphContext 후보와 evidence reconciliation: not-applicable(미색인 — codegraphcontext-context.md), evidence-reconciliation.md 미해소 불일치 0.
- 기존 동작 파악 근거: HEAD AST analysis/cited-ast/ExitObserver_clearTheSymbol.ast.json(분기 17), probe 재실행 analysis/probe/(취소가 원장에서 거절, flatten 모양은 접수), 기존 시험 a094_exit_clear*·exitloop_test(가짜 제출자)·exit_e2e(진짜 게이트웨이, 취소 미도달), 호출부 1.
- Function Logic Map / Branch Test Map: analysis/function-logic/internal-app-engine--exitobserver.clearthesymbol/ (pre-edit; 편집 뒤 재생성·재번호).
- upstream 상속 테스트 영향: no — internal/trading·official 무편집.
- 실패 테스트 선행 작성: yes — internal/app/engine/a198_clearing_cancel_test.go RED 를 HEAD 코드에서 먼저 실행(analysis/tdd/red.txt).
- 설정·DB·journal 변경과 rollback: 없음 — 스키마·설정·토글 0. 새로 쓰이는 행은 flatten 이 이미 쓰는 CANCEL attempt·위험 감소 결정 형식. rollback = 이전 이미지.
- 안전 불변식 §0 위반 여부 검토: 통과 — LIVE 주문 0(가짜 브로커 httptest), 손절 즉시성 약화 없음(판정 무변화, 막혀 있던 취소가 나가 손절이 나갈 수 있게 됨), 새 API 호출 종류 없음(retry matrix 행 31·33·37), 운영 토글 무변화, 로그에 계좌·수량·가격·Detail 없음.
```

## 2단계 RED → GREEN → 검증 (tasks 3.1~4.2, 2026-11-23, 팀메이트 Opus)
- 커밋: RED `459287086`(시험 + red.txt + Pre-Edit) · GREEN `09bb9e8c1`(production 2 파일 + 번들). push 없음.
- 편집 기존 함수: `ExitObserver.clearTheSymbol` 하나(수량·가격 해석을 발행 앞으로 이동 — 결정에 최대 수량을 싣기 위함, 해석 실패 시 결정 행을 안 남기는 것 외 판정 동일; 발행을 관측기 안 `execgw.Issuer` CANCEL 로; B10·B11 실패 기록 호출; 접수 확정 시 표시 삭제). 구조체 필드 1(`clearCancelLogged`). 새 파일 `a198_clear_cancel_log.go`(새 함수 2). submit·record·NewExitObserver·Context.ExitObserver·a129 함수·execgw·journal production 무편집. `check_analysis` required 1 = clearTheSymbol(번들 post-edit, 옛 B9↔B10 재번호).
- RED(analysis/tdd/red.txt): HEAD production 에서 A198 7 중 6 FAIL — 브로커 취소 0 · 원장 CANCEL attempt 0 · 실패 줄 0. ACKED 특성 시험은 HEAD 에서도 PASS(기존 규칙).
- 결합 시험이 잰 것(GREEN): 매수 미체결 → 브로커 취소 1(W-BUY-1), 원장 CANCEL attempt 1 = `CONFIRMED|W-BUY-1|key=false|…"side":"BUY"`, 같은 주기 손절 SELL 10 1건, 손절 결정은 멱등 키 있음. 무장된 매도 → 취소 1(SELL, 키 없음), 그 주기·다음 주기 손절 0·취소 여전히 1, 종결 기록(CANCELLED, 무체결) 뒤 주기 손절 10. 거절 400 → 줄 1(order.rejected · FAILED_CONFIRMED · dispatch_rejected_by_broker), 반복 1, 500 으로 바뀌면 2(order.in_doubt · IN_DOUBT · dispatch_outcome_unknown). 발행 실패(원장 트리거) → 취소 0 · 줄 1(cancel_decision_failed, 상태 빈 값). IN_DOUBT → 취소 1 · 손절 0 · 지연 경보·연속 실패 경보 둘 다. ACKED 고착 → 취소 0 · 손절 0 · **지연 경보 나감** · 연속 실패 경보 없음. 운영 조립(Context.ExitObserver, ExitOrderLog non-nil) → 취소 1 · 손절 1. design §2 의 기존 규칙과 어긋남 관측 없음.
- 4.1 변이: analysis/mutation/ledger.md — 8 중 7 CAUGHT(A198), M7(매도 취소 접수를 치움으로)은 A198 만으로 생존(뒤의 ReleaseClearedExitProposal 이 대신 막음), A198+A094 로 CAUGHT(A094 기존 시험 2, 가짜 제출자).
- 4.2 선택 시험(원문 analysis/verify/):
  - t1 `go test ./internal/app/engine -run 'A198|A094|ExitClear|ExitObserver' -count=1`(GOFLAGS=-trimpath) ok 53.5s
  - t2 같은 패키지 `-run 'A087|A129|A163|A092|A090|A091|A150|A159|Exit|Clear|Observer|Exits'`(-trimpath) → FAIL 5: 전부 runtime.Caller·패키지 경로로 소스를 여는 구조 시험이 -trimpath 에서 「no such file」(a087 ×4, a111 ×1). t2b 같은 묶음 GOFLAGS 없이 → ok 260.5s(HEAD 09bb9e8c1). → -trimpath 는 이 패키지의 구조 시험과 충돌(환경 결함, 이 change 무관)
  - t3 `./internal/execgw -run 'Cancel|RiskGuardian|Issue'` ok · t4 `./internal/journal -run 'Prepare|Decision'` ok · t5 `-race ./internal/app/engine -run A198` ok
  - `make lint` rc 0 · `openspec validate --strict` valid · `check_analysis.py` 「evidence complete」 rc 0
- 미실행: 전체 스위트 · `make gate` · `make sdd-sync` · `make sdd-check` · `make test-seams` · 패키지 전체(internal/app/engine 의 선택 밖 시험) · cmd/tossctl 시험 · post-edit 커버리지 재측정 · 4.3 리뷰(Manager 가 붙임).
- 남은 위험: 사전 단건 조회 GET 의 호출 수를 결합 시험이 세지 않음(가짜 브로커 되읽기 미계수) · 실패 줄 쓰기가 루프 안 동기(쓰기 대상이 막히면 같은 주기 뒤 포지션 판정이 늦어질 수 있음 — a129 는 주기 끝 대기열로 피함) · `clearCancelLogged` 는 접수 확정 때만 지워져, 종결로 사라진 주문의 항목은 프로세스 수명 동안 남음(주문 수만큼, 작음) · M7 의 실제-게이트웨이 시험 부재 · design §6 의 잔여(ACKED/IN_DOUBT 보류, 전일 결속, 매수 재취소).

## 구현 적대 리뷰 r1 (2026-11-24)
- 보이스: 별도 컨텍스트 Opus + codex 둘째 보이스(Manager 가 붙임). 탐침 원문 `analysis/review/impl-r1-probe_test.go.txt`(버리는 파일 사본).

| # | 발견 | 처분 |
|---|---|---|
| P0-1 | 취소 대상이 이미 끝난 주문이면 계약상 409 `already-filled·canceled·rejected·modified` → 409 는 확정 status 목록(journal/dispatch.go:351) 밖, code 는 `definitiveRefusalCodes`(refusal_code.go:46) 밖 → CANCEL attempt IN_DOUBT → B3·`checkSymbolFree` 가 그 종목 모든 변경을 막고 CANCEL 은 세션 해소 대상 밖 → **재시작까지 손절 막힘**. 탐침 P1(이미 체결)·P2(접수 뒤 손절 반대 미체결 → 재취소 409)·P3(둘 중 하나 400 → 접수된 쪽 재취소 409) 세 모양 실측, 편집 전 base 에서는 셋 다 통과 | **Manager 결정 2026-11-24: 범위에 넣음** — design D7(이미 끝난 주문의 취소 거절은 확정 거절, CANCEL 만) · D8(접수 확정 매수 재취소 금지), §2-5 정정, 정본 delta 둘. 증거 `analysis/d7-d8-evidence.md`. 구현은 보강분 freeze 뒤(tasks 4b) |
| P1-2 | 「결정 발행 실패에도 cleared 유지」 변이(N10) 생존 — 시험의 원장 트리거가 손절 자신의 결정까지 막아 손절이 어차피 안 나감 | 반영(구현 단계): 취소 결정만 실패시키는 주입으로 시험 보강 — design §4-15 |
| P2 | 수량·가격 해석 실패 갈래(B9) 무기록 — Requirement 문언과의 틈 | 반영: design D4a(B9 에도 실패 줄 한 번) |
| P2 | 접수 확정 뒤 표시 정리·수량 해석 실패의 시험 부재(N6·N9) | 반영: design §4-16 |
| P2 | 중복 억제 표시를 로그 쓰기 **앞에** 세워, 쓰기 패닉을 삼키면 그 실패가 다시 기록되지 않음 | 반영: design D4a(표시를 쓰기 뒤에) |
| P2 | withPending=false 매도 경로를 실제 게이트웨이로 재는 시험 없음(M7 이 A198 만으로 생존한 이유와 같은 공백) | 반영: design §4-16 |
| P2 | 최대 수량 ×2 변이(N2) 생존 — 결정 최대 수량이 정확히 주문 수량임을 단언하지 않음 | 반영: design §4-16(원장 preimage `max_quantity` 단언) |
| P2 | 정정 주문의 낡은 수량·가격 | 기록만(엔진에 정정 경로 없음 — design §6) |
| P2 | 확정 거절(400·404·422)의 매 주기 재시도가 결정·attempt 행을 쌓음 | 기록만 — 잔여(design §6) |
| P2 | 픽스처 `005930` 은 실제 종목 코드 | design §4 문장 정정: 「계좌·주문번호·수량·가격은 합성, 종목 코드는 저장소 관례」(101 개 시험 파일) |

## 보강분 문서 (2026-11-24, 팀메이트 Opus)
- 증거 `analysis/d7-d8-evidence.md`(계약 원문 · 분류 사슬 AST · 사유 코드/골든 · a165 순서 · flatten/기동 복구 · D8 분기 · 편집 표면 · 정본 충돌). 새 AST: classifyMutation · classifyRefusalCode · refusalCodes · ClassifyBrokerRefusal · ClassifyHTTPMutation · isDefinitiveRejection · alreadyProcessingRefusal · Resolver.Resolve · Resolver.resolveCancel · classifyCancel · Saga.resolveStep · Recovery.Run.
- 문서: design D7·D8·D4a, §2-5 정정·§2-6, §3, §4-10~17, 변이 보강, §6 · proposal What Changes 4·5·Impact·Capabilities(`order-execution` 추가) · 정본 delta `specs/order-execution/spec.md`(새 파일, ADDED 본문 328자) · exit-policy ADDED 「엔진의 취소가 접수 확정된 매수는 다시 취소하지 않는다」(239자) · tasks 4b · Story acceptance 1 문장(tracker `--check` rc 0).
- `openspec validate a198-a-clearing-cancel-is-issued-as-a-cancel-decision --strict`: valid (rc 0, 2026-11-24).

## 보강분 freeze 리뷰 r1 (2026-11-24, 1회차·마지막)
- 보이스: codex-cli 0.165.0 read-only `</dev/null`. 원문 `analysis/review/delta-freeze-r1-codex.md`, 프롬프트 `delta-freeze-r1-prompt.txt`. 판정 「동결 불가 — P1 1, P2 5, P0 미입증」.

| # | 발견 | 대조 | 처분 |
|---|---|---|---|
| P1-1 | `already-modified` 를 확정 거절로 두면 flatten 은 `Failed` 로 종결되고 `UnsettledCancelSymbols`(journal/flatten.go:324 — 리뷰 인용)가 `Failed` 를 매도 보류에서 빼 청산을 허용, 엔진은 옛 번호의 종결 기록 뒤 원장에 없는 후계 매도 주문을 치우지 못한 채 손절 — 후계가 살아 있는 매도면 추가 매도 경로. 기동 해소는 REPLACED 를 일부러 park(amend_indoubt.go:145) | 증거 §5 와 같은 우려, flatten 매도 보류 해제는 새 사실 | **Manager 결정 대기**(already-modified 를 D7 에서 뺄지 — 결정 변경) |
| P2-2 | flatten 재전송 묶음(`already-processing → already-canceled`)에서 마지막 거절을 attempt 전체의 「취소 미성립」으로 일반화 — 앞선 전송의 취소가 성립했을 수 있음, 재개·기동 복구가 종결 attempt 를 다시 보지 않음 | 증거 §4 의 우려와 같음 | **Manager 결정 대기**(재전송 결과에 D7 을 적용할지 — 결정 변경) |
| P2-3 | 정본 order-execution Scenario 「다른 사유의 409」(spec.md:590-592 — already-processing 이 아닌 취소 409 는 결과 불명)와 새 delta 가 같은 입력에 반대 결과 → ADDED 만으로는 둘 다 만족 불가. delta 는 정책·status 제한이 없는데 증거 안 (가)는 「모호인 409·정책 있음」 | 직접 확인 — 사실 | **Manager 결정 대기**(MODIFIED delta 필요, 적용 범위 하나로 확정). 증거 §8 에 사실 기록 |
| P2-4 | D8 체결 경합 설명 오류 — 손절 접수 뒤(CLOSING) 매수 체결은 E31~E33 `ENTRY_WHILE_CLOSING` 거절로 RECONCILE, 「다음 주기 프로젝션이 보호」는 틀림 | table.go:78-80 직접 확인 | **사실 정정 반영**: design §2-5, 증거 §6. 경합 시험은 결정 뒤 |
| P2-5 | §4-10~16·변이의 공백(네 code 각각, AMEND·nil 정책·읽을 수 없는 code, D8 범위 결속·조회 오류, 가격 해석 실패, 쓰기 패닉 뒤 재기록, 마지막 Scenario 의 전제) | 사실 | **반영**(결정에 묶이지 않는 것만): design §4-17, exit-policy Scenario 에 「다른 미치움 주문·미종결 attempt 없음」 전제. already-modified·재전송 묶음·늦은 체결 시험은 결정 뒤 |
| P2-6 | 증거가 인용한 `cited-ast/ExitObserver_clearTheSymbol.ast.json` 은 편집 전 판 — 현재 소스와 sha 불일치 | 사실(의도된 pre-edit 판이었으나 증거 머리말이 「같은 작업 트리」로 적음) | **사실 정정 반영**: 증거 머리말에 판 구분과 post-edit 번들 경로 |

## 보강분 Manager 결정 반영과 조건부 freeze (2026-11-24)
- 결정: D7 = 엔진 취소 정책에만 · 세 code(`already-filled`·`already-canceled`·`already-rejected`) · `already-modified`·flatten·정책 없음·PLACE·AMEND 무변 · 사유 코드 기존 `dispatch_rejected_by_broker` · 재시도 뒤 결과에도 같은 분류 · 정본 Scenario 「다른 사유의 409」 MODIFIED · D8 그대로 · P2-4 정정·P2-5 시험 전부.
- **사실 정정(결정 불변)**: 엔진 정책은 같은 attempt 안에서 재전송하지 않는다(정본 order-execution:531, `withCancelRefusal` B3 `Definitive` 즉시 반환) — 「already-processing 뒤 재전송」 은 엔진에서는 **다음 주기의 새 취소**다. design D7·§4-13b 를 그 모양으로 적었다(already-processing → FAILED_CONFIRMED → 다음 주기 재취소 already-canceled → D7 확정 거절 → 종결 기록 뒤 손절).
- **보강분 Manager 조건부 freeze 2026-11-24 — 조건 확인값**:
  1. 반영: design D7·§3·§4-13·13b·17·변이·§2-6·§6, proposal What Changes 4·Capabilities·Impact, tasks 4b.4~4b.6, delta 둘(order-execution MODIFIED+ADDED, exit-policy ADDED) — ✓
  2. `openspec validate a198-a-clearing-cancel-is-issued-as-a-cancel-decision --strict`: valid — ✓
  3. MODIFIED 블록: `analysis/modified-delta/make_modified.py` 가 정본을 떠서 치환 1건(원문 1회 단언), 대조 diff `compare.diff` = 바뀐 줄 2→2, 그 밖의 차이 0(스크립트 단언), 재실행 두 번 sha256 동일(멱등) — ✓. ADDED 본문 358자.
  4. D7 편집 계획 = `Gateway.withCancelRefusal` 1 + 새 파일의 새 함수. flatten 패키지·`internal/journal`·`classifyMutation`/`classifyRefusalCode`/`refusalCodes`·`AllReasonCodes` 무편집 — ✓(계획; 구현 diff 로 재확인)
  5. a166 census: shapes.json·sites.json 에 `cancel_refusal` 0 — ✓. 구조 고정: `TestA165CancelRefusalUsesOnlyInjectedClockForJudgement`(a165_clock_judgement_test.go:48)가 cancel_refusal.go **원문**에서 `time.Now(`·`time.Since(`·`context.WithDeadline(` 금지와 `context.WithTimeout(dctx, policy.Window)` 1회를 센다 — 계획한 편집은 새 함수 호출 1줄로 그 문자열을 바꾸지 않아 시험 무편집으로 통과 예상(걸림 아님으로 판단, 구현 뒤 실행으로 확인). 그 밖에 `withCancelRefusal` 를 부르는 시험 4 파일은 동작 시험(정책 nil·flatten 정책) — ✓
- 결론: 조건 (1)~(5) 충족 → 보강 구현(4b.5~)으로 진행.

## Pre-Edit 선언 갱신 (tasks 4b.5, 2026-11-24 — High-risk, 면제 없음)

```text
Pre-Edit Gate:
- change id / task id: a198-a-clearing-cancel-is-issued-as-a-cancel-decision / 4b.5 → 4b.6·4b.7
- 대상 심볼: execgw.Gateway.withCancelRefusal(기존, 새로 편집) · engine.ExitObserver.clearTheSymbol(기존, 추가 편집 — D8·D4a). 새 코드: execgw 새 파일의 D7 판정 함수 1, a198_clear_cancel_log.go 의 기존 새 함수 수정(D4a 표시 순서). 무편집: classifyMutation·classifyRefusalCode·refusalCodes·alreadyProcessingRefusal·AllReasonCodes·Gateway.submit·internal/journal·internal/flatten·cmd/tossctl.
- CodeGraph/AST: analysis/d7-d8-evidence.md §2(withCancelRefusal 호출 자리 gateway.go:733 1, CANCEL 만 가르는 유일한 기존 자리) · analysis/function-logic/internal-execgw--gateway.withcancelrefusal/(pre-edit, 분기 13) · clearTheSymbol post-edit 번들(분기 17).
- CodeGraphContext: not-applicable(1단계와 같은 사유).
- 기존 동작 파악 근거: a165 시험 24(cancel_refusal_test.go · a165_*_test.go) · 리뷰 탐침(analysis/review/impl-r1-probe_test.go.txt) · openapi 409 계약 원문.
- Function Logic Map / Branch Test Map: 위 두 번들.
- upstream 상속 테스트 영향: no.
- 실패 테스트 선행 작성: yes — 4b.6 RED(analysis/tdd/red-4b.txt).
- 설정·DB·journal 변경과 rollback: 없음 — 사유 코드 재사용(골든 무편집), 스키마 0. rollback = 이전 이미지.
- 안전 불변식 §0: 통과 — LIVE 0, 손절 즉시성 강화 방향(재시작까지 막힘 → 종결 기록 대기), 새 API 호출 0, 토글 0, flatten 동작 무변.
```

## 보강 RED → GREEN → 검증 (tasks 4b.6~4b.8, 2026-11-24, 팀메이트 Opus)
- 커밋: RED `4f32233d4` · GREEN `e92b7621b` · 검증 기록(이 절). push 없음.
- 편집 기존 함수: `Gateway.withCancelRefusal`(엔진 정책 갈래에 D7 판정 호출 1) · `ExitObserver.clearTheSymbol`(D8 · D4a). 새 함수: `settledCancelRefusal`(새 파일 internal/execgw/a198_settled_cancel_refusal.go). 수정한 a198 새 함수: `noteClearCancelFailure`(표시를 쓰기 뒤로). `check_analysis` required **2**(위 둘) — flatten·journal·classifyMutation·refusalCodes·AllReasonCodes 무편집(`git diff e1eaf720c --stat -- internal` 6 파일). 생성기 재실행 뒤 delta diff 0.
- RED(analysis/tdd/red-4b.txt, 1판 production): 엔진 A198 새 시험 6 FAIL — P1 `cancel attempts = [IN_DOUBT|dispatch_outcome_unknown]` · P2 `broker cancels = [W-BUY-1 W-BUY-1]` · P3 `[W-BUY-1 W-BUY-2 W-BUY-1 W-BUY-2]` · 13b `[FAILED_CONFIRMED|broker_rejected IN_DOUBT|dispatch_outcome_unknown]` · 해석 실패 줄 0(quantity·price) · 패닉 뒤 줄 0. execgw `TestA198EngineSettledOrderRefusalsAreDefinitive` 6 본문 전부 AMBIGUOUS.
- GREEN(원문 verify/4b-*.txt 와 시험 로그): P1 `CANCEL|FAILED_CONFIRMED|W-BUY-1|dispatch_rejected_by_broker`, 종결 기록 뒤 주기에 손절 1 · P2 `CANCEL|CONFIRMED|W-BUY-1` → 손절 `FAILED_CONFIRMED|opposite_pending_order_exists` → 다음 주기 취소 0·손절 `CONFIRMED` · P3 취소 `W-BUY-1(400) W-BUY-2(접수) W-BUY-1(접수)` → 둘째 주기 손절 1 · 13b already-processing `FAILED_CONFIRMED|broker_rejected` → 재취소 already-canceled `FAILED_CONFIRMED|dispatch_rejected_by_broker` → 종결 기록 뒤 손절 · already-modified `IN_DOUBT`.
- 변이(analysis/mutation/ledger.md 「4b」): 13/13 CAUGHT. V8(D8 을 매도에도 — 초과 매도 방향)은 A094 가짜 제출자 시험 1개만 죽임(실제 게이트웨이 시험은 ReleaseClearedExitProposal 이 대신 막아 못 죽임).
- 선택 시험(GOFLAGS 없음): engine `-run 'A198|A094|ExitClear|ExitObserver'` ok 48.9s · execgw `-run 'Cancel|RiskGuardian|Issue|Refusal|A165'` ok 57.3s(a165 구조 고정 `TestA165CancelRefusalUsesOnlyInjectedClockForJudgement` 포함) · `./internal/flatten` 전체 ok 20.6s · journal `-run 'Prepare|Decision'` ok · `-race -run A198` ok 124.4s · `make lint` rc 0 · strict validate valid · check_analysis 「evidence complete」.
- 미실행: 전체 스위트 · `make gate`·`sdd-sync`·`sdd-check`·`test-seams` · engine 패키지의 선택 밖 시험 · cmd/tossctl 시험 · post-edit 커버리지 재측정 · `ConfirmedCancelOf` 조회 오류 시험(읽기 오류 주입 수단 없음) · withPending=false 매도 경로의 실제 게이트웨이 시험(구성 미완 — A094 가짜 제출자 시험이 그 분기를 덮음).

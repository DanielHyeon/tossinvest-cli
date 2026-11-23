# 1.8 편집 표면 · 고정 장치 표

HEAD `e4165fa3d`(코드 = `9738a0c32`), a198 **미착지**(origin/main 미전진 — 2026-11-24 1단계 끝에 재확인). 근거: AST `cited-ast/`, 타입 전수 `verified-census/`,
CodeGraph `code-context/codegraph-callers.txt`(1.6.0), grep 원문은 이 파일의 표 안 `파일:줄`.

## 1. 편집할 기존 함수 (확정안 — 「조건부」는 Manager 결정에 달림)

| # | 함수 | 파일 | 무엇 때문에 | a166 census | 소스를 여는 시험 · 핀 | 비고 |
|---|---|---|---|---|---|---|
| P1 | `verifyGate` | internal/app/engine/interlock.go | B6@536 를 닫힌 집합 판정으로(낡음이면 반환하지 않고 표지 + 뒤 조항 계속), D1a 계좌 확인 | 키 아님(이름이 `NewContext` 동결 본문에 없음 — `runInterlock` 안에서만 호출) | 없음(`"interlock.go"` 를 여는 시험 0) | 첫 실패 반환 구조(AST 11 갈래 전부 `return`) — claim-check C2 |
| P2 | `runInterlock` | 같은 파일 | B2/B3: 청산 전용 결과(Verified=false · 새 표지 · EntryPermitted=false · 새 감사 action · 래치) | 키 아님, 이름이 `NewContext` 동결 본문에 있음 → **이름 유지** | 없음 | a105 3.2 가 B3 도달 시험을 계획(편집 아님) |
| P3 | `NewContext` | internal/app/engine/engine.go | B16@592 Guardian 유지 조건(+ 래치를 runInterlock 에 두면 `gateFacts` 에 entry 한 칸) | **키 1**(`internal/app/engine/engine.go‖NewContext`) — 본문 한 글자도 동결. a192 `census-repair` 절차로 그 키만 갱신. `sites.json` 의 이 함수 자리(`opts.BudgetFactory.Budget`)는 무변 | `a087_admission_dataflow_test.go`(NewContext AST 안 하한가 판 생성 위치) · `a130_trade_outcome_refusal_log_test.go`(engine.go) | 기억: 「a166 census 는 배선 함수 본문을 고정한다」 — 선택 시험에 `./internal/accountbudgetruntime -run A166` 를 반드시 넣을 것 |
| P4 | `runEngineRun` | cmd/tossctl/engine.go | B7@232 새 표지로 + 기동 출력 한 줄(D8) | 0 | `a109_the_lock_outlives_every_endpoint_close_test.go`(defer 등록 순서 — 새 defer 없으면 무관), `a108`(ready 발행 없음) | `errEngineGateOff` 문구 무변 |
| P5 | `Context.ExitObserver` | internal/app/engine/exitwiring.go | B2@323 새 표지로 | 0 | `a163_judgement_pin_test.go:259`(이 함수 안 덮어쓰기 1건 — 무관), `a071_readiness_assembly_test.go`(금지 문자열 `ProtectionReadiness`·`ReadinessAdapter`·`checkProtection`), `a092_structure_pins_test.go`(exitwiring.go 의 다른 함수) | **a198 이 같은 함수를 편집 중**(design D2/D3 — Guardian 을 `CancelIssuer` 로 단언). a198 착지 뒤에 편집 → 순차, 동시 아님. base 를 a198 착지 뒤로 잡아야 ⑤ 창에 a198 몫이 안 들어옴 |
| P6 | `Context.ReconcileDriver` | internal/app/engine/reconcileloop.go | B2@368 새 표지로 | 0 | `a071_readiness_assembly_test.go`(금지 문자열) | |
| P7 | `judgeAttestation` | internal/app/engine/attestation_watch.go | 낡음/거부 판정을 인터록과 **같은 함수**로(D7) — 지금은 `Verify` 첫 실패만 보므로 「만료 + 다른 계좌」를 `expired` 로 판정(인터록은 D1a 로 거부) | 0 | 없음 | claim-check C7 |
| P8 | `attestationWatch.eventFor` | 같은 파일 | 낡음 단계 문장(D7), 청산 전용으로 돌고 있음 문장 | 0 | 없음 | 날수 · 지남 단계 문장(「그 뒤 엔진 재시작이 거부된다」)은 D7 밖 — Q |
| P9 | `newAttestationWatch` | 같은 파일 | 「지금 청산 전용으로 돌고 있다」의 출처 = 기동 때 `c.Automation` 의 새 표지 | 0 | 없음 | 구조체 `attestationWatch` 에 칸 하나(함수 아님) |
| P10 | `AllReasonCodes` | internal/execgw/failclosed.go | 새 사유 코드 등록(닫힌 목록) | 0 | `a098_sender_down_reason_test.go`(failclosed.go 를 열고 길이·소속 단언), `failclosed_test.go`(golden) | 분기 0 함수지만 본문 편집이라 gate ⑤ 번들 필요 |
| P11(조건부) | `logGateDecision` | interlock.go | 청산 전용 기동에도 「보호는 프로세스 수명」 문장(정본 「보호 미배선 기동」) + 새 표지 필드 | 0 | 없음 | verified-census ⑦ — Q |
| P12(조건부) | `runSoakAttest` · `newSoakAttestCmd` | cmd/tossctl/soak.go | 「The automation gate will refuse to start」(565) · 「The gate will still refuse to start … protective」(577 — interlock-gates-entry-not-exit 이후 이미 거짓) · Long 도움말(220) | 0 | soak 시험(문구 단언 여부 2단계 확인) | tasks 2.5. verify-execution-capability 의 미완 task 는 soak.go 를 편집 대상으로 적지 않음(1.10·1.11 은 완료) |

**편집하지 않는 것(근거):** `Gateway.submit`·`checkEntry`(B1@860 이 이미 `raisesExposure` 일 때만 래치를 읽음), `EntryGate` 메서드(`Block` 그대로 씀),
`refuseStartup`, `UnmetInterlockClauses`(표 `interlockClauses` 는 패키지 변수), `attest.Attestation.Verify`(D1a 는 새 메서드로 — 1.6), `stageKey`·`prepare`(키 규칙 무변 — D7),
`productionStrategyWorker`(엄격 유지), `engineRuntime`(루프 집합 무변).

**새 함수(새 파일 권고 — gate ⑤ required 를 늘리지 않음):** `attest.Attestation.NamesAccount`(1.6), 닫힌 집합 판정 하나(인터록 · 감시 공용), 청산 쪽 판정 메서드
(`AutomationStatus` 의 「청산 루프를 돌려도 되는가」), 감사 Detail 조립, 기동 출력 문장.

**패키지 변수 편집(함수 아님 — 번들 불요):** `execgw.latchOrder`(새 사유를 `ReasonStrategyCentralIntegrity` 뒤 · `ReasonOperatingModeBlocked` 앞에 — 머리말 규칙
「appended, never inserted … ahead of the operating mode」), `reason.go` 상수, `audit.ActionGate…` 상수, `attestationWatch` 구조체 칸, `AutomationStatus` 칸,
`internal/execgw/testdata/reason_codes.golden`(데이터), `a098_sender_down_reason_test.go` 의 `reasonCodesRegisteredAfterA098`(시험 파일의 패키지 변수).

## 2. 뒤집힐 기존 시험 (예측 — 2단계 RED 가 확정)

| 시험 함수 | 파일 | 무엇이 뒤집히나 | 처리 |
|---|---|---|---|
| `TestGateOnRefusals` | interlock_test.go:342 | 사례 **셋**: 「attestation expired」(514), 「attestation misses an endpoint the engine calls」(533), 「attestation predates the price read」(547) — 셋 다 낡음 | design §3-6 은 「두 사례」라 적음 → **셋**(사실 정정). 사례를 청산 전용 기대로 옮기거나 새 시험으로 |
| `TestRefusedStartIsAudited` | interlock_test.go:649 | 픽스처가 만료 attestation(653) — 이제 거절이 아님 | 거절이 유지되는 결함(다른 계좌 등)으로 픽스처 교체 |
| `TestGateDecisionIsLoggedStructurally` 의 하위 「refused」 | interlock_test.go:798~ | 픽스처가 만료(804) | 같음 |
| `TestEachInterlockClauseHasALine` 의 「an expired attestation」 | runtime_wiring_test.go:95 | 만료 = 거절 열거 기대 | 만료 + 다른 계좌(D1a) 또는 만료 + 다른 조항으로 |
| `TestA196RefusalTellsTheConsequenceFirst` 의 scope_lost · expired | a196_attestation_deadline_internal_test.go:439 | 제목 · 본문 첫 문장 「기동이 거부된다」 | 새 귀결 문장으로(나머지 셋 incomplete · account_mismatch · unreadable 무변) |
| `TestA196ScopeLossWordingFollowsTheMissingKind` | 같은 파일:542 | 네 사례 모두 「기동이 거부된다」 접두 단언(575·578) | 새 귀결로 |
| (가능) `TestA135UnusableAttestationsAreCritical` · `TestA135ScopeLossIsCritical` | a135_attestation_watch_internal_test.go | 단계 값은 무변 예상(만료 + 다른 계좌 사례 없음) | 회귀 핀으로 남음 |

뒤집히지 않을 것으로 본 것: `TestNothingElseRefusesTheOperatorConfiguration`·guardian_test(유효 attestation — `errors.Is` 부정 단언만), `TestGateOffStartsAndDoesNoGateWork`·
`TestTheGateOffPathIsUntouched`·`TestAGateOffEngineRefusesWithoutEnumeratingClauses`(게이트 OFF), `TestTheDriverRefusesAnUnverifiedGate`·ExitObserver 미검증 시험(손으로
만든 `Automation{}` — 새 표지 false), cmd/tossctl `engine_test.go` 의 손으로 만든 Context(`Verified: true`), `a098_sender_down_reason_test.go` 의 길이 단언(변수에 새 코드를
더하면 통과 — 시험 **함수** 무편집).

## 3. 진행 중 change 와의 겹침

| change | 상태(tasks) | 같은 파일 | 같은 함수 | 판정 |
|---|---|---|---|---|
| a198 clearing cancel | 종결 중(1단계 끝 기준 미착지) | internal/app/engine/exitwiring.go | **`Context.ExitObserver`**(a198 design D2·D3 · tasks 3.2) | **겹침** — 순차(a198 착지 뒤 a199 편집). base 를 a198 착지 뒤로 다시 잡을 것(Q) |
| a197 external opposite buy | 등록 | exitloop 쪽(`ExitObserver.submit` 등 — `ExitObserver` **타입**의 메서드) | `Context.ExitObserver` 아님 | 겹침 없음(파일도 다름 — exitloop.go 예상) |
| a105 protection wired | 구현 0 | protection_wiring.go · guardian_test · interlock_entry_test | `runInterlock` 은 **시험 도달**(3.2)만 — 편집 아님 | 함수 편집 겹침 없음. 단 a199 뒤 `runInterlock` B3 의 뜻이 「검증된 런타임의 EntryPermitted」로 좁아짐 — a105 3.2 는 낡지 않은 attestation 으로 B3 에 닿아야 함 |
| a107 retire second protection core | 등록 | 봉인 시험 `a071_*` | — | a199 는 `a071_*` 를 편집하지 않음(금지 문자열만 지킴) |
| verify-execution-capability | 1.4 · 2.x · 3.x 미완(사람 실측 위주) | soak.go(완료된 1.10·1.11 만), engine gateway.go(완료 2.7 근거 인용) | 미완 task 의 편집 함수 없음 | 겹침 없음. P12(soak.go 문구)는 조건부 |
| a063 renewal profile | 4.x 운영 실측 | console templates | — | 겹침 없음 |

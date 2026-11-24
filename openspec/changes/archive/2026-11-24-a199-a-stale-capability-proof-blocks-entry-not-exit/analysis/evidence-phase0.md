# a199 사전 증거 — attestation 낡음 시 저하 모드(진입 차단·청산 유지)

> **1단계 재확인(tasks 1.3, 2026-11-24, worktree `~/tossos-work/a199-impl`, HEAD `e4165fa3d` — 코드는 `9738a0c32` 와 같음, 등록 커밋은 문서뿐).**
> 사전 증거의 AST 12개를 같은 도구(`tools/logic-map`, scratch 빌드)로 다시 뽑아 대조했다 — `analysis/cited-ast/extract.py --compare`,
> 원문 `analysis/cited-ast/summary.txt`. **12개 전부 `source_sha256` 같음 · 시작/끝 줄 같음 · 분기 (id, 종류, 줄) 같음.** 아래 본문의 줄 번호·분기 번호는
> 그대로 유효하다. 1단계가 더 뽑은 AST 21개(호출자 함수 · 편집 후보 · 1.6/1.7 대상)도 같은 폴더에 있다. 보조 원문은 `analysis/phase0-raw/`.
>
> | 대상 | 사전 증거 | 현재 HEAD | 달라짐 |
> |---|---|---|---|
> | AST 12개(줄·분기·sha) | 9738a0c32 | e4165fa3d | 없음 |
> | CodeGraph callers(1.6.0) | scratch 색인 | worktree 색인(`analysis/code-context/codegraph-callers.txt`) | 없음(같은 호출자) — 1단계가 심볼을 더 물음 |
>
> **본문 정정(사실, 1단계 실측):**
> 1. Q3 표 「Guardian 등급」 행의 「Guardian 체인이 래치를 먼저 읽고 결정 미기록(entrylatch.go:26-33)」은 **파일 주석의 서술일 뿐 생산 배선이 아니다** —
>    `EntryLatchFor`·`EntryLatch()` 의 비시험 호출자 0(grep), `guardian_wiring.go` 에 진입 게이트 인자 없음. 생산에서 진입 차단의 집행은 게이트웨이(`Gateway.checkEntry`)뿐이다.
> 2. Q6 「a163 핀: 별도 a163 핀 시험은 안 찾음」 → 있음: `internal/app/engine/a163_judgement_pin_test.go:259` 가 `Context.ExitObserver` 안의 무조건 덮어쓰기 1건을 셈(a199 의 B2 편집과 무관한 줄).
> 3. Q6 「인터록 행동 시험」의 뒤집힘 예상은 `analysis/edit-surface.md` 가 대체한다(`TestGateOnRefusals` 의 낡음 사례는 **셋** — 「predates the price read」 포함).

- 시작 HEAD: `9738a0c325cbd10b48112c2a9383108bf3080717` (main) · 끝 HEAD: 같음(재확인함) → 줄 번호는 이 HEAD 기준.
- 읽기 전용. 저장소 무편집. logic-map 바이너리는 scratch 에 빌드(`go build -o <scratch>/logic-map ./tools/logic-map`).
- CodeGraph: CLI **1.6.0** (`codegraph callers <sym>`), 원문 `codegraph-callers.txt`. `affected` 는 쓰지 않음.
- AST 산출물(이 폴더): `ast-interlock-verifyGate.json` · `ast-interlock-runInterlock.json` · `ast-interlock-refuseStartup.json` ·
  `ast-engine-NewContext.json`(internal/app/engine/engine.go) · `ast-engine-runEngineRun.json` · `ast-engine-engineRuntime.json`(cmd/tossctl/engine.go) ·
  `ast-runtime_wiring-UnmetInterlockClauses.json` · `ast-attest-Load.json` · `ast-attest-Attestation_Verify.json` ·
  `ast-attest-Attestation_MissingEndpoints.json` · `ast-attestation_watch-judgeAttestation.json` · `ast-protection-Gateway_checkProtection.json`.
- 보조 원문: `codegraph-callers.txt`, `source-open-tests.txt`, `test-name-hits.txt`, `grep-degraded.txt`.

---

## Q1. 기동 경로의 순서

### 호출 사슬 (CodeGraph 1.6.0 callers, 비시험만)
`newEngineRunCmd`(cmd/tossctl/engine.go:136) → `runEngineRun`(:196) → `engineAssemble` 변수(:576) → `assembleEngine`(cmd/tossctl/engine_assembly.go:15) → `engine.NewContext`(internal/app/engine/engine.go:451) → `runInterlock`(interlock.go:390) → `verifyGate`(interlock.go:496).
`verifyGate` 호출자 1(runInterlock), `runInterlock` 호출자 1(NewContext), 생산 `NewContext` 호출자는 assembleEngine 하나(grep `engine.NewContext(` 비시험 1건 engine_assembly.go:53). **flatten-all 은 NewContext 를 안 거침** — `engine.NewOrderPath` + 자체 `execgw.NewEntryGate`/`execgw.New`(cmd/tossctl/flatten.go:211,243-251) → 인터록 무관.

### runEngineRun 단계 (AST `ast-engine-runEngineRun.json`, 196-384)
| 단계 | 줄 | AST 분기 | 실패 시 |
|---|---|---|---|
| 1 journal flock | 209 | B3 if@210 | return err (211) |
| 2~4 조립+인터록 | 219 `engineAssemble` | B4 if@220, B5 if@221(인터록 거절이면 조항 열거) | 열거 출력 후 `errEngineInterlockUnmet` (226) / 그 밖 err (228) |
| 게이트 OFF | 232 | B7 if@232 `!ectx.Automation.Verified` | `errEngineGateOff` (233) |
| 5 verify runlock | 242-247 | B8/B9 | `errVerifyInProgress` (246) |
| 6 활성 마커 | 251-269 | B10~B12 | 거절 아님(note 출력) |
| 7 루프 조립 | 277 `engineRuntimeFactory`→`engineRuntime` | B13 if@278 | return err (279) |
| 7 표면 endpoint 4종 | 304-373 | B15~B26 | **강등 보고만**(`reportEngineEndpointDegraded` 495) — 엔진은 계속 |
| 실행 | 383 `rt.Run` | — | — |

### NewContext 순서 (AST `ast-engine-NewContext.json`, 451-631)
NewOrderPath(452, B1) → openAuditLog(459, B2) → `newAutomationStatus`(469) → 알림 publisher 해석(477) → `recordGateSettings` 감사(486, B5 — 실패 시 기동 거절 491) → **①계좌 해석 `resolveAccountRef`**(500, B6 → `refuseStartup` 502; 게이트와 무관하게 매 기동 GET /api/v1/accounts) → ②journal open(507, B7) → 2b apply hooks(517, B8) → ③`buildGateway`(531, B11 → 547) → 3b Guardian: **`gate.Enabled && guardian == nil` 일 때만** 생산 Guardian 생성(560, B12; 실패 B14@568) → ④`runInterlock`(581, B15@588 → jrn.Close + return err 590) → `!automation.Verified` 이면 guardian=nil(B16@592-594) → strategy projection(596) → Context 반환(601).
- `buildGateway` 안에서 기동 복원: RECONCILE `tracker.Restore`(gateway.go:279), 미전달 critical 진입 래치 `restoreAlertEntryLatch`(287), 운영 모드 투영 `bindOperatingModeProjection`(291), 보호 readiness provider 조립(308-317). → **이 복원들은 인터록보다 먼저 돈다.**

### runInterlock (AST `ast-interlock-runInterlock.json`, 390-421)
- B1 if@393 `!gate.Enabled` → `logGateDecision`(info) → return status, nil (397). attestation 안 읽음.
- B2 if@400 `verifyGate` 실패 → `refuseStartup`(401; 감사 `ActionGateRefused` 기록, interlock.go:357-363 — 게이트 ON 일 때만, AST refuseStartup B1 if@354) → `logGateDecision`(Warn, `obs.EventOperatingMode` + reason, interlock.go:485) → return err (403).
- 통과: Verified=true(406), `EntryPermitted = Protection==WIRED`(409) → 감사 `ActionGateAccepted`(410-417, B3: 기록 실패 시 return err 417).

### 거절되면 무엇이 일어나나
| 항목 | 사실 | 근거 |
|---|---|---|
| 프로세스 | `runEngineRun` 이 `errEngineInterlockUnmet` 반환 → main 이 `os.Exit(1)` | engine.go:226, main.go:25 |
| 루프 | 하나도 안 뜸(엔진 Context 자체가 nil) | engine.go:588-590, 219-229 |
| 감사 기록 | `audit.ActionGateRefused` 1줄(Detail=오류 원문) | interlock.go:357-363 |
| 구조 로그 | `engine.operating_mode` Warn 1줄(stderr=engine.log) | interlock.go:485 |
| 알림(ntfy/outbox) | **없음** — runInterlock 은 logger·audit 만 받음, Notifier 미사용 | interlock.go:390-421 (매개변수에 Notifier 없음) |
| 재시도 | 엔진 자체 재시도 없음. autostart 는 **콘솔 시작 시 1회** 호출, 거절이면 문자열만 반환 | engineautostart.go:57-87 |
| 콘솔 | 계속 실행(정본 engine-safety 「승인된 엔진 자동 기동」 Scenario 자동 기동의 인터록 거부, spec.md:737-739) | — |
| 엔진 밖 감시 | 마커는 인터록 뒤(6단계)에 잡으므로 거절된 기동은 마커를 안 씀 → `heartbeat-check`(매 분 타이머)가 「autostart ON·마커 없음」이면 down critical 1회. 거절 **사유**는 싣지 않음 | engine_heartbeat_check.go Long 설명, deploy/systemd/tossos-heartbeat.timer |
| 호스트 수준 1분 재기동 | engine.go:289,323 주석이 「autostart가 1분마다 다시 세웠다」(a108/a109 사고 서술)고 적지만, 현재 저장소에서 그 재시도 경로는 이번 검색(`deploy/`, engineautostart.go)으로 **미확인** | — |

### 게이트 OFF 흐름
NewContext: Guardian 미생성(560) → runInterlock B1(393) 즉시 반환, attestation·Guardian·한도 검사 0 → guardian=nil(592) → cmd B7(232) `errEngineGateOff`("기동할 루프 집합이 없다", engine.go:82-85). 즉 **게이트 OFF = 엔진 런타임 없음**(정본 「엔진 런타임 수명주기」② spec.md:399, Scenario 424-426). 시험: `TestGateOffStartsAndDoesNoGateWork`(interlock_test.go:277), `TestTheGateOffPathIsUntouched`(interlock_entry_test.go:103), `TestAGateOffEngineRefusesWithoutEnumeratingClauses`(cmd/tossctl/engine_test.go:148). 계좌 읽기(GET accounts)는 게이트 OFF 에서도 수행(interlock.go:44-50).

---

## Q2. 인터록 조항별 목적

**번호 체계가 셋이다(사실):** ① 파일 머리 주석 1~8(interlock.go:13-30, 8항 = ProtectionReadiness — **낡음**, 코드 562-576 이 「더는 거절 안 함」), ② `verifyGate` 본문 주석 1~8(+7b·7c), ③ 운영자 열거표 `interlockClauses` 1~8(runtime_wiring.go:215-238). 아래는 `verifyGate` AST 분기 순서(최초 실패에서 중단, 496 주석).

| AST | 줄 | 검사 | sentinel(errors.Is 가능) | 지키는 것 |
|---|---|---|---|---|
| B1 | 498 | Guardian 주입 | `ErrGuardianRequired` | **모든 주문** — ExitObserver 도 Guardian 의 `ReductionIssuer` 를 요구(exitwiring.go:326) |
| B2 | 506 | 한도 5필드 `Limits.Validate` | `ErrLimitsRequired` | **진입** — 「게이트웨이가 EXPOSURE_RAISING 결정에 쓰는 같은 술어」(502-505); RISK_REDUCING 결정엔 한도 스냅숏 없음(114-115) |
| B3 | 513 | 거래 정책 place·sell·cancel·live | `ErrTradingPolicyRefused` | **청산 가능성** — 「들어갈 수 있는데 못 나가는」 구성 거절(580-596) |
| B4 | 519 | `attest.Load` | Load 의 sentinel 을 `%w` 로 감쌈 | 아래 표 |
| B5 | 531 | 계좌 해석 결과 비어 있음 | `ErrAccountUnresolved` | 범위(조립이 걸렀어야 함) |
| B6 | 536 | `att.Verify(now, account, RequiredEndpoints())` | Verify sentinel `%w` | 아래 표 |
| B7 | 542 | Guardian 한도 = 감사 한도 | `ErrGuardianLimitsMismatch` | **진입**(EXPOSURE_RAISING 만, 114) |
| B8 | 546 | 7b 하루 매수 상한(a136) | (daily_buy_interlock.go, 미열람) | **진입** |
| B9 | 550 | 7c 최대 보유 수(a138) | (open_positions_interlock.go, 미열람) | **진입** |
| B10 | 555 | 게이트웨이 배선(Orders reader·Entry gate) | `ErrGatewayRequired` | **모든 주문** — 생성 id round-trip(633-656) |
| B11 | 558 | transport 가 clientOrderId 운반 | `ErrKeylessTransport` | **모든 place**(매수·매도) — replay 중복 방지(126-135) |
| — | 562-577 | (옛 9항) ProtectionReady | 없음 — 게이트웨이로 이동 | 진입(아래 Q3·Q4) |

모든 거절은 `ErrAutomationGateRefused` 로도 감싸짐 → `UnmetInterlockClauses` 가 표로 열거(runtime_wiring.go:252-266, AST B1 if@253 비인터록이면 nil).

### attest 오류 종류 (AST `ast-attest-Load.json`, `ast-attest-Attestation_Verify.json`)
| 함수 | AST | 원인 | sentinel | errors.Is 로 가를 수 있나 |
|---|---|---|---|---|
| Load | B2 if@164 | 파일 없음 | `ErrMissing` | 예 |
| Load | B2 else@167 | 읽기 실패(권한 등) | `ErrMalformed` | **파싱 실패와 같은 sentinel** — 문구로만 구분 |
| Load | B3 if@171 | JSON 파싱 실패 | `ErrMalformed` | 〃 |
| Load | B4 if@174 | format_version > 1 | `ErrFormatTooNew` | 예 |
| Verify | B2/B3/B4 case@213/215/218 | 계좌 칸·만료 칸·endpoint 목록 빔 | `ErrIncomplete` ×3 | 셋은 **문구로만** 구분 |
| Verify | B5 if@222 | `!now.Before(ExpiresAt)` | `ErrExpired` | 예 |
| Verify | B6 if@228 | 계좌 숫자열 불일치 | `ErrAccountMismatch` | 예 |
| Verify | B7 if@233 | 필요 endpoint 누락 | `ErrEndpointNotAttested` | 예 — **어느** endpoint 인지는 문구. 데이터로는 `MissingEndpoints(required)`(242-258)를 따로 불러야 함 |

- Verify 는 **첫 실패만** 반환(순서: 빈 칸 → 만료 → 계좌 → endpoint). 만료이면서 POST 누락인 파일은 `ErrExpired` 만 보인다(attestation_watch.go:99 도 같은 사실을 적음).
- `Expired(now)`(261)·`ExpiresWithin`(269)·`MissingEndpoints` 는 Verify 와 독립 호출 가능한 공개 메서드.

### RequiredEndpoints (interlock.go:219-230) 와 호출 자리 (rg/grep 기반, CodeGraph 미확인)
| endpoint | 엔진에서 부르는 자리 |
|---|---|
| GET /api/v1/accounts | `resolveAccountRef` 매 기동(interlock.go:699) |
| GET /api/v1/orders, /orders/{id} | `execgw.OfficialOrders` — 체결 감지(cmd/tossctl/engine.go:770-775), Resolver(gateway.go:297-304), 게이트웨이 round-trip |
| GET /api/v1/buying-power | 대사 스냅숏(reconcile/snapshot.go:294), 체결 감지(filldetect/detect.go:625), 매수 preflight(execgw/failclosed.go:129), IN_DOUBT(execgw/indoubt.go:475) |
| GET /api/v1/holdings | 계좌 sweep(runtime_wiring.go:60,67 — 대사·체결 감지), exit floor(exitwiring.go:208) |
| GET /api/v1/prices | exit 관측(exitloop.go:874), 편입(adoption.go:308) |
| POST /api/v1/orders | `Gateway.Place` — exit 매도(exitwiring.go:337 `opts.Submit = c.Gateway`), 전략 매수(휴면+보호 거절) |
| POST /api/v1/orders/{id}/cancel | `Gateway.Cancel` — exit 관측이 자기 발의 취소(interlock.go:590-591 주석) |

- 주석 215-218: 전략 전용 읽기(FX)를 이 목록에 넣으면 「US 진입 증거 부재가 reconcile·protection·exit·fill 루프를 거절」하게 된다는 이유로 제외 — **이 목록의 실패가 청산 루프까지 끈다는 점을 저자들이 이미 비용으로 인식**한 문장.
- 발급자: 감독 증거는 `acceptSupervised` 가 나이 ≥ 유효 기간(기본 30일, soak/attest.go:92·102)이면 **오류가 아니라 건너뜀**(soak/attest.go:331-333) → 갱신 타이머(6시간, deploy/systemd/tossos-attest.timer)가 POST 두 개 없는 파일을 정상 저장(cmd/tossctl/soak.go:541, 561-569 「gate will refuse to start」 출력 후 return nil). 즉 「감독 증거 30일 경과」는 파일상 **만료가 아니라 `ErrEndpointNotAttested`** 로 나타남(11-23 실제 파일의 sentinel 은 `~/.config/tossctl` 접근 금지로 **미확인**).
- cmd/tossctl/soak.go:574-577 출력 문구 「interlock clause 9 … gate will still refuse」는 낡음(interlock-gates-entry-not-exit 이후 기동 거절 안 함).

---

## Q3. 이미 있는 「진입만 막는」 장치

공통 집행점: `Gateway.submit`(execgw/gateway.go:470) — `raisesExposure` 는 형태에서 계산(Place: `side=="buy"` 381, Cancel: false 420, Amend: 수량·가격 증가 451). 순서 2a `checkProtection`(586) → 2b `checkEntry`(592) → … 디스패치 직전 둘 다 재검사(662, 669). `checkEntry`/`checkProtection` 모두 `!plan.raisesExposure` 면 즉시 통과(gateway.go:860; protection.go:90 AST B1 if@90). EntryGate 주석: 「Blocks apply to new entries only. Cancels and liquidations are never gated」(retry.go:479-481).

| 장치 | 상태가 사는 곳 | 세우는 자 / 푸는 자 | 진입 차단 자리 | 청산 통과 방식 | 재시작 |
|---|---|---|---|---|---|
| 운영 모드 NORMAL/ENTRY_BLOCKED/HALT_ALL | 원장 `operating_modes`(journal/core_domain.go:183) → EntryGate 래치 `ReasonOperatingModeBlocked` 투영(execgw/modegate.go:35-75) | 자동 강화: 닫힌 트리거 6종(journal/operating_mode.go:75-96 — 일일손실·401/403·알림 미전달·exit 관측 두절·대사 5주기·체결 5주기; **attestation 트리거 없음**), HALT_ALL 자동 금지 / 완화: OPERATOR+승인+audit(`tossctl engine mode-release`, 엔진 프로세스 안 소켓 engine.go:359-373) | `EntryGate.checkAccountEntryLocked`(retry.go:630) via `Gateway.checkEntry` | `ModeAllows`: RISK_REDUCING 은 모든 모드 허용(operating_mode.go:232-234) + raisesExposure=false 는 checkEntry 미진입 | 유지 — 기동 시 원장에서 복원(gateway.go:291) |
| 진입 래치(사유별) | EntryGate 메모리 `latches`(retry.go:488) | Block/Clear(558,569); 해제 세대(585-609) | 같음 | 같음 | 사유별: 알림 미전달 래치는 원장에서 복원(`restoreAlertEntryLatch` gateway.go:287), 나머지 메모리 래치는 미확인 |
| 조회 staleness | EntryGate `lastOK`(retry.go:487) | RecordSuccess 로 자동 해제 | checkAccountEntryLocked 642-657 | 같음 | 메모리(재시작 시 「never observed」로 차단 시작, 650-651) |
| RECONCILE(계좌·종목) | 원장 reconcile_states → Tracker·`symbolLatches` 투영(symbolgate.go:183) | 대사 / `reconcile-resolve`·재조회 일치 | `CheckEntryFor`(symbolgate.go:229) | 정본: 위험 축소는 **확정 하한 수량으로만**(order-execution spec.md:346-352) — 완전 무차단 아님 | 유지(gateway.go:279) |
| Guardian 등급 | 결정 Safety Class EXPOSURE_RAISING / RISK_REDUCING / PROTECTION_WEAKENING(journal/execution_contract.go:51-53) | RiskGuardian 발급 | Guardian 체인이 래치를 먼저 읽고 결정 미기록(entrylatch.go:26-33) | RISK_REDUCING 엔 한도 스냅숏 없음 | — |
| ProtectionReady | 서명 protection attestation + 시장별 supervisor assembly(아래 Q4) | 컴파일 상수·조립(`Wired:false`) | `Gateway.checkProtection`(protection.go:89) | raisesExposure=false 즉시 통과 | — |
| strategy-entry | 휴면 — 두 시장 비활성(spec.md:399) | — | worker 생성 시 `Journal·Gateway·Guardian != nil && Automation.Verified`(strategy_entry_supervisor.go:469), 스냅숏에 `AutomationVerified`·`EntryPermitted`·`ProtectionWired`(392-394) | — | — |

정본 Requirement(줄): risk-management 「Kill switch와 운영 모드 — 모드×클래스 표」(121); engine-safety 「자동화 게이트 기동 인터록」(106) 중 「ProtectionReady는 기동 조건이 아니라 진입 허가 조건」(117-135); 「자동 진입은 모든 안전 권한의 교집합이다」(521); 「재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다」(1148); 「운영 모드는 산 프로세스의 진입 게이트에 닿는다」(1809); 「진입 게이트는 사유별 해제 세대를 가진다」(1765); 「배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」(1640); order-execution 「RECONCILE 상태」(346), 「Fail-closed 분기」(309); exit-policy 「관측 경로와 fail-safe」(63), 「Automatic adoption obeys every authoritative RECONCILE state」(277).

**현 빌드 사실:** 생산 supervisor assembly 가 두 시장 모두 `Wired: false`(internal/app/engine/protection_wiring.go:41-42) → provider 가 Wired 아닌 assembly 를 거부(protectionreadiness/production.go:293) → 게이트 ON 엔진의 **모든 매수는 이미 게이트웨이에서 `broker_protection_not_wired` 로 거부**(execgw/reason.go:107). 시험 `TestProductionEngineAssemblesPairedUnwiredReadinessProvider`(a071_readiness_assembly_test.go:10), `TestEntryIsPermittedOnlyWhenProtectionIsWired`(interlock_entry_test.go:82). 즉 지금의 검증된 런타임은 사실상 이미 「진입 0·청산 가능」이다.

---

## Q4. 손절·청산 경로가 attestation·게이트에 기대는 자리

**`internal/attest` 비시험 import 자(grep):** cmd/tossctl/soak.go, cmd/tossctl/httpapi_reader.go(Mask 만), internal/app/engine/{interlock,runtime_wiring,attestation_watch}.go, internal/console/{data,portfolio,performance_history}.go, internal/protection/controller.go(protection 서명 타입), internal/soak/attest.go, internal/flatten/confirm.go(Mask 만), internal/verifylive/*. → **execgw·reconcile·filldetect·exitloop 는 capability attestation 을 읽지 않는다.** 의존은 전부 「기동이 되느냐」를 통한 간접 의존.

| 경로 | 게이트·인터록 결과를 읽는 자리 | 효과 |
|---|---|---|
| 런타임 전체 | `NewContext` 가 인터록 실패 시 Context 없음(engine.go:588-590); cmd `!Verified` 면 거절(cmd/tossctl/engine.go:232) | 거절 = 아래 전부 없음 |
| exit 관측(손절·익절) | `ExitObserver`: `!c.Automation.Verified` 면 `ErrExitObserverUnavailable`(exitwiring.go:323-324); Guardian 이 `ReductionIssuer` 여야(326) | Verified + Guardian 필수 |
| Guardian | 게이트 ON 일 때만 생성(engine.go:560), `!Verified` 면 nil(592-594) | 저하 모드에서 Guardian 존재 여부가 설계 질문 |
| 대사 + 편입(adoption) | `ReconcileDriver`: `!Verified` 면 거절(reconcileloop.go:368-369); 편입은 대사 드라이버 안(opts.Adoption 374) | Verified 필수 |
| 체결 감지 | `engineFillDetector` 는 Verified 검사 없음(cmd/tossctl/engine.go:763-788), 단 런타임 자체가 Verified 뒤에만 조립 | 간접 |
| 보조 실행자 | `engineRuntime` Auxiliary: 알림 배달·일반 알림 relay·`AttestationWatch`·수동 매수 갱신·하한가 갱신·IN_DOUBT 정산(cmd/tossctl/engine.go:749-750) | 런타임과 함께 |
| 비상 청산 flatten-all | NewContext·인터록 **안 거침**(flatten.go:211-264), TTY 확인 필요 | attestation 무관(사람 명령) |
| 게이트웨이 | capability attestation 안 읽음; 진입만 `checkProtection`/`checkEntry` | 매도·취소 통과 |

- 「ProtectionReady는 attestation 범위에서만 WIRED다」(engine-safety spec.md:476-519)의 attestation 은 **capability attestation 이 아니다** — 시장별 **서명 protection attestation**(internal/attest/protection_signature.go, protection_matrix.go; 파일 `protection-readiness-manifest.json`·시장별 AttestationFile, protectionreadiness/production.go:21-22, 286, 349) 이고 `buildGateway` 가 provider→`NewPairedReadinessAdapter` 로 게이트웨이에 꽂음(gateway.go:308-317). 인터록과의 연결: **없음** — `interlockClauses` 에 보호 항목 없음(runtime_wiring.go:233-237), `verifyGate` 562-577 에서 제거, 남은 것은 보고용 상수 `ProfileProtection = UNWIRED`(execgw/protection.go:68) → `AutomationStatus.Protection`·`EntryPermitted`(interlock.go:409). 머리 주석 8항(interlock.go:29-30)은 이와 어긋나는 낡은 서술.
- 정본은 이미 「readiness 실패 시 기존 보유의 reduce-only exit·reconciliation 은 계속」을 시장 단위로 요구(spec.md:493-507). capability attestation 쪽에는 같은 분리가 없다.
- **게이트 OFF 일 때 청산 루프:** 돌지 않는다. 런타임 조립 전 `errEngineGateOff`(engine.go:232-233), 그리고 ExitObserver·ReconcileDriver 생성자도 `!Verified` 거절. 루프 등록은 `engineRuntime` 의 고정 4개 감독 루프(reconcile·exit·filldetect·strategy-entry, cmd/tossctl/engine.go:723-745) 하나뿐이고 게이트별 분기 없음(AST `ast-engine-engineRuntime.json` 분기 B1~B6 은 전부 생성자 오류 반환).

---

## Q5. 현재의 감시·알림

**attestation_watch.go (a135·a196)** — 보조 실행자 `attestation-watch`(39), 1시간 주기(41), 게이트 OFF 면 취소만 기다림(252-256), 브로커 호출 0.
- `judgeAttestation`(100-144, AST `ast-attestation_watch-judgeAttestation.json`): Load 오류 → `unreadable`; Verify 오류를 errors.Is 로 `expired`/`scope_lost`(누락 목록 = MissingEndpoints)/`account_mismatch`/`incomplete`/기타 `unreadable`(106-121) — **인터록과 같은 술어를 재사용**(사본 없음, 주석 11-15). 통과면 기한 = min(만료, 감독증거 기한=`ProofAt+(Expires−Issued)`)(`restartDeadline` 151-159); 기한 지남 → `proof_overdue`; D-8+ healthy, D-7~3 notice(normal), D-2~1 imminent(critical)(124-143).
- 알림: `checkOnce`(271) → `prepare`(309, 판정 패닉 회수) → `notifyOutsideRecovery`(301, 알림기 패닉 비회수). 하루 한 번·에피소드 키(`stageKey` 358). 타입 `broker.credential_expiring`(a196 D6). 알림 꺼짐이면 critical 도 normal(spec.md:2223). Verified·EntryPermitted·게이트·운영 모드 **쓰기 0**(파일 머리 22-23, a135 D4).
- 정본: engine-safety 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」(spec.md:2213-2297) — 「실행 중인 엔진의 동작은 바뀌지 않는다」(Scenario 2239-2241), 「감시의 정지는 진입·청산·손절·운영 모드…에 닿아서는 안 된다」(2225).

**콘솔:** `Console.readAttestation`(internal/console/data.go:266-305) — `attest.Load` 후 **Verify 를 쓰지 않고 자체 규칙**으로 Reasons 를 만듦(계좌 칸 빔·endpoint 빔·만료·누락 endpoint; 계좌 일치 검사 없음, 72h 경고 296). 렌더: templates.go:838-860 「능력 증명 attestation」 섹션(상태 유효/사용 불가, 「게이트가 열리지 않는 이유」), 요약 `attestSummaryText`(settings_tabs.go:354-363). RequiredEndpoints 는 `engine.RequiredEndpoints()` 주입(cmd/tossctl/console.go:411).

**기동 거절의 기록 자리:** 감사 `ActionGateRefused`(interlock.go:357-363), stderr 구조 로그 `engine.operating_mode` Warn(485), cmd 가 조항 열거 출력(cmd/tossctl/engine.go:222-225). **alert outbox·ntfy 전송 없음**(Q1 표). 엔진 밖에서는 heartbeat 감시가 「엔진 down」만 알림(사유 없음).

---

## Q6. 변경 표면의 고정 장치

### a166 census (internal/accountbudgetruntime/testdata/a166-wiring/shapes.json 90키, sites.json) — 시험 `internal/accountbudgetruntime/a166_wiring_census_test.go`: `TestA166WiringCensusBaseline`(301), `TestA166WiringCensusRejectsBypasses`(308)
| 함수 | census 키로 동결? | 비고 |
|---|---|---|
| `internal/app/engine/engine.go‖NewContext` | **예** | 본문 한 줄만 바꿔도 깨짐(기억: a189·a192 사고) |
| `internal/app/engine/engine.go‖NewOrderPath` | **예** | |
| `internal/app/engine/gateway.go‖buildGateway` | **예** | |
| `internal/app/engine/gateway.go‖newGatewayReplay` | 예 | |
| `cmd/tossctl/engine_assembly.go‖assembleEngine` | **예** | |
| `cmd/tossctl/flatten.go‖buildFlattenWiring` | 예 | |
| `cmd/tossctl/engine_reconcile.go‖buildReconcileResolveRuntime` | 예 | |
| verifyGate · runInterlock · refuseStartup · newAutomationStatus | 키 아님(NewContext 본문 문자열에 이름만 등장) | 이름 바꾸면 NewContext 동결 본문이 바뀜 |
| engineRuntime · runEngineRun · ExitObserver · ReconcileDriver · checkProtection · judgeAttestation | 0 적중 | |

### 소스를 직접 여는/이름을 고정하는 시험 (grep, `source-open-tests.txt`·`test-name-hits.txt`)
- `TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources`(internal/app/engine/a071_readiness_assembly_test.go:27): exitloop.go·exitwiring.go·reconcileloop.go·runtime_wiring.go 에 `ProtectionReadiness`·`ReadinessAdapter`·`checkProtection` 문자열 금지.
- `TestA087ProductionBoardAdmissionFlowIsOnlyLiteralNil`(a087_admission_dataflow_test.go:19): NewContext AST 안의 하한가 판 생성 위치 고정.
- `"engine.go"` 를 여는 cmd/tossctl 시험 10파일: a090_engine_runtime_unobserved_log, a092_exit_options_pin, a094_boot_order, engineproc, a129_engine_runtime_exit_order_log, a163_engine_runtime_source_age_log, a109_the_lock_outlives_every_endpoint_close, engine_runtime_branch, engine_strategy_entry_dormant(`TestProductionRuntimeIncludesOneDormantStrategyEntryOuterLoop` — 감독 루프 4개 **순서까지** DeepEqual, spec.md:452-457), engine_test(`TestTheLoopSetIsTheSpecifiedThree` 364 — 부분 일치) + internal/app/engine/a130_trade_outcome_refusal_log.
- `exitwiring.go` 를 여는 시험: a071_readiness_assembly, a092_structure_pins. `gateway.go` 를 여는 시험: a071_security_review, protection/dormant_test, execgw/a165_retry_structure, journal/position_campaign_hardening.
- 인터록 행동 시험: interlock_test.go(`TestGateOnRefusals` 342 — 「attestation expired」514·「misses an endpoint」533 사례 포함, `TestRefusedStartIsAudited` 649, `TestGateDecisionIsLoggedStructurally` 771, `TestTheEngineRequiresThePriceRead` 956), interlock_internal_test.go, interlock_entry_test.go(`TestTheGateStartsWithoutBrokerSideProtection` 30 등), cmd/tossctl/engine_test.go(`TestAnUnmetInterlockIsEnumerated` 171), a136/a138/a139 *_interlock_test.go. `RequiredEndpoints` 는 시험 13파일 40적중(attestation 드리프트 가드 포함 — 정본 spec.md:110 「목록 확장 change 는 가드 갱신」).
- 이름 고정 적중 수: verifyGate 0 · runInterlock 0 · refuseStartup 0 · interlockClauses 0 · UnmetInterlockClauses 1파일 · errEngineGateOff 2파일 · errEngineInterlockUnmet 2파일 · ExitObserver 32파일 · ReconcileDriver 5파일 · engineRuntime 9파일 · EntryPermitted 6파일 · judgeAttestation 2파일.
- a163 핀: 이번 검색은 `recordGateSettings` 의 a163 항목(interlock.go:804-805)만 확인, 별도 a163 핀 시험은 **안 찾음**.

---

## Q7. 선례

| change | 결정 한 줄 |
|---|---|
| harden-execution-base (archive 2026-07-26) | 인터록 도입. design.md 에는 「저하 모드 금지」 문장이 **없다**(grep `degrad`/`warning` 0). 그 문장은 코드 주석뿐: 커밋 `30563ae53`(2026-07-26) interlock.go:32-34 원문 *"Any failure is a startup refusal … not a warning. There is no degraded mode: an engine that trades with an unverified capability is the thing this interlock exists to prevent."* review.md:29 「attestation 기동 인터록으로 기계 연결」 |
| verify-execution-capability (진행 중) | proposal Why: 「자동화 게이트는 capability attestation 없이 켤 수 없다 … 이 change가 완료되어야 실전 자동 주문이 열린다」 — attestation = 무인 자격 갱신·rate·조회 완전성·주문 경로의 증명 |
| **interlock-gates-entry-not-exit** (archive 2026-08-29) | **직접 선례.** 보호 미배선 조항을 「기동 거절」→「노출 증가 mutation 거절」로 옮김. D1 「관측은 코드가 바뀌면 만료되므로 조항을 없애지 않고 발동 지점으로 옮긴다」, D2 집행점은 `Gateway.Place`(형태에서 계산, 한 곳), §0.4 「현재 손절 즉시성은 0이다(엔진이 뜨지 않으므로). 이 change 후 exit observer가 돈다」, §0.9 「완화하는 것은 하나 … 그 외 여덟 조항, 한도, 정책 수치, **attestation 요구는 전부 불변**」, D4 각 조항이 여전히 거절함을 개별 시험, 위험 「로컬 판단 오작동 매도 … 사용자가 교환을 명시적으로 선택(2026-07-30)」 |
| add-core-domain (archive 2026-07-26) | 운영 모드 3종. **EXIT_ONLY 삭제** 이유 원문(design.md:31): 「EXIT_ONLY 삭제 — ENTRY_BLOCKED와 행동이 동일해 §0.7 사다리의 무의미 단계였다(2c가 실제 차이를 만들면 재도입)」. 정본 risk-management spec.md:131 동일 |
| a063 (진행 중) | D3 「A refused renewal remains failed and visible … The warning is advisory: it must not stop a running engine or relax the startup interlock.」 |
| a125 (archive 2026-09-28) | a063 예외 퇴역 — 상세 미열람 |
| a105 (진행 중, 구현 0) | 보호 감독 `Wired` 생산 전환 — `productionProtectionAssemblies` wired 파라미터화(proposal What Changes 2) |
| a107 (진행 중, 등록만) | 두 번째 보호 core 퇴역 — 봉인 시험 파일(a071_security_review_test 등)을 같이 편집하므로 순서 겹침 금지 |
| a134 (archive 2026-10-14) | 엔진 침묵은 엔진 밖(호스트 타이머 heartbeat-check)에서 판정; ntfy 자체 침묵 경보 전제는 거짓(D0) |
| a135 (archive 2026-10-14) | 감시는 인터록 술어 그대로(D1), 보조 실행자(D3), 엔진 영향 0(D4: Verified·EntryPermitted·게이트·운영 모드 쓰기 0), 「갱신하면 POST 를 잃는다」 예측 안 함(D6) |
| a196 (archive 2026-11-24) | 기한 = min(만료, 감독증거+유효기간)(D1), 알림 문장 「지금 엔진을 재시작하면 기동이 거부된다 … 자동 손절·익절이 돌지 않는다」(D7), 재검증은 엔진과 같은 잠금 → 엔진 정지 필요(D7 G10) |

`grep -rni "degraded mode|저하 모드|exit-only|EXIT_ONLY|청산 전용"` (openspec·docs, md) 13건 — 전부 EXIT_ONLY 삭제 서술(add-core-domain·risk-management·guardian-chain.md:103·ROADMAP.md:27) 또는 foundation 의 WTS 기능 「성능 저하 모드로 강등」(add-tossos-foundation/design.md:44). 「reduce-only」는 이 grep 에 넣지 않음(정본 다수 등장, spec.md:495·503·507·527·539 등).

---

## 안 본 것 / 한계
- daily_buy_interlock.go·open_positions_interlock.go 본문(7b·7c) AST 미추출.
- RiskGuardian 의 reduction 발급 내부(모드 표 적용 지점) 미열람 — import 근거로 attest 미의존만 확인.
- 콘솔 [엔진 시작] 버튼 경로(engineproc.go)의 거절 표시 미열람.
- 메모리 래치 중 재시작 복원되는 것의 전수 목록 미작성(알림 미전달·모드·RECONCILE 만 확인).
- endpoint→호출 자리 매핑은 grep 기반(CodeGraph 미교차).
- 11-23 실제 attestation 파일의 sentinel(만료 vs 범위 상실) — 운영 파일 접근 금지로 미확인.
- 시험은 하나도 돌리지 않음.

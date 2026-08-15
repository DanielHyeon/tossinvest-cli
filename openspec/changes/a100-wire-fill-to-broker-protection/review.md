# Review — a100 proposal-freeze

- **날짜**: 2026-08-11
- **게이트**: `docs/WORKFLOW.md:144` proposal-freeze — change의 첫 구현 task 착수 전 1회.
  `:149` 위험 등급 가중 — 주문 실행·위험관리·원장·reconciliation을 건드리므로 **adversarial Eng
  관점 필수**.
- **대상**: 리뷰 이전 판의 `proposal.md`, `design.md`, `tasks.md`,
  `specs/{protection-orders,engine-safety,fill-detection}/spec.md`.
- **결과**: **범위 재조정 + 재조정본의 2차 개정.**
  - Phase 1 (CEO): critical 8 / high 8. 두 독립 보이스가 6개 차원 전부에서 「전제가 유효하지
    않다」로 합의했고, 반대 방향 의견은 없었다.
  - Phase 3 (Eng, adversarial): 재조정본에 대해 critical 4 / high 6 / medium 1 + 자체 3.
    **그중 셋은 계획대로면 컴파일·롤백·배포가 불가능한 수준이었다**(X1·X2·X9).
  - 두 라운드 모두 코드로 재확인한 것만 수용했다. 기각한 지적과 사유도 아래에 적었다.

## 보이스 구성

| Phase | 관점 | 보이스 |
| --- | --- | --- |
| 1 | CEO (전제·범위·대안) | 코드 기반 자체 검증 + Codex + 독립 Claude 서브에이전트 — **3인** |
| 2 | Design | **not-applicable** — 아래 「생략한 단계」 참조 |
| 3 | Eng (**adversarial 필수**) | Codex 적대적 리뷰 + 코드 기반 자체 검증 — **2인** |
| 3.5 | DX / 운영자 | 자체. 범위 축소로 남은 표면(상주 주문 가시성)만 |

Phase 1의 모든 하중 인용은 HEAD에서 직접 재검증했다. 재검증 없이 수용한 발견은 없다.

## 사용자 결정 (전제 게이트)

autoplan이 자동 결정하지 않는 유일한 항목. 네 선택지를 평문으로 제시했고
(`AskUserQuestion`은 이 환경에서 조용히 실패한다 — 프로젝트 기억 `askuserquestion-tool-fails-here`),
사용자가 **1번 — 분할 + 선행 실측**을 선택했다.

- `a100` = 브로커 상주 보호 설치 (기존 보유 + 신규 체결). `Wired`는 계속 false.
- `a105` = supervisor · `Wired` 생산 · coverage latch · manifest 재서명 · 서명 도구.
- 착수 전 실측 2건: 조건주문 발동(M-A), 엔진 사망 노출 창(M-B).

## Phase 1 — CEO 합의표

| # | 차원 | Claude | Codex | 합의 |
| --- | --- | --- | --- | --- |
| 1 | 전제가 유효한가 | 아니오 | 아니오 | **CONFIRMED** |
| 2 | 지금 풀 문제가 맞는가 | 범위가 잘못 잘렸다 | 대상 population이 틀렸다 | **CONFIRMED** |
| 3 | 범위 조정이 맞는가 | 분할 | 분할 | **CONFIRMED** |
| 4 | 대안을 충분히 봤는가 | 아니오 — 보유 경로가 후보에도 없었다 | 아니오 — 더 작은 change 미검토 | **CONFIRMED** |
| 5 | 시장·경쟁 리스크 | 해당 없음(사내 단일 사용자 제품) | 해당 없음 | N/A |
| 6 | 6개월 궤적이 건전한가 | 아니오 | 아니오 | **CONFIRMED** |

## 발견과 처리

### CRITICAL

| # | 발견 | 출처 | 근거 | 처리 |
| --- | --- | --- | --- | --- |
| C1 | **오늘의 계좌에 손절을 한 건도 남기지 않는다.** 촉발이 「체결 delta > 0」인데 6개 레인이 모두 OFF여서 신규 체결이 없다. proposal 자신의 Why가 기존 보유를 "더 위험한 쪽"이라 부르면서 촉발이 거기 닿지 않았다 | Codex + 자체 | 레인 상태, `journal/adoption.go:289`(`InitialStop`) | **수용.** D2를 「이벤트가 아니라 상태 수렴」으로 재설계. 대상에 기존 보유 포함 |
| C2 | **보호를 설치하려다 체결 감지를 잃는다.** D8의 동기 `Ledger` 데코레이터가 (a) 보호 실패를 `Apply` 에러로 만들면 `pollLocked` B6이 **같은 사이클의 남은 스냅샷을 폐기**하고, (b) 성공해도 왕복 시간이 뒤 스냅샷의 커밋을 밀어 신선도 표본을 오염시켜 `evaluateSLO` → `Gate.Block`. 원안 tasks 4.6은 (a)만 격리했다. **이 change 자신의 `fill-detection` spec이 이미 금지하던 것** | 자체 + Codex | `detect.go:316-342`, `:349`, `:523`, `ledger.go:106-107`, 측정된 B6 미실행 | **수용.** D3을 독립 수렴 워커로 재설계 |
| C3 | **supervisor/`Wired` 절반이 안전 이득에 불필요하다.** 보호주문은 매도이고 `checkProtection`은 `!raisesExposure`에서 즉시 반환하므로 readiness를 조회하지 않는다 | 서브에이전트 | `execgw/protection.go:89-91`, `gateway.go:377,416,447` (AST 5분기 전수) | **수용.** 그 절반 전체를 a105로 이관 |
| C4 | **켤 수 없는 능력을 만든다.** `supervisor_digest`를 담은 서명 manifest를 발행하는 코드가 저장소에 없다. `attest/protection_signature.go`는 검증 전용이고 서명자는 `_test.go`에만 있다 | 서브에이전트 (자체 확인) | `internal/attest/protection_signature.go`, `cmd/` 전수 검색 | **수용.** a105 이관 목록에 서명 도구를 명시 |
| C5 | **중심 전제가 한 번도 관측된 적이 없다.** 조건주문 발동은 양 시장 미측정인데 D2 조건 (1)이 fixture 일치로 대체하고 있었다 | 서브에이전트 (자체 확인) | `verify-execution-capability/measurements.md:197` | **수용.** M-A를 착수 전 게이트로. D10이 주장 범위를 제한 |
| C6 | **한 포지션에 매도 권한이 둘이 되는데 계약이 없다.** 브로커가 조건주문에 수량을 예약하지 않으므로 초과 매도가 가능하고, a091이 "매도 0인 손절은 critical"을 이미 spec에 넣어 두어 **정상 경합이 critical을 만든다** | 서브에이전트 (자체 확인) | `measurements.md:55`(M13), a091 `engine-safety` spec | **수용.** D8 계약 + spec 요구사항 추가. 단 「critical 예외」가 아니라 **시도 자체를 안 하는** 형태로 — 예외를 파면 delta가 MODIFIED가 되어 a071·a091과 archive 순서로 얽힌다 |
| C7 | **문서가 AST 없이 분기를 주장했다.** proposal 머리말이 "모든 분기 주장의 근거"라고 적었으나 design의 Context 표와 D8-(5)는 산출물이 없는 함수에서 유도했다 | 자체 | `.claude/CLAUDE.md` 「단계 건너뛰기 금지」 4항 | **수용.** `pollLocked` 산출물 생성. **그 즉시 D8이 인용한 함수 이름이 틀렸음이 드러났다** — `PollOnce`는 분기 0의 5줄 래퍼이고 로직은 `pollLocked`에 있다. `initialize`는 a105 이관으로 해소 |
| C8 | `engine-safety` spec이 요구한 「digest 불일치를 지목하는 refusal」에 대응 task가 없었다. `Current`/`Assess`가 tasks에도 FLM 목록에도 없고, 재서명 누락이 "아무것도 구성 안 됨"과 같은 `RefusalMissingEvidence`로 나온다 | 자체 | `protectionreadiness/production.go:187`, `:117`, `:252-299` | **수용.** 해당 요구사항 전체를 a105로 이관하여 해소 |

### HIGH

| # | 발견 | 출처 | 처리 |
| --- | --- | --- | --- |
| H1 | **봉인 해제가 D6의 주장보다 넓다.** `dormant_test.go:59`의 import walk는 `internal/protection`만 매칭한다. `protectionofficial`·`protectionlifecycle`은 walk 대상이 아니어서 아무 app 파일에서나 import할 수 있다 | 자체 | **수용.** 하나를 열면서 walk를 세 패키지 + 워커로 **넓힌다**(D6, tasks 4.2). 순 효과는 봉인 강화 |
| H2 | refusal taxonomy가 원인 4개를 말하지만 `initialize`의 bare return은 6개이고 그중 하나만 code를 설정한다 | 자체 | **a105 이관** |
| H3 | coverage latch가 이 change에서 구조적으로 무동작이다. 진입은 네 겹으로 닫혀 있고 a100은 `Wired`를 만들지 않는다 | 자체 | **수용.** D5에서 삭제하고 a105로 |
| H4 | 무보호 상태의 **최대 허용 시간**이 없다 | Codex | **수용.** D2 + spec 요구사항 + tasks 3.4 |
| H5 | **수렴 계약이 정의되지 않았다** — 언제 "수렴했다"인지 판정 불가 | Codex | **수용.** D2 정의 + tasks 3.3 + M-A 4b(판정 데이터 존재 확인) |
| H6 | coverage 게이트의 TOCTOU | Codex | **a105 이관** (latch 자체가 이 change에서 사라짐) |
| H7 | 죽은 코드 1,364줄(`Controller`+`Repository`)을 a100 **이전에** 정리하라 | 양쪽 | **부분 수용.** 삭제는 `execgw`가 쓰는 도메인 타입과 얽혀 별도 리뷰 단위다. **정리 change 등록을 a100 완료 후가 아니라 지금으로 당긴다**(tasks 6.5) |
| H8 | `ProtectionWired`가 entry supervisor에서 조립 시점 정적 bool이라 포지션별 coverage를 실을 수 없다 | 자체 | **a105 이관** |

### Phase 3 — adversarial Eng에서 새로 나온 것

리뷰가 만든 **새 설계**를 다시 공격한 결과다. 자기 설계를 자기가 리뷰하는 약점이 있으므로
Codex 적대적 보이스를 함께 돌렸다.

| # | 발견 | 근거 | 처리 |
| --- | --- | --- | --- |
| E1 | **상주 trigger가 엔진의 현재 청산선보다 느슨하다.** `InitialStop`은 t0에 얼린 값인데 실효 청산 수준은 ratchet와 `RunnerTrailPct`로 올라간다. 숨기면 "브로커에 손절이 있으니 안전하다"는 잘못된 안심이 생긴다 | `exitpolicy/ratchet.go:245`, `ladder.go:138-140`, `:392` | **수용.** D9 — 상주 보호는 **재난 하한**이며 영속된 스칼라에서만 유도한다. 콘솔·운영 문서가 차이를 표시한다(tasks 6.4.2). ratchet 수준이 스칼라로 영속되는지는 tasks 0.8이 확인 |
| E2 | **flat 포지션에 상주 주문이 남는 창.** 비-보호 경로(수동 매도·익절·대사)로 닫히면 다음 수렴 주기까지 상주 주문이 살아 있고, M13대로 수량 예약이 없으므로 그 창에 매수가 들어오면 **방금 산 주식에 남의 손절이 걸린다** | `measurements.md:55` | **부분 수용.** 레인이 전부 OFF인 현재는 수동 매수로 한정되므로 운영 문서로 처리(6.4.3). **a105가 진입을 열기 전에 닫는 것을 이관 목록에 명시** |
| E3 | 수렴 판정(tasks 3.3)이 브로커 조회의 trigger·수량 값에 의존하는데 **그 값이 비교 가능한 형태로 돌아오는지 미확인**이다 | M-A가 아직 미실행 | **수용.** M-A에 4b 단계 추가. 실패 시 수렴 정의를 등록 응답 기준으로 다시 쓰되 **느슨하게 바꾸지 않는다** |

### Phase 3 — Codex 적대적 Eng 보이스 (재조정된 계획 대상)

**평결: reject at HEAD `ce78b0d`.** 재조정된 계획을 별도 프로세스가 코드로 공격한 결과다.
아래는 전부 **내가 코드에서 직접 재확인한 것**만 실었다. 재확인 못 한 것은 그렇게 적었다.

| # | 발견 | 확인한 근거 | 처리 |
| --- | --- | --- | --- |
| X1 | **CRITICAL — 조립이 컴파일되지 않는다.** `protectionlifecycle`의 lifecycle 함수가 전부 unexported이고 `external_api_test.go:11`이 exported 패키지 레벨 함수를 금지한다. exported 메서드도 없다. `dependency_test.go:12`가 journal·protection·execgw·app 의존을 금지하므로 워커를 그 안에 둘 수도 없다 | 소스 전수 확인 — exported 함수 0, `State` 메서드 3개 전부 unexported | **수용.** 봉인이 넷이 아니라 **다섯**이었다. D1이 다섯 번째를 최소 해제 + 대체 단언으로 열고, D6 표에 행을 추가. tasks 4.0 |
| X2 | **CRITICAL — 롤백 계약이 성립하지 않는다.** journal이 더 새로운 스키마를 거부한다("An older binary **must not** touch it"). 백업 복원은 상주 주문의 기록을 잃고, 구버전 엔진이 모르는 채 자기 청산을 내어 매도 권한이 둘이 된다 | `journal/journal.go:23-27`, `:240`. 프로젝트 기억 `tossos-branch-behind-main-schema`의 2026-08-04 실발생과 같은 기전 | **수용.** spec 요구사항이 **불가능한 것을 요구하고 있었다.** D4 재작성, 롤백을 사람 절차로. tasks 6.1 |
| X3 | **CRITICAL — 대사와의 경합.** "커밋된 것만 읽는다"는 최신을 보장하지 않는다. 워커가 10주를 읽고 브로커 왕복 중에 대사가 0주를 커밋하면 워커가 SELL 10을 등록한다 | `reconcileloop.go:413`, `reconcile/converge.go:209`, `:244` | **수용.** 전송 직전·응답 후 재확인을 D3에 추가. tasks 3.8, fixture 5.9 |
| X4 | **CRITICAL — 발동한 child의 체결을 원장이 귀속할 수 없다.** 어댑터가 `TriggeredOrderID`를 bool로 접고, 체결 감지는 `mutation_attempts`가 소유한 id만 추적한다. 브로커가 팔았는데 journal은 보유로 읽고 재등록으로 간다 | `protectionofficial/gateway.go:271`, `journal/fills.go:1538-1548` | **수용.** child id를 `BrokerProtection`에 싣고 terminal 행을 「보유 재확인 후에만 재등록」으로. tasks 4.5.7, 5.11 |
| X5 | **HIGH — 「ACTIVE」가 어댑터로 판정 불가.** `WATCHING/PAUSED/ORDERING/ORDERED`가 전부 같은 값으로 접힌다. 일시정지와 무장이 구별되지 않는다. 그리고 "수렴하면 다시 묻지 않는다"면 이후 취소·만료를 영영 못 본다 | `protectionofficial/gateway.go:308-310`, `protection/domain.go:452` | **수용.** 수렴 판정을 어댑터가 실제 노출하는 값으로 재정의. 캐시 대신 주기 확장. tasks 0.11, 3.6 <br>단 codex가 "수량·trigger도 없다"고 한 것은 **틀렸다** — `BrokerProtection`은 `Quantity`·`Trigger`를 싣는다 |
| X6 | **HIGH — 전송 전 durable pending 커밋이 없다.** `recoverSubmit`은 영속된 pending이 없으면 시장을 latch한다 | `protectionlifecycle/lifecycle.go:63`, `:71`, `:23` | **수용.** D4에 저장 순서 고정, tasks 2.6, spec 요구사항 |
| X7 | **HIGH — 워커의 실행 주체·기동 순서가 없다.** 감독 루프는 넷이고 gateway 조립에 워커 필드가 없다. `buildGateway`에서 띄우면 automation interlock 평가보다 먼저 돈다 | `cmd/tossctl/engine.go:377`, `app/engine/gateway.go:92`, `engine.go:489`, `:533` | **수용.** 조립과 기동을 분리, gate 미verified면 미기동. tasks 3.9 |
| X8 | **HIGH — 「오늘의 보유」를 전부 덮지 않는다.** 편입은 토글이고 기본값 off. 제외·미편입 보유는 unmanaged 알림만 난다. 대사 루프는 gate 미verified면 존재를 거부한다 | `config/engine.go:79-90`, `adoption.go`, `reconcileloop.go:342` | **수용.** proposal의 주장을 정정. 이 배포의 현재 `adoption.enabled`는 **true**로 확인했으나 **토글이라는 사실을 감추지 않는다.** tasks 0.9 |
| X9 | **HIGH — 새 엔드포인트가 attestation 카탈로그 밖이다.** `RequiredEndpoints()` 8개에 조건주문 create/get/modify/cancel도 sellable-quantity도 없다. 갱신 안 하면 미증명 호출, 재발급 없이 갱신하면 **두 시장의 모든 루프가 거부된다** | `interlock.go:206-229`, `official/conditional_writes.go`, `protection_reads.go` | **수용.** D11 신설. **a100이 배포되지 못하게 만들 수 있는 유일한 경로.** tasks 0.10, 7.7b |
| X10 | **HIGH — D8이 「청구권 하나」를 강제하지 못한다.** 취소 확인 실패에도 매도를 진행시키므로 둘이 동시에 유효할 수 있다. 그리고 익절은 완전 청산이지만 보호 경로가 아니라 규칙에 걸리지 않는다 | `exitloop.go:1207-1213`(`isFullExit`에 `ActionLadderTakeProfit` 포함, `isProtective`에는 없음) | **수용.** 요구사항 제목과 본문을 **실제로 강제하는 것**으로 낮춤(「창을 최소화하고 남는 창을 밝힌다」). 취소 기준을 `isFullExit`로 확대. tasks 4.5.6, 5.10 |
| X11 | **MEDIUM — FLM 규칙 재위반.** 산출물 없는 내부 동작 주장 목록 | 그중 `dispatch`의 "분기 14 / return 16"은 이 change에 산출물이 **없다** | **부분 수용.** 수치를 지웠다 — 면제 사유에 분기 수를 쓰는 것 자체가 산출물 없는 분기 주장이다. 나머지 지적 다수는 **경계 사실**(파일 존재·타입 필드·상수 목록)이라 AST 대상이 아니라고 판단했고, 그 판단을 여기 남긴다 |

**codex의 지적 중 기각한 것**은 X5의 일부(수량·trigger 부재 주장)와 X11의 대부분이다.
전자는 코드가 반증했고, 후자는 「함수 내부 분기」가 아니라 타입·상수·파일 경계에 대한 주장이라
`tools/logic-map`의 대상이 아니다. 경계가 애매한 것은 착수 시점에 만든다(tasks 0.4).

### Phase 3.5 — DX / 운영자

범위 축소로 원래 예정했던 표면(manifest 재서명 마찰, refusal 진단성)이 a105로 갔다. 남은 것:

- **없던 것이 생긴다.** 배포 후 브로커 계좌에 조건주문이 나타난다. 운영자가 그것을 자기 주문으로
  오해하지 않도록 운영 문서에 적는다(tasks 6.4.1).
- **보이지 않는 실패가 가장 위험하다.** 수렴이 계속 실패하는 포지션은 아무 화면에도 나타나지
  않으면 "보호된 줄 알았는데 아닌" 상태다 → 보호 미설치 시간 + 상한 알림(tasks 3.4, 6.2).
- **trigger가 엔진 청산선과 다르다**는 사실은 운영자가 반드시 알아야 한다(E1, tasks 6.4.2).

### 착수 조건 조사 (tasks 0.8~0.10) — 자기 반증 2건

리뷰가 아니라 **착수 조건을 실제로 확인한 결과**다. 둘 다 이 문서가 이미 승인한 결정을 뒤집었다.
문서가 코드보다 먼저 말했고, 코드가 다르게 답했다.

| # | 결정 | 확인 결과 | 처리 |
| --- | --- | --- | --- |
| P1 | D9 「ratchet 수준이 스칼라로 영속되는지 **미확인**」 | **영속된다.** `exit_states.baseline_price`가 exit 판정마다 갱신되고(`journal/exit_state.go:563`), `RunnerTrailPct`는 `CandidateHighWaterRunner` 후보로 그 안에 합성되며(`ladder.go:391-411`), 단조 비감소는 이미 불변식이다(`candidate.go:86-88`) | D9 전면 재작성, spec 요구사항 5 재작성. 「재난 하한」 프레이밍 폐기 |
| P2 | D11 「재발급이 a071 서명 절차를 요구하면 C4에 걸린다」 | **무관하다.** a071의 Ed25519는 protection capability attestation이고 시동 interlock이 보는 것은 서명 없는 capability attestation이다. 대신 **soak 기록이 6일 stale이라 오늘 재발급 자체가 거부된다**(`MaxRecordAge` 48h) | D11에 실측 절 추가, 배포 순서 (a)~(e) 고정, 0.12 신설 |

**P1이 P2를 키웠다.** trigger가 `baseline_price`를 따라가므로 교체가 예외가 아니라 자동 경로의
상시 동작이 되고, `POST /conditional-orders/{id}/modify`가 카탈로그 대상이 된다. 그 엔드포인트는
verify 기록에 증거가 없다.

**부수 발견 2건(리뷰 대상 밖이지만 기록한다).**

1. attestation은 **2026-08-29 만료**이고 soak는 2026-08-05 이후 돌지 않았다. a100이 없어도
   그때까지 새 기록이 없으면 automation gate를 켠 엔진이 뜨지 않는다.
2. 2026-07-31 verify 실행은 **완료되지 않았다** — `conditional-persist`가 `awaiting-restart`,
   `sell-boundary`가 422(NXT 미지원)로 fail. 그 결과 **333430에 등록된 조건주문을 취소할 단계가
   실행되지 않았다.** 333430은 편입 포함 목록 안이므로, 수렴 워커가 만나는 첫 orphan이 된다.

### 편집 전 FLM (tasks 0.4) — 설계를 네 곳 바꿨다

편집 대상 6건의 AST와 **측정된 커버리지**를 만들었다. 「FLM은 편집 전이 아니라 주장 전」
(프로젝트 기억 `flm-before-claiming-not-before-editing`)의 후반부 — 실제 편집 직전 — 이고,
결과는 앞의 두 라운드와 같다. **문서가 코드보다 먼저 말했고 코드가 다르게 답했다.**

| # | 이전 판 | AST/측정 결과 | 처리 |
| --- | --- | --- | --- |
| F1 | D8-3b 「완전 청산 경로가 취소를 시도하도록 **넓힌다** — 기준은 `isFullExit`」 | `record` L1117이 **이미** `orderable && (CancelPendingFirst \|\| isFullExit(proposal))`이다. 진입 조건은 처음부터 익절을 포함했다 | D8-3b 재작성. 넓힐 것은 조건이 아니라 **취소 대상**(working order → 상주 조건주문) |
| F2 | D8-2 「취소 확인 실패는 매도를 막지 않는다」(원칙만) | `clearTheSymbol`의 에러는 `return err`로 **매도 없이 반환**(L1142, **미실행**), `!cleared`는 `orderable = false`로 **보류**(L1145) | D8-2에 기전 추가 — 취소를 그 함수/반환값에 **섞으면 안 된다**. `ArmSuppressedReason`에도 새 값을 넣지 않는다 |
| F3 | D6 「허용 파일 목록을 `gateway.go`와 **워커 조립 지점**으로 명시한다」 | 허용 목록은 파일 단위 1개(`dormant_test.go:61`). 워커를 추가하면 **두 번째 조립 경로**가 생긴다 — 가드가 막으려던 그것 | D6 재작성. **허용 목록 불변**, 워커는 `internal/protection`을 import하지 않고 좁은 인터페이스만 받는다 |
| F4 | D4 「additive-nullable」(원칙만) | `scanExitStateResult` 22분기 중 20개가 부패 판정. 보호 컬럼을 `v10Evidence`·`full`·평탄화 비교 중 하나에라도 넣으면 멀쩡한 행이 부패 → **exit 정책 정지 = 손절 정지** | D4에 세 판정 리스트 표 추가. tasks 2.3의 명시적 조건 |

**측정이 드러낸 것.** 손절 발행 함수 `ExitObserver.submit`은 **판정 가능한 10분기 중 6개가
미실행**이고, 그 6개에 **발행을 막는 네 경로가 전부 포함된다**. ⇒ a100은 이 함수에 제출을
막는 새 조건을 넣지 않는다. 「이미 flat」은 기존 수량 0 분기가 이미 잡으므로 바꾸는 것은
결과의 이름뿐이며 분기 수는 늘지 않는다.

`buildGateway`는 분기 4개가 **전부 에러 검사**이고 조건부 조립이 하나도 없다(그리고 4개 전부
미실행이다). ⇒ 워커는 무조건 생성하고 기동은 호출자가 한다 — 토글 분기를 넣으면 이 함수의
첫 조건부 조립이 된다.

### 조건주문 probe 구현 (tasks 0.10 (a)) — 0.10이 적은 (a)가 세 곳 틀렸다

0.10은 조사였고 (a)는 그 조사가 처방한 작업이다. 처방을 실행하려고 코드를 읽었더니 처방
자체가 틀려 있었다. **조사와 구현 사이에도 같은 역전이 있었다** — 0.10은 `soak/attest.go`와
`interlock.go`를 읽고 썼지만 `protectionofficial/gateway.go`는 읽지 않았다.

| | 0.10이 적은 것 | 코드가 답한 것 | 근거 |
|---|---|---|---|
| S1 | read 1개(`GET /conditional-orders`) | **3개.** `ConditionalOrderRaw`·`ProtectionConditionalOrdersRaw`·`SellableQuantityRaw`가 자동 경로다 | `gateway.go:103,122,173,193,221,246` |
| S2 | (a)에서 `RequiredEndpoints()`도 확장 | **확장하면 안 된다.** 둘 다 거부 기준이고, probe만으로 attestation에 실린다 | `attest.go:130`(거부) vs `:183`(적재), `interlock.go:518` |
| S3 | (d) attest는 아무 바이너리로나 | **a100 바이너리라야 한다.** `LiveOnlyEndpoints()`가 조건주문을 모르면 M-A 증거가 조용히 버려진다 | `cmd/tossctl/soak.go:456-465` |
| S4 | (a)가 3일 시계를 출발시킨다 | **아니다.** `Window`는 streak 창이라 창 안 1회 성공이면 증명된다 | `summary.go:107,162` |

그리고 조사에 없던 사실 둘이 일정 자체를 바꿨다.

- **S5 — 배포가 엔진 정지다.** soak는 `tossos:local` 컨테이너 안에서 돌고(pid 576246)
  **같은 컨테이너가 엔진**이다(pid 575876). probe 배포 = 이미지 재빌드 + 컨테이너 재시작이므로
  **손절 없는 창**이 생긴다. 기억 `tossos-engine-stop-removes-stoploss`에 따라 KST 05:00~09:00에서만
  한다 ⇒ **0.3(M-B)과 같은 작업이 됐다**(tasks 0.10b).
- **S6 — 고아 조건주문이 증거의 유일한 원천이다.** `GET /conditional-orders/{id}`는 계좌에
  조건주문이 있어야 증명된다(`probeOrderByID`와 같은 계약). 지금 그것을 가능하게 하는 것은
  333430의 2026-07-31 고아뿐이고, **M-A와 `verify run --resume`이 그것을 취소한다.**
  ⇒ **probe 배포는 M-A보다 먼저다.** 순서를 뒤집으면 by-id는 다음 조건주문이 생길 때까지
  증명되지 않는다.

**FLM이 예측을 검증했다.** `RunCycle`의 편집 전 AST는 분기 3개(전부 계정 해석·취소 확인)였고
"probe 추가는 분기를 바꾸는 편집이 아니라 순서 목록에 항목을 더하는 편집"이라고 적었다.
편집 후 재측정도 분기 3개다. 대신 위험은 분기가 아니라 **부작용**에 있었고, 그것이 Pre-Edit
선언 7번의 통과 조건 4개(credential 격리·completeness 격리·순서·거부 목록 불변)이며 각각
RED 테스트가 됐다. 새 probe 3개는 100.0% 측정됐다.

**배치가 게이트를 두 번 움직였다.** 어댑터를 `cmd/tossctl/soak.go` 안에 둔 것은
`static_test.go:136`이 그 파일 하나만 검사하기 때문이고(새 파일로 빼면 조건주문 어댑터가
가드 밖으로 나간다), 그 안에서 위치를 옮긴 것은 인접 삽입이 `soakReads.Order`를 "수정된 함수"로
끌어왔기 때문이다. 편집하지 않은 함수에 FLM을 만드는 대신 배치를 옮겨 **게이트가 사실대로
말하게** 했다.

## 이 리뷰가 바꾼 것

| 파일 | 변경 |
| --- | --- |
| `proposal.md` | 제목·Why·What 전면 개정. 기존 보유 대상 명시, reduce-only 근거, 선행 실측 절, a105 이관 목록, **자기 위반 기록**(C7) |
| `design.md` | D2·D3·D5·D8·D9·D10 신규/재설계, D6 경계 확장, 원안 D2·D3 삭제(이관), Migration Plan 0단계 추가 |
| `tasks.md` | 0절에 실측 게이트와 FLM 대상 재확정, 3절을 수렴 워커로, 4.5 이중 권한, 6.4 운영 문서 3항, 6.5 정리 change 선등록, 「a105로 이관」 절 신설 |
| `specs/protection-orders/spec.md` | 요구사항 5건 — 상태 수렴, 미설치 시간 상한, journal 컬럼, 단일 매도 청구권, trigger 유도 |
| `specs/engine-safety/spec.md` | supervisor 요구사항 삭제(이관), 봉인 요구사항에 **import 경계 확장** 추가 |
| `specs/fill-detection/spec.md` | 「보호 계획 단계」를 「보호를 계획하지 않는다」로 강화, 근거를 측정으로 교체 |
| `analysis/.../detector.polllocked/` | **신규 산출물** — AST 10분기, 측정된 branch-test-map(5개 미실행) |
| `a071/tasks.md` §6 | supersession 분할 기록 (gateway → a100, supervisor → a105) |
| `STORY-TOS-a100.yaml` | 제목·acceptance 11건 전면 개정, `rescoped`·`prerequisites` 추가 |

### 2차 개정 (Phase 3 결과 반영)

| 파일 | 변경 |
| --- | --- |
| `design.md` | D1에 다섯 번째 봉인, D2에 「ACTIVE 판정 불가」와 재확인 주기, D3에 대사 경합 재확인·워커 실행 주체, D4 롤백 전면 재작성 + 전송 전 pending 커밋, D8에 익절 경로·child 귀속·**보장하지 못한다는 명시**, **D11 신설**(attestation 카탈로그) |
| `tasks.md` | 0.9~0.11, 2.6, 3.6 정정, 3.8~3.9, **4.0(다섯 번째 봉인)**, 4.5.6~4.5.7, 5.9~5.11, 6.1 재작성, 7.7b |
| `specs/protection-orders/spec.md` | 수렴 판정을 어댑터 노출값으로, 재확인 의무, 전송 직전 재확인, child 귀속, 롤백 요구사항 재작성, 매도 청구권 요구사항을 **실제로 강제하는 수준으로 하향** |
| `proposal.md` | 「모든 보유」→「exit 관리로 편입된 보유」 정정 + 토글 사실 명시, `dispatch` 분기 수치 삭제 |

### 3차 개정 (착수 조건 0.8~0.10 실측 반영)

| 파일 | 변경 |
| --- | --- |
| `design.md` | **D9 전면 재작성**(얼린 `InitialStop` → `exit_states.baseline_price`, 「재난 하한」 폐기, 교체가 상시 경로가 된다는 파생 비용 명시), **D11에 재발급 실측 절 추가**(a071 서명 무관, soak 48h 신선도, verify 증거의 있고 없음, read는 supervised에서 조달 불가) |
| `specs/protection-orders/spec.md` | trigger 요구사항 재작성 — 유도 원천을 baseline으로, 단조성을 **상속되는 불변식**으로, 「영구 간극 표시」를 「수렴 지연 표시」로. 반사실 시나리오 1건 교체, 재계산 금지 시나리오 신설 |
| `tasks.md` | 0.8·0.9·0.10을 실측 결과로 종결, 배포 순서 (a)~(e) 고정, **0.12 신설**(soak 재시작은 a100과 무관하게 필요) |

### 4차 개정 (편집 전 FLM, tasks 0.4)

| 파일 | 변경 |
| --- | --- |
| `analysis/function-logic/` | **신규 산출물 6건** — `buildgateway`(분기 4), `scanexitstateresult`(22), `journal.openexitstates`(4), `exitobserver.record`(14), `exitobserver.submit`(11), `dormant` 가드(20). 전부 AST + **측정된 branch-test-map** |
| `design.md` | D4에 세 판정 리스트 + 「목록을 읽는 쪽도 하나」, D6 허용 목록 불변으로 정정, D8-2에 취소 비차단 기전, D8-3b 정정(`isFullExit`는 이미 쓰이고 있었다) |
| `tasks.md` | 0.4~0.7 종결. 0.4에 6건의 결론 요약 |
| `pre-edit-gate.md` | **신규.** High-risk 6대상 Pre-Edit 선언. **5건이 「조건부 통과」**이며 각 조건이 그 편집의 통과 요건 |
| `.claude/CLAUDE.md` | (a100 외) agent-config drift 해소 — codex mirror에만 있던 YAGNI·KISS를 source에 맞춰 넣었다. 규칙을 지우는 `--generate` 대신 |

### 5차 개정 (조건주문 probe 구현, tasks 0.10 (a))

| 문서 | 바뀐 것 |
|---|---|
| `tasks.md` | 0.10의 배포 순서에 정정 블록 추가(S1~S4), 0.10a(구현·완료)·0.10b(배포=엔진 정지, M-B와 동시)·0.10c(첫 사이클 확인) 신설 |
| `pre-edit-gate.md` | 7번 신설. 앞선 "High-risk 경로가 아니다"는 **틀렸다** — §0-5 인증 경로다. 통과 조건 4개 명시. journal 2건의 "(진행 중)" 측정값도 확정치로 교체 |
| `measurement-prereq.md` | M-A 단계 0 앞에 probe 배포 선행을 명시(S6) |
| `analysis/function-logic/internal-soak--runner.runcycle/` | 신설(편집 전 작성 → 편집 후 재생성·재측정) |

## 생략한 단계와 사유 (`not-applicable`)

**침묵한 생략은 금지이므로 전부 적는다.**

- **Phase 2 (Design 리뷰): not-applicable.** 이 change는 UI·시각 표면을 만들지 않는다. 운영자
  가시성은 Phase 3.5(DX)에서 봤고 그 결과가 tasks 6.2·6.4다. 콘솔 화면 변경이 생기면 그 시점에
  Design 리뷰를 별도로 받는다.
- **`internal/filldetect` 편집 FLM: not-applicable(편집 없음).** 다만 **분기를 근거로 인용하므로**
  `pollLocked` 산출물은 만들었다. 「편집하지 않음」은 인용의 면제 사유가 아니다.
- **`strategyDispatchCycle.dispatch` FLM: not-applicable(조건부).** 편집하지 않고 내부 분기를
  근거로 쓰지 않는다. 근거는 D7. 조건이 깨지면 그 시점에 만든다.
- **`ProductionProvider.initialize`·`Current`·`Assess` FLM: a100 범위 아님.** a105가 편집하므로
  a105의 FLM 대상이다. a100은 이 함수들의 분기를 **인용하지도 않는다** — 리뷰 이전 판의
  Context 표가 그것을 인용했고, 그 절이 삭제됐다.
- **`make gate` / `sdd-check`: 미실행.** 이 리뷰는 proposal-freeze이고 구현 전이다. 게이트는
  tasks 7절이 담당한다. `openspec validate --all --strict`(85/85 통과)와
  `check_analysis.py`(evidence complete)는 이 리뷰에서 실행했다.

## 남은 위험 (수용하고 진행)

1. **M-A가 실패할 수 있다.** 그러면 이 change는 멈춘다. 설계가 아니라 전제가 무너지는 것이다.
   그 가능성을 알고 시작한다 — 모르고 구현을 끝낸 뒤 아는 것보다 낫다.
2. **표본 1의 측정.** M-A는 "발동이 가능한가"에 답하고 "항상 발동하는가"에는 답하지 않는다.
   US는 이 측정 밖이며 KR 결과를 US로 옮겨 쓰지 않는다.
3. **수렴 워커의 브로커 호출량.** 주기·백오프 기본값이 보수적이어야 한다(tasks 3.6).
4. **a105가 열 때 닫아야 할 것 둘.** flat 포지션 창(E2)과 진입 권위 전체. 이관 목록에 있다.
5. **자기 설계를 자기가 리뷰한 부분.** 재조정본의 D2·D3·D8·D9는 이 리뷰가 만든 것이다.
   Codex 적대적 보이스가 그중 **넷을 실제로 깨뜨렸다**(X1·X3·X4·X10). 한 라운드 더 돌리면 더
   나올 가능성이 높다 — **구현 후 리뷰(tasks 7.6)는 형식이 아니라 필수다.**
6. **`PAUSED`가 실재하며 발동하지 않는 상태라면** raw status 노출이 필수 task가 되고 범위가
   `internal/protection`·`protectionofficial` 편집까지 늘어난다. M-A가 답한다(tasks 0.11).
7. **X9(attestation 카탈로그)가 a100을 배포 불가로 만들 수 있다.** 재발급이 a071의 서명 절차를
   요구하면 서명 도구 부재(C4)에 걸려 a105 없이는 배포할 수 없다. tasks 0.10이 착수 조건으로
   이것을 먼저 확인하는 이유다.

## 2026-08-15 current-main 재동결 기록

2026-08-11의 범위 축소는 유지하되 완료 표시는 재사용하지 않았다. 실행 base는 current main
`882a0b490b0b6d2eb7abe5c5040c514776f49f3e`로 재고정했고, 다른 change로 이관하지 않는다.

독립 A0가 찾은 착수 차단과 처리는 다음과 같다.

- M-A partial/4b 실패는 더 이상 착수를 허용하지 않는다. parent child-id의 **local receipt**가
  child fill local receipt보다 엄격히 먼저인 완전 PASS와 evidence-backed raw-status 표만 0.2/0.11을
  닫는다. 2026-08-14 runbook은 만료된 역사본이며 매 세션 read-only 재발행과 주문별 사람 승인이 필요하다.
- `PAUSED`/unknown은 ACTIVE가 아니고 mutation 금지+operator alert다. raw status/child id는
  손실 없이 보존한다.
- worker-only `protection_child_orders` registrar가 fill 전에 exact parent/client/scope/generation
  owner를 기록한다. `TrackedFillOrders`, `confirmedFillOwners`, `resolveFillOrigin`이 ordinary/protection
  owner를 exact union한다. child-before-registration은 소급 apply/backfill 없이
  `ATTRIBUTION_FAILED` reconcile/alert와 mutation block으로 끝나고 account reconciliation만 복구한다.
- current `engine.Runtime.Loops`는 하나의 non-cancel return이 전체 runtime을 내리므로 worker를
  넣지 않는다. recovery 뒤 `AuxiliaryExecutor`로 기동·취소·drain하고 ordinary cycle error는
  반환하지 않는다. panic/예상 밖 return은 `protection.convergence.worker_stopped`와 durable
  reconcile/alert만 남기며 다른 safety loop와 entry gate를 바꾸지 않는다.
- lifecycle public API seal을 T3 뒤에 두던 compile-order 순환을 없앴다. T2-B가 planned worker
  call graph의 exact pure allowlist/authority-minting 0을 먼저 RED/GREEN하고 A2-B accepted된 뒤 T3를 연다.
- config keys/defaults/invalid-block refusal, stable alert identity/episode, `/position-management` shared
  projection을 normative spec에 고정했다.

current-main read-only baseline은 `protectionlifecycle` 83.4%, `journal` 75.1%, `app/engine` 63.7%,
`cmd/tossctl` 55.2%로 모두 PASS했다. 과거 a099 RED blocker는 current main의 사실이 아니다.
R0 evidence는 기존 경계와 `engineRuntime`, `Gateway.adapt`, `TrackedFillOrders`,
`confirmedFillOwners`, `resolveFillOrigin`, auxiliary stop seam, lifecycle API guard를 current source SHA로
관리한다. `filldetect`는 계속 편집하지 않고 `ProjectPosition`은 citation-only/no-fallback이다.

Manager는 OpenSpec/분석/최종 증거만 소유한다. 각 Terra 구현 로트는 별도 Terra 적대 리뷰어가
검토하고 구현자 수정 뒤 같은 리뷰어의 `P0=0, P1=0`과 Manager 문서 acceptance를 받아야 다음
로트로 간다. 전체 구현 후 gstack review, `make sdd-sync`, `make sdd-check`,
`make gate CHANGE=a100-wire-fill-to-broker-protection`를 실행하고 Manager가 완료 여부를 대조한다.

### R0 최종 독립 재검토 영수증

2026-08-15 A0 Terra가 최신 shared 상태를 처음부터 다시 읽고 다음을 독립 재실행했다.

- HEAD/base `882a0b490b0b6d2eb7abe5c5040c514776f49f3e` 일치
- strict OpenSpec validate PASS, `check_analysis.py` PASS, `git diff --check` PASS
- current `ast.json` 17개의 source SHA-256 전부 현재 소스와 일치
- `make sdd-sync`는 hard index를 갱신하고 `all indexes current`로 종료(GBrain busy advisory만 유지)
- T4 실행 순서는 **T4-A(4.5~4.6 authority/exit/child) accepted → T4-B(3.9·4.1~4.4 runtime wiring)**이며,
  그 전에는 worker construction/start 및 broker mutation 도달 경로를 열 수 없음

최종 판정은 **ACCEPT — P0=0, P1=0, P2=0**이다. R0만 닫혔으며 구현 착수 권위는 아니다.
남은 hard gate는 사람의 주문별 승인을 받은 M-A complete PASS, parent child-id local receipt의
child fill local receipt 엄격 선행, 그리고 그 결과로 동결한 0.11 raw-status 판정표다.

### 2026-08-15 M-A 승인 후 read-only preflight와 M0 결정

사용자가 M-A를 승인했지만 Terra preflight와 별도 A-M 적대 감사는 **HOLD/NO-GO**로 판정했고
주문·preview·config·engine·Docker·journal mutation은 0건이었다.

- 시점은 토요일 장외였고 다음 KR 세션 직전 재검증이 필요하다.
- official OPEN conditional은 마스킹 기준 `PAUSED` 2건·`WATCHING` 3건이었다. 기존 주문은 M-A
  cleanup과 섞지 않는다.
- 046890/466100은 이미 OPEN/adopted 관리 상태이고 1주 108490에는 PAUSED conditional이 있어
  fresh 비관리 최소위험 후보가 없었다. 관리 보유 1주 사용은 자동 선택하지 않고 별도 사람 결정이다.
- official parent read는 가능했으나 WTS child surface는 invalid였고 installed/deployed auth·soak
  authority가 갈렸다. 당일 fresh authority가 아니면 place 금지다.
- 결정적 P0는 현 verify record가 parent/child successful response receipt의 monotonic sequence,
  parent fsync → child request barrier와 raw payload digest를 남기지 않는다는 점이었다. shell wrapper와
  wall/server timestamp는 process 내부 response 경계를 증명하지 못한다.

따라서 다른 change를 만들지 않고 M0(tasks 0.2a)를 A100에 추가했다. 기존 human-confirmed
`verify run --include-trigger --confirm-each`의 trigger step에만 official raw causal receipt를 추가하고,
runtime·engine·protection·attestation·journal에는 연결하지 않는다. Terra-M0가 RED/GREEN을 구현하고
별도 A-M0가 `P0=0/P1=0`으로 재검토하기 전에는 승인 상태와 무관하게 M-A place를 열지 않는다.

### 2026-08-15 A-M0 1차 설계 리뷰 — REJECT 후 보강

A-M0는 transport body-read 경계가 caller에 노출되지 않음, `readRetry`가 429를 성공으로 덮을 수 있음,
create 직후 crash 시 exact parent cleanup identity가 durable하지 않음, confirm-each의 replay 예외,
parent/child exact identity 필드 부족을 P0로 판정했다. 구현을 열지 않고 같은 A100의 M0 계약을 다음처럼
보강했다.

- M0는 exact trigger-only resume/redo mode이며 ttl-edge·다른 redo를 pre-broker 거부한다.
- official transport가 every-attempt body-read-complete/status/raw-result를 decode 전에 제공한다.
- parent/child exact ID는 causal receipt가 아니라 owner-only verify cleanup checkpoint에 즉시 fsync한다.
- durable parent child-ID receipt부터 durable child fill까지의 첫 401/429/read/decode/identity gap은 irreversible
  HOLD이며 retry 성공이 지우지 못한다.
- exact raw field schema/tag equality, file+directory fsync, symlink/owner/mode 및 kill-point RED를 추가했다.

이 보강은 아직 구현 착수 권위가 아니다. 같은 A-M0가 문서 재검토에서 `P0=0/P1=0`을 확인해야 한다.

2차 재검토는 create response→checkpoint 사이의 unavoidable crash window와 `Runner.Run` cleanup
prologue를 추가 P0로 찾았다. 이에 M0-only pre-create pending client intent를 owner-only verify record에
먼저 fsync하고, crash resume은 official all-page unique match를 checkpoint한 뒤 mutation 없이 HOLD하도록
정했다. parent raw child ID는 exact child checkpoint를 먼저 fsync하고 그 뒤 causal receipt를 fsync한다.
ordinary outstanding가 있으면 cleanup prologue/broker factory 전에 거부하며, 401도 critical gap이고
parent/child equality는 서로 다른 두 ID group으로 분리했다. 이 2차 보강 역시 같은 A-M0의 재수용 전에는
구현 권위가 아니다.

### 2026-08-15 A-M0 최종 문서 재검토 — ACCEPT

동일 A-M0 reviewer가 최신 문서·current source seam과 strict validation을 다시 대조했다. 결론은
**P0=0, P1=0**이며 Terra-M0 구현 착수를 승인했다. 승인 범위는 measurement-only M0뿐이다.
구현 뒤 실제 diff/RED-GREEN/kill-point mutation을 별도 Terra adversary가 재검토하고 P0/P1을 0으로
닫기 전에는 M-A place를 실행하지 않는다.

- strict OpenSpec: 93/93 PASS
- `check_analysis.py --change a100-wire-fill-to-broker-protection`: PASS
- `git diff --check`: PASS

### 2026-08-15 M0 구현 독립 리뷰와 Manager acceptance

Terra 구현을 core/fixture와 receipt 하위 소유권으로 분리했고, 각 구현과 분리된 Terra adversary가
실제 diff를 공격했다. 최초 core 리뷰는 재시작 중복 POST, arbitrary Broker+trace 위조, parent checkpoint
writer 실패의 ordinary FAIL, exact-mode full catalogue 실행, pending recovery 불가, 거짓 zero-fill PASS를
찾았다. receipt 리뷰는 parent path TOCTOU, concurrent/duplicate lease, arbitrary append, persistence 실패 뒤
재사용, symlink ancestor, empty-body digest 누락을 찾았다. 구현자가 RED-first로 수정한 뒤 같은 reviewer가
재검토했다.

- A-M0 core 최종: **ACCEPT, P0=0, P1=0**. same concrete `official.Client`가 mutation과 raw+attempt를
  함께 소유하고, exact prerequisites/pending recovery/manual HOLD/terminal checkpoint failure/causal ordering을
  고정했다.
- A-M0-R receipt 최종: **ACCEPT, P0=0, P1=0, P2=0**. 모든 path component dirfd/no-follow, leaf O_EXCL,
  file→same-dirfd fsync, exclusive run lease, lease-only typed writer, poison latch와 empty-body digest를 확인했다.
- core reviewer P2 하나는 첫 trace seam의 독립 재생 가능한 pre-GREEN commit 영수증이 없다는 과정상
  한계다. 구현 결과 테스트와 mutation 대응은 검증됐으나 역사적 RED 순서를 새로 만들어내지 않고 그대로
  기록한다.

Manager 독립 재실행:

- `go test ./internal/official ./internal/verifylive ./cmd/tossctl -count=1`: PASS
- 같은 3 package `-race`: PASS
- scoped `go vet`, `git diff --check`, strict OpenSpec validate: PASS
- `check_analysis.py --change a100-wire-fill-to-broker-protection`: PASS

따라서 tasks 0.2a와 8.0만 완료로 인정한다. M-A 실주문은 별도의 fresh session preflight와 주문별 사람
승인이 필요하고 현재 장외/후보·residual 상태 때문에 여전히 HOLD다. M-A complete PASS와 0.11 raw-status
판정표 전에는 T1을 열지 않는다.

### 2026-08-15 M-A 다음 진행 재검토 — HOLD 유지

M0 acceptance 뒤 Terra preflight와 구현자와 분리된 A-M Terra가 current state를 read-only로 다시
감사했다. 두 결과는 일치했다.

- 시장 폐장, OPEN conditional 5건(`PAUSED` 2/`WATCHING` 3), 후보 보유의 managed/residual 겹침,
  incomplete sellable receipt, `trading.conditional=false`, stale soak/authority 때문에 실제 주문을
  열 수 없다.
- PATH 설치 binary는 M0 receipt flag가 없는 2026-07-30 build다. reviewed source에서 임시 candidate를
  만들어 source 36건·SHA·help·local tests를 봉인했지만, `commit: unknown` dirty-source artifact이고
  설치되지 않아 장중 execution authority가 아니다.
- 사용자의 일반적인 진행 승인은 구체적인 symbol·trigger·expiry·1주·confirm token을 보여 준 뒤의
  주문별 승인으로 대체할 수 없다.

판정은 **A-M HOLD — P0: 실행 가능한 동일성 있는 M0 설치 artifact 부재와 시장 폐장;
P1: residual/후보·sellable·auth/rate·soak·conditional gate 미완결**이다. preview/place/modify/cancel,
verify resume, config/lifecycle mutation은 0건이었다. tasks 0.2와 0.11은 계속 미완료이고 T1은 열지 않는다.

### 2026-08-15 M0 release artifact 재검토 — ACCEPT, 설치는 별도 gate

Terra 구현 담당은 기존 `commit: unknown` 후보를 설치하지 않고, current reviewed source를 완전히
봉인한 `/tmp/a100-m0-artifact.M8EeNi`를 만들었다. exact HEAD, binary-safe tracked patch, 206개
untracked input, 14,540-entry source archive와 NUL-safe manifest, Go toolchain/build command/ldflags,
candidate self-report, M0 help와 focused test evidence를 포함한다. 별도 source extraction 두 곳의
binary SHA는 모두
`899a74aca0d72411de80bb92fe89cfbdf31948dbf008635a1c4bb7125ddae882`로 같았다.

Manager가 1차 대조에서 top-level manifest 부재를 찾아 구현자에게 돌려보냈다. 구현자는 receipt의
immutable payload를 self-excluding `SHA256SUMS`로 다시 봉인했고 Manager와 별도 Terra adversary가
각각 50/50 PASS를 재현했다. adversary 최종 판정은 **P0=0, P1=0, P2=0**이다. source archive path
set과 NUL manifest, embedded version/commit/date, installed preimage와 same-filesystem atomic
install/rollback 계획도 독립 일치했다.

이 acceptance는 artifact provenance와 향후 원복 절차에만 해당한다. `/home/daniel/.local/bin/tossctl`
은 여전히 SHA `b0e805f3…da96f9`의 기존 binary이며 실제 install, runtime/config/Broker/verify mutation은
0건이다. 설치는 별도 명시 승인과 fresh preimage 확인이 필요하고, 설치 승인은 주문별 LIVE 승인을
대체하지 않는다. 시장 폐장과 residual/후보/auth/rate/soak gate도 그대로이므로 M-A, 0.11, T1은
계속 HOLD다.

### 2026-08-15 M0 atomic install 리뷰 — ACCEPT, live scope는 계속 HOLD

후속 승인 범위를 CLI 한 파일의 원자 설치로 한정했다. Terra 설치 담당은 candidate 50/50 seal과
기존 PATH preimage를 다시 확인하고 same-directory backup·staged candidate를 각각 fsync한 뒤 단일
atomic rename과 directory fsync를 수행했다. live target은 candidate SHA `899a74ac…e882`, backup은
preimage SHA `b0e805f3…96f9`이고 둘 다 regular non-symlink `0755 daniel:daniel`, 같은 filesystem이다.
postcheck는 isolated HOME의 identity/help뿐이었고 rollback은 필요하지 않았다.

별도 reviewer의 1차 판정은 install integrity `P0=0`이었지만, receipt 직후 계속 갱신된 token/marker/
log/WAL을 사전 snapshot하지 않은 비간섭성 P1을 남겼다. 별도 Terra가 read-only attribution receipt를
봉인했다. installation보다 수시간 먼저 기동한 동일 engine PID와 containers(restart 0), 설치 직전부터
이후까지 이어진 `reconcile.clean` cadence, 설치 전 config hash/mtime, audit/order/verify mutation 0건이
기존 엔진의 자연 쓰기를 강하게 증명했다. 같은 reviewer 재검토는 **P0=0, P1=0, P2=1**로 P1을
폐쇄했다. P2는 historical 개별 syscall owner를 사후 확정할 수 없다는 비차단 한계다.

따라서 installed M0 identity gate는 ACCEPT다. 그러나 이 설치는 `verify run --include-trigger`/`--resume`,
fresh Broker preflight, 주문, config flip, Docker/engine action을 승인하지 않는다. M-A와 0.11, 제품 lot은
계속 시작 금지다.

### 2026-08-15 fresh M-A read-only preflight 리뷰 — evidence ACCEPT, live HOLD

후속 승인은 Broker/local read-only preflight에만 한정했다. Terra 수집 담당은 installed M0 identity,
market/auth/rate, official conditional OPEN, holdings/sellable, local verify/config, soak/attestation을 읽고
어떤 preview·주문·verify run/resume·config·lifecycle mutation도 실행하지 않았다.

별도 Terra adversary는 최초 receipt의 “pagination incomplete”를 P1로 반려했다. 같은 시점의 독립
forced-OpenAPI `--status OPEN --limit 100` read는 `count=5`, `has_next=false`, cursor 없음,
`PAUSED` 2·`WATCHING` 3의 terminal 결과였다. 수집 담당은 추가 Broker 호출 없이 그 독립 증거를
귀속한 `/tmp/a100-ma-preflight-corrected.kSn14D`를 새로 봉인했고, 같은 reviewer가 mode/owner,
3/3 hash, redaction과 pagination 정정을 재검토했다. 최종 receipt verdict는
**ACCEPT — P0=0, P1=0, P2=1**이다. P2는 host process survey의 runtime 귀속 한계다.

그러나 운영 M-A는 **HOLD/NO-GO**다. KR 폐장, 기존 OPEN 5건의 사람별 retain/cancel 결정 부재,
official sellable 불완전과 미관리 whole-share KR 후보 부재, soak `ready=false`/streak 0/window 0,
attestation binary-unbound, `conditional-trigger=deferred`, `trading.conditional=false`가 남았다.
pagination 정정은 이 gate들을 닫지 않는다. tasks 0.2/0.11과 T1은 계속 미완료이며 다음 live action,
preview 또는 toggle은 허용되지 않는다.

### 2026-08-15 R0/M0 selective source landing review

사용자는 시장 폐장과 독립적인 R0 current-main evidence와 M0 measurement tooling만 별도 랜딩하도록
승인했다. 선별 diff는 `cmd/tossctl/verify*`, M0 read-only `internal/official`, `internal/verifylive`, 이 change의
분석·명세로 한정하며 engine/runtime product worker, journal schema, protection lifecycle, Docker/compose와
A112/PM 파일을 포함하지 않는다. 이 landing은 A100 완료·archive·배포가 아니며 0.2 M-A, 0.11과 T1 이후
checkbox를 바꾸지 않는다.

두 Terra landing reviewer와 기존 M0/receipt adversary는 same official-client binding, exact CLI mode,
pending recovery/manual HOLD, dirfd/no-follow/O_EXCL/fsync/exclusive lease/poison, causal child barrier를 다시
검증해 P0=0/P1=0을 보고했다. GitHub full test의 최초 attempt에서 기존 execgw test가 닫힌 임시 port의
재사용 불가능성을 가정해 실제 403을 받은 flaky failure가 발생했다. Terra 구현자는 테스트만 deterministic
pre-write RoundTripper로 바꾸고 정확히 `POST /oauth2/token` 한 번만 transport에 도달하며 order POST는
도달하지 않음을 고정했다. 별도 reviewer 재검토는 P0=0/P1=0/P2=0이고, 후속 GitHub CI는 PASS다.

전체 `make gate CHANGE=a100-wire-fill-to-broker-protection`의 task-completeness 단계는 의도대로 실패한다.
그 gate는 A100 전체 완료를 판정하므로 이 부분 랜딩에서 통과했다고 주장하거나 checkbox를 조작하지 않는다.
부분 랜딩의 hard gate는 scoped tests/race/vet, strict OpenSpec, logic-map analysis, current CodeGraph
`make sdd-check`, exact staged-path audit, final gstack P0/P1=0과 GitHub CI다. A100 전체 gate는 모든 제품
lot이 실제 완료될 때만 통과시킨다.

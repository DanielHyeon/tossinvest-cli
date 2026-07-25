# Review: add-core-domain (proposal-freeze)

리뷰 시점: 구현 착수 전(proposal-freeze). 보이스 3개 독립 투입 — codex CLI(ENG/SAFETY/SPEC-QUALITY 렌즈, 22건), Claude Eng 적대 리뷰(P1 코드 대조, 13건), Claude CEO+DX(범위·실행가능성, 10건). 총 45건.

Manager는 각 보이스의 구조적 주장을 **P1 코드에서 직접 재검증**했다. 아래 "코드 근거"는 그 검증 결과이며, 리뷰어 주장을 그대로 옮긴 것이 아니다.

## 결론 요약

**이 change는 구현 착수 불가 판정.** 두 개의 change로 분리하고 범위를 재구성한다.

리뷰의 핵심 발견은 개별 결함이 아니라 **하나의 구조적 오류**다: 이 change는 무인 자동매매의 안전 근거 전체를 "네이티브 조건주문이 브로커에 상주한다"(design D2)에 걸었는데, **조건주문 경로는 P1이 만든 안전 아키텍처 바깥에 통째로 있다.** Gateway에도, attestation에도, 체결 귀속 모델에도 없다. 세 보이스가 서로 다른 각도에서 독립적으로 같은 지점에 도달했다.

## A. 조건주문 경로가 P1 안전 아키텍처 밖에 있다 (3/3 보이스)

| 주장 | 보이스 | 코드 근거 (Manager 검증) | 판정 |
|---|---|---|---|
| A1. 보호주문 mutation이 ExecutionGateway를 우회 | codex#1, Eng#1 | `trading.Service.ConditionalPlace` (internal/trading/conditional.go:123) = 확인 토큰 검사 → `s.conditional.CreateConditionalOrder` 직접 호출. `internal/trading`·`internal/official`은 execgw를 import하지 않음. journal 선기록·MutationAttempt·nonce·IN_DOUBT·심볼 in-flight 전부 없음 | **확인 · 수용** |
| A2. 보호 제출의 응답 유실 crash window 미정의 | codex#2, Eng#1 | A1의 귀결. saga는 "에러"만 보고 재제출 → 브로커측 stop 중복 → 발동 시 oversell | **확인 · 수용** |
| A3. **브로커측 조건주문 체결이 포지션·체결 귀속에서 보이지 않는다** | Eng#2 | `journal.NetPositions` (internal/journal/fills.go:457-463) SQL이 `JOIN intents i ON i.id = a.intent_id`. 주석 자체가 "An order with no local intent contributes nothing: that is the definition of an external order". 조건주문 발동은 로컬 intent 없는 신규 브로커 주문 → **외부 주문으로 분류 → 포지션이 영원히 CLOSED에 도달하지 못함** → reconcile 영구 불일치 → 성과 미기록 → saga는 계속 보호가 필요하다고 믿음 | **확인 · 수용 (최중대)** |
| A4. attestation이 조건주문 능력을 증명하지 않은 채 게이트 ON 가능 | codex#13, CEO/DX F1 | `engine.RequiredEndpoints()` (internal/app/engine/interlock.go:81-91)에 조건주문 엔드포인트 없음. verify-execution-capability도 place/cancel/조회만 검증. 즉 손절 메커니즘이 **한 번도 실증되지 않은 채** 완전히 유효한 attestation으로 자동매매가 켜질 수 있다 | **확인 · 수용** |
| A5. 조건주문 안전 속성을 API 존재만으로 가정 | codex#13 | 프로세스 사망 후 존속·시장별 지원·트리거 기준가·정규장 밖 동작·expiry·OCO sibling 취소·부분체결 잔량·보유수량 예약 의미 중 어느 것도 코드가 보증하지 않음. design.md는 이를 Risks에 적었을 뿐 활성화 전제조건으로 만들지 않음 | **확인 · 수용** |
| A6. 심볼 in-flight 규칙이 보호 즉시성과 충돌 | codex#9, Eng#1 | `gateway.go:333` "One in-flight mutation per symbol". 보호를 Gateway에 올리면 **진입 IN_DOUBT가 같은 심볼의 손절 제출을 막는다**(§0.3 위반). 우회하면 단일 Gateway 계약이 깨진다 | **확인 · 수용** |

A3가 이 리뷰의 가장 깊은 발견이다. D2는 "브로커 상주라서 안전하다"를 근거로 네이티브 조건주문을 택했는데, P1의 체결 귀속 모델은 "로컬 intent 없는 브로커 주문 = 외부 주문"으로 정의한다. **두 설계 결정이 구조적으로 양립하지 않는다.** 조건주문을 쓰려면 발동 주문을 lineage에 편입하는 경로를 먼저 만들어야 한다.

## B. Guardian·게이트 계약의 fail-open 구멍 (3/3 보이스)

| 주장 | 보이스 | 코드 근거 | 판정 |
|---|---|---|---|
| B1. 한도 부분 미설정으로 기동 가능 | codex#6 | `Limits.IsZero()` (execgw/guardian.go:62) = `MaxQuantity<=0 **&&** MaxNotional<=0`, 개별 검사는 `if l.MaxQuantity > 0 &&` → 0 = "검사 안 함". `interlock.go:203`은 **둘 다** 0일 때만 거부. 총 노출·일일 손실 한도는 `Limits`에 존재조차 안 함 | **확인 · 수용** |
| B2. **게이트 ON이 청산·보호 경로를 보장하지 않는다** | Eng#6 | `verifyGate` (interlock.go:194-229) 6단계 중 trading policy를 보는 단계가 없음. `policy.Sell=false` 또는 `policy.Conditional=false`여도 게이트는 ON+verified. → 엔진이 **매수는 하는데 손절도 청산도 못 하는 naked long** | **확인 · 수용** |
| B3. 청산 결정에 한도 스냅샷을 찍으면 큰 청산이 거부된다 | Eng#7 | `verifyLimits` (guardian.go:180-208)는 cancel만 면제, 빈 스냅샷은 노출 미증가 mutation만 통과. 비어있지 않은 스냅샷은 **SELL에도 적용** → `MaxOrderQuantity` 초과 청산이 `ReasonGuardianLimitExceeded`로 거부(§0.3 위반). `flatten.decisionFor` (flatten/flatten.go:625-633)가 이미 빈 Limits 패턴을 쓰지만 스펙이 이 발급자 계약을 명시하지 않음 | **확인 · 수용** |
| B4. 총계 한도가 제출 시점에 강제 불가 — 동시 결정이 합산 한도를 뚫는다 | codex#4, Eng#5 | in-flight 락은 **심볼 단위**(gateway.go:341). 서로 다른 심볼 두 건이 5초 창 안에서 같은 노출·현금·일손실 스냅샷을 각각 통과. 체결 감지는 비동기(3s 폴링) | **확인 · 수용** |
| B5. GuardianDecision이 위험 판정 입력에 결합되지 않음 | codex#5 | `AuthorizationRequest`에 stop·target·RR·equity·기존 노출 없음. `Limits`는 수량·notional·통화뿐. "매수 10주" 승인은 손절 존재 여부를 아무것도 말하지 않음 | **확인 · 수용** |
| B6. 인터록의 config 한도와 Guardian이 찍는 한도가 서로 다른 객체 | Eng#11 | `verifyGate`는 `gate.MaxOrder*`를 검사(interlock.go:203), 결정 `Limits`는 주입된 Guardian이 독립적으로 생성. 둘을 잇는 것이 없음 | **확인 · 수용** |
| B7. NonceStore가 in-memory | Eng#10, CEO/DX F7 | `NewMemoryNonceStore` (guardian.go:124) 주석이 직접 명시: "Phase 2, where decisions may outlive a process, is expected to back this with the journal database". 태스크 2.4에 누락 | **확인 · 수용** |
| B8. 비상 차단까지 사람 승인에 묶임 | codex#15 | 스펙 결함(코드 아님). 손실 한도·credential 실패 시 즉시 강화되어야 할 전환이 승인 대기 | **확인 · 수용** |

B2가 특히 아프다. "No Stop = No Trade"를 스펙 최상단에 써놓고, 정작 **손절을 낼 수 있는지를 게이트가 확인하지 않는다.**

## C. Position 권위 모순 (2/3 보이스)

| 주장 | 보이스 | 판정 |
|---|---|---|
| C1. "체결에서만 파생" vs "reconcile 시 계좌 값 우선"이 서로 모순이고 배선도 없음 | codex#12, Eng#3 | **수용** — `QuantityMismatch.Authority()`가 브로커 값을 돌려주지만 아무도 쓰기로 소비하지 않음(mismatch는 진입 차단만). 불일치 시 포지션이 영구히 틀린 채 유지되고, §0.3으로 청산은 안 막히므로 과대 수량 기준 청산 = oversell |
| C2. 포지션 진실이 두 개 (`reconcile.LocalStateFromJournal` vs 신규 `internal/position`) | Eng#4 | **수용** — 단일 권위 명시 필요. 체결 delta는 부호 없음(`Applied.Delta >= 0`)이라 상태기계가 intent side에서 부호를 재도출해야 함 |
| C3. 상태기계가 누적 스냅샷에서 결정적으로 파생되지 않음 | codex#10 | **수용** — 즉시 전량체결 전이, OPENING 종료 판단, SCALING 진입·종료, 다중 주문 lineage, 외부 포지션, CLOSED→다음 FLAT 관계 전부 미정의. 완전한 전이표 + position-instance ID 필요 |
| C4. 체결 반영·Position 갱신·보호 enqueue가 원자적 커밋이라는 요구 없음 | codex#11 | **수용** |

## D. 보호 saga 상태·정정 규칙 (2/3 보이스)

| 주장 | 보이스 | 판정 |
|---|---|---|
| D1. "보호 수량은 항상 체결 수량과 일치" + "취소 후 재등록"은 동시 체결 환경에서 구현 불가 | codex#8 | **수용(수정)** — "항상 일치"를 측정 가능한 transient invariant + 최대 허용 시간으로 대체 |
| D2. 원자적 modify가 있는데 스펙이 취소-후-재등록을 강제 | Eng#8 | **수용** — `official.ModifyConditionalOrder` (conditional_writes.go:63)가 존재. 취소-후-재등록은 네이티브 OCO를 택한 이유인 무보호 창을 스스로 만든다. modify 원자성은 A5의 capability 검증 항목으로 |
| D3. saga 상태 전이표 부재 | codex#19 | **수용** — ACTIVE 외 상태·허용 mutation·timeout·재시작 행동·crash point 매트릭스 필요 |
| D4. synthetic 폴백 전환 조건이 "거부"뿐 | codex#14 | **수용** — 명시적 unsupported 응답만 폴백 허용. 모호한 결과 뒤 synthetic 추가 = 이중 청산 |
| D5. HALT_ALL과 No Stop = No Trade가 양립 불가 | codex#3 | **수용** — HALT_ALL에서 늦게 감지된 부분체결·재시작 시 발견된 미보호 포지션에 신규 stop을 낼 수 없다. 보호 생성·증량과 reduce-only 청산은 별도 safety class로 |

## E. 분석 지표 (3/3 보이스 만장일치)

| 주장 | 보이스 | 판정 |
|---|---|---|
| E1. **MFE/MAE의 데이터 소스가 존재하지 않는다** | codex#20, CEO/DX F2, Eng#9 | **수용 — P3로 삭제** |

세 보이스가 독립적으로 같은 결론. `filldetect`는 주문의 `AveragePrice`(체결가)만 관측하고 **보유 심볼의 시장가 시계열을 만들지 않는다**. design D5와 trade-analytics 스펙이 인용한 "filldetect 가격 관측 스냅샷"은 실재하지 않는 소스다. 태스크 3.1은 테이블만 만들고 채우는 주체가 없다. 신규 가격 폴러를 filldetect에 넣으면 §0.4 rate budget과 주문 감지 SLO를 놓고 경쟁한다.

**Manager 결정: MFE/MAE를 이번 change에서 삭제하고 P3(시세 스트림)로 이관한다.** 실현손익·R 배수·보유 시간만 남긴다 — P3 배분 로직이 시작하는 데 필요한 것은 그것이다. 없는 소스를 근거로 스펙을 쓰는 것보다 범위를 줄이는 쪽이 정직하다. (사용자가 MFE/MAE를 P2에 유지하길 원하면 전용 best-effort 시세 수집기를 별도 예산·테이블로 추가하는 안이 대안이다.)

| E2. analytics 계산·retention이 주문 경로에서 격리되지 않음 | codex#21 | **수용** — close 이벤트만 원자적으로 기록, 집계·retention은 outbox 비동기 |

## F. 스펙 품질·앵커링

| 주장 | 보이스 | 판정 |
|---|---|---|
| F1. 일일 손실·총 노출의 계산 계약 부재 | codex#16 | **수용** — 권위 데이터·통화 정규화·거래일 경계(시장별 DST)·실현/미실현·미체결 예약 포함 여부·stale 시 fail-closed |
| F2. long-only 가정이 명시되지 않음 | codex#17 | **수용** — TossOS는 long-only, SELL은 보유수량 이하 reduce-only, short 구조적 금지를 명시 |
| F3. journal "v5+"가 migration 산출물로 불충분 | codex#18 | **수용** — 버전별 immutable migration·스키마 계약 테스트·백업/복구를 태스크 분리. 또한 design.md의 "롤백 = 구버전 바이너리"는 **롤백이 아니다**(ErrSchemaTooNew로 기동 거부) — 백업/복구 절차로 교체 |
| F4. StockOS 이식 범위가 과소 명시 | CEO/DX F8 | **수용** — 경로 읽기 가능·인용 테스트 수 전부 정확(guardian 20·target_stop 29·a090 36·structural_rr 14) 확인됨. 그러나 `guardian.py` 714줄에 제외 목록에 없는 분기(레버리지/인버스, ETF/ETN, 미국장 진입 시간창, 당일 재진입 쿨다운)가 있어 과다·과소 이식 위험. 태스크에 in-scope/제외를 열거하고 절대 경로를 선행조건으로 |
| F5. 게이트 인터록이 두 곳에서 다른 조건으로 규정됨 | CEO/DX F9 | **수용** — engine-safety 인터록을 MODIFIED로 선언 |
| F6. 태스크 2.5·4.2·3.2·4.1이 각각 하루 이상 + 설계 결정을 숨김 | Eng#13 | **수용** — 특히 2.5는 **엔진 측 Gateway 배선 전체**를 숨긴다: `execgw.New`는 `cmd/tossctl/flatten.go:221`에만 있고 엔진 Context에는 Gateway 필드조차 없다 |
| F7. StockOS 상수가 annotated-but-active | Eng#12 | **수용** — 등급 배수·비용 bps 열거, 비용은 과대 추정이 보수 방향, "검증됨" 전환은 verify/tracer 결과에 결속 |
| F8. tracer 태스크가 계좌·시장·notional 상한·중단 기준 미정 | codex#22 | **수용** — 구현 검증과 live capability attestation을 분리, live 미완료 시 게이트 OFF 유지가 명시적 산출물 |
| F9. 태스크 2.1·2.4·4.1에 파일 앵커 부족 | CEO/DX F6·F7 | **수용** — `internal/trading/conditional.go`, `internal/app/engine/broker.go`, `internal/execgw/guardian.go` 인용 |

## 반영하지 않는 주장

| 주장 | 보이스 | 사유 |
|---|---|---|
| 게이트 활성화(2.5)를 별도 change로 분리 | CEO/DX F3이 **분리 불필요** 판정 | 동의. 2.5는 배선과 "미충족 조합 기동 거부" 통합 테스트이고, 게이트는 기본 OFF·flip은 사람 승인 + 별도 verify change의 attestation을 요구하므로 병합 자체가 돈을 움직이지 못한다. 단 A4가 고쳐진다는 전제 하에서만 성립 |
| tracer 코드 now / 실전 later 분리가 잘못됨 | 없음 (CEO/DX F4가 **옳다** 판정) | 유지. httptest로 코드 경로를 증명하고 실전은 사람 승인 트랙으로 |
| 운영 모드 4축·provenance lineage가 과설계 | CEO/DX F5가 **과설계 아님** 판정 | 유지. EXIT_ONLY와 HALT_ALL은 실제로 다른 운영 상태이고, lineage 단일 질의는 §0.5 감사 요구 |

## Manager 구조 결정: change 2분할

A·B 클러스터(14건)는 전부 **"P1의 실행 계약 자체를 확장해야 보호주문 자동화가 가능하다"**로 수렴한다. 이것은 새 capability 추가가 아니라 기존 메인 스펙(order-execution·engine-safety) 수정이고, Guardian 판정 로직과는 독립적으로 구현·검증할 수 있다. 하나의 change로 묶으면 40개 넘는 태스크에 게이트 하나가 걸리고, 가장 위험한 배관이 도메인 로직에 섞여 리뷰된다.

**분할:**

1. **`extend-execution-contract`** (선행) — 강제 장치. 조건주문을 Gateway mutation으로 승격(journal·attempt·IN_DOUBT·fingerprint), 발동 주문의 lineage 편입(A3), mutation safety class와 클래스별 직렬화(A6·D5), RiskIntent 해시 결합(B5), 한도 fail-closed 비트와 총계 한도(B1), 청산 발급자 계약(B3), 원자적 위험 예약(B4), journal NonceStore(B7), 게이트 전제조건에 trading policy·단일 한도 출처 추가(B2·B6), `RequiredEndpoints()`와 attestation에 조건주문 속성 추가(A4·A5). **MODIFIED**: order-execution, engine-safety.

2. **`add-core-domain`** (후행, 재범위) — 판단 정책. 비용 모델, 사이징·손절·RR, Guardian 판정 체인, 운영 모드·kill switch 우선순위표, 포지션 원장(전이표·조정 이벤트·단일 권위), 보호 saga(상태표·정정 규칙), provenance lineage, 성과(실현손익·R·보유시간), tracer.

경계 원칙: **change 1 = 실패해도 안전한 레일(강제), change 2 = 그 레일 위의 판단(정책).** change 1은 P1이 오늘 하듯 합성 결정으로 완전히 테스트 가능하다.

3. **`verify-execution-capability`** (기존 진행 중 change) — attestation 계약에 조건주문 등록·트리거 관측·modify 원자성·시장별 지원을 추가한다(A4·A5·D2). 이 검증이 끝나기 전에는 자동 **진입**이 불가하다는 것을 명시적 산출물로.

## 후속

- change 1·2 각각 별도 proposal-freeze 리뷰를 받는다(본 리뷰는 분할 전 범위에 대한 것).
- E1(MFE/MAE 삭제)은 사용자가 뒤집을 수 있는 결정으로 남긴다.
- 운영 파라미터 수치는 여전히 사용자 확정 대기 — 미확정 시 small_live 보수 기본값.

---

# Review 2라운드 (재범위본)

선행 change `extend-execution-contract`가 리뷰 후 재작성되면서 계약이 크게 바뀌었는데, **이 change의 스펙은 옛 계약을 기준으로 동결된 채 갱신되지 않았다.** 그 정합성 검토를 포함한 2라운드.

## 판정

**착수 불가.** 두 부류다 — (1) 선행 계약 변경이 전파되지 않은 정합성 부채, (2) 판단 계층 자체의 안전 구멍.

## A. 선행 계약과의 불일치 (전파 누락 — Manager 책임)

| # | 내용 | 판정 |
|---|---|---|
| A1 | **HALT_ALL이 여전히 모든 취소를 면제한다** — risk-management 스펙이 "위험 감소 mutation(…, **취소**)은 계속 허용"이라 쓰는데, 선행 계약은 활성 보호의 취소를 PROTECTION_WEAKENING으로 분류하고 HALT_ALL에서 금지한다. **재작성이 닫은 fail-open을 정책 계층에서 되연다** | **확인 · 최중대** |
| A2 | **기각된 단건 `min(로컬, 계좌)` 상한이 네 곳에 살아 있다** (design.md, position-ledger, tasks 5.4, protection-execution). 선행 계약은 이를 원자적 청산 예약으로 대체하고 로컬 파생을 상한 근거로 쓰는 것을 명시적으로 기각했다. protection-execution의 "보호 수량 합계 ≤ 보유"는 예약이 이미 소유한 것의 두 번째 구현 | **확인 · 최중대** |
| A3 | 결정 발급 계약이 재작성 이전 것 — safety class가 없고 "빈 한도 스냅샷"을 위험 감소의 표지로 쓴다(선행 계약이 그 표지로는 구별 불가하다고 판정한 바로 그것). **RiskIntent preimage를 쓰는 주체가 어느 change에도 없다** | **확인 · 수용** |
| A4 | **PROTECTION_WEAKENING 발급 경로가 두 change 어디에도 없다** — saga는 보호 수량 축소와 CLOSED 시 잔여 취소를 해야 하는데 그것이 그 클래스다. 정책 계층이 발급하지 않는 클래스의 판단을 소유한다 | **확인 · 수용** |
| A5 | 선행 리뷰 C1이 해결한 "단일 한도 출처" 모순이 이 change 안에서 재발(무조건 동등성 요구 vs 위험 감소의 빈 스냅샷) | **확인 · 수용** |
| A6 | **다섯 영역이 양쪽 change에 중복 소유**(총계 한도 계산 계약·게이트 배선·단일 출처 동등성·§0.3 회귀·예약 race 테스트), **세 영역이 어느 쪽도 소유하지 않음**(발동 주문 체결의 방향, flatten의 결정 발급, flatten의 조건주문 취소) | **확인 · 수용** |
| A7 | saga의 "재제출 금지" 문구가 이제 **1차 해소 절차를 금지한다** — 선행 계약의 첫 단계가 동일 키·본문 재요청인데, 이 문구는 보호주문이라는 가장 위험한 클래스에서 멱등 재생을 금지하는 것으로 읽힌다 | **확인 · 수용** |
| A8 | "방향은 intent side에서 재도출"이 발동 주문에 적용 불가 — 포지션을 닫는 바로 그 체결(브로커측 stop 발동)에 intent가 없다. 방향은 조건주문 leg·예상 주문에서 와야 하는데 어느 change에도 그 요구가 없다 | **확인 · 수용** (선행 R1과 동일 뿌리) |
| A9 | "Modified Capabilities 없음"이라 선언했지만 조정 이벤트 설계가 reconciliation 의미를 바꾼다(확정 스펙은 운영자 확인 해제, 이 change는 자동 해제 추가). `make validate`는 잡지 못한다 | **확인 · 수용** |

## B. 판단 계층의 구현 가능성

| # | 내용 | 판정 |
|---|---|---|
| B1 | **판정 체인과 예약 트랜잭션의 관계가 미정의** — 체인은 스냅샷 위 순수 함수인데 선행 계약은 같은 한도를 트랜잭션 안에서 판정·예약하고 as-of 재검증·롤백·재수집을 요구한다. 어느 평가가 권위인지, 체인이 ALLOW를 낸 뒤 트랜잭션 측 거부의 reason-code가 무엇인지, 재수집 시 체인을 다시 도는지가 없다 → 구현자가 한도를 두 번 평가해 답이 갈리거나 체인 단계를 잃는다 | **확인 · 최중대** |
| B2 | **체인의 두 단계가 이 Phase에 입력 생산자가 없고 둘 다 fail-closed다** — 구조적 RR은 세션 고가·VWAP·저가·시가를 요구하고(StockOS structural_rr.py), 위험 기반 수량은 등급배수를 요구하는데(a090 계약이 `grade_missing`에 fail-close) 둘 다 신호 계층 산물이고 전략은 이 change의 Non-Goal이다. **명세대로면 체인이 ALLOW를 낼 수 없고 tracer의 진입 leg도 못 돈다.** MFE/MAE를 자른 것과 정확히 같은 "데이터 소스 없음" 논리가 여기엔 적용되지 않았다 | **확인 · 최중대** |
| B3 | "StockOS `evaluate_guardian`의 검증된 순서를 보존(SHALL)"이 충족 불가 — 실제 순서를 확인한 결과 내가 쓴 순서와 다르고(주문 크기 한도와 손절 계약이 뒤바뀜, 중복 검사 두 개를 하나로 합침), **최소 RR은 `evaluate_guardian`에 아예 없다**(전략 계층 소관) | **확인 · 수용** |
| B4 | "같은 질의 위의 투영"이라는 단일 권위 주장이 달성 불가 — `NetPositions`는 심볼→순수량만, 시장 차원 없음, float64 누적, 전 기간. position-ledger는 심볼·시장 단위·인스턴스 식별·평균단가·순서 있는 이벤트 위 상태기계를 요구한다. 게다가 조정 이벤트가 하나라도 생기면 "두 값이 같다" 시나리오가 거짓이 되어 **진입 차단 해제 조건이 도달 불가**가 된다 | **확인 · 수용** |
| B5 | saga 상태 이름(RECORDED/DISPATCHED)이 attempt 단계와 충돌하고, ACTIVE의 권위가 폴러라는 것·폴러 장애가 10초 SLO에 무엇을 뜻하는지가 없다 | **확인 · 수용** |

## C. 안전 구멍

| # | 내용 | 판정 |
|---|---|---|
| C1 | **자동 강화가 청산 경로를 교착시키고 사람만이 풀 수 있다** — HALT_ALL에서 PROTECTION_WEAKENING 금지 + 청산 가용수량이 WATCHING 예약을 차감 → 보유 50·stop 100 WATCHING이면 가용이 음수라 청산 불가, 유일한 해법(보호 축소·취소)이 금지된 클래스. **무인 시스템이 스스로 도달하고 사람만 풀 수 있는 상태** | **확인 · 최중대** |
| C2 | **`internal/flatten`(비상 청산)이 새 레일에 의해 깨지고 어느 change도 소유하지 않는다** — flatten은 `ScanOrders(Status:"OPEN")`로 취소 대상을 열거하므로 **브로커측 WATCHING 조건주문을 찾지도 취소하지도 못한다.** 청산은 class 없는 결정으로 Gateway를 통과하는데, 계약 적용 후엔 거부되거나(청산 불능) 조용히 기본값 처리(fail-open)된다 | **확인 · 최중대** |
| C3 | HALT_ALL 예외와 필수 산출물이 2-클래스 어휘로 쓰여 있다 — 필요한 산출물은 모드 × 클래스 표이고, **EXIT_ONLY × PROTECTION_WEAKENING이 완전 미정의**인데 그것이 EXIT_ONLY에서의 부분 청산이 요구하는 바로 그 mutation | **확인 · 수용** |
| C4 | "비용은 과대 추정이 보수 방향(SHALL)"이 청산 측에서 fail-closed가 아니다 — StockOS의 `SELL_COST_BUFFER_EXCEEDED`는 추정 비용이 버퍼를 넘으면 **매도를 거부**한다. 과대 추정이 청산을 막으면 §0.3 위반. "비용 모델은 청산 게이트로 적용하지 않는다"는 요구가 필요 | **확인 · 수용** |
| C5 | "outbox"가 두 가지를 가리킨다 — 전달 실패가 자동 강화 트리거인데 분석 retention도 outbox에 얹혀 있어, **막힌 180일 정리 작업이 운영 모드를 사람만 풀 수 있는 상태로 승격시킬 수 있다** | **확인 · 수용** |
| C6 | 최소 RR 1.5가 StockOS에서 **가장 느슨한 값**이고 provenance가 오귀속 — 실제 값은 2.0(§22 lock)·1.5·1.6·1.3이고 라이브 게이트는 setup별 2.0~4.0. 티어를 명시하지 않고 1.5를 고른 것은 §0.9 위반(완화 방향) | **확인 · 수용** |
| C7 | **이식하려는 비용 모델이 KIS 것이다** — StockOS `costs.py`가 KIS 소매 수수료표와 FSC 매도세를 기본값으로 문서화하고 모든 override 키가 `KIS_*`인데, 내 design.md는 "KIS 고유 항목 제외"를 명시한다. **자기모순**이고 과소 추정은 진입 현금·손익분기 검사에서 fail-open | **확인 · 수용** |

## D. 태스크

| # | 내용 | 판정 |
|---|---|---|
| D1 | **네 태스크가 각각 "journal v6"를 만든다** — 선행 change에서 C8로 지적받고 단일 원자 마이그레이션으로 고친 바로 그 위반이 여기서 반복 | **확인 · 수용** |
| D2 | 약 8개 태스크가 하루를 넘거나 설계 결정을 숨긴다(2.2·3.2·4.4·5.2·5.3·5.5·6.2/6.3/6.5·7.4). 특히 4.4는 선행 9.1과 **같은 작업**(엔진 Gateway 배선) | **확인 · 수용** |
| D3 | D3의 보류 3항목에 소유자·앵커·기본값이 없어 구현자가 멈추거나 추측한다. 그중 레버리지/인버스는 StockOS 체인의 **두 번째** 게이트(kill switch 다음)라 보류가 곧 최우선 검사 하나를 조용히 제거하는 것 | **확인 · 수용** |
| D4 | 이식 범위 열거가 불완전 — `CASH_ONLY_REQUIRED`·`SELL_COST_BUFFER_EXCEEDED`·`DAILY_TURNOVER_EXCEEDED`·`MAX_POSITIONS_EXCEEDED`·`CANCEL_RATE_EXCEEDED`·`TARGET_BELOW_BREAK_EVEN`이 in-scope·제외·보류 어디에도 없다 | **확인 · 수용** |
| D5 | 테스트 수 인용은 전부 정확(guardian 20·target_stop 29·a090 36·structural_rr 14, guardian.py 714줄). 단 `test_costs.py`는 4케이스뿐이고 검증 동작은 `test_costs_env_override.py`(16)에 있어 이식 범위가 과소 | **확인 · 수용** |

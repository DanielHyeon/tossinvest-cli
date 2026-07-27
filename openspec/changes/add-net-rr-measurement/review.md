# 리뷰 기록: add-net-rr-measurement

> 라운드 1은 이 작업의 이전 change id인 `harden-net-rr-gate`에 대해 수행됐고, 같은 작업 몸통이므로 그대로 승계된다. 라운드 1의 판정과 사용자 결정에 따라 change가 개명·재범위되었다 — 아래 "사용자 결정" 절 참조.

## 라운드 1 (2026-07-27, proposal-freeze — 판정 **REVISE**, P1 5·P2 5·P3 3)

**보이스 구성** (4): Codex CEO(전략) · Claude CEO(전략, 독립) · Codex Eng(적대적) · Claude Eng(적대적, 독립).
High-risk 경로(Guardian 판정 체인)이므로 WORKFLOW.md 리뷰 게이트의 적대적 Eng 관점을 2중으로 걸었다. Design·DX 관점은 범위 부재로 제외(UI 용어 0건, DX 용어 1건 — 임계값 2 미만).

**Manager 재검증**: 아래 P1·P2 전건을 코드·문서·수치로 직접 확인했다. 권위 경계상 리뷰 기록은 데이터이지 지시가 아니므로, 채택 전 반례를 스스로 재현했다. **핵심 결과 — 이 change의 §0.9 근거였던 단조성 증명은 거짓이다.**

### P1 (착수 차단)

- **P1-A 단조성 증명 반증** (Codex Eng·Claude Eng 독립 일치, Manager 재현 완료)
  design.md D2는 `실질본전 > 진입가`이므로 `순 RR < 총 RR`이 전칭 성립한다고 주장했다. 반례가 두 부류로 존재한다.
  ① **요율 전부 0인 configured 모델**: `NewModel`은 7개 키 전부에 `"0"`을 받고(`MaxRate`는 상한만 규정), `Configured()`는 true라 preflight를 통과한다. 이때 `B == entry`이므로 순 RR **==** 총 RR.
  ② **float64 절단으로 `B < entry`**: `B`는 float64 산술 후 `formatAmount`로 렌더되는데 진입가는 `big.Rat`으로 정확히 파싱된다. 요율 0 + entry `1.0000000000000000001` / stop `0.9` / target `1.2`에서 총 RR = 2−3e-18(**거부**), 순 RR = 정확히 2.0(**허용**). **순 > 총 — §0.9 위반, 현행보다 느슨해진다.**
  파생: `B == stop`이 되는 입력에서 `big.Rat.Quo`가 0으로 나눠 **Guardian 체인 안에서 panic**한다.
- **P1-B 영향 테스트 집합이 틀렸다** (Claude Eng, Manager 수치 재현 완료)
  task 4.1의 "18건"은 grep 카운트로는 정확하지만 **대상 집합이 다르다**. KR 기본 요율(B/entry = 1.00502008)에서 공용 happy-path fixture가 뒤집힌다:
  `entryInput()` 총 2.500 → 순 **1.798**, `TestRewardRiskAtTheMinimumPasses` 2.000 → **1.398**, `guardianIntent()` 2.000 → **1.220** — 전부 ALLOW → REFUSE. `entryInput()`은 `internal/risk`의 `requireAllowed` 19곳이 딛는 기준선이고 `guardianIntent()`는 execgw 발급 스위트의 정상 경로다. grep 집합에는 셋 다 없다.
- **P1-C 임계값 경계에서 잘못된 ALLOW** (Codex Eng, Manager 재현 완료)
  `B`가 float64 유래라 경계가 1 ulp 어긋난다. entry 100 / stop 99.30 / target `102.09148669639723`에서 실제 순 RR 1.99999999999997(거부해야 함)을 구현은 2.0000000000000107로 계산해 **허용**한다. 실물 호가는 틱 양자화되어 도달 확률은 낮지만, 이 패키지가 `big.Rat`을 쓰는 이유([contract.go:17-25](../../../internal/risk/contract.go) — "경계가 가격의 이진 전개의 부산물이 되면 안 된다")를 이 change가 스스로 무효화한다.
- **P1-D 인과 주장 과장** (Codex CEO·Claude CEO 독립 일치, Manager 확인 완료)
  proposal은 "이것은 가설이 아니라 StockOS의 실측 손실 구조다"라고 썼다. 058 요약(line 9)은 원인을 **셋**으로 열거한다 — ①필터 없는 시초 추격 진입 ②노이즈보다 좁은 0.70% 손절 ③비용 차감 후 RR 0.88. 비용은 ③이다. 표본이 **8건 0승**이므로 비용이 0이었어도 전건 손실이었다. 비용은 이미 지는 구조를 악화시킨 것이지 계좌를 비운 주원인이 아니다.
  또한 "어떤 종목 발굴 개선도 음의 기대값 산술을 이기지 못한다"는 논리 오류다 — 기대값은 `p·W − (1−p)·L`이고 발굴은 `p`를 바꾼다. 058 자신의 손익분기 승률 53.1%는 발굴이 승률을 그 위로 올리면 RR 0.88도 흑자라는 뜻이다.
- **P1-E 최소 손절 폭은 이연 가능한 잔여 위험이 아니다** (4개 보이스 전원 일치)
  요구 target은 `B + 2(B−stop)`이므로 **손절을 좁힐수록 요구 target이 내려간다**. 동시에 같은 요구사항의 `floor(위험예산/(entry−stop))` 사이징은 좁은 손절에 더 많은 수량을 준다. 좁은 손절이 이중으로 보상받는 구조이며, 이는 RC-K3(노이즈 안쪽 손절 → 11~30초 손절)의 기하 그 자체다. "ATR이 없어서 이연"은 비논리 — 비용 상대 하한(`entry − stop ≥ k×(B − entry)`)은 ATR 없이 landed `BreakEvenSellPrice`만으로 성립한다.

### P2 (archive 전 필수)

- **P2-F 0.88 재현은 두 자리 우연** (Codex Eng·Claude Eng·Claude CEO 일치, Manager 확인)
  058은 왕복 0.23%를 각 레그에서 **가감**해 `0.82/0.93 = 0.881720`을 얻었다. 이 change의 식은 `1/(1−매도측요율)`로 **그로스업**해 `0.880718`을 얻는다. **다른 산식이 두 자리에서만 일치**한 것이라 ±0.6% 안의 어떤 식이든 매칭된다. task 2.2가 `≈0.88`을 고정해봐야 잡으려는 오류를 못 잡는다.
- **P2-G Detail 자유텍스트는 약속한 측정을 지탱하지 못한다** (Codex CEO·Codex Eng 일치, Manager 확인)
  `decisions` 테이블([execution_contract.go:140](../../../internal/journal/execution_contract.go))에 reason/detail 열이 없고, **거부된 판정은 애초에 발급되지 않아 이 테이블에 남지 않는다**. 게다가 두 값을 거부 시에만 기록하면 markout·MFE/MAE에 필요한 **ALLOW 모집단이 비어 있다**. design.md의 "P3 markout이 두 값의 차이를 입력으로 쓴다"는 성립하지 않는다.
- **P2-H 임계값 provenance 오귀속** (Claude CEO·Claude Eng, Manager 확인)
  인용한 두 출처는 전부 **총/구조 RR** 게이트다: `live_entry_contract.py:53`의 2.0은 **미국 전용** RR 플로어이며 최적화 메뉴로 1.0~5.0 조정 가능, `default_lock.py`의 2.0은 Plan 044가 1.3으로 낮춘 값. **순 RR의 선례는 따로 있다** — StockOS는 이 게이트를 이미 출시했다: `early_entry_geometry.py`의 `net_risk = distance + cost`, `net_reward = target − entry − cost`, 차단 코드 `NET_RR_INSUFFICIENT`, 임계값 **KRX 1.5 / US 2.0**. 058 P1-2의 처방은 **순 RR ≥ 1.3**이다. design.md는 이 선례를 "대안 A(기각)"로 몰랐던 채 기각했다.
- **P2-I "임계값 2.0 무변경"은 사실상 임계값 변경** (Claude Eng, Manager 확인)
  총 RR 환산 = `2 + 3·(B/entry − 1)/s`. KR: 손절 1% → **3.51**, 0.5% → 5.01. US(매도측 1.1%): 손절 1% → **8.37**, 최소 target은 손절 폭과 무관하게 **+6.37% 고정**. 과대추정 요율 위에서 이는 미국 경로 사실상 폐쇄이며, design.md는 이를 "의도된 동작" 한 줄로 통과시켰다.
- **P2-J target은 실행에 구속력이 없다** (Claude CEO, Manager 부분 정정)
  리뷰어의 "소비처가 contract.go뿐"은 **틀렸다** — target은 `execgw/issue.go`·`journal/decision.go`(정규 JSON preimage에 해시됨)·`tracer.go`에도 흐른다. 그러나 결론은 맞다: **`internal/exitpolicy`는 target을 읽지 않는다**(ladder는 고정 % rung 1.5/2.5/4.0/6.0, ratchet은 target 참조 0건). 즉 게이트를 조이면 **더 큰 target을 선언할 유인**이 생기고 그 선언은 어떤 청산도 구속하지 않는다. RC-K4의 "손절은 가깝게, 익절은 멀게" 비대칭이 그대로 재현된다.

### P3 (문서 정합)

- **P3-K 단 번호·낡은 서술**: proposal/design이 `min_reward_risk`를 "6단"이라 부르나 `entryChain`상 **7단**이고 `docs/guardian-chain.md:20`은 이미 7로 쓴다. `reason.go:111`·`input.go:23`이 총 RR 산식을 문서화하고 있으며 task 목록에 없다. `contract.go` 패키지 doc과 `guardian-chain.md:45`의 정밀도 근거도 이 change가 무효화한다.
- **P3-L "순 RR 계산 불가" 시나리오는 도달 불가**: rung 5가 동일한 `BreakEvenSellPrice` 호출을 이미 `INPUT_UNAVAILABLE`로 거부한다. 하나의 사실에 두 reason code.
- **P3-M 비용 범위 표기 불일치**: proposal은 56.4%(슬리피지 포함)를, 스펙 델타는 53.1%(비포함)를 쓰는데 Non-Goals는 슬리피지를 제외한다. 보증 명칭도 "비용 차감 후"가 아니라 **"수수료·세금 차감 후"**여야 정확하다.

### Manager 처분 — 자동 채택 (2판 반영 대상)

P1-A·B·C, P2-F·G·I, P3-K·L·M은 전건 채택. 근거: 전부 정확성·완전성 결함이며 취향 판단이 아니다.

- P1-A → D2를 `순 RR ≤ 총 RR`(등호 조건 명시)로 재서술 + 분모 부호 가드(`RewardRisk`가 이미 하는 방식) + `B < entry`를 증명으로 없애지 말고 명시 거부 + 요율 전부 0/언더플로 모델 테스트
- P1-B → task 4.1을 "rung 7에 도달하는 모든 진입 fixture"로 재정의, `entryInput()`·`guardianIntent()` 재기준선 수립을 RED 이전 선행 작업으로
- P1-C → `B`를 게이트 직전 **보수 방향(위)으로 올림**하거나 유리수 end-to-end. 경계 바로 위/아래 적대적 테스트
- P2-F → "058 재현" 증거 주장 삭제, TossOS 자체 산출 `0.880718` 고정, 두 산식 차이를 기록
- P2-G → ALLOW·거부 양쪽에 구조화 필드(또는 additive 텔레메트리 이벤트)를 이 change에서 정의하거나, 측정 주장을 design에서 삭제
- P2-I → gross 환산표와 US +6.37% 하한을 D3에 명시
- P3-K·L·M → 문서 정합, 시나리오 정리, 명칭을 "수수료·세금 차감 후 RR"로

### Manager 처분 — 사용자 결정 대기 (User Challenge 4건)

아래는 4개 보이스가 **사용자가 승인한 방향 자체를 바꿔야 한다**고 본 항목이다. autoplan 규칙상 자동 결정하지 않는다. 사용자 원안이 기본값이며, 바꾸려면 모델 쪽이 입증해야 한다.

- **UC1 (P1-E)**: 최소 손절 폭을 이 change에 포함할 것인가(또는 활성화 조건으로 못 박을 것인가), 이연을 유지할 것인가
- **UC2 (P2-H)**: 순 기준 2.0을 유지할 것인가, StockOS 실측 선례(KRX 1.5 / US 2.0)나 058 처방(1.3)으로 맞출 것인가, 아니면 "2b 실측 전까지의 임시 활성화 상한"으로 명시할 것인가
- **UC3 (P2-J)**: target이 청산을 구속하지 않는 상태에서 target 게이트를 조이는 것이 타당한가 — 선행으로 exit ladder가 target을 소비하게 하거나, 게이트를 "전략 타당성 검사(실행 의미 없음)"로 스펙에 명시할 것인가
- **UC4 (P1-D + Claude CEO F7)**: 이 change를 **측정 전용**(순 RR 계산·기록만, 게이트 없음)으로 먼저 낼 것인가 — 4개 보이스 중 3개가 이것이 더 낫다고 봤다. P2-G가 요구하는 측정 배관을 만들면서 진입을 억제하지 않고 임계값 추측도 불필요하며, P2-J를 즉시 드러낸다

**라운드 1 결론: REVISE. 구현 착수 불가.** P1 5건은 전부 "고치면 되는" 결함이지만, UC1~UC4는 change의 형태 자체를 바꿀 수 있으므로 사용자 결정 후 2판을 작성한다.

---

## 사용자 결정 (2026-07-27) — 형태 A: 측정 전용

라운드 1의 User Challenge 4건을 사용자에게 제시했다(AskUserQuestion 도구가 내부 오류로 실패해 산문으로 제시). 선택: **A — 측정 전용으로 축소.** change id를 `harden-net-rr-gate` → `add-net-rr-measurement`로 개명하고 범위를 재작성했다. 폐기 사유는 새 proposal 상단에 한 줄로 기록(WORKFLOW 예외 경로 — change 폐기).

제시한 네 형태와 기각 근거:

| 형태 | 내용 | 처분 |
|---|---|---|
| **A** | 순 RR 계산·기록만, 게이트 무변경 | **채택** |
| B | 게이트 전환 + 최소 손절 폭 동시 포함 | 이연 — 추측 상수가 3~4개로 늘고, 비용 상대 형태에는 두 시장에 동시에 통하는 `k`가 없다(KR 유효 하한 `k>1.3944` ↔ 같은 `k`가 US에 최소 손절 2.97% 강제, `k=2`면 US 최소 손절 4.25% + 요구 target +14.86%). 그 숫자들이 전부 `[미검증]` 요율에서 나온다 |
| C | 게이트 전환, 손절 폭 이연 (원안) | 기각 — 4보이스 전원이 지목한 구멍(손절을 좁혀 통과)을 열어둔 채 P1-B(공용 fixture 3개 + `requireAllowed` 19곳 재기준선)를 감수한다. 재기준선은 추측 임계값을 모든 테스트에 박아 다음 조정 때 또 전멸시킨다 |
| D | 폐기하고 발굴 먼저 | 기각 — 리뷰 산출물(순 기준 선례 발견, 측정 구멍, 손절 폭 결합)이 문서로만 남는다. 발굴은 파일 표면이 겹치지 않아 이 change와 병행 가능하므로 배타 선택이 아니다 |

**A 선택이 라운드 1 지적에 대해 하는 일:**

| 라운드 1 | A에서의 처분 |
|---|---|
| P1-A 단조성 증명 거짓 | **격하** — 게이트가 없어 §0.9 미개입. 두 반례(요율 0 등호 / float64 유래 `B<entry`)를 task 3.5의 회귀 테스트로 고정해 게이트 승격 change가 반드시 보게 한다 |
| P1-B fixture 전멸 | **소멸** — ALLOW/REFUSE 불변, 재기준선 금지(task 3.3) |
| P1-C 경계 잘못된 ALLOW | **격하** — 기록값 정밀도 문제. 게이트 승격의 선행 조건으로 design D3에 명시 |
| P1-D 인과 주장 과장 | **수정** — proposal Why에서 "비용은 원인 셋 중 ③", "8건 0승이므로 비용 0이어도 전건 손실", 기대값 논리 오류 정정 |
| P1-E 최소 손절 폭 | **이관, 단 데이터와 함께** — task 5.2가 `k`의 근거인 손절 폭 분포를 산출 |
| P2-F 0.88 재현 | **수정** — 증거 주장 삭제. 두 산식(가감 `0.881720` vs 그로스업 `0.880718`)이 두 자리에서만 일치함을 task 3.4 주석에 기록 |
| P2-G 측정 배관 부재 | **이 change의 본체** — 거부 경로에 첫 영속(task 4.1), 자기완결 테이블(D1) |
| P2-H 임계값 provenance 오귀속 | **수정 + 이관** — task 6.3이 문서 정정, 임계값 확정은 후속 change가 측정 결과로 |
| P2-I US 하한 | **이관** — 계산표가 아니라 task 5.1의 실측 거부율로 |
| P2-J target 비구속 | **측정으로 드러냄** — 목표가를 기록해 두고(task 2.1) 선언 target vs 실제 청산가를 후속 측정 대상으로. 고치지는 않음(Non-Goals) |
| P3-K 단 번호·낡은 서술 | **수정** — task 6.1·6.2 |
| P3-L 도달 불가 시나리오 | **소멸** — 관측에서는 본전 산출 실패가 결측으로 기록되고 새 reason code가 생기지 않는다(D5, task 3.6) |
| P3-M 비용 범위 표기 불일치 | **수정** — `FEE_TAX_ONLY` 표기를 SHALL로, 지표 명칭을 "수수료·세금 차감 후 RR"로 |

**라운드 2 리뷰는 추가분(관측 기록의 실행 경로 결합·스키마·하네스 격리)에 대해 착수 전 실행한다 — task 1.1.**

---

## 라운드 2 (2026-07-27, 추가분 — 판정 **REVISE**, P0 2·P1 8·P2 4)

**보이스 구성** (2, 적대적 Eng 2중 — High-risk 경로): Codex Eng · Claude Eng(독립). CEO 관점은 재실행하지 않았다 — 형태 A는 사용자 결정으로 확정됐고 재논의는 정착된 결정의 재litigation이다. 범위는 추가분 3축(관측 기록의 판정 경로 결합 / schemaV8 쓰기 순서 / 하네스 격리).

**Manager 재검증**: 아래 P0·P1의 사실 주장을 전부 코드로 확인했다.

### P0

- **R2-1 측정 실패가 라이브 진입을 막을 수 있다** (2보이스 일치, 전 링크 코드 확인)
  연쇄: 관측 쓰기 실패 → 델타의 "실패는 알림 대상이다(SHALL)" → 유일한 durable 알림 경로는 critical(`obs/notifier.go` `notifyCritical`) → `deliver()` 재시도 소진 → `Gate.Block(ReasonAlertUndelivered)`로 **즉시 in-process 진입 latch** + `EscalateOperatingMode(ModeTriggerCriticalAlertUndelivered)` → **ENTRY_BLOCKED durable** → 해제는 OPERATOR 승인·audit 필요.
  즉 디스크 풀·`STRICT` 타입 위반·v8 열 오타 같은 **순수 측정 결함이 거래를 멈추고 사람 승인을 요구한다.** 메인 스펙 line 88 "분석·성과 작업 실패는 트리거가 아니다(SHALL NOT)"와 그 시나리오, trade-analytics 델타 자체와 정면 충돌 — 이 change가 도입한 모순이다.
  **함정**: `SeverityNormal`로 낮추면 best-effort·droppable이라 "조용히 넘어가서는 안 된다"와 충돌한다. **기존 두 등급 어느 것도 요구를 만족하지 않는다.**
  파생(R2-2, HIGH): `outbox.UndeliveredCount`는 전역이라(계좌·타입·등급 필터 없음) 측정 알림 1건이 운영자의 진짜 IN_DOUBT latch 해제를 막는다. `deliver()`가 뮤텍스를 ~4초 잡아 실제 인시던트 알림 경로를 직렬화하고 그 지연을 진입 경로(재수집 10초·TTL 60초)에 주입한다.
  **처분(채택)**: 관측 실패는 **관측 테이블 자체의 degradation 행/카운터**로 durable하게 남기고 알림은 normal 등급으로 한다. 명시 `SHALL NOT` 추가 — 관측 기록 실패는 outbox critical 경로·`CRITICAL_ALERT_UNDELIVERED`·`EntryGate.Block`에 진입하지 않는다. 관측 실패 + 알림 전달 소진 후에도 모드·EntryGate 불변임을 e2e 테스트로 고정.

- **R2-3 ALLOW 영속의 쓰기 순서** (2보이스가 **처방에서 충돌** — Manager 판정)
  Codex: 완전성을 위해 원자 트랜잭션 **안**. Claude: `defer tx.Rollback()`(issuance.go:121)이 전부를 되돌리므로 관측 삽입 오류가 체인 ALLOW를 거부로 바꾼다 → 델타 자신의 SHALL NOT 위반 → **밖**이 정답이고 결손을 복원 가능하게 할 것.
  **Manager 판정: Claude 채택.** 트랜잭션 안은 측정 결함이 거래를 **직접** 막게 만들어 R2-1이 CRITICAL로 지적한 것을 더 짧은 경로로 재현한다. 두 P0를 동시에 만족하는 유일한 형태는 "밖 + 복원 가능한 결손"이다.
  성립 근거: preimage(`journal.RiskIntent`)가 이미 entry/stop/target/market/policy_version을 담으므로 결정만 있고 관측 없는 행은 **결정론적으로 backfill 가능**하다. 단 backfill은 그 시점 요율을 쓰므로 재구성 행임을 표시하고 자기 지문을 남긴다.
  거부 측은 트랜잭션이 없어 단독 쓰기이고, 결손 시 preimage가 없으므로 **복원 불가**다 — 돈이 걸리지 않으므로 수용하되 명시한다.
  세 번째 실패 순서(Codex 추가): 관측을 발급 **전에** 커밋하면 발급이 뒤에 `LIMIT_REACHED`/`DECISION_EXPIRED` 등으로 거부될 때 결정 없는 거짓 ALLOW 관측이 남는다.

### P1

- **R2-4 결정 참조는 실제 FK여서는 안 된다** (Claude; Codex는 FK를 주장 — Manager: Claude 채택)
  `foreign_keys(on)`이며 같은 파일의 `spent_nonces.decision_id`가 정확히 이 이유로 FK를 두지 않는다("pruning expired decisions must not be blocked by (or cascade into) these rows" — execution_contract.go:188). FK는 분석 테이블의 수명을 계약 테이블에 얹고, 관측 선행 쓰기를 불가능하게 하며, D1의 "자기완결" 주장을 깨뜨린다.
- **R2-5 결과 열거가 없어 null 참조가 3중 모호** (Claude): `IssueEntry`는 체인을 한 번 평가한 뒤 발급을 재수집 루프로 돌리므로 체인 ALLOW가 `StageIssuance`에서 죽을 수 있다. null 참조는 "거부" / "체인 ALLOW·발급 거부" / "크래시 결손" 세 가지를 뜻한다. **필수 결과 열거**(`REFUSED_CHAIN`/`ALLOWED_ISSUED`/`ALLOWED_ISSUANCE_REFUSED`) + `IssueRefusalReason` 코드 추가.
- **R2-6 "정지한 체인 단계"는 산출 불가** (Claude, Manager 확인): `risk.Decision`은 `{Allowed, Reason, Detail}`뿐이고 단계 필드가 없다. `ReasonInputUnavailable`은 `internal/risk`에서 **42곳**이 발생시키므로 reason→rung이 다대일이다. "셋업이 없어서"와 "임계값이 틀려서"를 구분한다던 그 열이 조용히 틀린다. 판정에 additive 단계 필드가 필요하고 어느 task도 다루지 않았다.
- **R2-7 "발급 절차와 예약의 권위"에 MODIFIED가 없다** (Claude): 메인 스펙의 그 요구사항은 순서 있는 SHALL(ALLOW → 원자 tx → 제출)인데 이 change가 그 시퀀스에 쓰기를 끼운다. seam을 지목한 MODIFIED 필요.
- **R2-8 하향 마이그레이션 rollback 주장이 원장 계약과 충돌** (Claude, Manager 확인): `schema.go`는 "There is no down-migration and there will not be one"이고 rollback은 "run the previous binary"(`ErrSchemaTooNew` 거부) + 실패 시 자동 사전 백업(`backup.go`)이다. task 2.2와 design Migration Plan의 "rollback: 테이블 drop"은 **틀렸다**.
- **R2-9 US는 감사된 정책으로 도달 불가** (Claude, Manager 확인): `checkOrderSize`(min_reward_risk보다 **앞**)가 교차 통화를 `INPUT_UNAVAILABLE`로 거부하고 `DefaultPolicy()`는 전 항목 KRW다. US 격자점은 provenance 없는 USD 한도를 발명해야 하므로 "정책 수치의 provenance"(SHALL)와 충돌 — US 산출물이 조작된 정책 위에 있음을 표기해야 한다.
- **R2-10 4개 산출물 중 2개가 이 change에서 불가** (2보이스 일치): ② 손절 폭 분포는 **순환**이다 — 격자의 손절 폭은 작성자가 고른 값이므로 거기서 `k`를 유도하면 proposal이 금지한 "추측 노브"가 된다. ④ 선언 target vs 실제 청산가는 종결 포지션이 필요하다. ①은 분포가 아니라 경계면 지도, ③은 미측정 placeholder의 재서술. 그런데 두 델타가 "산출물이 착수 조건을 충족(SHALL)"이라고 써서 **충족 불가한 SHALL 쌍**을 만들었다.
- **R2-11 provenance 모순이 메인 스펙에 남는다** (Claude, Manager 확인): 메인 스펙 line 33은 "1.5는 최저 티어 값이라 기각"인데 이 델타는 1.5가 출시된 순 기준 KRX 게이트라고 쓴다. sync 후 한 capability가 둘을 동시에 갖는다. **미래 change가 2.0→1.3 완화를 정당화할 때 인용할 문장이 바로 이것이다.** task 6.3은 문서·주석만 고치고 메인 스펙 본문을 건드리지 않는다 → MODIFIED 필요.

### P2

- **R2-12 분포가 구조적으로 좌측 절단** (Claude): rung 5가 `target < B`를 이미 거부하므로 어떤 ALLOW에도 순 RR ≤ 0이 나타날 수 없다. 명시하지 않으면 후속 change가 절단을 증거로 오독한다.
- **R2-13 `internal/obs`가 tasks·Impact에서 누락**: 스펙이 알림을 요구하므로 `EventType` 추가·등급 결정·"모든 critical 이벤트가 표에 있다" 테스트가 필요하다.
- **R2-14 보존 정책 없음**: 매 평가마다 쓰는 테이블인데 trade-analytics는 180일 보존 정책을 SHALL로 요구하고 `trade_outcomes`에는 있다.
- **R2-15 관측 범위를 진입으로 한정해야 함**: `IssueReduction`이 같은 `evaluateChain` seam을 공유하고 stop/target이 없으므로 거기 배선하면 전부 null인 RR 행이 나온다. `Evaluate` 안에 배선하면 순수 체인에 journal 의존이 생겨 task 5.4가 주장하는 하네스 격리가 깨진다.

**라운드 2 결론: REVISE.** 전건 형태 변경 없이 스펙·design·tasks 수정으로 해소된다(P0 2건은 "밖 + 복원 가능한 결손" + "비강화 degradation 기록" 두 결정으로 동시 해소). 3판 반영 후 착수.

---

## 라운드 3 (2026-07-27, requirement 수정분 — 판정 **REVISE**, P0 3·P1 5·P2 3)

**보이스 구성** (2, 적대적 Eng 2중): Codex Eng · Claude Eng(독립). 범위: 3판이 만든 requirement 텍스트 — MODIFIED 2건과 개정된 ADDED 3건. CEO 관점 미실행(형태 A 확정, 재논의는 정착 결정의 재litigation).

**Manager 재검증**: 아래 전건을 수치·코드로 확인했다. **이번 라운드의 최대 발견은 Manager가 3판에서 직접 써넣은 산식 오류다.**

### P0

- **R3-1 3판이 §0.9 완화 경로를 열었고, 그것을 막으려 쓴 산식이 틀렸다** (Manager 자체 발견 → 2보이스 독립 확인)
  3판의 provenance 문단이 "총 기준 1.5 기각이 순 기준 1.5를 기각하는 근거는 아니다"라고 써서 순 1.5 채택 경로를 열었다. **순 1.5는 손절 폭이 교차점을 넘으면 총 2.0보다 느슨하다** — KR 손절 −5%에서 총 2.0은 target +10.000%를 요구하는데 순 1.5는 +8.755%만 요구하고, 그 target의 총 RR은 **1.751**로 현행 거부 대상이다.
  그리고 그것을 경고하려 적은 일반형 `((1+r)c − 1 − R)/(r − R)`이 **틀렸다**: KR에서 97.49%를 준다(올바른 값 2.51%). 후속 change가 그 식을 적용하면 "완화는 손절 폭 97% 초과에서만 발생 = 사실상 없음"으로 읽는다 — **오류가 안전하지 않은 방향이다.** 올바른 닫힌형: `s* = (1 + r)(c − 1) / (R − r)`, `r=1.5·R=2.0`에서 `5(c−1)`. KR 2.5100% / US 10.6168%(재검증 완료).
  **처분(채택)**: 산식 정정 + 귀결 명시 — 손절 폭이 커지면 `r·s`와 `R·s`가 지배하므로 **`r < 2.0`인 어떤 순 임계값도 충분히 넓은 손절에서 반드시 완화된다.** §0.9를 만족하는 형태는 둘뿐이다: 순 임계값 ≥ 2.0, 또는 총 2.0을 **유지한 채 순 검사를 논리곱으로 추가**(구성상 단조 강화). 후속 change의 한 사이클을 절약한다.
- **R3-2 교차점이 `[미검증]` 요율에서 나온 상한이다** (Claude, Manager 재계산 확인)
  `s* ≈ 5 × (매수측 + 매도측 요율)`이므로 요율이 낮아지면 교차점도 낮아져 **완화 구간이 넓어진다**: 수수료 0.015%·거래세 0.18%면 `c` = 1.0021041, `s*` = **1.05%**(현행 2.51%의 절반 미만). 게다가 그 수치들이 비용모델 지문 없이 적혀 있어 자매 델타(trade-analytics)의 "산출 수치에는 지문·범위 병기(SHALL)"를 스스로 위반했다. **처분(채택)**: 상한임을 SHALL로 표기 + 지문 병기 의무.
- **R3-3 열화 계수를 관측 테이블에 쓰는 설계는 동일 실패영역이다** (Codex)
  라운드 2 처분(D7)이 "실패는 관측 테이블 자체의 열화 기록으로 durable하게 남긴다"였는데, `SQLITE_FULL`·I/O 오류·스키마 오류로 관측 INSERT가 실패하면 **같은 저장소의 열화 쓰기도 함께 실패한다.** 라운드 2 P0가 요구한 durable gap count가 보장되지 않는다. **처분(채택)**: 독립 실패영역 저장 + 그마저 불가하면 구조화 로그·프로세스 내 단조 카운터로 강등, **재시작 후 durable하다고 주장 금지**(SHALL NOT).

### P1

- **R3-4 재구성 작업의 배치 제약이 없고, 유일한 자연스러운 집이 치명적이다** (Claude — CRITICAL로 제기, Manager P1 분류: 스펙 텍스트 결함이며 구현 전이므로)
  요구사항이 "탐지·재구성되어야 한다"만 쓰고 어디서 도는지 말하지 않는다. landed 주기 작업의 집은 엔진 런타임이고 거기 `SupervisedLoop`으로 등록하면 그 요구사항이 막겠다던 P0가 재현된다: 감독 루프의 반환은 **전 루프 종료**라 측정 오류가 exit observer·filldetect를 함께 죽이고(§0.3), 연속 실패 임계는 ENTRY_BLOCKED 강화이며, 트리거가 폐쇄 열거라 측정 작업이 **대사 실패 트리거를 차용**해 원인을 오귀속한다. **처분(채택)**: 감독 루프 등록 금지 + 별개 프로세스·스케줄 경계 + 어떤 ModeTrigger에도 미사상.
- **R3-5 provenance 문단 안에서 provenance가 퇴행했다** (Claude, Manager 확인 — 자기모순)
  3판이 `live_entry_contract.py:53` 인용을 **떨어뜨리고**, 기각 근거 "1.5는 최저 티어 값이라 기각"을 **삭제**한 뒤 그 기각을 총 기준 전용으로 조용히 재범위했다. 그런데 "설정 범위 1.0~5.0의 최저 티어"는 **기준 무관**한 사실이므로 재범위를 지지하는 근거가 원문에 없었다. 동시에 새로 들여온 세 수치(KRX 1.5 / US 2.0 / 처방 1.3)는 파일·검증 상태 없이 등장했다 — 나쁜 provenance를 비판하면서 나쁜 provenance를 썼다. **처분(채택)**: 인용 복원 + 기각 근거의 기준 무관성 명시 + 순 선례 수치에 경로·검증 상태 병기 의무.
- **R3-6 유예창·유일성 부재로 중복 행이 생긴다** (Claude): 안티조인이 진행 중 쓰기를 결손으로 오인해 재구성하면 실제 쓰기가 뒤이어 착지해 한 결정에 두 행이 남고, **라이브 임계값이 도출될 분포가 이중 계수된다.** 요구사항 1은 FK를 금지했을 뿐 유일성을 명시하지 않았다. **처분(채택)**: 쓰기 데드라인 초과 결정에만 재구성 + 결정 참조 유일 인덱스(FK 아님).
- **R3-7 재구성 가능성이 결정 정리 지평에 갇힌다** (Claude): FK를 금지한 이유가 정리를 막지 않기 위함인데, 결정이 정리되면 결손은 **탐지도 복원도 불가**해지고 조용히 사라진다. **처분(채택)**: 재구성 주기 < 정리 지평 + 영구 손실 계수.
- **R3-8 관측이 제출을 막을 수 있었다** (Codex): "커밋 이후 수행"만 써서 동기 관측 쓰기가 `ALLOW → 원자 tx → 제출` 사이에 끼어 원장 지연이 결정 TTL(60초)을 소진할 수 있었다. **처분(채택)**: 제출의 선행조건·동기 대기 지점 금지 + 실패·지연이 제출·HELD 검증을 지연·취소 금지(SHALL NOT).

### P2

- **R3-9 재구성 과장** (Codex·Claude 일치): preimage에 비용모델 지문·관측 시각·결과 구분이 없다. **처분**: 복원/재산출 항목을 열거로 분리, 순 RR·실질본전·지문은 "재구성 시점 모델로 새로 산출"이며 원래 관측값의 복원이라 표현 금지. 발급 시각·재구성 시각 둘 다 기록.
- **R3-10 "비강화 등급"은 존재하지 않는 등급** (Claude): `obs.Severity`는 `critical`·`normal` 둘뿐이다. **처분**: `SeverityNormal` 명시 + `criticalEvents` 표 **비구성원성 테스트 신설**(기존 표 테스트는 포함만 단언 — 누락되면 v8 열 오타가 `Gate.Block`+ENTRY_BLOCKED로 번지는 스위치가 열린다).
- **R3-11 "US는 현실 구간에서 전부 엄격"이 증명 의무 면제로 읽힌다** (Claude): TossOS는 손절 폭 상한이 없으므로 15% 손절도 허용 입력이고 거기서 순 1.5는 느슨하다. **처분**: 해당 서술 제거 + 면제 불가 명시. 목표가 복원이 최소 RR rung의 거부에 의존함(preimage 계약은 목표가를 선택 항목으로 둔다)도 명시.

### 청정 확인 (두 보이스 일치)

- MODIFIED "발급 절차와 예약의 권위"는 메인 스펙 문단을 **바이트 동일**로 승계했고 SHALL 손실 0.
- MODIFIED "No Stop = No Trade"에서 **게이트는 여전히 총 기준 2.0**이고 시나리오도 `(target−entry)/(entry−stop) < 2.0`로 온존.
- 관측을 원자 트랜잭션 밖에 두는 것은 `engine-safety`의 HELD 예약 검증(spec:67)·`position-ledger`의 원자 apply와 **충돌하지 않는다**(Manager 교차 검증 + 2보이스 독립 확인).

**라운드 3 결론: REVISE → 반영 완료.** 전건이 텍스트 정정으로 해소됐고 3판 반영 후 `openspec validate --strict` 통과. 스펙에 적힌 산식·수치(KR 2.5100% / US 10.6168% / 5(c−1) / 손절 −5%의 총 RR 1.751)는 재계산으로 재검증했다.

**라운드 4는 걸지 않는다.** 라운드 3 처분은 전부 (a) 산식·인용 정정 (b) 이미 라운드 2·3이 본 요구사항에 SHALL NOT 추가 (c) 테스트 항목 추가이며, 새 requirement를 도입하지 않는다. 다음 게이트는 구현 후 Manager diff 리뷰 + `make gate`다.

---

## task 1.3 — Manager 자체 검토 (2026-07-27, 4판 반영 후)

기계적 검사 4종을 돌렸다. **반증 서술 잔존 4건을 발견해 수정했다** — 라운드 3 처분을 스펙에는 반영했으나 design·proposal·tasks가 따라오지 않은 것이다.

| 검사 | 결과 |
|---|---|
| 라운드 1~3이 반증한 서술 잔존 | **4건 발견 → 전건 수정.** ① design.md D7·proposal.md의 "관측 테이블 자체의 열화 기록"(R3-3 동일 실패영역) ② proposal.md의 "rollback은 테이블 drop"(R2-8 원장 계약 위반) ③ **tasks 6.6이 틀린 산식 `((1+r)c−1−R)/(r−R)`을 그대로 보유**(R3-1) ④ design.md의 "비강화 등급"(R3-10). 재검사 후 실질 잔존 0(남은 grep 히트 2건은 반증 서술을 *인용*한 정상 문장) |
| MODIFIED 축자 승계 | "발급 절차와 예약의 권위" 시나리오 2→3·SHALL 5→9·원문 문장 손실 0. "No Stop = No Trade" 시나리오 3→5·SHALL 5→15. 핵심 문장 개별 확인: 첫 규범 문장·t0 기준선 문장·`기본값 2.0 미달·계산 불가는 거부한다(SHALL — 0 대체 금지)`·복원된 `live_entry_contract.py:53`·복원된 "최저 티어 값이라는 사실" 전건 존재. **게이트는 여전히 총 기준 2.0** |
| 스펙 수치 독립 재계산 | `s* = (1+r)(c−1)/(R−r)`, `5(c−1)`, KR `c`=1.0050201·`s*`=2.5100%, US `c`=1.0212336·`s*`=10.6168%, KR 손절−5% 순1.5 +8.755% / 총2.0 +10.000% / 그 target의 총 RR 1.751, 실측추정 `c`=1.0021041·`s*`=1.05% — **전건 일치** |
| 처분 ↔ 스펙 SHALL ↔ tasks 커버리지 | 라운드 2·3 처분 18개 지점 전부 스펙에 SHALL/SHALL NOT로 존재하고 대응 task 존재. **결함 0** (스펙 SHALL 89개 / tasks 48항목) |

부수 정정: tasks 6.6이 "스펙에 적힌 2.5100%"라고 참조했으나 스펙 표기는 2.51%였다 — 검증 불가능한 참조가 되므로 양쪽 표기를 병기하도록 수정.

**교훈(다음 change에 적용)**: 리뷰 처분을 스펙에만 반영하고 proposal·design·tasks를 갱신하지 않는 드리프트가 라운드마다 발생했다. 처분 반영 시 4개 파일을 함께 훑고, 반증된 표현 목록으로 기계 검사를 돌리는 것을 절차로 둔다.

**task 1.3 완료. Manager 작업 종료 — 구현(2~7그룹)은 별도 세션 Teammate로 위임한다.**

---

## task 2.0 — Pre-Edit 선언 (2026-07-27, 구현 착수 직전)

High-risk 2개 경로에 해당한다: **원장 스키마**(schemaV8)와 **Guardian**(기록 배선이 판정 경로에 붙는다).
`base-commit.txt` = `33ea82a4ed181d648735051d796634a621290441` (이 선언 직전 `capture_change_base.py`로 고정 — 착수 전 미고정 상태였다).

### 착수 직전 버전 확인 (task 2.0 명시 요구)

| 확인 | 결과 |
|---|---|
| `journal.SchemaVersion` | **7** ([schema.go:6](../../../internal/journal/schema.go)) — 병행 세션이 8을 claim하지 않았다. 버전 재조정 불필요, v8을 이 change가 claim한다 |
| `migrations` 마지막 스텝 | `{Version: 7, SQL: schemaV7}` (adoption.go) — v8은 append |
| 병행 세션 미커밋 변경 | `internal/verifylive/{fake_broker,http}_test.go`, `openspec/changes/verify-execution-capability/tasks.md` — journal·risk·execgw와 무교차 |

```text
Pre-Edit Gate:
- change id / task id: add-net-rr-measurement / 2.x·3.4·4.x
- 대상 심볼(패키지.함수):
    internal/journal   — SchemaVersion(const), migrations(var) [additive append]
    internal/risk      — Decision(struct, additive field), Evaluate
    internal/execgw    — RiskGuardian.IssueEntry
    ※ checkMinRewardRisk·RewardRisk·DefaultMinRewardRisk·checkStopContract 는 **무변경 대상**
- 기존 동작 파악 근거:
    · 마이그레이션 계약: schema.go:8-26 (additive 규칙 4개 + "There is no down-migration
      and there will not be one" — 롤백은 이전 바이너리 ErrSchemaTooNew(journal.go:23-26,227),
      중간 실패 복구는 backup.go 자동 사전 백업)
    · FK 부재 선례: spent_nonces (execution_contract.go:188-197) — "pruning expired decisions
      must not be blocked by (or cascade into) these rows". 관측 테이블의 decision 참조도 동일 근거
    · preimage 계약: RiskIntent.Canonical() (decision.go:79-118) — 필드 순서 고정 canonicalJSON,
      target_price는 required=false(선택 항목). HashPreimage = SHA-256(canonical)
    · 체인 순서: entryChain 12 rung (chain.go:67-80). min_reward_risk는 index 6 = **7단**
      (task 6.1의 근거 — contract.go:170 "sixth"는 오기, contract.go:105 "fourth"도 오기: stop_contract는 5단)
    · 판정 값: Decision{Allowed, Reason, Detail} (chain.go:28-32) — 단계 필드 없음
    · reason→rung 다대일 확증: ReasonInputUnavailable 비테스트 발생 **42곳**
      (chain.go 35 + contract.go 4 + reason.go 3) — 역산 금지 근거 실측
    · 본전 단일 정의: checkStopContract가 BreakEvenSellPrice(LimitPrice, "1", Market) 소비
      (contract.go:147-160). 순 RR도 같은 호출을 쓴다(task 3.2가 고정)
    · 발급 seam: IssueEntry (riskguardian.go:318-421) — evaluateChain 1회(riskguardian.go:66,340)
      → EntryExposureValue → RecordDecisionAndReserveWithRecollection(376) → return.
      관측은 376의 err 처리 **이후**에 둔다(트랜잭션 밖)
    · 등급표: criticalEvents map (event.go:219-238, 17종), SeverityOf(248) 미지 이벤트는 normal.
      기존 표 테스트는 **포함만** 단언 → task 4.6의 비구성원성 테스트가 신설분
- upstream 상속 테스트 영향: **no**. internal/risk·internal/journal의 결정 계약 표면은 TossOS 신규다.
  회귀 방지: (a) 3.3이 checkMinRewardRisk·RewardRisk·DefaultMinRewardRisk 무변경을 고정하고
  (b) 2.3 골든 테스트가 Canonical()·risk_hash 바이트 불변을 고정하며
  (c) 7.3이 기존 스위트 **무수정** 통과를 판정 불변의 증거로 삼는다.
  entryInput()·guardianIntent() 재기준선은 **금지**(D5 — ALLOW/REFUSE 불변이므로 필요 없다)
- 실패 테스트 선행 작성: yes (각 task RED 먼저 — 2.1 스키마, 3.1 순 RR, 4.1 배선 순서)
- 안전 불변식 §0 위반 여부 검토: **통과**
    §0.1 LIVE 주문 side effect 0 — 하네스는 읽기 전용, 관측은 쓰기 전용
    §0.3 무관 — 청산 경로 미접촉. 단 2.5b가 SupervisedLoop 등록 금지를 구조 테스트로 고정
    §0.4 새 외부 호출 0
    §0.6 additive only — 기존 테이블·열·인덱스·preimage·해시 무변경, 하향 마이그레이션 없음
    §0.9 미개입 — 어떤 의도의 ALLOW/REFUSE도 바뀌지 않는 것이 이 change의 계약
- Function Logic Map: **필요**. 기존 함수 내부 로직을 수정하는 대상은 `risk.Evaluate`와
  `execgw.RiskGuardian.IssueEntry` 둘이다(High-risk이므로 면제 불가).
  12개 rung 함수는 수정하지 않는다 — 단계 스탬프를 Evaluate 한 곳에 두어
  "판정이 단계를 직접 보고한다"를 만족시키면서 rung 함수 12개의 내부 편집을 피한다.
```

---

## task 7.3 — 판정 불변 증명과 수정된 기존 테스트 근거 (2026-07-27)

### 판정 불변의 증거

이 change의 계약은 "어떤 의도의 ALLOW/REFUSE도 바꾸지 않는다"이고, 그 증거는 **기존 스위트가 무수정으로 통과한다**는 사실이다.

| 대상 | 결과 |
|---|---|
| `internal/risk` 기존 테스트 파일 | **무수정 0건 변경.** `chain_test.go`·`contract_test.go` 등 전부 그대로 통과(패키지 127건) |
| `internal/execgw` 기존 테스트 파일 | **무수정.** 272건 통과 — `guardianIntent()`·`entryInput()` 재기준선 **없음**(D5 예측대로 P1-B 미발생) |
| `checkMinRewardRisk` · `RewardRisk` · `DefaultMinRewardRisk` | 본문 **무변경**(diff 0줄). `input.go`·`reason.go`·`contract.go` 변경은 전부 주석이다 |
| 국소 단언 | `TestTheStepIsAdditiveToTheVerdict`(reason·detail 불변), `TestAGuardianWithNoObserverIssuesAsBefore`(observer 미주입 시 change 이전과 동일) |

`internal/risk/chain.go`의 production diff는 +42/−4이며 내용은 ① `Decision.Step` 필드 추가 ② `StepPreflight`/`StepReduction` 상수 ③ `at()` 신규 leaf ④ `Evaluate`의 3개 return에 `at(...)` 래핑이다. rung 함수 12개는 **한 줄도 건드리지 않았다**.

### 수정한 기존 테스트 — 각각의 근거

기존 테스트는 두 파일 5개 함수만 수정했다. 둘 다 스키마 버전 상승의 기계적 귀결이며, 단언 자체는 하나도 바꾸지 않았다.

| 파일 / 함수 | 수정 내용 | 근거 |
|---|---|---|
| `migration_v7_test.go` — `TestMigrationV6ToV7PreservesEveryRow`, `TestOlderBuildRefusesTheV7Journal`, `TestV7MigrationBacksUpBeforeApplying`, `TestFailedV7MigrationLeavesTheJournalRestorable` | `openTestJournalAt`(head까지 마이그레이션) → `openJournalAtSchema(t, path, 7)` + 기대 버전 리터럴 정합 | `SchemaVersion`이 8이 되면 이 4개는 **v6→v8 전이의 테스트로 조용히 바뀌어** v6→v7 스텝이 무검증 상태가 된다. `migration_v6_test.go:92-95`가 v7 착수 때 같은 조치를 하며 근거를 주석으로 남겼다 — *"a later change adding v7 must not silently turn it into a test of a different transition"*. 그 선례를 그대로 따랐고 **단언·기대값·오류 문구는 전건 그대로**다 |
| `schema_test.go` — `TestSchemaTablesAndColumns` | `wantTables`에 `entry_decision_observations` 1줄 삽입(정렬 위치) | 이 테스트의 목적이 "스키마 추가를 의도적 행위로 만드는 것"이므로 additive change에서 이 리터럴을 갱신하는 것이 정상 흐름이다. **기존 항목의 삭제·개명·재정렬 0건.** 추가성 주장은 이 목록에 의존하지 않는다 — `TestV8IsPurelyAdditive`가 v7 스키마 객체 전부의 DDL을 v8 저널의 같은 객체와 바이트 대조한다 |

두 수정 모두 Function Logic Map 증거를 첨부했다(`analysis/function-logic/`).

### 부수 발견 — SDD 도구 결함 1건 (수정하지 않음)

`tools/logic-map/check_analysis.py`는 분기 없는 함수에서 크래시한다: 추출기가 `"branches": null`을 내보내는데 검사기는 `value.get("branches", [])`로 읽어 `None`을 그대로 순회한다. 이 change가 그 도구를 실제 수정 함수로 처음 돌린 change라 처음 드러났다.

**수정하지 않고 회피했다.** 추출기를 고치면 `extract_go_ast.go:analyze`가 수정 함수가 되어 그 자체로 Function Logic Map을 요구하는 재귀가 생긴다. 대신 `at()`을 유일한 호출자인 `Evaluate` 바로 아래로 옮겼다 — 삽입 지점이 `refuse()`의 `end+1`이라 `intersects()`의 `count == 0` 규칙에 걸려 **변경되지도 않은 `refuse`가 수정 함수로 오탐**되고 있었기 때문이다. 배치는 독립적으로도 옳다(3줄 헬퍼는 호출자 옆). 도구 결함은 별건으로 남긴다.

---

## task 7.5 — 완료 보고 (2026-07-27)

### 실행한 명령과 실제 결과

| 명령 | 결과 |
|---|---|
| `openspec validate add-net-rr-measurement --strict --no-interactive` | `Change 'add-net-rr-measurement' is valid` |
| `go test ./...` | **3042 passed / 55 packages**, 실패 0 |
| `go vet ./...` | `No issues found` |
| `go test -race ./internal/{risk,journal,execgw,measure/...,obs,costs}` | **1108 passed / 7 packages** |
| crash 테스트(SIGKILL 자식 프로세스) | `TestACrashBetweenCommitAndObservationIsRecoverable` PASS |
| `python3 tools/logic-map/check_analysis.py --change add-net-rr-measurement` | `evidence complete or diff-proven exempt` |
| `python3 tools/sdd/check_agent_config_sync.py` | `synchronized` |
| `python3 tools/pm/generate_master_tracker.py --check` | `hierarchy and generated trackers are current` |
| `make gate CHANGE=add-net-rr-measurement` | **GATE PASS — 8/8** (base `21400c7`) |

### 변경 파일

기존 production 파일 수정은 10개 / +198 −29이고, 그중 절반이 주석이다.

| 파일 | 성격 |
|---|---|
| `internal/journal/schema.go` | `SchemaVersion` 7→8, `migrations`에 v8 append |
| `internal/risk/chain.go` | `Decision.Step` additive 필드 + `at()` + `Evaluate` 3개 return 래핑 |
| `internal/execgw/riskguardian.go` | `Observer` 옵션·필드 + `IssueEntry`의 관측 hand-off 3곳 |
| `internal/obs/event.go` | measurement 이벤트 타입 2개(둘 다 normal 등급) |
| `internal/risk/{contract,input,reason}.go` · `docs/guardian-chain.md` | **주석·문서만**(6.1~6.4) |
| `internal/journal/{schema_test,migration_v7_test}.go` | 스키마 버전 상승의 기계적 귀결 — 근거는 위 task 7.3 |

신규 파일 19개: `internal/journal/entryobservation.go`, `internal/risk/netrr.go`, `internal/costs/fingerprint.go`, `internal/execgw/observation.go`, `internal/measure/{counterfactual,population,reconstruct}.go`, `internal/measure/degrade/degrade.go`, 대응 테스트, `tools/boundarymap/`.

### DoD·게이트

- change/task DoD: tasks.md 48항목 전건 완료, 미완료 0
- Function Logic Map: **적용**. 수정한 기존 함수 9개 전부 증거 첨부(`analysis/function-logic/`) — 면제 없음
- High-risk Pre-Edit Gate: 수행·기록(위 task 2.0). 착수 시 `SchemaVersion == 7` 확인 후 v8 claim
- upstream 회귀: 없음. 상속 테스트 650개는 `internal/{risk,journal,execgw}`의 이 표면에 도달하지 않으며, 전체 3042건 green
- 산출물: `analysis/boundary-map.md`(경계면 지도, `tools/boundarymap`로 재생성 가능)

### 남은 위험 (task 7.5 필수 5항목)

1. **이 change는 StockOS의 손실 기하를 거부하지 않는다.** 차단은 후속 게이트 change 소관이다. 다만 산출물이 보여준 사실 하나는 기록해 둔다 — 058의 기하(총 RR 1.5)는 **현행 총 2.0 게이트가 이미 거부한다**(`TestTheRealPopulationSuppliesTheStopWidths`). 순 기준 전환이 없어도 그 8건은 오늘 통과하지 못한다.
2. **라이브 관측 모집단은 아직 0건이다.** `evaluateChain` 도달 경로는 `RiskGuardian` 발급뿐이고 그 호출자 `Tracer`에 프로덕션 호출자가 없다. 즉시 산출물은 합성 경계면 지도뿐이며, 관측 배관을 먼저 둔 이유는 첫 판정이 기록 없이 지나가지 않게 하는 것이다.
3. **비용 요율 7개가 전부 미측정이다.** 모든 산출 수치는 2b 실측 후 재산출 대상이며, 지문(`costs/71d81b5150330fd2`)이 실측 전후 관측의 혼합 집계를 구조적으로 막는다. 6.6 테스트가 요율 변경 시 실패해 스펙의 교차점 수치(KR 2.51% / US 10.62%) 갱신을 강제한다.
4. **후속 게이트 change의 착수 조건 4개 중 이 change가 주는 것은 ①(경계면 지도)뿐이다.** ②(손절 폭 분포·`k`)는 실거래 추출 성패에 달렸고 — 오늘은 058 전사 8행뿐이라 `k`는 **미결** — ③④는 별도 선행이다. 산출물 머리의 선언표가 이것을 그대로 담는다.
5. **거부 측 관측 결손은 복원 불가다.** 거부에는 결정이 없고 따라서 preimage도 없다. 계수만 하며(`refusal_observation_lost`), 발급 측의 복원 가능한 결손(`observation_write_failed`)과 별개 kind로 집계한다.

### 추가로 기록해 둘 위험 2건

6. **`k` 산출은 실질적으로 미달성이다.** 실거래 DB 경로를 실행 인자로 받는 배관은 있으나, 이 세션에서 선행 시스템 DB에 접근하지 않았다. 사용한 것은 058 문서 전사 8행이고 손절 폭이 전부 0.007로 동일해 **분포라고 부를 수 없다.** 산출물이 그 사실을 행 수와 함께 표기한다.
7. **재구성 배치의 스케줄러가 아직 없다.** `measure.Run`과 그 격리 계약(2.5b 구조 테스트)은 landed지만, 이를 별개 프로세스·스케줄 경계에서 주기 실행하는 배선은 이 change의 범위가 아니다. 그때까지 크래시 결손은 탐지 가능하되 자동 복원되지 않는다.

### 외부 서비스 상태

- CodeGraph: sync 완료, worktree fingerprint 일치(hard evidence 게이트 통과)
- GBrain: **동기화 실패** — `Timed out waiting for PGLite lock`. 병행 세션이 lock을 보유한 것으로 보이며, GBrain은 advisory이므로 게이트는 WARN으로 통과한다. 강제 해제하지 않았다.
- codegraphcontext: advisory WARN

### 병행 세션과의 간섭 (절차 기록)

`base-commit.txt`를 **두 번 재고정**했다(`33ea82a` → `76f0098` → `21400c7`). 이 change는 커밋이 하나도 없는 순수 working-tree 변경인데, 병행 세션이 작업 중 4회 커밋(`c652c27`·`76f0098`·`b560af1`·`27deac8`·`21400c7`)하면서 그들의 수정 함수가 이 change의 Function Logic Map 요구 집합으로 잘못 귀속됐기 때문이다. 재고정 방향은 **타인의 변경을 배제하는 쪽**이며 자기 변경을 숨기는 쪽이 아니다 — 재고정 후에도 이 change의 수정 함수 9개는 전부 요구 집합에 남아 증거를 첨부했다.

`check_index_freshness.py`의 fingerprint가 전 파일 내용 해시라, 병행 커밋 1회마다 CodeGraph 인덱스가 stale이 된다. 게이트는 `make sdd-sync` 직후 재실행으로 통과시켰다.

### 이 change가 하지 않은 것

`checkMinRewardRisk`·`RewardRisk`·`DefaultMinRewardRisk` 무변경, `entryInput()`·`guardianIntent()` 재기준선 없음, preimage·`risk_hash` 무변경(골든 고정), 하향 마이그레이션 없음, 게이트 임계값 무변경. **어떤 의도의 ALLOW/REFUSE도 바뀌지 않았다.**

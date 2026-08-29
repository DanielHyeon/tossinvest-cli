# Change: add-net-rr-measurement

> `harden-net-rr-gate`를 폐기하고 이 change로 대체한다 — proposal-freeze 리뷰(4보이스)가 순 RR 게이트 전환의 §0.9 단조성 근거를 반증하고 약속한 측정 배관이 실재하지 않음을 확인했다. 리뷰 기록은 이 디렉터리의 `review.md`(라운드 1)에 그대로 승계된다.

## Why

Guardian 체인 7단(`min_reward_risk`)은 **총 RR** `(target − entry) / (entry − stop)`를 검사한다([contract.go:175](../../../internal/risk/contract.go)). 비용은 이 비율의 양쪽을 악화시키므로 총 RR 2.0을 통과한 의도가 수수료·세금 차감 후에는 훨씬 낮은 비율일 수 있다. StockOS 058 RC-K4가 그 사례를 기록한다 — 명목 1.50이 차감 후 0.88.

**그러나 비용은 StockOS 손실의 원인 셋 중 하나(③)다.** 058 요약이 열거하는 원인은 ①필터 없는 시초 추격 진입 ②시초 노이즈보다 좁은 0.70% 고정 손절 ③차감 후 RR 0.88이며, 표본은 **8건 0승**이다. 비용이 0이었어도 전건 손실이었다. 비용은 이미 지는 구조를 악화시킨 요인이지 계좌를 비운 주원인이라는 증거는 없다. 또한 기대값은 `p·W − (1−p)·L`이므로 발굴이 승률 `p`를 손익분기(058 기준 53.1%) 위로 올리면 낮은 RR도 흑자가 된다 — "발굴로는 음의 기대값을 이길 수 없다"는 주장은 성립하지 않는다.

따라서 이 change는 게이트를 조이지 않는다. **먼저 잰다.** 근거는 셋이다.

1. **임계값을 고정할 데이터가 없다.** 순 기준의 선례는 StockOS가 실제로 출시한 `early_entry_geometry.py`의 `NET_RR_INSUFFICIENT`(임계값 **KRX 1.5 / US 2.0**)이고 058 P1-2의 처방은 **1.3**이다. TossOS가 인용해 온 2.0 두 출처는 전부 총/구조 RR 게이트이며 하나는 미국 전용 조정 가능 값이다. 순 2.0에는 선례가 없다.
2. **비용 요율 자체가 미측정이다.** `DefaultModel`의 7개 요율 전부 `[미검증 — 2b 실측 대체 대상]`이며 의도적 과대추정이다. 게이트를 걸면 이 placeholder가 제품 행동으로 굳는다 — US는 매도측 1.10%(FX가 양 레그) 때문에 최소 target이 손절 폭과 무관하게 +6.37%로 고정된다.
3. **측정 배관이 없다.** `decisions` 테이블은 발급된(ALLOW) 판정만 담고 reason/detail 열이 없다. 거부는 `chainRefusal`이 `IssueRefusal`을 메모리에 만들어 반환할 뿐 **어디에도 영속되지 않는다**. 이 상태로 게이트를 걸면 거래 0건일 때 "셋업이 없어서"와 "임계값이 틀려서"를 구분할 수 없다 — 058이 8일간 오진한 패턴이 정확히 그것이다.

## What Changes

- **판정 관측 기록**(신규): 진입 체인이 판정될 때마다 총 RR·순 RR·실질본전·비용모델 지문·정지한 rung·reason·**결과 구분**을 **ALLOW과 거부 양쪽** 모두 구조화 기록한다. 거부는 현재 영속 경로가 아예 없으므로 이것이 첫 기록이 된다. 자기완결 테이블(additive, schemaV8)이며 `decisions`의 봉인된 preimage·해시는 건드리지 않고 결정 참조는 **외래키가 아니다**. 범위는 **진입 한정**(청산 판정은 stop·target이 없어 전 항목 결측 행만 만든다).
- **결과 구분과 정지 단계는 새 정보다**(라운드 2): 체인 ALLOW과 발급 성공은 다른 사실이므로 `REFUSED_CHAIN`/`ALLOWED_ISSUED`/`ALLOWED_ISSUANCE_REFUSED`를 열거로 담는다. 정지한 rung은 `risk.Decision`에 단계 필드가 없고 하나의 reason이 42곳에서 발생하므로 판정이 직접 보고하도록 additive 확장한다.
- **관측은 원자 트랜잭션 밖, 결손은 복원 가능**(라운드 2 P0): 트랜잭션 안에 두면 관측 실패가 결정을 롤백해 측정 결함이 거래를 멈춘다. 밖에 두고, 관측 없는 발급 결정을 안티조인으로 탐지해 preimage에서 재구성한다(재구성 표지·재구성 시점 지문 병기). **관측 실패는 critical 등급·outbox·미전달 자동 강화·진입 게이트 차단에 진입하지 않는다**(알림 등급은 `SeverityNormal`) — 대신 **관측 행과 독립된 실패영역**에 계수한다(라운드 3: 같은 테이블에 열화 행을 쓰는 설계는 디스크 소진 시 함께 실패하는 자기모순이다). 재구성 작업은 **엔진 감독 루프로 등록하지 않는다** — 감독 루프의 반환은 전 루프 종료라 측정 오류가 청산 관측·체결 감지를 함께 죽인다.
- **순 RR을 관측값으로 산출**: `(target − 실질본전) / (실질본전 − stop)`, 실질본전은 7단 앞의 손절 계약 rung이 이미 소비하는 `costs.Model.BreakEvenSellPrice(entry, "1", market)`. **게이트 입력이 아니다.**
- **비용 기준 표기**: 관측마다 비용 범위를 명시한다(`FEE_TAX_ONLY`). 슬리피지(058 실측 ~0.13%p)는 포함하지 않으므로 이 지표의 이름은 "비용 차감 후"가 아니라 **수수료·세금 차감 후 RR**이다.
- **반사실 측정 하네스**(신규): 합성 격자와 StockOS 실거래 3가격을 체인에 통과시켜 임계값 후보별 통과/거부 경계와 손절 폭 분포를 산출한다. 읽기 전용 산출물이며 주문 경로·journal 쓰기·실계좌 접근 0. 058이 `strategy_optimization_runs` 0행으로 끝내 하지 않은 측정이다.
- **게이트는 무변경**: `checkMinRewardRisk`는 총 RR·임계값 2.0 그대로다. 이 change는 어떤 의도의 ALLOW/REFUSE도 바꾸지 않는다.
- **리뷰 라운드 1의 문서 결함 수정**: 단 번호(6단→7단), `reason.go:111`·`input.go:23`의 산식 서술, `contract.go` 패키지 doc과 `docs/guardian-chain.md:45`의 정밀도 근거, `guardian-chain.md:74`의 2.0 provenance(순 기준 선례 부재를 명시).

## Non-Goals

- **최소 RR의 게이트 기준 전환** — 이 change의 출력이 후속 change의 입력이다. 착수 조건을 아래 Impact에 열거한다.
- **최소 손절 폭 계약** — 비용 상대 형태(`entry − stop ≥ k×(B − entry)`)에는 두 시장에 동시에 통하는 `k`가 없다: KR에서 쓸모가 있으려면 `k ≥ 1.4`(StockOS의 0.70% 손절을 거부하는 최소값 1.3944)인데 같은 `k`가 US에 최소 손절 2.97%를 강제한다. `k`는 이 change의 손절 폭 분포 측정에서 나와야 한다. 정공법인 ATR 상대는 TossOS에 ATR이 없어 별도 change 소관이다.
- 임계값 재보정 자체, 슬리피지 모델, `decisions` preimage·해시 변경, 새 주문 경로.
- **target을 청산이 소비하게 만드는 일** — `internal/exitpolicy`는 target을 읽지 않는다(ladder는 고정 % rung, ratchet은 참조 0건). 이 change는 그 사실을 **측정으로 드러내되**(선언 target vs 실제 청산가) 고치지는 않는다.

## Capabilities

### New Capabilities

(없음)

### Modified Capabilities

- `risk-management`: 관측 기록·결손 복원·순 RR 관측 요구사항 **추가 3건**. 여기에 **MODIFIED 2건**(라운드 2) — ① "발급 절차와 예약의 권위"는 순서 있는 SHALL이고 이 change가 그 시퀀스 뒤에 쓰기를 두므로 seam을 명시한다 ② "No Stop = No Trade와 위험 기반 수량"의 임계값 provenance를 정정한다: 인용된 2.0은 총·구조 RR 기준이며 순 기준으로 이전 적용할 수 없다(순 기준 선례는 시장별 KRX 1.5 / US 2.0, 사후 분석 처방 1.3). **게이트 판정 로직은 무변경** — 순 RR은 관측값이다.
- `trade-analytics`: 반사실 판정 측정 요구사항 추가. 기존 "분석 경로의 격리"(실행 경로 무영향·강화 트리거 아님·보존 기간)에 종속되며, 격자의 기하 외 단계 비구속 고정과 **산출 가능/불가의 선언**을 포함한다.

## Impact

- Affected code: `internal/journal`(schemaV8 신규 테이블 — additive, 기존 테이블·preimage 무변경, 보존 정책, 안티조인 재구성), `internal/risk`(순 RR 관측 산출 함수 신설 + 판정에 정지 단계 additive 확장, `checkMinRewardRisk` 무변경), `internal/execgw`(관측 기록 배선 — 거부 경로가 처음으로 영속, 진입 한정), `internal/obs`(비강화 등급 이벤트 타입 — 라운드 2 R2-13), 반사실 하네스, 문서 4곳.
- **위험 등급: High-risk**. WORKFLOW High-risk 열거의 두 항목에 해당한다 — **원장 스키마**(schemaV8 신규 테이블)와 **Guardian**(기록 배선이 판정 경로에 붙는다). 판정 로직 자체는 바뀌지 않지만 등급은 경로로 정해지므로 full TDD + race/crash 테스트 + Pre-Edit 선언 + 적대적 Eng 리뷰가 요구된다. 라운드 1이 적대적 Eng 2보이스로 이미 걸렸으므로 2판 리뷰는 추가분(관측 기록의 실행 경로 결합·스키마·하네스 격리)에 대해 실행한다(WORKFLOW: requirement 변경 재리뷰).
- §0 검토: §0.9 미개입(어떤 판정도 완화·강화되지 않음 — 이 change의 정의), §0.3 무관, §0.4 새 외부 호출 0, §0.6 additive 신규 테이블(하향 마이그레이션을 만들지 않는다 — 원장 계약대로 롤백은 이전 바이너리 실행·`ErrSchemaTooNew` 거부이고 중간 실패 복구는 자동 사전 백업이다), §0.5 관측 기록 자체가 audit 강화.
- upstream 회귀: `internal/risk`·`internal/journal`의 해당 표면은 TossOS 신규이므로 상속 테스트 650개와 무관. **라운드 1의 P1-B(공용 fixture 전멸)는 발생하지 않는다** — ALLOW/REFUSE가 불변이므로 `entryInput()`·`guardianIntent()` 재기준선이 필요 없다.
- **후속 게이트 change의 착수 조건 4개 중 이 change가 주는 것과 못 주는 것**(라운드 2 R2-10 — 원안은 넷 다 산출한다고 과대 선언했다):

  | 착수 조건 | 이 change | 근거 |
  |---|---|---|
  | ① 총·순 RR **경계면 지도** | ✅ 합성 격자 | 실제 의도의 *분포*는 아니다. ALLOW 쪽은 본전 미달 거부로 **좌측 절단**돼 순 RR ≤ 0이 나타나지 않는다 |
  | ② 손절 폭 분포(`k` 근거) | ⚠️ **실거래 모집단에만 의존** | 격자의 손절 폭은 작성자가 고른 값이므로 거기서 `k`를 유도하면 순환이다. 추출 실패 시 `k`는 **미결로 남는다** |
  | ③ 시장별 실측 `B/entry` | ❌ | 요율 7개가 미측정이므로 지금 산출값은 placeholder의 재서술이다. 2b 요율 실측 이후 |
  | ④ 선언 target vs 실제 청산가 | ❌ | 종결 포지션이 필요하다. 이 change는 목표가를 기록해 두는 것까지만 한다 |

  즉 이 change 완료만으로 후속 게이트 change가 임계값과 `k`를 모두 정할 수 있게 되지는 않는다. ③④는 별도 선행이며 ②는 실거래 추출 성패에 달렸다.
- **측정 모집단의 한계(명시)**: 오늘 프로덕션 Guardian **진입** 판정은 0건이다 — `evaluateChain` 도달 경로는 `RiskGuardian` 발급뿐이고 그 호출자인 `Tracer`에 프로덕션 호출자가 없다. `tossctl engine run`은 landed되었으나(`add-engine-runtime` archive) 그 루프 집합은 reconcile·exit·체결 감지이고 진입 의도를 만들지 않으며, 인터록 조항 6(`ProtectionReady=UNWIRED`)이 자동 진입을 기계적으로 막는다. 따라서 라이브 관측은 tracer 슬라이스 또는 게이트 ON 이후에 쌓인다. 반사실 하네스가 이 change의 즉시 산출물인 이유이며, 관측 배관을 먼저 두는 이유는 **첫 판정이 기록 없이 지나가지 않게** 하는 것이다.

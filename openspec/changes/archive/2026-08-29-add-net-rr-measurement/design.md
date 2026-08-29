## Context

라운드 1 리뷰(4보이스, `review.md`)가 `harden-net-rr-gate`의 세 핵심 주장을 반증했다. 그중 이 change의 형태를 정한 것은 둘이다.

- **§0.9 단조성 증명이 거짓**: `순 RR < 총 RR`이 전칭 성립한다고 주장했으나 ① 요율 전부 0인 configured 모델에서 등호가 되고 ② `B`가 float64 유래라 고정밀 진입가에서 `B < entry`가 되어 **순 > 총**이 된다(재현 확인: 총 2−3e-18 거부 → 순 2.0 허용). 게이트를 걸지 않으면 §0.9가 개입하지 않으므로 이 반례들은 안전 결함에서 **기록값 정밀도 문제**로 내려간다.
- **측정 배관이 실재하지 않음**: `decisions`는 발급된 판정만 담고 reason/detail 열이 없다. 거부는 `chainRefusal`이 `IssueRefusal{Stage,Reason,Detail}`을 메모리에 만들어 반환할 뿐 **영속되지 않는다**. 따라서 "두 값을 Detail에 담아 P3 markout이 쓴다"는 원안의 설계는 성립하지 않았다.

한편 **ALLOW 모집단의 절반은 이미 landed다**: `journal.RiskIntent` preimage가 `entry_price`·`stop_price`·`target_price`·`market`·`policy_version`을 정규 JSON으로 영속하고 해시한다([decision.go:63](../../../internal/journal/decision.go)). 빠진 것은 **어떤 요율로 본전을 계산했는지**다 — 요율이 바뀌면 사후 재계산이 당시 판정과 달라진다.

현재 journal `SchemaVersion = 7`([schema.go:6](../../../internal/journal/schema.go)).

## Goals / Non-Goals

**Goals:**
- 거부 모집단에 첫 영속 경로를 만든다. 오늘은 판정의 절반이 흔적 없이 사라진다.
- 순 RR을 산출해 기록한다 — 게이트가 아니라 관측값으로.
- 임계값·`k`를 추측이 아니라 데이터에서 정할 수 있게 한다(반사실 하네스).
- 어떤 의도의 ALLOW/REFUSE도 바꾸지 않는다.

**Non-Goals:**
- 게이트 기준 전환, 최소 손절 폭, 임계값 확정, 슬리피지 모델, preimage·해시 변경, target을 청산이 소비하게 만드는 일.

## Decisions

### D1. 자기완결 관측 테이블 (schemaV8), preimage는 불가침

`RiskIntent.Canonical()`에 필드를 더하면 `risk_hash`가 바뀌어 결정 계약과 골든 테스트가 깨진다. 그래서 **preimage는 건드리지 않는다**. 관측은 별도 additive 테이블에 둔다.

관측 테이블은 **자기완결**로 한다 — 가격 3개·시장을 직접 담고, 발급된 판정에만 결정 참조를 nullable로 채운다.

*대안(기각) — 결정 참조만 두고 preimage를 조인*: 거부에는 결정이 없어 애초에 절반이 불가능하고, 분석 쿼리가 계약 테이블을 조인하게 되어 trade-analytics의 "분석 경로의 격리"와 어긋난다.
*대안(기각) — `decisions`에 additive-nullable 열 추가*: `decisions`는 계약 테이블이다. 관측을 섞으면 계약의 표면이 분석 요구에 따라 자라게 된다.

가격 3개가 preimage와 중복되는 것은 의도적이다. 둘은 같은 입력에서 같은 시점에 쓰이므로 어긋나지 않고, 계약(해시된 진실)과 관측(분석 사본)의 역할이 다르다. 이 중복이 D6의 결손 재구성을 가능하게 만든다.

**결정 참조는 외래키가 아니다**(라운드 2 R2-4). `foreign_keys(on)`이므로 FK를 걸면 계약 테이블의 만료·정리가 분석 행에 막히거나 전파된다 — 같은 파일의 `spent_nonces.decision_id`가 정확히 그 이유로 제약을 두지 않는다("pruning expired decisions must not be blocked by (or cascade into) these rows"). FK는 관측 선행 쓰기도 불가능하게 만들고 자기완결 주장과도 충돌한다.

**결과 구분 열거가 필요하다**(라운드 2 R2-5). `IssueEntry`는 체인을 한 번 평가한 뒤 발급을 재수집 루프로 돌리므로 체인 ALLOW이 `LIMIT_REACHED`·`DECISION_EXPIRED`·`VERSION_CONFLICT`·`SNAPSHOT_RECOLLECTION_EXHAUSTED`로 죽을 수 있다. 참조가 null이라는 사실만으로는 "거부" / "체인 ALLOW·발급 거부" / "크래시 결손" 셋을 구별할 수 없으므로 `REFUSED_CHAIN`·`ALLOWED_ISSUED`·`ALLOWED_ISSUANCE_REFUSED`를 명시 열거로 담는다.

**정지 단계는 판정이 직접 보고해야 한다**(라운드 2 R2-6). `risk.Decision`은 `{Allowed, Reason, Detail}`뿐이고 단계 필드가 없다. `ReasonInputUnavailable`은 `internal/risk`에서 **42곳**이 발생시키므로 reason→rung 역산은 다대일이 되어 "셋업이 없어서"와 "임계값이 틀려서"를 구분한다던 열이 조용히 틀린다. 판정 값에 단계를 담는 additive 확장이 필요하다.

### D2. 비용모델 지문 + 비용 범위 표기

관측에 담을 두 메타데이터:

- **비용모델 지문**: 산출에 쓰인 요율 집합의 식별자. 요율이 실측으로 교체되면(2b) 이전 관측과 이후 관측을 섞어 집계하는 것이 오류임을 지문이 드러낸다. `DefaultModel`의 7개 요율은 전부 `[미검증 — 2b 실측 대체 대상]`이므로 이 표기가 없으면 미검증 수치가 검증된 것처럼 읽힌다.
- **비용 범위**: `FEE_TAX_ONLY`. 슬리피지(058 실측 ~0.13%p)를 포함하지 않으므로 지표 이름은 "수수료·세금 차감 후 RR"이다. 원안이 proposal에서 56.4%(슬리피지 포함)를 인용하면서 Non-Goals로 슬리피지를 제외한 불일치를 이 표기가 구조적으로 막는다.

### D3. 순 RR 산출식과 그 정밀도 한계

`순 RR = (target − B) / (B − stop)`, `B = BreakEvenSellPrice(entry, "1", market)`. 손절 계약 rung이 목표가를 비교하는 값과 **동일한 본전**을 쓴다 — 본전 정의가 두 곳으로 갈라지면 두 소비자가 같은 사실을 다르게 말한다. 수량 1은 이익 하한 0에서 수량이 약분되기 때문이며, 손절 계약 rung이 이미 같은 근거로 그렇게 한다.

`RewardRisk`(총)는 존치하고 시그니처를 바꾸지 않는다. 두 값을 함께 기록하는 것이 이 change의 목적이므로 둘 다 필요하다.

**정밀도(라운드 1 P1-C의 정직한 처분)**: `B`는 float64 산술 후 `formatAmount`로 렌더되고 `internal/risk`는 그 문자열을 유리수로 정확히 파싱한다. 따라서 관측 순 RR은 float64의 bounded-relative-error를 물려받는다. 관측값에서 이는 마지막 유효자리의 문제이고 어떤 판정도 좌우하지 않는다. **게이트로 승격하는 change는 이것을 먼저 해결해야 한다** — 재현된 사례에서 실제 1.99999999999997인 의도를 구현은 2.0000000000000107로 계산했다. 보수적 처방은 `B`를 게이트 직전 **위로** 올림하는 것(본전이 커지면 순 RR이 작아지므로 단조 안전)이거나 유리수 end-to-end다. 이 change는 그 선택을 하지 않고 한계를 기록한다.

### D4. 반사실 하네스 — 두 모집단, 외부 데이터는 선택적

라이브 진입 관측은 오늘 0건이다: `evaluateChain` 도달 경로는 `RiskGuardian` 발급뿐이고 그 호출자 `Tracer`에 프로덕션 호출자가 없다. `tossctl engine run`은 landed되었지만 그 루프 집합(reconcile·exit·체결 감지)은 진입 의도를 만들지 않고, 인터록 조항 6이 자동 진입을 막는다. `internal/replay`는 99바이트 스텁이다. 배관만 두면 tracer 슬라이스 또는 게이트 ON까지 계기판에 눈금이 없다.

- **합성 격자**: 진입가 × 손절 폭 × 목표가를 격자로 만들어 시장별 통과·거부 경계면과 임계값 후보(1.3 / 1.5 / 2.0)별 거부율을 산출. 외부 의존 0이므로 **이것만으로 측정이 성립한다**.
  - **fixture 계약**(라운드 2 R2-C): `risk.Evaluate`는 순수 함수이고 `Input`은 전부 평문 값이라 journal·브로커가 필요 없다. 다만 12개 rung 전부가 fixture를 소비하므로 기하 외 단계를 **비구속으로 못 박아야** 한다 — kill switch off·mode NORMAL·latch false·심볼 allowlist 포함·주문 크기 여유·현금 충분·당일 진입 0·개방 노출 0·일손실 0·중복·미체결 없음. 그러면 구속되는 것은 `stop_contract`(본전 미달)와 `min_reward_risk` 둘이고, **이 둘은 분리 보고**한다. 합치면 "본전을 못 넘음"과 "RR이 낮음"이 한 숫자로 뭉개진다.
  - **US 도달 가능성 한계**(라운드 2 R2-9): `checkOrderSize`가 `min_reward_risk`보다 **앞**이고 교차 통화를 `INPUT_UNAVAILABLE`로 거부하는데 `DefaultPolicy()`는 전 항목 KRW다. 따라서 US 격자점은 provenance 없는 USD 한도를 발명해야 성립하고, 그 산출물은 **조작된 정책 위에 있음을 표기**해야 한다("정책 수치의 provenance" SHALL과의 정합).
  - **좌측 절단**(라운드 2 R2-12): rung 5가 본전 미달 목표가를 이미 거부하므로 ALLOW 쪽에는 순 RR ≤ 0이 구조적으로 나타나지 않는다. 명시하지 않으면 후속 change가 절단을 분포의 성질로 오독한다.
- **실거래 기록** — 선택이 아니라 **`k`의 유일한 비합성 출처**(라운드 2 R2-10): 선행 시스템의 실제 진입 3가격을 통과시켜 "그 게이트였다면 거부했는가"를 보고, 그 손절 폭이 실제 분포를 준다. 격자의 손절 폭은 작성자가 고른 값이라 거기서 `k`를 유도하면 순환이므로, **추출이 실패하면 `k`는 미결로 남는다고 산출물에 표기한다.** 경로는 실행 인자로 받고 읽기 전용이며 **저장소에 반입하지 않는다**(로컬 DB 커밋 금지). 추출 불가 시 사후 분석 문서의 표(≤8행)를 fixture로 쓰고 그 출처와 표본 수를 표기한다.

하네스는 주문을 만들지 않고 원장에 쓰지 않는다. 실패는 분석 실패이며 운영 모드 강화 트리거가 아니다(risk-management의 열거형뿐).

### D5. 게이트 무변경이 만드는 것

`checkMinRewardRisk`를 건드리지 않으므로 라운드 1의 **P1-B가 소멸한다**: `entryInput()`(`requireAllowed` 19곳)·`guardianIntent()`(execgw 발급 스위트)·`TestRewardRiskAtTheMinimumPasses`·US 케이스의 ALLOW/REFUSE가 전부 그대로다. 재기준선 작업이 없다.

또한 라운드 1의 **P3-L이 해소된다**: 원안의 "순 RR 계산 불가 → MIN_RR_NOT_MET" 시나리오는 손절 계약 rung이 동일 호출을 이미 `INPUT_UNAVAILABLE`로 거부하므로 도달 불가였다. 관측에서는 본전 산출 실패가 판정을 바꾸지 않고 **관측 항목의 결측**으로 기록된다 — 하나의 사실에 두 reason code가 생기지 않는다.

### D6. 쓰기 순서 — 원자 트랜잭션 **밖**, 결손은 복원 가능 (라운드 2 P0, 두 보이스가 충돌한 지점)

`RecordDecisionAndReserve`는 `BeginTx`(BEGIN IMMEDIATE) → precheck → 결정 삽입 → 예약 → `Commit`이며 `defer tx.Rollback()`이 걸려 있다. 관측을 어디에 두느냐가 두 P0를 동시에 결정한다.

- **안에 두면**: 관측 삽입 오류가 결정·예약을 함께 롤백한다. 디스크 풀이나 `STRICT` 타입 위반 같은 **순수 측정 결함이 진입을 막는다** — 이것이 정확히 R2-1이 CRITICAL로 지적한 "측정 실패가 거래를 멈춘다"를 더 짧은 경로로 재현한다. 기각.
- **밖에 두면**: 커밋 후 크래시 시 관측 없는 결정이 남아 ALLOW 모집단이 결손된다 — 이 change가 고치려는 결함 그 자체.

**채택: 밖 + 복원 가능성.** 결손을 수용하지 않고 상계한다 — 관측 없는 발급 결정을 계약 테이블 안티조인으로 탐지하고, preimage가 이미 가격 3개·시장·정책 버전을 담으므로 **결정론적으로 재구성**한다. 재구성은 그 시점 요율을 쓰므로 원본과 같게 읽혀서는 안 되고, 재구성 표지와 재구성 시점 지문을 함께 남긴다.

거부 측은 트랜잭션이 없어 단독 쓰기이고 참조할 preimage가 없으므로 **결손이 복원 불가**다. 돈이 걸리지 않으므로 수용하되 계수한다 — 이 비대칭을 스펙에 명시한다.

관측 실패의 오류 분류는 landed 규칙을 그대로 쓴다: `IssueRefusalReason`이 인식하지 못하는 오류는 위험 거부가 아니다("a database error is a bug or an outage"). 따라서 관측 실패에 reason code가 붙지 않고, 체인 판정은 여전히 ALLOW으로 남는다.

### D7. 관측 실패 알림은 비강화 등급 — 기존 두 등급 어느 것도 맞지 않는다 (라운드 2 P0)

`SeverityCritical`은 durable outbox에 들어가고, 전달 소진은 `Gate.Block(ReasonAlertUndelivered)`로 **즉시 in-process 진입 latch** + `EscalateOperatingMode(ModeTriggerCriticalAlertUndelivered)` → **ENTRY_BLOCKED durable**을 유발한다. 해제는 OPERATOR 승인·audit가 필요하다. 즉 critical로 올리면 v8 열 오타가 사람 승인을 요구하는 거래 정지가 된다 — 메인 스펙의 "분석·성과 작업 실패는 트리거가 아니다(SHALL NOT)"와 정면 충돌.

부수 결합도 있다: `outbox.UndeliveredCount`는 전역이라 계좌·타입·등급 필터가 없고, 측정 알림 1건이 운영자의 진짜 IN_DOUBT latch 해제를 막는다. `deliver()`는 뮤텍스를 쥔 채 재시도하므로 실제 인시던트 알림 경로를 직렬화하고 그 지연을 진입 경로(재수집 10초·TTL 60초)에 주입한다.

`SeverityNormal`로 낮추면 best-effort·droppable이라 "조용히 넘어가서는 안 된다"와 충돌한다. **그래서 durability를 알림에서 분리한다**: 실패는 계수하고 알림은 `SeverityNormal`로 보낸다. 알림의 전달 여부가 거래 가능성을 좌우하지 않으면서 기록되지 않은 판정의 수를 조회할 수 있다. `obs.Severity`에는 `critical`과 `normal`만 있으므로 "비강화 등급"이라는 제3의 등급을 만들지 않고(라운드 3 R3-10), 관측 실패 이벤트 타입의 **`criticalEvents` 표 비구성원성**을 테스트로 고정한다 — 그 표의 구성원 여부가 진입 게이트 차단과 자동 강화로 이어지는 유일한 구조적 스위치이고, 기존 표 테스트는 포함만 단언한다.

**열화 계수를 관측 테이블에 두면 안 된다**(라운드 3 R3-3 — 라운드 2 처분의 자기모순): 관측 INSERT가 `SQLITE_FULL`·I/O 오류·스키마 오류로 실패한 상황이라면 **같은 저장소의 열화 쓰기도 함께 실패한다.** 계수는 관측 행과 **독립된 실패영역**에 두고, 독립 저장마저 불가능하면 구조화 로그와 프로세스 내 단조 카운터로 강등하되 거래 경로는 계속하며 그 강등된 계수를 재시작 후에도 durable하다고 주장하지 않는다.

## Risks / Trade-offs

**[오늘은 아무것도 막지 못한다]** → 이 change는 StockOS의 손실 기하(0.70% 손절 + 차감 후 RR 0.88)를 거부하지 않는다. 완화 근거는 인터록이다: `ProtectionReady=UNWIRED`로 자동 진입이 기계적으로 막혀 있어 지금 통과시킬 의도가 없다. 게이트가 실제로 필요해지는 시점은 진입이 열리는 시점이고, 그때까지 이 change가 그 게이트의 숫자를 만든다. **이 트레이드오프를 받아들이는 것이 A안 선택의 내용이다.**

**[측정이 늦게 도착한다]** → 라이브 모집단은 엔진 런타임 이후다. 반사실 하네스가 즉시 산출물을 주지만 합성 격자는 실제 시장 분포가 아니다. 완화: 임계값 확정을 라이브 관측 이후로 못 박고(risk-management delta의 SHALL), 합성 산출물은 "경계면 지도"로만 쓴다.

**[미검증 요율이 산출물을 오염시킨다]** → 7개 요율 전부 미측정 과대추정이다. 완화: 지문·범위 표기를 SHALL로 걸어 2b 실측 전후 관측을 섞을 수 없게 한다. 게이트가 없으므로 오염의 결과는 잘못된 숫자이지 잘못된 거부가 아니다.

**[관측 기록이 실행 경로에 붙는다]** → 판정마다 쓰기가 생긴다. 완화: 기록 실패가 판정을 바꾸지 않고(SHALL NOT) 알림만 발생한다. 다만 조용히 넘기지도 않는다 — 기록되지 않은 판정은 측정에서 영구히 사라진다.

**[중복 저장]** → 가격 3개가 preimage와 관측에 이중으로 존재한다. 완화: D1의 근거(계약 vs 분석 사본, 같은 입력·같은 시점). 어긋남을 잡는 테스트를 둔다.

## Migration Plan

schemaV8 = 신규 테이블 1개 + 인덱스 + 보존 정책. 기존 테이블·열·인덱스·preimage·해시 무변경. 마이그레이션 순서: v7 → v8 단일 additive 스텝.

**롤백은 "테이블 drop"이 아니다**(라운드 2 R2-8 — 원안이 원장 계약을 어겼다). `internal/journal/schema.go`가 명시한다: *"There is no down-migration and there will not be one"* — 하향 스텝은 열을 drop해야 하고 그것은 규칙 3이 금지한다. 원장의 롤백은 **이전 바이너리를 실행하는 것**이고, 그 바이너리는 새 `user_version`을 오독하지 않고 `ErrSchemaTooNew`로 거부한다. 마이그레이션이 중간에 실패했을 때의 복구는 **자동 사전 백업**(`backup.go`)이다. 이 change도 그 계약을 그대로 따른다.

## Open Questions

- 선행 시스템 실거래 DB에서 진입 3가격을 추출할 수 있는지는 실행 시 확인한다. 불가하면 사후 분석 문서의 표를 fixture로 쓰고 출처를 표기한다(D4).
- 라이브 관측이 몇 건 쌓여야 임계값을 정할 수 있는지(표본 수 하한)는 이 change가 정하지 않는다. 후속 change가 분포를 보고 정한다 — 지금 숫자를 박으면 그것이 또 하나의 추측 노브가 된다.
- 선언 목표가와 실제 청산가의 관계는 포지션이 종결되어야 측정된다. 이 change는 목표가를 기록해 두는 것까지만 한다.

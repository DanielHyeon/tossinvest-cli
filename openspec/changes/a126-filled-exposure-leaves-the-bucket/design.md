# a126 design — 체결 노출은 bucket 을 떠난다 (freeze)

> 위험 등급 **High-risk** (사이징·원장). 이 문서는 freeze 로트 산출물이며 생산 코드는 한 줄도 바꾸지 않았다.
> 분기·early return 을 근거로 쓰는 문장은 전부 `analysis/freeze-ast/census.md`(AST 가 열거한 전부, base
> `989ab031`)의 `함수 Bn` 을 인용한다. 손으로 읽은 좌표는 "(읽음)" 으로 표시한다 — 그것은 분기 주장이 아니라
> SQL 문자열·쓰기 자리의 위치다.

## 증거 기반

| 종류 | 무엇 | 어디 |
|---|---|---|
| 기준 커밋 | `989ab031` (0.1 재고정 `d8ed3de1`) | `base-commit.txt` |
| AST 열거 | 18 함수, base blob 과 sha 대조 후 추출 | `analysis/freeze-ast/{extract.py,census.md,ast/}` |
| CodeGraph 1.6.0 callers | `ReadJournalBucketUsage` 6(생산 4: `refuseStaleBucketUsage`·`riskBucketSharedUsage`·`RevalidateQFinalAdmission`·`loadProductionRiskEntries`), `smallestRecordedBucketLimit` 1, `releaseRiskBucketOwner` 9(**전부 `_test.go`**) | 2026-09-30 질의 |
| 쓰기 자리 census (`rg`, 비시험) | `risk_bucket_reservations.filled_minor` 를 0 이 아닌 값으로 쓰는 자리 = `persistRiskBucketFillTransition` 의 UPDATE 하나(`risk_bucket_fill.go:1068`, 읽음). `risk_bucket_owners.released_at` 과 `risk_bucket_owner_release_receipts` 의 쓰기 = `releaseRiskBucketOwner` 안 `risk_bucket_owner.go:1017`·`:1029` 뿐. `risk_bucket_scope_latches` 의 DELETE 는 0 | 2026-09-30 |
| 운영 원장 실측 | `~/.config/tossctl/journal.db` read-only(`mode=ro`), 건수만: `user_version=32`, `risk_bucket_owners` 0 · `risk_bucket_owner_release_receipts` 0 · `risk_bucket_reservations` 0 · `risk_bucket_fills` 0 · `risk_bucket_scope_latches` 0 | 2026-09-30 10:01 KST |
| memory recall | `scripts/memory-recall.sh "risk bucket filled usage release"` → 0건. GBrain MCP 는 이 세션에서 연결 실패(advisory, 생략) | — |

## 현재 동작 (문제의 기전)

- 사용량의 **유일한** 계산은 `riskbucket.ReadJournalBucketUsage` → `readProductionRiskUsage` 의 SQL(계좌·dimension·value
  의 모든 예약 행, owner 해제 여부 필터 없음, 읽음 `production_snapshot_authority.go:453`) → `aggregateProductionRiskUsage`
  B1(모든 행 합산)이다. 한도 비교 모집단은 별도 함수 `smallestRecordedBucketLimit` 의 SQL
  `state IN ('HELD','FILLED') OR filled_minor<>'0'`(읽음 `risk_bucket_usage.go:91`)이다.
- `filled_minor` 는 단조 증가만 한다: 쓰기 자리는 `persistRiskBucketFillTransition` B22 하나이고, 쓰는 값은 저장값에
  `riskMinorMonotoneDelta(next, previous)`(L1013, 음수면 B6 오류)를 `addRiskMinor` 로 더한 것이다(L1032, 읽음). 감소 쓰기는
  없다. 그래서 owner 가 해제돼도 그 행은 계좌 bucket 합에 영원히 남는다.

## Q1 — 떠나는 사건: **② owner 해제(release receipt)**

**결정.** 포지션 귀속 노출이 소멸했다는 원장 사실은 `risk_bucket_owner_release_receipts` 행(이하 "영수증")이다. 영수증이
커밋된 owner 에 귀속된 예약 행은 사용량 합과 한도 모집단을 떠난다.

**귀속 열 (스키마 영수증).**
- 예약 행은 Position id 를 갖지 않는다. 귀속 열은 owner 키 `(account_ref, market, symbol, owner_prospective_generation)`
  와 `decision_id` 다(`schemaV22` `risk_bucket_reservations`, 읽음 `risk_bucket.go:703–710`).
- owner → Position 결속은 `risk_bucket_owners.actual_generation`(set-once bind, `bindRiskBucketOwnerActualInTx`)이다.
- 예약 행과 **종결된 Position generation** 을 한 행에서 잇는 것은 영수증뿐이다: PK 가 owner 키이고 `actual_generation`·
  `position_id`·`position_version`·`reconcile_state_id`·`observation_id`·`predecessor_state_digest` 를 들고 있으며
  UPDATE/DELETE 가 트리거로 막혀 있다(`schemaV24`, 읽음 `risk_bucket_owner.go:79–105`).
- 따라서 D5 의 "authoritative Position/fill projection 에 **귀속된** filled exposure" 가 끝났다는 사실의 원장 표현은 영수증이다.

**②가 ①·③을 포함한다 (AST 인용, `Journal.releaseRiskBucketOwner`).** 영수증 INSERT(B47)에 닿으려면:
- ① 포지션 종결: B22 `positionState != "CLOSED" || closedAt == ""` → 거절, B23 수량 ≠ 0 → 거절, B21 최신 generation 아님 → 거절.
- ③ 대사 확인 0 보유: B36 활성 reconcile → 거절, B37 공식 broker-zero 관측 권위 없음 → 거절,
  B40 그 관측이 모든 선행 사건(예약·주문·체결·보호·intent·mutation·watermark·조정 시각, B38 순회)보다 엄밀히 뒤가 아님 → 거절.
- 그 외 청결: B14 campaign CLOSED, B16 campaign claim 0, B26–B28 검사표(legacy_held · bucket_held · pending_entry ·
  owner_latch · scope_latch · unresolved_fill ×2 · protection_saga · protection_order · sell_claim), B31 미해소 BUY
  mutation, B34 미해소 SELL mutation, B9 상태 봉인 대조, B45 owner 경합.

**①·③ 단독을 버린 이유.** ① `positions.state='CLOSED'` 는 예약 행과 owner 경유로만 이어지고, 단독으로는 late fill ·
보호 saga · 미해소 SELL · 미확정 actual(UNKNOWN) 을 배제하지 못한다(위 검사표가 그 목록). ③ reconcile 0 단독은
generation·체결 해소를 묶지 않는다. 둘 중 하나로 감소하면 원장이 아직 노출을 품을 수 있는 상태에서 진입 cap 이 열린다(fail-open).

**현재 도달성.** `releaseRiskBucketOwner` 는 package-private 이고 CodeGraph callers 9 가 전부 시험이다(a066 6.5 잔여 #5:
"owner release is unreachable in production today"). 운영 원장의 영수증은 0 이다(실측). 그러므로 **이 change 의 생산 효과는
owner 해제가 배선될 때(활성화 로트)까지 0 이다.** 이것은 결함이 아니라 순서다: 떠나는 사건을 해제보다 느슨한 사실에
걸면 배선 없이도 효과가 나지만 그것은 위 fail-open 이다.

## Q2 — 소급인가 전방인가: **파생(저장값 불변), 스키마 변경 없음**

**결정.** 감소는 저장된 `filled_minor` 를 바꾸지 않는다. 사용량 reader 가 영수증 존재로 **파생**한다.
- `filled_minor` 는 역사적 사실로 남는다. 상태 봉인(`verifyRiskBucketStateDigest`)·해제 영수증의
  `predecessor_state_digest`·완화 경로의 `ExpectedStateDigest` 가 그 값을 봉인하므로, 행을 고치면 봉인이 깨진다.
- 새 테이블·열·트리거가 필요 없다(영수증 v24·scope latch v22 가 이미 있다). 마이그레이션 0, rollback 은 코드 되돌림뿐.

**소급성.** 파생 규칙은 정의상 이미 있는 영수증에도 적용된다(소급). 적용 대상 = 운영 원장의 영수증 = **0 건**(2026-09-30
10:01 KST, v32 원장 실측). 영수증을 만드는 생산 경로도 0 이다(Q1 도달성). 그래서 오늘 소급으로 바뀌는 사용량은 없다.

> **사람 승인 항목 H1 (표기만 — 실행 금지).** 활성화 로트가 owner 해제를 배선하기 직전, 운영 원장의 영수증 수를
> 다시 재야 한다. 0 이 아니면 그 owner 들의 사용량이 배포 순간 한꺼번에 떠나므로(소급) 사람이 그 목록을 보고
> 승인한다. 운영 원장 재계산·쓰기는 이 change 가 하지 않는다.

## Q3 — overage latch 재계산과의 순서: **latch 해제 → owner 해제(=떠남) → 다음 체결의 재계산**

AST 사실:
- `recomputeOverageLatches` 는 `ApplyFill` 의 두 자리에서만 불린다: B23(actual 보완) · B38(새 체결). 다른 생산 호출자 없음.
- 그 함수는 latch 를 **세우기만** 한다: B9 overage>0 일 때만 `anyOverage=true`, B11 기존 overage 보다 클 때만 올림,
  B12 `!anyOverage` 면 아무것도 지우지 않고 반환. 지우는 갈래가 없다.
- 공유 bucket 의 타인 사용량은 B6 `SharedUsedMinor` 로 들어오며, 그 값은 `riskBucketSharedUsage`(B1 순회, B4 합<자기몫 →
  ReplayMismatch) 가 `ReadJournalBucketUsage` 로 센다. **즉 이 change 가 reader 를 바꾸면 재계산은 자동으로 떠난 뒤의 합을 본다.**
- owner 해제는 그 owner 의 latch 가 남아 있으면 거절된다: `releaseRiskBucketOwner` B26–B28 검사표의 `owner_latch`
  (RISK_OVERAGE·UNKNOWN_ACTUAL_RISK) 와 `scope_latch`.
- RISK_OVERAGE 를 푸는 유일한 경로 `ReleaseRiskOverageLatch` 는 **활성** owner 만 받는다: B5(released_at IS NULL 행 없음 → Stale).

**정의된 순서.**
1. (latch 가 있으면) 운영자 해제 `ReleaseRiskOverageLatch` — OPERATOR·승인·audit(D8). 활성 owner 에서만.
2. owner 해제 = 영수증 커밋 = 사용량 떠남(같은 트랜잭션, 파생이므로 추가 쓰기 없음). latch 가 남아 있으면 1 로 되돌아간다.
3. 이후 **다른** owner 의 체결에서 `recomputeOverageLatches` 가 줄어든 공유 합으로 overage 를 판정한다.

**불변 조건.**
- 떠남은 어떤 latch 도 지우지 않는다. 이미 다른 owner 에 선 RISK_OVERAGE 는 그대로다(자동 경로는 latch 를 풀지 않음, D8).
  줄어든 합은 **다음** 체결의 판정과 신규 진입 여유에만 영향한다.
- reader 의 latch 집계(`Latched`·`OverageLatched`·`UnknownLatched`)는 떠난 행도 계속 센다 — 합에서만 뺀다. 정상 원장에서
  떠난 행의 latch 플래그는 0 이지만(해제가 owner latch 0 을 요구, 운영자 해제는 예약 플래그도 지움 — 읽음
  `risk_bucket_relaxation.go:356`), 손상 원장에서 그 플래그가 1 이면 진입 차단이 유지된다(보수).

## Q4 — 부분 체결·부분 종결: **감소 없음. 귀속 가능한 최소 단위는 owner generation 전체**

- 부분 **체결**(BUY)은 이미 수량 단위로 귀속된다: `risk_bucket_fills.delta_quantity`·`risk_bucket_fill_allocations
  (transfer_minor, filled_minor)`(schemaV23) — a066 D5 의 transfer/actual 규칙. 이 change 는 그것을 건드리지 않는다.
- 부분 **종결**(일부 매도)을 예약 행에 귀속하는 열은 **없다**: `applyRiskBucketFillInTx` B1 이 SELL 을 즉시 nil 로
  돌려보내고, `applyRiskBucketOwnerBindingInTx` B1 도 SELL 을 무시한다. risk bucket 테이블 어디에도 SELL 체결 행이 없다.
- 수량 비례 감소를 하려면 (a) SELL 체결 → owner 귀속 행과 (b) 여러 가격의 BUY 로 쌓인 금액을 어느 기준(FIFO·평균)으로
  떼는지의 정책이 필요하다. 둘 다 원장에 없고, 고르는 순간 사이징 정책 변경이다(§0.9 — 불명확하면 변경 금지).
- **결정:** 부분 종결은 사용량을 줄이지 않는다(보수, 과잉 차단 방향). owner 해제 때 그 owner 의 모든 예약 행(모든 leg ·
  모든 dimension)이 한꺼번에 떠난다. 부분 종결 감소는 잔여 **R2** 로 이름 붙여 남긴다.

## 설계 결정

### D1. "떠난 행" 의 정의는 reader 한 곳에만 있다

예약 행 r 이 **떠났다** ⇔ 다음이 모두 참:
1. r 의 owner 키로 영수증 행이 있다.
2. 같은 owner 키의 `risk_bucket_owners.released_at` 이 NULL 이 아니고 영수증 `released_at` 과 같다(두 쓰기가 한 트랜잭션,
   `risk_bucket_owner.go:1017`·`:1029`).
3. 같은 owner 키에 `risk_bucket_scope_latches` 행이 **하나도 없다**(D4).

이 판정은 `readProductionRiskUsage` 의 SQL 이 행마다 한 열(`departed`)로 계산하고 `aggregateProductionRiskUsage` 만 소비한다.
다른 곳에 같은 규칙을 다시 쓰지 않는다 — 판정이 둘이면 서로의 시험을 통과시켜 둘 다 살아남는다(코드와 DB 트리거 쌍의 선례).

### D2. 떠난 행은 합에서만 빠지고 검증과 digest 에는 남는다

`aggregateProductionRiskUsage` 에서:
- 떠난 행도 지금의 행 검증(B2: 금액·상태·RELEASED 의 held=0·snapshot·policy·식별자)을 그대로 받는다.
- 추가 fail-closed: 떠난 행이 `held_minor≠0` 이거나 `state='HELD'` 면 `ErrJournalUsageInvalid`(해제는 `bucket_held`=0 을
  요구했으므로 이것은 손상이다).
- 떠난 행의 `filled_minor` 는 합에 더하지 않는다. latch 플래그는 집계에 그대로 든다(Q3).
- `RowDigest` 의 parts 에 행마다 `departed` 값을 넣는다 — 같은 행 집합이 다른 합을 내면 digest 도 달라야 한다.

### D3. 한도 모집단은 사용량 모집단과 **같은 행, 같은 루프**다

a066 6.5 수리(#3a)는 "원장이 세는 행과 같은 모집단" 의 가장 작은 한도로 cap 한다고 적었다(읽음 `risk_bucket_usage.go:59–62`).
떠난 행이 합을 떠나고 한도 모집단에는 남으면 그 문장이 거짓이 되고, a066 6.5 좁힌 재리뷰가 이 결정에 묶어 둔 활성성 잔여
("이력 있는 bucket 에서 한도 상향이 실효가 없다")가 그대로 남는다.
- 구현: `readProductionRiskUsage` 가 snapshot `limit_minor` 도 읽고, `aggregateProductionRiskUsage` 가 같은 루프에서
  "모집단 행"(지금 규칙 `state IN (HELD,FILLED) OR filled≠0` **그리고** 떠나지 않음)의 최소 한도를 `JournalBucketUsage` 에
  돌려준다. `smallestRecordedBucketLimit` 는 그 필드로 대체되어 삭제된다 — 두 SQL 이 문자열로 같은지를 시험하는
  존재 검사가 아니라 구조로 같게 만든다.
- 이것은 한도 **선언** 체계를 바꾸지 않는다(비목표). 어떤 행이 "활성" 인지만 a066 문장대로 맞춘다.

### D4. 해제 뒤 scope latch 는 떠남을 되돌린다

- 해제된 owner 에 late BUY 체결이 오면 `latchReleasedOwnerLateFillInTx` 가 그 옛 owner 에 `ORPHAN_FILL` scope latch 를
  쓴다(B5 해제 owner 없음 → 반환, 아니면 B7). 그 late 체결 금액은 bucket 에 들어가지 않는다 — 주문 조회
  `queryRiskBucketOrder` 의 두 호출 WHERE 가 `ow.released_at IS NULL` 이라(읽음 `risk_bucket_fill.go:574`·`:577`)
  `applyRiskBucketFillInTx` 는 B4 `!found` 로 반환한다.
- 오늘은 해제된 owner 의 filled 가 합에 남아 **우연히** 그 공유 bucket 여유를 덮는다. 떠남이 그 덮개를 걷으면 late
  체결 노출이 공유 bucket(sector·strategy·horizon·market)에서 보이지 않게 된다(a066 잔여 #8 과 합쳐 fail-open).
- 그래서 D1-3: scope latch 가 하나라도 있으면 그 owner 의 행은 떠나지 않은 것으로 센다. scope latch 는 비시험 코드에서
  DELETE 되지 않으므로(census) 되돌림은 영구이며 보수 방향이다. 결과는 **오늘과 정확히 같은 합**이다(개선도 악화도 아님).
- 잔여 #8(late 체결이 공유 bucket 을 latch 하지 않음)은 그대로 활성화 로트 몫이다.

### D5. 저장값은 바뀌지 않으므로 체결·해제·완화 경로 코드는 편집하지 않는다

편집 대상은 reader 두 함수(`readProductionRiskUsage`·`aggregateProductionRiskUsage`), `JournalBucketUsage` 타입,
`refuseStaleBucketUsage`(한도 필드 소비), `smallestRecordedBucketLimit`(삭제)로 예상한다. `persistRiskBucketFillTransition`·
`releaseRiskBucketOwner`·`ReleaseRiskOverageLatch`·`recomputeOverageLatches` 는 편집하지 않는다(편집이 필요해지면 그
함수의 gate FLM 번들을 먼저 만든다 — "Logic Map 은 계획보다 항상 많다").

### D6. 소비자 영향

`ReadJournalBucketUsage` 의 생산 소비자 4 곳이 모두 떠난 뒤의 합을 받는다:
- `refuseStaleBucketUsage`: B7 snapshot 이 원장보다 **적게** 주장하면 stale 거절. 떠남 뒤 원장이 줄면 옛 snapshot 은
  **많게** 주장한다 → 받음(보수 — 그 snapshot 으로 계산한 q_final 은 더 작다). B12 한도 비교는 D3 의 한도로.
- `riskBucketSharedUsage` → Q3.
- `RevalidateQFinalAdmission`: `latchedUsageRefusal` 만 쓴다 → latch 집계 불변이므로 동작 불변.
- `loadProductionRiskEntries`: B10 latch 거절 불변, 합과 digest 는 D2 대로. B6 는 계좌·시장·종목의 **모든 generation**
  scope latch 를 거절하므로 D4 의 ORPHAN_FILL 종목은 여기서도 막힌다. 이 reader 는 스키마 27 에 고정돼 있어(a066 잔여 #7)
  지금 v32/v35 원장에서는 어차피 거절한다.

### D7. §0 안전 불변식 대조

- §0.3 손절·비상 청산: 무접촉. reader 는 EXPOSURE_RAISING 경로에서만 불린다(a066 D7). 체결·해제 경로는 편집하지 않는다.
- §0.9 사이징은 보수 방향만: 감소는 진입을 여는 방향이다. 그러므로 감소는 오직 영수증(=a066 D6 의 가장 강한 청결 증명)
  에만 걸리고, 모르는 것(scope latch, 손상 행)은 전부 "안 떠남" 또는 거절로 떨어진다.
- §0.6 스키마: 변경 없음.
- D8 "자동은 조이기만 한다" 와의 관계: D8 의 대상은 잠금·latch 의 **해제**다. 원장 사실에 결속된 사용량 감소는 a066
  D5 자신이 이미 자동으로 한다(cancel/expiry 의 HELD 해제). 떠남은 그 filled 짝이다. 이 해석은 사람 확인 항목 **H2** 로 둔다.

## 반증 설계 (구현 로트 1.x 가 세울 것)

"감소를 무조건 수행" 축의 변이는 **반드시** 잡혀야 한다(fail-open 축). 무변이 대조군이 GREEN 인지 먼저 보고, 양성 대조군
(떠나야 하는 행이 떠나는지)도 함께 둔다.

| 변이 | 뚫는 것 | 잡는 시험 |
|---|---|---|
| M1 `departed` 를 항상 참 | 활성 owner 의 filled 가 사라짐 | 활성 owner 사용량이 그대로인 시험 |
| M2 영수증 조건 제거(owner released_at 만 봄) | 영수증 없는 해제 표식 | released_at 만 있고 영수증 없는 행은 안 떠남 |
| M3 scope latch 되돌림 제거 | D4 fail-open | 해제 뒤 late BUY → 사용량 복원(RecordFill 경로 **와** 전략 정산 경로 각각) |
| M4 한도 모집단이 떠남을 따르지 않음 | D3 활성성 | 떠난 행의 작은 한도가 더는 cap 하지 않음 |
| M5 한도 모집단만 떠나고 합은 남음 | 모집단 분리 | 같은 시험 쌍 |
| M6 떠난 행의 held≠0 을 받음 | 손상 은폐 | 손상 원장 → `ErrJournalUsageInvalid` |
| M7 떠난 행의 latch 플래그를 집계에서 뺌 | latch 해제 우회 | 떠난 행 플래그 1 → 진입 거절 유지 |
| M8 부분 매도로 감소 | Q4 | 부분 매도 뒤 사용량 불변 |

실값 픽스처: a066 owner-lifecycle 픽스처(`closeRiskBucketOwnerLifecycle`, 주석 `risk_bucket_owner_test.go:802–804`)는
`filled_minor='0'` 으로 갭을 가렸다 — 1.4 에서 실제 체결 금액을 쌓는 픽스처로 바꾸고 주석을 지운다.

## 롤백

코드만 되돌린다. 스키마·저장값 변경이 없으므로 원장 되돌림이 없다. 되돌리면 사용량은 누적 합(오늘 동작)으로 돌아가며
이것은 더 보수적인 방향이다.

## 잔여 (이 change 가 닫지 않는 것)

- **R1** 생산 효과 0 — owner 해제 배선(a066 잔여 #5, 활성화 로트) 전까지.
- **R2** 부분 종결 감소 — SELL→owner 귀속 행과 금액 기준 정책이 필요(Q4).
- **R3** late 체결이 공유 bucket 을 latch 하지 않음 — a066 잔여 #8 그대로(D4 는 오늘 수준을 지킬 뿐).
- **R4** 이 change 의 창에 형제 커밋(`55963f29` a092, `internal/obs`)이 이미 들어왔다. 첫 Go 편집 전(1.0)에 사람 승인 base
  재고정 절차로 다시 옮긴다.

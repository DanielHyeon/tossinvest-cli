# a126 design — 체결 노출은 bucket 을 떠난다 (freeze)

> 위험 등급 **High-risk** (사이징·원장). 이 문서는 freeze 로트 산출물이며 생산 코드는 한 줄도 바꾸지 않았다.
> 분기·early return 을 근거로 쓰는 문장은 전부 `analysis/freeze-ast/census.md`(AST 가 열거한 전부, base
> `989ab031`)의 `함수 Bn` 을 인용한다. 손으로 읽은 좌표는 "(읽음)" 으로 표시한다 — 그것은 분기 주장이 아니라
> SQL 문자열·쓰기 자리의 위치다.
>
> 2판(2026-09-30): freeze 적대 리뷰 R1·R2·R3 의 처분을 반영했다(review.md). 1판과의 차이는 D3(한도 모집단 **불변**),
> D1 조건 4·부패 검출, D2(RowDigest 변경 철회), D4(되돌림 범위 축소와 잔여 명명), Q3(UNKNOWN 자동 해제 포함)다.

## 증거 기반

| 종류 | 무엇 | 어디 |
|---|---|---|
| 기준 커밋 | `989ab031` (0.1 재고정 `d8ed3de1`) | `base-commit.txt` |
| AST 열거 | 24 함수, base blob 과 sha 대조 후 추출(1판 18 + 리뷰가 지목한 6) | `analysis/freeze-ast/{extract.py,census.md,ast/}` |
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
  `riskMinorMonotoneDelta(next, previous)`(L1013 호출 — 그 함수 B3 `a.Cmp(b) < 0` 이면 오류, 호출자 B6 반환)를 `addRiskMinor`
  로 더한 것이다(L1032, 읽음). 감소 쓰기는 없다. 그래서 owner 가 해제돼도 그 행은 계좌 bucket 합에 영원히 남는다.

## Q1 — 떠나는 사건: **② owner 해제(release receipt)**

**결정.** 포지션 귀속 노출이 소멸했다는 원장 사실은 `risk_bucket_owner_release_receipts` 행(이하 "영수증")이다. 영수증이
커밋된 owner 에 귀속된 예약 행은 사용량 **합**을 떠난다(한도 모집단은 떠나지 않는다 — D3).

**귀속 열 (스키마 영수증).**
- 예약 행은 Position id 를 갖지 않는다. 귀속 열은 owner 키 `(account_ref, market, symbol, owner_prospective_generation)`
  와 `decision_id` 다(`schemaV22` `risk_bucket_reservations`, 읽음 `risk_bucket.go:703–710`). 같은 owner 키의 사본이
  `risk_bucket_final_decisions` 에도 있고 둘을 묶는 FK·CHECK 는 없다(D1 조건 4).
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
owner 해제가 배선될 때(활성화 로트)까지 0 이다.** 그 배선 로트의 착수 조건은 아래 「잔여」 R3 에 박는다.

## Q2 — 소급인가 전방인가: **파생(저장값 불변), 스키마 변경 없음**

**결정.** 감소는 저장된 `filled_minor` 를 바꾸지 않는다. 사용량 reader 가 영수증 존재로 **파생**한다.
- `filled_minor` 는 역사적 사실로 남는다. 상태 봉인(`verifyRiskBucketStateDigest`)·해제 영수증의
  `predecessor_state_digest`·완화 경로의 `ExpectedStateDigest` 가 그 값을 봉인하므로, 행을 고치면 봉인이 깨진다. 이 봉인들은
  owner 하나의 저장값만 덮고 계좌 전체 사용량을 봉인하지 않는다(R2 Q3 확인) — 파생 감소가 어떤 봉인도 흔들지 않는다.
- 새 테이블·열·트리거가 필요 없다(영수증 v24·scope latch v22 가 이미 있다). 마이그레이션 0, rollback 은 코드 되돌림뿐.

**소급성.** 파생 규칙은 정의상 이미 있는 영수증에도 적용된다(소급). 적용 대상 = 운영 원장의 영수증 = **0 건**(2026-09-30
10:01 KST, v32 원장 실측). 영수증을 만드는 생산 경로도 0 이다(Q1 도달성). 그래서 오늘 소급으로 바뀌는 사용량은 없다.

> **사람 승인 항목 H1 (표기만 — 실행 금지).** 활성화 로트가 owner 해제를 배선하기 직전, 운영 원장의 영수증 수를
> 다시 재야 한다. 0 이 아니면 그 owner 들의 사용량이 배포 순간 한꺼번에 떠나므로(소급) 사람이 그 목록을 보고
> 승인한다. 운영 원장 재계산·쓰기는 이 change 가 하지 않는다.

## Q3 — latch 와의 순서: **latch 해소 → owner 해제(=떠남) → 다음 체결의 재계산**

**latch 작성자 전수 (AST).** owner 플래그와 예약 플래그는 언제나 같은 트랜잭션에서 함께 쓰인다(R2 Q1 확인):
- `persistRiskBucketFillTransition` B23(예약 플래그, :1071) · B29(owner 플래그, :1087) — 값은 `ApplyFill` 의 결과.
- `latchRiskBucketFillFailure` B2(owner UNKNOWN=1) · B3(예약 UNKNOWN=1), 그리고 B1 → `latchRiskBucketScope`(B1: REPLAY_MISMATCH·
  ORPHAN_FILL 만 허용).
- `ReleaseRiskOverageLatch` B11(owner RISK_OVERAGE=0) · B12(예약 RISK_OVERAGE=0) — 운영자 완화, 활성 owner 만(B5).
- `clearResolvedUnknownLatches`(B1–B3 모든 체결이 actual 확정일 때만 B4 로 UNKNOWN 을 지움) — `ApplyFill` B23 경로(actual
  보완)에서만 불리고 결과는 위 persist B23·B29 로 쓰인다. **즉 UNKNOWN_ACTUAL_RISK 는 a066 에서 이미 원장 사실(actual 증거
  완성)에 결속되어 자동으로 풀린다.**

**재계산 AST 사실.**
- `recomputeOverageLatches` 는 `ApplyFill` 의 두 자리에서만 불린다: B23(actual 보완) · B38(새 체결).
- 그 함수는 RISK_OVERAGE 를 **세우기만** 한다: B9 overage>0 일 때만 `anyOverage=true`, B11 기존 overage 보다 클 때만 올림,
  B12 `!anyOverage` 면 아무것도 지우지 않고 반환.
- 공유 bucket 의 타인 사용량은 B6 `SharedUsedMinor` 로 들어오며, 그 값은 `riskBucketSharedUsage`(B1 순회, B4 합<자기몫 →
  ReplayMismatch) 가 `ReadJournalBucketUsage` 로 센다. **reader 를 바꾸면 재계산은 자동으로 떠난 뒤의 합을 본다.** 활성 owner
  의 행은 D1 조건 2 로 떠날 수 없고(`uq_risk_bucket_active_owner`, 읽음 `risk_bucket.go:702`) own 은 total 의 부분집합이므로
  건강한 원장에서 B4 는 서지 않는다(R2 Q2).
- 재계산의 한도는 owner 자기 snapshot 의 최소 한도다(`loadRiskBucketFillTransition`, 읽음 `risk_bucket_fill.go:800–812`) —
  admission 의 한도 모집단(D3)과 다른 셋째 모집단이며 a066 부터 그렇다. 이 change 는 그것을 건드리지 않는다.

**정의된 순서.**
1. owner 에 latch 가 있으면 먼저 해소된다: RISK_OVERAGE 는 운영자 해제 `ReleaseRiskOverageLatch`(OPERATOR·승인·audit, D8),
   UNKNOWN_ACTUAL_RISK 는 actual 증거 완성 시 `clearResolvedUnknownLatches` 자동 해제. scope latch(REPLAY_MISMATCH·ORPHAN_FILL)
   는 지우는 경로가 없으므로 그 owner 는 **해제될 수 없다**(검사표 `scope_latch`).
2. owner 해제 = 영수증 커밋 = 사용량 떠남(같은 트랜잭션, 파생이므로 추가 쓰기 없음). latch 가 남아 있으면 거절(B26–B28
   `owner_latch`·`scope_latch`).
3. 이후 **다른** owner 의 체결에서 `recomputeOverageLatches` 가 줄어든 공유 합으로 overage 를 판정한다.

**불변 조건.**
- 떠남은 어떤 latch 도 지우지 않는다. 이미 다른 owner 에 선 RISK_OVERAGE 는 그대로다. 줄어든 합은 **다음** 체결의 판정과
  신규 진입 여유에만 영향한다.
- reader 의 latch 집계(`Latched`·`OverageLatched`·`UnknownLatched`)는 떠난 행도 계속 센다 — 합에서만 뺀다. 해제 뒤 행 플래그를
  세우는 정상 경로는 없다(fill 은 B4 로 해제 owner 를 못 찾고, late fill 은 scope latch 만 쓰고, 완화는 B5 로 거절). 그러므로
  떠난 행의 플래그 1 은 손상일 때만 있고, 그때 진입 차단이 유지되며 **복구 경로는 없다**(잔여 R6 — 오늘과 같다).

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
2. 같은 owner 키의 `risk_bucket_owners.released_at` 이 영수증 `released_at` 과 같다. 두 값은 단일 작성자가 한 트랜잭션에서
   쓰므로(`risk_bucket_owner.go:1017`·`:1029`) 건강한 원장에서는 항상 참이다 — 이 조건의 역할은 **부패 검출**이다:
   영수증이 있는데 owner 의 `released_at` 이 NULL 이거나 다르면 `ErrJournalUsageInvalid`(fail-closed)로 거절한다
   ("안 떠남" 으로 조용히 넘기지 않는다).
3. 같은 owner 키에 `risk_bucket_scope_latches` 행이 **하나도 없다**(D4).
4. owner 키는 **예약 행의 열**(`r.account_ref, r.market, r.symbol, r.owner_prospective_generation`)로 영수증과 잇고, 같은 행이
   `risk_bucket_final_decisions` 의 사본(`d.*`, `r.decision_id` 로 조인)과 같아야 한다. 다르면 `ErrJournalUsageInvalid`.
   공유 사용량의 `own`(`loadRiskBucketFillTransition`)과 late-fill latch 는 `d.*` 로 owner 를 고르므로, 두 사본이 갈라진 원장은
   떠남 판정과 소유 판정이 서로 다른 owner 를 보게 된다 — 그 상태를 판정 전에 거절한다(R1 #4 · R2 #5).

이 판정은 `readProductionRiskUsage` 의 SQL 이 행마다 한 열(`departed`)과 불일치 표식으로 계산하고
`aggregateProductionRiskUsage` 만 소비한다. 다른 곳에 같은 규칙을 다시 쓰지 않는다 — 판정이 둘이면 서로의 시험을 통과시켜
둘 다 살아남는다(코드와 DB 트리거 쌍의 선례).

### D2. 떠난 행은 합에서만 빠지고 검증에는 남는다

`aggregateProductionRiskUsage` 에서:
- 떠난 행도 지금의 행 검증(B2: 금액·상태·RELEASED 의 held=0·snapshot·policy·식별자)을 그대로 받는다.
- 추가 fail-closed: **영수증이 있는 owner 의 행은 떠남 여부와 무관하게**(scope latch 로 되돌려진 행 포함) `held_minor≠0`
  이거나 `state='HELD'` 면 `ErrJournalUsageInvalid`(해제는 `bucket_held`=0 을 요구했으므로 이것은 손상이다). D1 조건 2·4 의
  불일치도 같은 오류다. 검증은 D1 조건 3(scope latch) 판정 **앞**에 서므로 되돌림이 손상 검사를 우회하지 못한다(codex #4).
- 떠난 행의 `filled_minor` 는 합에 더하지 않는다. latch 플래그는 집계에 그대로 든다(Q3).
- **파급 범위(의도).** 이 거절은 그 bucket 을 쓰는 모든 소비자에게 간다: `riskBucketSharedUsage` B3 이 `ErrJournalUsageInvalid`
  를 ReplayMismatch 로 바꾸고, 그 공유 bucket 에서 체결 중인 **무관한 활성 owner** 가 `applyRiskBucketFillInTx` B12 로 scope
  latch 를 받는다(체결은 커밋됨). 기존의 잘못된 행 처리와 같은 방향(보수)이며 계좌 단위로 넓다(R2 #6).
- `RowDigest` 형식은 바꾸지 않는다(1판의 "parts 에 departed 추가" 철회 — R2 #7 · R3 #5). snapshot digest 는 이미 파생된
  `filled`·`held` 를 담으므로(`production_snapshot_authority.go:399–400`, 읽음) 떠남은 digest 에 반영된다. 형식 변경은
  근거 없는 digest 도메인 변경이었다.

### D3. 한도 모집단은 **바꾸지 않는다** — 떠난 행은 한도 비교에 남는다 (Manager 판정 2026-09-30)

1판은 떠난 행을 한도 모집단에서도 빼려 했다. 리뷰가 그것이 **완화**임을 보였다(R1 #3 · R2 #1):
- a066 #3(a) 는 공유 bucket 에 서로 다른 매니페스트가 다른 한도를 선언할 때(horizon 은 KR·US 매니페스트가 따로) 기록된
  가장 작은 한도로 cap 한다. 떠난 행이 모집단을 떠나면, 작은 한도를 선언한 행이 전부 떠난 bucket 에서 큰 한도의 진입이
  자기 snapshot 한도로만 cap 되고(`refuseStaleBucketUsage` B10 `!found` → continue), 모든 행이 떠나면 기록 한도 검사 자체가
  사라진다. 이것은 #3(b)(상류 단일 한도 검증) 해소 전의 방어 약화이고 §0.9 에 어긋난다.
- **결정:** `smallestRecordedBucketLimit` 와 그 모집단(`state IN (HELD,FILLED) OR filled≠0`)은 그대로 둔다. 떠나는 것은 사용량
  합뿐이다. fail-closed 이고 #3(a) 방어를 보존한다.
- **비용:** a066 6.5 좁힌 재리뷰의 활성성 잔여("이력 있는 bucket 에서 한도 상향이 실효가 없다")는 닫히지 않는다 — 잔여 **R5**.
  #3(b) 가 해소되면 그때 한도 모집단의 떠남을 따로 연다. 사람 항목은 불요(fail-closed 쪽 선택).
- a066 주석 "활성은 원장이 세는 행과 같은 모집단"(읽음 `risk_bucket_usage.go:59–62`)은 이 change 뒤 사실이 아니게 된다 —
  구현 로트가 그 주석을 "한도 모집단은 떠난 행을 포함한다(a126 D3)" 로 고친다(코드 동작 불변).

### D4. 해제 뒤 ORPHAN_FILL scope latch 는 떠남을 되돌린다 — 되돌림이 서는 경로는 하나뿐이다

**서는 경로.** 해제된 owner 의 등록 주문에 late BUY 체결이 오고, 체결 증분을 읽을 수 있고, Campaign hook 이 결선된 경우:
`apply_hook.go:280–287`(읽음 — `hooks.Campaign != nil` 일 때만 `applyRiskBucketOwnerBindingInTx` 호출) →
`applyRiskBucketOwnerBindingInTx` B1(BUY) · B2/B3 통과(증분 판독) · B4 통과(증분>0) · B9 `len(keys)==0`(활성 owner 없음) →
`latchReleasedOwnerLateFillInTx` B5 통과(해제 owner 발견) · B7 ORPHAN_FILL scope latch. 그 late 체결 금액은 bucket 에
들어가지 않는다 — `queryRiskBucketOrder` 두 호출의 WHERE 가 `ow.released_at IS NULL`(읽음 `risk_bucket_fill.go:574`·`:577`)이라
`applyRiskBucketFillInTx` 는 B4 `!found` 로 반환한다.

**서지 않는 경로(잔여 R3 에 명명, R1 #2 · R3 #1 · R2 #4).**
- (a) 소유가 모호한 체결: `fills.go:540–543`(읽음)의 `latchRiskBucketFillFailureForScope` 는 SQL 이 `ow.released_at IS NULL`
  (읽음 `risk_bucket_fill.go:1159`) 이라 해제 owner 를 못 찾는다(B5 순회 대상 0).
- (b) 증분 판독 불가: binding B2/B3 → 같은 `latchRiskBucketFillFailureForScope` → 같은 이유로 해제 owner 에 아무것도 안 씀.
- (c) terminal 경로의 ReplayMismatch: `releaseTerminalRiskBucketOrderInTx` 도 `riskBucketOrderForFill`(released 필터)을 쓴다.
- (d) 미등록 order id: `latchReleasedOwnerLateFillInTx` 의 조회가 비어 B5 로 반환.
- (e) Campaign hook 미결선: binding 자체가 안 불린다(R1 #8). 결선 여부는 배포 구성 의존이다.
- (f) broker order id 의 세대 간 재사용: 주문 등록의 충돌 검사는 **같은 owner generation 안**만 본다(읽음
  `risk_bucket_fill.go:107–113`, `strategy_dispatch_runtime.go:1144–1149`; 스키마 UNIQUE 는 `(decision_id, order_id)`,
  `risk_bucket.go:770`). 뒤 generation 이 같은 order id 를 쓰면 옛 generation 의 late 체결을 `riskBucketOrderForFill`(활성 owner)
  이 가로채고 binding 은 B9 에 닿지 않는다 — 옛 owner 는 떠난 채로 남는다(codex #1). broker order id 의 세대 간 유일성은
  브로커 의미론 의존이며 이 저장소에서 증명되지 않았다.

**되돌림이 서도 남는 구멍 — 발급된 주문의 제출(codex #2).** 예: 한도 100, A 의 60 이 떠남 → 다른 종목 B 가 80 을 예약·발급
→ A 의 late 체결로 60 이 되돌아와 합 140. A 는 scope latch 만 받고, B 에는 bucket latch 도 재계산도 없다. 제출 재검증
`RevalidateQFinalAdmission` 은 B27 latch 만 보고 합을 보지 않으므로(D6) B 는 제출된다. 되돌아온 금액도 옛 filled 이지 late
체결 금액이 아니다. 오늘은 사용량이 줄지 않으므로 이 순서가 생기지 않는다 — **떠남이 새로 여는 순서**다.

**1판 주장의 정정.** 1판은 "결과는 오늘과 정확히 같은 합" 이라 했다. 거짓이다(R1 #1):
- 위 (a)–(e) 에서는 되돌림이 없고, 오늘은 해제 owner 의 옛 filled 가 합에 남아 **우연히** 그 공유 bucket 여유를 덮는다.
- 되돌림이 서는 경로에서도, 영수증 커밋부터 late 체결까지의 **창** 동안 공유 bucket 여유가 풀리고 그 창에 들어온 진입이
  소비한 여유는 되돌림이 취소하지 못한다. 해제는 ACTIVE 주문 부재만 요구하므로(검사표 `pending_entry`, 읽음
  `risk_bucket_owner.go:929`) late fill 은 실재 경로다.
- 오늘의 덮개도 우연이었다: 옛 filled 금액이지 late 체결 금액이 아니다(R2 #4).
- 이 창과 (a)–(e) 는 a066 잔여 #8(late 체결이 공유 bucket 을 latch 하지 않음)과 한 가족이다. 생산 효과가 0 인 동안(R1)
  실재하지 않으며, **owner 해제 배선 로트의 착수 조건**으로 막는다(잔여 R3, tasks.md 3.1).
- 1판이 검토한 대안 "해제 뒤 미해소·미귀속 체결이 있으면 떠나지 않음" 류로 창을 닫는 것은 해제 경로 편집이 필요해 D5 의
  편집 금지와 충돌하므로 이 change 에서는 기각한다(Manager 판정).

**owner 플래그는 되돌림 트리거가 아니다(R1 #5).** 해제 뒤 owner 플래그를 세우는 정상 경로는 없다(Q3 불변 조건). 떠난 행의
latch 플래그는 되돌림이 아니라 latch 집계로 진입을 막는다.

### D5. 저장값은 바뀌지 않으므로 체결·해제·완화 경로 코드는 편집하지 않는다

편집 대상은 reader 두 함수(`readProductionRiskUsage`·`aggregateProductionRiskUsage`)와 `JournalBucketUsage` 타입으로 예상한다.
`smallestRecordedBucketLimit`·`refuseStaleBucketUsage` 는 동작 불변(D3, 주석 한 줄만). `persistRiskBucketFillTransition`·
`releaseRiskBucketOwner`·`ReleaseRiskOverageLatch`·`recomputeOverageLatches`·`latchReleasedOwnerLateFillInTx` 는 편집하지
않는다(편집이 필요해지면 그 함수의 gate FLM 번들을 먼저 만든다 — "Logic Map 은 계획보다 항상 많다").

### D6. 소비자 영향

`ReadJournalBucketUsage` 의 생산 소비자 4 곳이 모두 떠난 뒤의 합을 받는다:
- `refuseStaleBucketUsage`: B7 snapshot 이 원장보다 **적게** 주장하면 stale 거절. 떠남 뒤 원장이 줄면 옛 snapshot 은
  **많게** 주장한다 → 받음(보수 — 그 snapshot 으로 계산한 q_final 은 더 작다). B8–B12 한도 비교는 불변(D3).
- `riskBucketSharedUsage` → Q3.
- `RevalidateQFinalAdmission`: B25 순회에서 B26(판독 오류 → 거절) · B27 `latchedUsageRefusal` 만 사용량을 쓴다 — 합을 읽지
  않으므로 코드 동작은 불변이고 B26 은 D1·D2 의 새 `ErrJournalUsageInvalid` 도 거절로 받는다(보수). **그러나 "영향 없음" 은
  아니다:** 발급 뒤 파생 사용량이 **늘 수 있게** 되었고(D4 되돌림) 이 재검증은 그 증가를 보지 않는다 — D4 「되돌림이 서도
  남는 구멍」, 잔여 R3(codex #2 가 1판의 "동작 불변" 결론을 반박).
- `loadProductionRiskEntries`: B10 latch 거절 불변, 합은 D2 대로. B6 은 계좌·시장·종목의 **모든 generation** scope latch 를
  거절하므로 D4 의 ORPHAN_FILL 종목은 여기서도 막힌다. 이 reader 는 스키마 27 에 고정돼 있어(a066 잔여 #7) 지금 v32/v35
  원장에서는 어차피 거절한다.

### D7. §0 안전 불변식 대조

- §0.3 손절·비상 청산: 무접촉. reader 는 EXPOSURE_RAISING 경로에서만 불린다(a066 D7). 체결·해제 경로는 편집하지 않는다.
- §0.9 사이징은 보수 방향만: 감소는 진입을 여는 방향이다. 그러므로 감소는 오직 영수증(=a066 D6 의 가장 강한 청결 증명)
  에만 걸리고, 모르는 것(scope latch, 부패·불일치 행)은 전부 "안 떠남" 또는 거절로 떨어진다. 한도 모집단은 건드리지 않는다(D3).
- §0.6 스키마: 변경 없음.
- D8 "자동은 조이기만 한다" 와의 관계: D8 의 대상은 잠금·latch 의 **운영자 해제**다. 원장 사실에 결속된 자동 감소는 a066
  자신이 이미 둘 한다 — cancel/expiry 의 HELD 해제(D5)와 actual 증거 완성 시 UNKNOWN_ACTUAL_RISK 자동 해제
  (`clearResolvedUnknownLatches`, Q3). 떠남은 앞의 것의 filled 짝이다.
- **H2 처분 (Manager 판정 2026-09-30 — 새 사람 질의 불요).** 사용자 결정 ③(2026-09-28)이 이 축 분리를 이미 담고 있다:
  사용량 수명주기를 완화 패키지가 아니라 독립 change 로 뗀 이유가 "자동 회계 의미론이지 사람 승인 해제가 아니다" 였다
  (STORY-TOS-a126 found_by · proposal.md 「What Changes」 명문). 따라서 D8 「자동은 조이기만」의 대상은 운영자 완화 축
  (잠금·latch 해제)이고, 원장 사실에 결속된 회계 감소는 cancel/expiry HELD 해제와 같은 부류다. 사용자 최종 보고에 이
  해석을 명시해 거부권을 남긴다 — 뒤집히면 이 절과 두 요구를 다시 연다.

## 반증 설계 (구현 로트 1.x 가 세울 것)

"감소를 무조건 수행" 축의 변이는 **반드시** 잡혀야 한다(fail-open 축). 무변이 대조군이 GREEN 인지 먼저 보고, 양성 대조군
(떠나야 하는 행이 떠나는지)도 함께 둔다.

| 변이 | 뚫는 것 | 잡는 시험 |
|---|---|---|
| M1 `departed` 를 항상 참 | 활성 owner 의 filled 가 사라짐 | 활성 owner 사용량이 그대로인 시험 |
| M2 영수증 조건 제거(owner released_at 만 봄) | 영수증 없는 해제 표식 | released_at 만 있고 영수증 없는 행은 안 떠남 |
| M3 scope latch 되돌림 제거 | D4 fail-open | 해제 뒤 late BUY(서는 경로) → 사용량 복원 — RecordFill 경로 **와** 전략 정산 경로 각각 |
| M3b 서지 않는 경로를 "선다" 로 바꾸는 시험 부재 | 잔여 은폐 | (a) 소유 모호 · (b) 증분 판독 불가 경로에서 해제 owner 에 scope latch 가 **없음**을 단언하는 잔여 핀 — 배선 로트가 바꾸면 깨지게 |
| M4 떠난 행을 한도 모집단에서 뺌 | D3 완화 | 작은 한도를 선언한 행이 떠난 뒤에도 큰 한도를 선언한 진입이 작은 한도로 거절됨 |
| M6 떠난 행의 held≠0 · HELD 를 받음 | 손상 은폐 | 손상 원장 → `ErrJournalUsageInvalid` |
| M6b D1 조건 2·4 불일치를 "안 떠남" 으로 삼킴 | 부패 은폐 | 영수증 有·released_at NULL, `r.*≠d.*` → `ErrJournalUsageInvalid` |
| M7 떠난 행의 latch 플래그를 집계에서 뺌 | latch 해제 우회 | 떠난 행 플래그 1 → 진입 거절 유지 |
| M8 부분 매도로 감소 | Q4 | 부분 매도 뒤 사용량 불변 |

(1판의 M5 "한도 모집단만 떠나고 합은 남음" 은 D3 결정으로 의미가 없어져 삭제. M6 에는 무관한 활성 owner 의 체결이 커밋되고
scope latch 만 받는 하위 사례를 둔다 — D2 파급 범위.)

실값 픽스처: a066 owner-lifecycle 픽스처(`closeRiskBucketOwnerLifecycle`, 주석 `risk_bucket_owner_test.go:802–804`)는
`filled_minor='0'` 으로 갭을 가렸다 — 1.4 에서 실제 체결 금액을 쌓는 픽스처로 바꾸고 주석을 지운다.

## 롤백

코드만 되돌린다. 스키마·저장값 변경이 없으므로 원장 되돌림이 없다. 되돌리면 사용량은 누적 합(오늘 동작)으로 돌아가며
이것은 더 보수적인 방향이다.

## 잔여 (이 change 가 닫지 않는 것)

- **R1** 생산 효과 0 — owner 해제 배선(a066 잔여 #5, 활성화 로트) 전까지.
- **R2** 부분 종결 감소 — SELL→owner 귀속 행과 금액 기준 정책이 필요(Q4).
- **R3** late 체결 가족: a066 #8 + D4 의 영수증~late 체결 창 + 되돌림이 서지 않는 경로 (a)–(f) + 되돌림 뒤 발급 주문 제출
  (codex #2). **owner 해제 배선 로트의 착수 조건 = R3 해소 또는 사람 수용**(Manager 판정, tasks.md 3.1). codex #3 은 이
  조건이 사람 수용으로 **면제 가능**해서는 안 된다고 본다 — 면제 가능성 여부는 Manager 재판정 대기(review.md 「0.5.2」).
  생산 효과 0 인 동안 실재하지 않는다.
- **R4** 이 change 의 창에 형제 커밋이 이미 들어왔다: `55963f29`·`f48e7865`(a092, `internal/obs`). journal·riskbucket 파일은
  건드리지 않았다(R3 리뷰 확인). 첫 Go 편집 전(1.0.1)에 사람 승인 base 재고정 절차로 다시 옮긴다.
- **R5** 한도 상향의 활성성 — 이력 있는 bucket 에서 가장 작은 기록 한도가 계속 묶는다(D3). a066 #3(b) 해소 뒤 따로 연다.
- **R6** 떠난 행의 손상 latch 플래그(owner 0·행 1)는 복구 경로가 없다(완화 B5 가 해제 owner 를 거절). 정상 작성자로는 생기지
  않고 오늘과 같다. M7 시험은 이것을 **기능이 아니라 보수적 막힘**으로 기록한다.

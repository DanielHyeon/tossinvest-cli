# a127 design (2판 — freeze 리뷰 1라운드 반영, review.md 「0.5」)

> 분기 · early return 을 근거로 쓰는 문장은 `analysis/freeze-ast/` 의 AST 산출물(편집 전, base `f9a25549`, `tools/logic-map`)에서 읽었다 —
> `loadProductionRiskEntries` 14 분기 · `openProductionRouteSnapshot` 4 분기 · `LoadProductionRiskSnapshotAuthority` 7 ·
> `LoadProductionRouteAuthorityBatch` 16 · `LoadProductionRouteAuthority` 2. 두 적재기 파일의 source sha256 은 `65fe66e7` 에서도 같다.

## 증거 기반

- **핀 두 자리**(AST): `loadProductionRiskEntries` B5 `:375` — `PRAGMA user_version` 읽기 실패 **또는** `version != productionRiskJournalSchema`
  이면 `:376` 에서 `errors.New("risk bucket: exact journal schema unavailable")` 반환. `openProductionRouteSnapshot` B4 `:609` — 같은 조건이면
  tx rollback · db close 뒤 `:612` `ErrProductionRouteUnavailable`(`:614` 는 성공 반환). 둘 다 원장 판독 **앞**에 선다.
- **핀은 출하된 어느 스키마와도 맞은 적이 없다.** 두 리터럴은 같은 커밋 `8022f578`(2026-08-04)에서 태어났고, 그 커밋의 `journal.SchemaVersion` 은
  이미 29 였다(`git show 8022f578:internal/journal/schema.go`). 27 은 그 직전 판본 사이의 값이다 — 가린 픽스처(전부 27)가 처음부터 결함을 가렸다.
- **생산 진입점**: supervisor 가 `journalPath = c.Journal.Path()`(`strategy_entry_supervisor.go:300-301`)를 route 적재기(`:305`)와 risk 적재기
  (`:328`)에 같이 넘긴다. `c.Journal` 은 `journal.Open` 이 연 핸들이고 Open 은 마이그레이션 뒤 반환한다(`journal.go:196`); migrate 는 더 새
  원장을 `ErrSchemaTooNew` 로 거절하고 같으면 곧바로 반환한다. → **엔진 안에서 두 적재기가 여는 파일의 `user_version` 은 `journal.SchemaVersion`**.
  예외 하나: 엔진 밖 명령 `cmd/tossctl/flatten.go:226` 은 engine lock 없이 `journal.Open`(마이그레이션 가능)을 부른다 — 더 새 바이너리의
  flatten 이 엔진 실행 중 원장을 올리는 창(D7 의 근거).
- **읽기 집합은 현재 원장에서 성립한다**(`analysis/measurements/readset-probe.log` 2판, 하네스 `readset-probe_test.go.txt`): 새로 연 v35 원장에서
  route 적재기의 두 SQL(원문 그대로) · risk scope latch 질의가 prepare · 실행된다. risk 의 큰 사용량 질의(`readProductionRiskUsage`,
  LEFT JOIN · EXISTS — `:473-488`)는 이 측정에 없다: 그 성립은 journal admission(`refuseStaleBucketUsage`) · 제출 재검증
  (`RevalidateQFinalAdmission`)이 **현재 원장에서** 매번 같은 함수를 부른다는 사실이 근거다(a126 시험 다수가 v35 원장에서 직접 부름).
- **없는 열은 그 열을 참조하는 질의가 prepare 될 때만 실패한다**(같은 측정의 대조 — `position_campaigns.entry_blocked` 를 지운 원장: owners 질의
  성공, campaign 질의 `no such column: entry_blocked`). 모든 적재기 SQL 은 상관 부질의 안에서도 한정 이름을 쓰고 큰따옴표 식별자가 없다 — 이름
  해석이 문자열 리터럴로 떨어지는 경로가 없다. **그러나 route 의 campaign 질의는 조건부다**: `loadProductionRouteOwnersFrom` 은 active owner 가
  없으면 `:652-653` 에서 성공으로 반환해 `:665` 의 campaign 질의를 prepare 하지 않는다(codex freeze P1) → D7.
- **v28~v35 의 읽기 집합 변경** — 두 갈래로 센다:
  - DDL(`internal/journal/*_v2[8-9].sql` · `*_v3[0-5].sql` 전수): 읽기 집합 표에 대한 `ALTER` · `CREATE TRIGGER` 는 v34 하나 —
    `risk_bucket_reservations.policy_record_digest`(nullable)와 삽입 · 불변 트리거. 적재기는 그 열을 읽지 않는다.
  - **값의 의미**(DDL 스캔이 구조상 못 보는 것 — Go 작성자): v35 이후 RISK_OVERAGE 는 운영자 해제로 지워질 수 있다(`journal/risk_bucket_relaxation.go:352` ·
    `:356` — owners · reservations 의 `risk_overage_latched=0`). 두 적재기가 바로 그 열을 읽는다(route `production.go:621` → 판정 `:656`,
    risk `production_snapshot_authority.go:474` → 판정 `:404`). 운영자 승인 해제이고 admission 도 같은 열을 읽으므로 fail-open 은 아니다. route
    owner 이력 digest(`:697-705`)는 latch 열을 담으므로 해제 뒤 digest 가 바뀌지만 revision 은 행 수 기반(`:642`)이라 매니페스트 `OwnerRevision`
    대조는 유지된다. v34 는 `risk_bucket_policies` 의 의미를 바꿨다 — v34 이후 행에서 그 표는 예약 자신의 가격이 아니라 키의 **첫** 레코드다
    (v34 SQL 머리 주석). risk 적재기의 `COALESCE(p.record_digest,'')` 는 RowDigest 에만 들어가고 합에는 무관(admission 과 같은 함수).

## 설계 결정

### D1. 결속 방식 — **주입된 현재 버전과 정확 일치**(Manager 재판정 2026-10-01)

결함의 본질은 「정확 일치」가 아니라 **마이그레이션과 함께 움직이지 않는 동결 리터럴**(27)이다.

**규칙**: `user_version == 주입 값(journal.SchemaVersion)` 일 때만 판독한다. 불일치는 방향을 문구로 가른다 — 더 새 원장(이 빌드보다 새 스키마)과
더 옛 원장(마이그레이션되지 않은 원장). 표 · 열 존재는 별도 목록으로 검사하지 않고 SQLite prepare 에 맡긴다 — 조건부 질의는 D7 로 무조건화.

**판정 교체 사슬.** 배정(2026-10-01) 때 방향은 (a) 「정본 선례 `journal.ReadOnly.checkSchema` 패턴 — `user_version > 상한` 이면 TooNew
거절 + 필요한 표 · 열 존재 확인」이었다. 착수 실측 뒤 (b) 를 권고로 올렸고 Manager 가 (b) 로 재판정해 (a) 지시를 대체했다. 근거 넷:

1. **엔진 안의 불변식이 `==` 다.** 두 적재기가 여는 파일은 같은 프로세스가 연 · 마이그레이션한 원장이다(위 증거). (a) 가 추가로 받는
   「더 옛 원장」은 엔진 안에서 생기지 않고, 받아 주면 열은 있으나 트리거 · 제약이 그 버전 이전인 원장(예: v34 정책 레코드 트리거 이전)을
   읽는 **허용 방향 구멍**만 남는다.
2. **선례의 상황이 다르다 — 교차 바이너리 vs 동일 프로세스.** `journal.ReadOnly.checkSchema`(`readonly.go:229-`)는 **콘솔** 등 엔진 밖
   열람용이다 — 다른 바이너리가 다른 판본의 원장을 열 수 있는 자리라 범위 수락이 맞다. `strategyevidence/readonly.go:43` 은 **같은 빌드의 자기
   저장소** — `version != SchemaVersion` 거절, 더 새면 `ErrSchemaTooNew`. a127 의 적재기는 후자의 상황이다.
3. **필요 표 · 열 목록은 읽기 집합의 둘째 철자다.** SQL 원문과 목록이 어긋나도 서로의 시험을 통과시킨다(「판정이 둘이면」). 표 · 열 존재는
   SQLite prepare 가 강제한다(증거).
4. **결함은 동결 리터럴이다.** 상한을 `journal.SchemaVersion` 상수로 주입하면 마이그레이션마다 값이 함께 움직인다.

### D2. 현재 버전은 호출자가 주입한다 — 0 이하면 열기 전에 거절

riskbucket · strategyrouter 는 journal 이 import 한다(`journal/risk_bucket*.go` → riskbucket, `journal/strategy_first_leg_atomic.go` → strategyrouter).
두 패키지가 `journal.SchemaVersion` 을 읽으면 순환이다. 그래서:

- `ProductionRiskSnapshotConfig` · `ProductionRouteConfig` 에 원장 스키마 필드(이름은 구현 로트) 하나씩.
- engine 의 두 호출 자리(`strategy_risk_authority.go:211` · `strategy_route_authority.go:179` 의 config 리터럴)가 `journal.SchemaVersion`
  **상수 선택자**를 넣는다. 런타임에 파일에서 읽은 값은 같은 파일을 자기와 비교하는 공허한 검사이고, 리터럴(예: 35)은 다음 마이그레이션에서
  같은 결함을 재현한다 — 둘 다 금지(구조 단언, S7 · S10).
- 필드가 0 이하면 두 적재기는 **원장 파일을 열기 전에** 거절하고, 문구가 주입 누락을 말한다(관측 방법: 존재하지 않는 원장 경로로 불러도 오류가
  열기 실패가 아니라 주입 누락이어야 함 — S6).

### D3. 거절 신원은 그대로, 방향 문구는 공개 경계까지 운반

- 스키마 불일치는 결함이다(`ErrProductionRiskScopeRefused` 주석 `production_snapshot_authority.go:38-42`). 범위 국소 거절로 재표식하지 않는다 —
  엔진은 그 신원에서만 범위를 건너뛴다(`strategy_owner_scope_authority.go:48`, S11).
- risk: `LoadProductionRiskSnapshotAuthority` 가 이미 `%w: %w`(`:174`)로 원인을 보존하고 엔진이 `entry.cause = err`(`strategy_risk_authority.go:219`)로
  싣는다 — 방향 문구가 그대로 관측된다.
- **route: 지금은 지워진다.** `openProductionRouteSnapshot` 은 맨 sentinel 을 돌려주고(`:612`), `LoadProductionRouteAuthorityBatch` 는 어떤 내부
  오류든 `fmt.Errorf("%w: owner snapshot", ErrProductionRouteUnavailable)`(`:350-352`)로 바꾼다. a127 은 opener 가 방향 문구를 담은 오류(Unavailable
  을 `%w`)를 돌려주고, Batch 의 `:352` 감싸기가 원인을 보존하게 바꾼다 — `errors.Is(…, ErrProductionRouteUnavailable)` 는 유지. 시험은 문구를
  **Batch(공개) 경계**에서 단언한다. 엔진 route 적재기는 오류를 버리고 `StrategyRouteAuthorityInvalid` 로 접는다(`strategy_route_authority.go:179-183`) —
  관측 표면 확장은 a127 범위 밖(잔여).

### D4. 수락은 실제 원장으로 — 축소 픽스처는 단독 수락 근거가 아니다

- **risk 양성**: 엔진 패키지(서명 위험 정책 픽스처가 있는 곳)에서 a112 트립와이어 `TestTheRiskStubBridgeIsStillNeededBecauseTheLoaderRefusesTheRealJournal`
  (`a112_owner_scope_trading_test.go:423`, 단언 `:436`)을 `journal.Open` 원장 경로의 양성 시험으로 뒤집는다 — 트립와이어 자신이 「핀이 고쳐지면 실패한다」로
  설계돼 있다.
- **route 양성 — 새 기반이 필요하다.** 서명 route 매니페스트 픽스처는 `strategyrouter/production_test.go`(내부 시험)에만 있고, strategyrouter 내부
  시험은 journal 을 import 할 수 없다(순환). 엔진 시험은 `loader.load` 를 stub 으로 바꾼다(`strategy_route_authority_test.go:26/55/73`), a112 는
  `readyRouteAuthority(...)` 를 쓴다. → `tossos_testseams` 빌드에 서명 매니페스트 작성 seam(기존 내부 픽스처의 이동)을 두고, **외부 시험 패키지**
  `strategyrouter_test` 가 `journal.Open` 원장으로 `LoadProductionRouteAuthorityBatch` 를 부른다.
- **음성**(두 적재기): 주입 0 · 미설정 거절(열기 전), 더 새 · 더 옛 `user_version` 거절 + 방향 문구, 읽기 집합 열을 지운 원장 거절 — route 는
  **active owner 가 없는 범위**에서도(D7). 열 삭제 픽스처 주의: SQLite `DROP COLUMN` 은 인덱스 · 트리거가 쓰는 열을 거절한다(예: `released_at` 은
  부분 인덱스 `uq_risk_bucket_active_owner`, `entry_blocked` 는 트리거 `strategy_first_leg_binding_insert_guard`) — 측정 하네스처럼 표를 다시
  만들어야 한다.
- **구조**: engine 두 호출 자리가 `journal.SchemaVersion` 선택자를 넘긴다는 AST 단언(리터럴 · 런타임 읽기 금지).
- **축소 픽스처 셋**(`riskbucket/production_snapshot_authority_test.go:219`, `strategyrouter/production_test.go:216`, `app/engine/strategy_risk_authority_test.go:224`)은
  주입 값과 같은 `user_version` 을 쓰게 바꾼다. riskbucket · strategyrouter 시험은 journal 을 import 할 수 없으므로 **journal 버전처럼 보이지 않는
  값**(예: 1)을 주입 · 기록한다 — 새 동결 리터럴 35 를 시험에 심지 않는다. 행 · 단언은 바꾸지 않는다.
- a112 시험 다리 `a112MirrorLedgerIntoRiskStub` 를 지우고 a112 거래 픽스처의 위험 적재기를 실제 원장으로 단일화(a112 소유 파일 — Manager 통지).

### D5. 바꾸지 않는 것

판독 SQL 의 문언, 사용량 판정(`aggregateProductionRiskUsage`), owner 재구성 판정(`loadProductionRouteOwnersFrom` 의 분기), 원장 스키마 ·
마이그레이션, 엔진의 진입 판정 순서.

### D6. §0 안전 불변식 대조 — 무엇이 오늘 진입을 0 으로 막는가

- LIVE 주문 side effect: 없음 — 적재기는 읽기 전용(`mode=ro` · `query_only`). 시험은 원장 픽스처만.
- **오늘 진입이 0 인 실제 이유**(1판의 「4-가족 관문」은 틀렸다 — 그 관문은 서명 4-가족 활성화가 있는 시장에만 선다): 시장 승격은
  `strategy_entry_supervisor.go:451-452` 에서 schedule 활성화(`s.restore.Activation == nil` → dormant)와 schedule · candidate · route · FX · risk ·
  account 권한 각각의 Ready 를 모두 요구한다. 오늘은 서명 route · risk · account 매니페스트와 schedule 활성화가 없어 dormant 이고, **설령 매니페스트가
  있어도 route · risk 는 핀 때문에 Ready 가 될 수 없었다**.
- **a127 뒤 새로 도달 가능해지는 경로**: schedule 활성화 + 서명 route · risk · account 매니페스트가 있고 **4-가족 활성화가 없는** 시장은
  단일 범위 handoff(시장당 1, `strategy_entry_supervisor.go:455-457` · `strategy_account_first_leg_authority.go:159-162`)로 1차 레그까지 갈 수 있다.
  지금까지 이 경로는 핀 때문에 도달 불가였다. 이것은 설계된 동작이고 새 구멍이 아니다 — 그러나 사람 승인(불변식 3 · 7)이 올바른 그림 위에서
  이루어지도록 명시한다: **a127 착지 뒤에는 서명 매니페스트 발급과 schedule 활성화가 곧 실진입 경로를 연다.** 그 발급 · 활성화는 사람 승인 항목이다.
- 손절 즉시성: 무관(진입 권한만).
- **권한을 넓히지 않는다**: 원장 내용 판단은 그대로이고, 엔진이 이미 admission · 재검증에서 같은 원장을 같은 함수(risk)로 읽는다. 열리는 것은
  「서명이 갖춰져도 영원히 거절」이라는 결함 상태뿐이다.

### D7. 판독은 한 읽기 트랜잭션에서, route 의 조건부 질의는 판독 전에 prepare

- **risk**: 버전 확인(`:375`) · scope latch(`:380`) · 다섯 dimension 사용량(`:399`)이 지금은 각각 autocommit 이다. 엔진 밖 `flatten` 이 engine lock
  없이 마이그레이션할 수 있으므로(증거) 확인과 판독 사이에 새 스키마가 커밋되는 창이 있고, 다섯 dimension 이 서로 다른 커밋을 볼 수 있다.
  섞인 snapshot 은 admission 이 쓰기 트랜잭션에서 다시 읽어(`journal/risk_bucket_usage.go:30-56`) fail-closed 하지만, a127 이 같은 줄을 편집하므로
  route 처럼(`production.go:603`) **읽기 전용 트랜잭션 하나**로 묶는다(비용 ≈ 0, `ReadJournalBucketUsage` 는 `UsageQueryer` 를 받음).
- **route**: `openProductionRouteSnapshot` 이 tx 위에서 owners · campaign 두 질의를 **판독 전에 prepare** 한다(같은 SQL 상수를 prepare 와 실행에
  쓴다 — 둘째 철자 금지). 그래야 「읽는 열이 없으면 fail-closed」가 active owner 가 없는 범위에서도 참이다.

## 반증 설계 (구현 로트가 세울 것)

| id | 변이 | 잡아야 할 시험 |
|---|---|---|
| S1 | risk 비교를 `version != 27` 로 되돌림 | 실제 원장 양성(risk) |
| S2 | route 비교를 `version != 27` 로 되돌림 | 실제 원장 양성(route, 외부 시험 패키지) |
| S3 | 버전 검사 삭제(risk) | 더 새 원장 거절(risk) |
| S4 | 버전 검사 삭제(route) | 더 새 원장 거절(route) |
| S5 | `!=` → `>` (더 옛 원장 수락 — (a) 로의 후퇴) | 더 옛 원장 거절(두 적재기) |
| S6 | 0 이하 주입 수락(가드 삭제) | 주입 0 + **존재하지 않는 원장 경로** → 오류가 열기 실패가 아니라 주입 누락 문구(가드가 열기 전에 섬을 관측) |
| S7 | engine 이 상수 대신 런타임 `user_version` 을 넘김 | 구조 단언(선택자 `journal.SchemaVersion`) |
| S8 | route 의 판독 전 prepare 삭제 | active owner 없는 범위 + 읽기 집합 열 삭제 원장 → 거절 |
| S9 | 더 새 · 더 옛 문구 뒤바꿈 | 방향 문구 단언(두 적재기, route 는 Batch 경계) |
| S10 | engine 이 리터럴(예: 35)을 넘김 | 구조 단언(선택자) — 값 비교로는 오늘 같아서 못 잡음 |
| S11 | 스키마 거절을 `ErrProductionRiskScopeRefused` 로 재표식 | 스키마 거절 오류가 그 신원을 갖지 않음 + 엔진이 범위를 건너뛰지 않음 |
| S12 | route `:352` 감싸기가 원인을 다시 버림 | Batch 경계 방향 문구 단언 |
| S13 | risk 판독을 tx 밖으로(autocommit 복귀) | 구조 단언(사용량 판독의 queryer 가 tx) |

## 롤백

이전 바이너리로 돌아가면 두 권한은 다시 거절한다(오늘 상태 — 진입 0). 원장 · 스키마 변경이 없으므로 롤백은 바이너리 교체뿐.

## 잔여

- 경로 결속: 적재기는 경로를 다시 열어 읽는다 — 열린 뒤 경로가 다른 파일로 바뀌는 경우는 inode 결속이 아니다(codex). 원장 파일 검증
  (`validateProductionRiskJournalFile` · `validateProductionRouteJournalFile` — 소유자 · 모드 · 심링크)은 그대로.
- 엔진 route 적재기는 오류 원인을 버리고 `StrategyRouteAuthorityInvalid` 로 접는다 — 방향 문구의 운영 관측은 risk 쪽만(범위 밖).
- v34 의 정책 레코드 의미 변화가 RowDigest 에 남기는 것(키의 첫 레코드)은 admission 과 공유 — a127 무관.
- 레인 활성화 · 서명 정책 · route 매니페스트 발급 · schedule 활성화는 a112 · 사람 승인 항목(D6).
</content>

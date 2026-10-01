# a127 design (4판 — freeze 리뷰 1~3라운드 반영, review.md 「0.5.1」~「0.5.3」)

> 분기 · early return 을 근거로 쓰는 문장은 `analysis/freeze-ast/` 의 AST 산출물(편집 전, base `f9a25549`, `tools/logic-map`)에서 읽었다 —
> `loadProductionRiskEntries` 14 분기 · `openProductionRouteSnapshot` 4 분기 · `LoadProductionRiskSnapshotAuthority` 7 ·
> `LoadProductionRouteAuthorityBatch` 16 · `LoadProductionRouteAuthority` 2. 두 적재기 파일의 source sha256 은 `65fe66e7` 에서도 같다.
>
> **좌표 기준**: 본문의 `production_snapshot_authority.go` · `strategyrouter/production.go` 줄 번호는 `65fe66e7`(= `f9a25549` 와 같은 바이트) 기준이다.
> 그 뒤 a112 6.1 (C) 가 두 파일 앞쪽을 편집해 `18109568` 에서 risk 는 **+9**(버전 확인 `:384` · latch `:389`/`:393` · 사용량 `:408`), route 는
> **+7**(`BeginTx` `:610` · 버전 확인 `:616` · 성공 반환 `:621` · owner 없음 반환 `:659` · campaign 질의 `:672`) 줄 밀렸다(codex 3R P3). 구현 로트는 그 해시 위
> 편집 전 AST 로 다시 잡는다(tasks 1.0.2).

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
  해석이 문자열 리터럴로 떨어지는 경로가 없다. **그러나 두 적재기 모두 조건부 질의가 있다**: route 의 `loadProductionRouteOwnersFrom` 은 active
  owner 가 없으면 `:652-653` 에서 성공으로 반환해 `:665` 의 campaign 질의를 prepare 하지 않고(codex 1R P1), risk 는 scope latch 가 선 범위에서
  사용량 질의 전에 `ErrProductionRiskScopeRefused` 로 돌아간다(2R 보이스 P2-1) → D7.
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
- **음성**(두 적재기): 주입 0 · 음수 · 미설정 거절(열기 전), 더 새 · 더 옛 `user_version` 거절 + 방향 문구, 조건부 질의 전용 열을 지운 원장 거절 —
  route 는 **active owner 가 없는 범위**, risk 는 **latch 가 선 범위**(D7, 반증표의 픽스처 규율). 열 삭제 픽스처 주의: SQLite `DROP COLUMN` 은 인덱스 · 트리거가 쓰는 열을 거절한다(예: `released_at` 은
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

### D6. §0 안전 불변식 대조 — 무엇이 진입을 막는가(주문 경로 기준)

- LIVE 주문 side effect: 없음 — 적재기는 읽기 전용(`mode=ro` · `query_only`). 시험은 원장 픽스처만.
- **승격은 진입 관문이 아니다**(1 · 2판 정정 — 2판은 supervisor `:451-452` 의 시장 승격을 「오늘 0 인 이유」로 들었으나 그것은 **화면 · 승격 판정**
  경로다). 생산 감독자는 `NewRefreshingPairedStrategyEntrySupervisor`(`cmd/tossctl/engine.go:682`)이고 그 worker 는 dormant + `RefreshesAuthority`
  라 `evaluationState` 가 승격과 무관하게 사이클을 허용한다(`strategy_entry_supervisor.go:1044-1046`). 주문은 매 사이클
  `runProductionStrategyMarketCycle` → `dispatchStrategyMarketHandoffs` → `strategyDispatchCycle.dispatch` 로 나가고, 코드 자신이 「`Effective` 는
  화면과 승격 판정만 움직인다 … 그 경로는 이 서술자를 읽지 않는다」라고 적는다(`:500-502`). **승격이 막힌다고 주문이 막히지 않는다.**
- **a127 뒤 주문 경로에 남는 조건** — 두 무리로 나눈다(3라운드 codex: 「전수 1~12, 자동은 11 · 12 뿐」은 거짓이었다 — 자동 조건이 더 많고 사람 조건 하나가
  빠졌다). a127 이 없애는 것은 **4 · 8 의 핀뿐**이다.
  - **(A) 사람 · 운영 조건** — 리뷰 세 라운드가 찾은 전부(전수 증명은 아님):
    1. 엔진 기동: `engine.automation_gate.enabled` + 검증된 attestation(`ectx.Automation.Verified`, 아니면 `errEngineGateOff` — `cmd/tossctl/engine.go:226`).
    2. 거래 정책: `trading.place` 등 필요한 허용과 LIVE 마스터 스위치(`internal/app/engine/interlock.go` `checkTradingPolicy` ~`:571`, 실행 시점
       `internal/trading/service.go:244` `AllowLiveOrderActions`) — Gateway 가 실행 옵션을 스스로 만들므로(`internal/execgw/gateway.go:461`) 주문마다의
       새 사람 승인이 아니라 **설정**이다.
    3. 공식 자격 증명(OpenAPI 키 · 토큰) — 없으면 공식 경로가 서지 않는다.
    4. route: 서명 route 매니페스트(+ 원장 owner snapshot — **a127 이 고치는 자리**).
    5. schedule: 서명 scheduler 활성화 매니페스트로 복원된 `restore.Activation` · calendar(`strategy_dispatch_cycle.go:86-88`, `strategy_proposal_authority.go:314`).
    6. candidate: 서명 threshold · evidence 매니페스트.
    7. FX: 서명 FX 정책 매니페스트 + 공식 FX 관측.
    8. risk: 서명 위험 정책 매니페스트(+ 원장 사용량 — **a127 이 고치는 자리**).
    9. proposal: 제안 서명 키 · 매니페스트 digest · evidence DB identity 환경값(`strategy_proposal_authority.go:323-350`, `TOSSOS_STRATEGY_EVIDENCE_*_ID`).
    10. account: 서명 계좌 권한.
    11. 보호 배선 · readiness 배포(a100 — `ReasonProtectionNotWired`, `internal/execgw/protection_refusal.go:4`).
    12. (선택) 4-가족 활성화 서명 — 없으면 유효 제안이 **정확히 하나**여야 하고(`strategy_account_first_leg_authority.go:161-163`, 둘이면 거절)
        handoff 는 시장 단위 하나(`strategy_dispatch_handoff.go:38-39`), 있으면 범위마다 handoff(다중 범위 — 이 경로도 a127 뒤 도달 가능).
  - **(B) 자동 런타임 조건**(스위치가 아님 — 원장 · 관측 상태가 정함, 예시이며 전수 아님): 감독자 accepting · 레인 미잠금 · 사이클 존재
    (`strategy_entry_supervisor.go:1041`), 캠페인 미점유 · FLAT/CLOSED(`strategy_market_handoff_delivery.go:42`), 진입 관문(latch · 신선도,
    `strategy_dispatch_cycle.go:143`), 보호 readiness 의 런타임 증거(`:96`), dispatch owner · lease 발급 · claim · fencing(`:147` · `:190` · `:196`,
    `internal/execgw/strategy_gateway.go:93`), 1차 레그 admission(q_final · 다섯 bucket · owner · 손실 잠금)과 Guardian, 충돌 attempt · 미사용
    결정 · 지원 주문 모양 · 읽을 수 있는 매수 여력(`internal/execgw/gateway.go:525` · `:535` · `:594`, `failclosed.go:40` · `:129`), 제출 전후 보호 재확인.
- **경보 수위**(Manager 판정 2026-10-01): a127 은 진입을 「연다」가 아니다 — **핀이라는 우연한 차단이 사라지고 설계된 조건 사슬(위 1~12)만 남는다**.
  (A) 는 전부 사람 서명 또는 운영자 설정이고, (B) 는 스위치가 아니라 상태 판정이다. (A) 밖의 비사람 스위치는 세 라운드 리뷰에서 찾지 못했다.
- **불변식 3(토글 OFF = upstream)**: automation gate OFF 면 엔진이 기동을 거절한다((A)1) — upstream 과 같다. a127 은 이 경로를 건드리지 않는다.
  **불변식 7(사람 승인)**: (A) 의 서명 · 발급 · 설정 전부가 사람 승인 항목이다.
- **오늘 동작 변화 0 의 영수증**: 생산 설정 실측은 a127 문서에 없다(가장 최근 기록은 a112 8.7.1 「생산에 서명 매니페스트 0건(측정)」, 2026-09-04).
  배포 전 재실측을 사람 항목 H1 로 둔다(tasks 2.0) — 매니페스트 digest 환경값 · scheduler 활성화 · automation gate 상태.
- 손절 즉시성: 무관(진입 권한만).
- **권한을 넓히지 않는다**: 원장 내용 판단은 그대로이고, 엔진이 이미 admission · 재검증에서 같은 원장을 같은 함수(risk)로 읽는다.

### D7. 판독은 한 읽기 트랜잭션에서, 두 적재기 모두 판독 전에 prepare

- **같은 트랜잭션**: risk 의 버전 확인(`:375`) · scope latch(`:380`) · 다섯 dimension 사용량(`:399`)이 지금은 각각 autocommit 이다. 엔진 밖
  `flatten` 이 engine lock 없이 마이그레이션할 수 있으므로(증거) 확인과 판독 사이에 새 스키마가 커밋되는 창이 있고, 다섯 dimension 이 서로 다른
  커밋을 볼 수 있다(섞인 snapshot 은 admission 이 쓰기 tx 에서 다시 읽어 fail-closed — `journal/risk_bucket_usage.go:30-56`). a127 은 route 처럼
  (`production.go:603`) **버전 · latch · 사용량 판독 전부를 읽기 전용 트랜잭션 하나**로 묶는다(`ReadJournalBucketUsage` 는 `UsageQueryer` 를
  받고 `*sql.Tx` 가 만족 — 판정 · 오류 신원 불변). `SetMaxOpenConns(1)` 이라 tx 밖에 남은 판독은 그 연결을 기다리다 ctx 기한까지 멈춘다 — 막는
  쪽 실패이지만 S13 이 모든 판독의 수신자를 단언한다.
- **판독 전 prepare**: 각 적재기는 버전 확인 직후 · 첫 판독 전에 자기 SQL 상수 전부를 그 tx 위에서 prepare 한다(같은 상수를 prepare 와 실행에 —
  둘째 철자 금지).
  - route: owners · campaign 두 질의. campaign 은 active owner 가 있을 때만 실행되므로(`:652-653`) prepare 가 없으면 그 범위에서 열 부재가 안 드러난다(codex 1R P1, 측정 2판).
  - risk: scope latch · 사용량 두 질의. 사용량 질의는 scope latch 가 0 일 때만 실행된다 — latch 가 선 범위는 사용량 판독 전에
    `ErrProductionRiskScopeRefused` 로 돌아가므로, prepare 가 없으면 **사용량 전용 열이 없는 원장이 범위 국소 거절로 재표식**돼 엔진이 다음 범위로
    넘어간다(2R 보이스 P2-1). prepare 를 latch early return 앞에 두면 스키마 결함이 결함 신원으로 먼저 나온다.
  - **오류 우선순위의 적용 범위**(3R codex P2): 위 순서(주입 → 파일 → 버전 → prepare 결함 → **원장에서 나온** 범위 국소 거절)는 **정책 · 입력
    결속 뒤** 원장 적재 안에서만 성립한다. `LoadProductionRiskSnapshotAuthority` 는 원장을 열기 전에 `bindProductionRiskInputs` 를 부르고
    (`18109568` `:169` → 적재 `:173`), 서명 정책에 종목 섹터 매핑이 없으면 `:311` 에서 `ErrProductionRiskScopeRefused` 를 돌려준다 — 그 범위에서는
    원장 결함보다 정책 범위 거절이 먼저다. 이것은 원장을 읽지 않은 정확한 범위 국소 거절이므로 바꾸지 않는다(같은 원장 결함은 매핑이 있는 범위에서
    결함으로 드러난다). 주입 누락(D2)만 정책 결속보다 앞에 둔다(열기 전 거절).
- **의존 기록**: 「prepare 가 없는 열을 드러낸다」는 modernc sqlite v1.54.0 이 prepare 를 즉시 수행한다는 사실에 기댄다(`stmt.go:23-45`
  `newStmt` → `prepareV2`, codex 2R 확인). 드라이버 판본이 바뀌어 지연 prepare 가 되면 S8 · S14 가 깨져 알린다.

## 반증 설계 (구현 로트가 세울 것)

픽스처 규율: **더 새 · 더 옛 원장 픽스처는 읽기 집합을 온전히 갖추고 `user_version` 만 바꾼다**(그렇지 않으면 prepare 가 대신 거절해 S3 · S4 · S5 가
생존). **열 삭제 픽스처는 조건부 질의 전용 열**을 지운다(route: `position_campaigns.entry_blocked` + active owner 없는 범위, risk:
`risk_bucket_final_decisions.owner_prospective_generation` 같은 사용량 전용 열 + latch 가 선 범위) — owners · latch 질의도 읽는 열을 지우면 prepare
삭제 변이가 생존한다. 표 재생성 방식은 측정 하네스(`readset-probe_test.go.txt`).

| id | 변이 | 잡아야 할 시험 |
|---|---|---|
| S1 | risk 비교를 `version != 27` 로 되돌림 | 실제 원장 양성(risk) |
| S2 | route 비교를 `version != 27` 로 되돌림 | 실제 원장 양성(route, 외부 시험 패키지 — 변이 하네스가 `tossos_testseams` 스위트를 돌려야 함) |
| S3 | 버전 검사 삭제(risk) | 더 새 원장 거절(risk) — 읽기 집합 온전 · 버전만 증가 |
| S4 | 버전 검사 삭제(route) | 더 새 원장 거절(route) — 같은 픽스처 규율 |
| S5 | `!=` → `>` (더 옛 원장 수락 — (a) 로의 후퇴) | 더 옛 원장 거절(두 적재기) |
| S6 | 0 이하 주입 수락(가드 삭제 · `<=0`→`==0`) | 주입 0 **과 음수** + 존재하지 않는 원장 경로 → 오류가 열기 실패가 아니라 주입 누락 문구(가드가 경로 검증보다 앞) |
| S7 | engine 이 상수 대신 런타임 `user_version` 을 넘김 | 구조 단언(import 경로를 해석한 AST — 두 config 필드 값이 journal import 의 `SchemaVersion` 선택자. 지역 식별자가 import 이름을 가리는 경우는 범위 밖 — 1.6 리뷰 P3-4, go/types 로 올리지 않음) |
| S8 | route 의 판독 전 prepare 삭제 | active owner 없는 범위 + `position_campaigns` 전용 열 삭제 원장 → 거절 |
| S9 | 더 새 · 더 옛 문구 뒤바꿈 | 방향 문구 단언(두 적재기, route 는 Batch 경계) |
| S10 | engine 이 리터럴(예: 35)을 넘김 | 구조 단언(S7 과 같은 시험) — 값 비교로는 오늘 같아서 못 잡음 |
| S11 | 스키마 거절을 `ErrProductionRiskScopeRefused` 로 재표식 | 스키마 거절 오류가 그 신원을 갖지 않음 + 엔진이 범위를 건너뛰지 않음 |
| S12 | route `:352` 감싸기가 원인을 다시 버림 | Batch 경계 방향 문구 단언 |
| S13 | risk 판독 일부를 tx 밖으로(버전 · latch · 사용량 중 하나라도) | 구조 단언 — 세 판독의 수신자가 모두 같은 tx(수신자 동일성과 tx 수명) |
| S14 | risk 의 판독 전 prepare 삭제(또는 latch early return 뒤로 이동) | latch 가 선 범위 + 사용량 전용 열 삭제 원장 → 거절이 `ErrProductionRiskScopeRefused` 가 아님 |
| S13d · S13e | risk BeginTx 의 ReadOnly 제거 · 버전 확인 뒤 tx 를 닫고 같은 이름으로 다시 엶 | 구조 단언 — BeginTx 하나 · ReadOnly · tx 대입 하나 · defer 된 Rollback 외 닫기 없음(구현 리뷰 1.6.2) |
| S15 · S16 | 원장 데이터 질의를 prepare 앞에 둠(risk) · opener 가 prepare 앞에서 데이터를 읽음(route) | 구조 단언 — 데이터 질의는 마지막 prepare 뒤, opener 는 데이터 질의 없음 · 모든 prepare 뒤에만 성공 반환(1.6.2) |

## 롤백

이전 바이너리로 돌아가면 두 권한은 다시 거절한다(오늘 상태 — 진입 0). 원장 · 스키마 변경이 없으므로 롤백은 바이너리 교체뿐.

## 잔여

- 경로 결속: 적재기는 경로를 다시 열어 읽는다 — 열린 뒤 경로가 다른 파일로 바뀌는 경우는 inode 결속이 아니다(codex). 원장 파일 검증
  (`validateProductionRiskJournalFile` · `validateProductionRouteJournalFile` — 소유자 · 모드 · 심링크)은 그대로.
- 엔진 route 적재기는 오류 원인을 버리고 `StrategyRouteAuthorityInvalid` 로 접는다 — 방향 문구의 운영 관측은 risk 쪽만(범위 밖).
- v34 의 정책 레코드 의미 변화가 RowDigest 에 남기는 것(키의 첫 레코드)은 admission 과 공유 — a127 무관.
- 레인 활성화 · 서명 정책 · route 매니페스트 발급 · schedule 활성화는 a112 · 사람 승인 항목(D6).

# a127 design


> 분기 · early return 을 근거로 쓰는 문장은 `analysis/freeze-ast/` 의 AST 산출물(편집 전, base `f9a25549`, `tools/logic-map`)에서 읽었다 —
> `loadProductionRiskEntries` 14 분기 · `openProductionRouteSnapshot` 4 분기 · `LoadProductionRiskSnapshotAuthority` 7 ·
> `LoadProductionRouteAuthorityBatch` 16 · `LoadProductionRouteAuthority` 2.

## 증거 기반

- **핀 두 자리**(AST): `loadProductionRiskEntries` B5 `:375` — `PRAGMA user_version` 읽기 실패 **또는** `version != productionRiskJournalSchema`
  이면 `:376` 에서 `errors.New("risk bucket: exact journal schema unavailable")` 반환. `openProductionRouteSnapshot` B4 `:609` — 같은 조건이면
  tx rollback · db close 뒤 `:614` `ErrProductionRouteUnavailable`. 둘 다 이 분기가 함수의 원장 판독(B6 이후 · 호출자의 `loadProductionRouteOwnersFrom`)
  **앞**에 선다.
- **생산 진입점**: supervisor 가 `journalPath = c.Journal.Path()`(`strategy_entry_supervisor.go:300-301`)를 route 적재기(`:305`)와 risk 적재기
  (`:328`)에 같이 넘긴다. `c.Journal` 은 `journal.Open` 이 연 핸들이고 Open 은 마이그레이션 뒤 반환한다(`journal.go:196`), 원장이 더 새로우면
  `ErrSchemaTooNew` 로 실패한다(`schema.go:20` 규칙). → **생산에서 두 적재기가 여는 파일의 `user_version` 은 항상 `journal.SchemaVersion`**.
- **읽기 집합은 현재 원장에서 성립한다**(`analysis/measurements/readset-probe.log`): 새로 연 v35 원장에서 route 적재기의 두 SQL(원문 그대로) ·
  risk 적재기의 scope latch 질의가 prepare · 실행된다. risk 의 나머지 판독은 `riskbucket.ReadJournalBucketUsage` 하나인데, 이 함수는 이미
  journal admission(`refuseStaleBucketUsage`) · 제출 재검증(`RevalidateQFinalAdmission`)이 **현재 원장에서** 매번 부른다 — 핀만이 생산
  snapshot 경로를 막고 있었다.
- **v28~v35 가 읽기 집합을 건드린 것**: 마이그레이션 SQL 전수(`internal/journal/*_v2[8-9].sql` · `*_v3[0-5].sql`)에서 읽기 집합 표
  (`risk_bucket_reservations` · `_snapshots` · `_policies` · `_owner_release_receipts` · `_owners` · `_final_decisions` · `_scope_latches` ·
  `position_campaigns`)에 대한 `ALTER` · `CREATE TRIGGER` 는 v34 하나 — `risk_bucket_reservations` 에 nullable `policy_record_digest` 와 그
  삽입 · 불변 트리거. 적재기는 그 열을 읽지 않는다. 행 의미(상태 값 · 금액 · latch)를 바꾼 단계는 없다.
- **없는 열은 prepare 에서 실패한다**(같은 측정의 대조 줄 — `no such column`). 두 적재기의 질의 오류는 각각 B10/`:401`·B6/`:382` 와
  호출자 오류 분기로 fail-closed 한다.

## 설계 결정

### D1. 결속 방식 — **(b) 주입된 현재 버전과 정확 일치**(Manager 재판정 2026-10-01)

결함의 본질은 「정확 일치」가 아니라 **마이그레이션과 함께 움직이지 않는 동결 리터럴**(27)이다.

**규칙**: `user_version == 주입 값(journal.SchemaVersion)` 일 때만 판독한다. 불일치는 방향을 문구로 가른다 — 더 새 원장(이 빌드보다 새 스키마)과
더 옛 원장(마이그레이션되지 않은 원장). 표 · 열 존재는 별도 목록으로 검사하지 않고 SQLite prepare 에 맡긴다(D4 의 열 삭제 변이로 고정).

**판정 교체 사슬.** 배정(2026-10-01) 때 방향은 (a) 「정본 선례 `journal.ReadOnly.checkSchema` 패턴 — `user_version > 상한` 이면 TooNew
거절 + 필요한 표 · 열 존재 확인」이었다. 착수 실측 뒤 (b) 를 권고로 올렸고 Manager 가 (b) 로 재판정해 (a) 지시를 대체했다. 근거 넷:

1. **생산 불변식이 `==` 다.** 두 적재기가 여는 파일은 같은 프로세스가 방금 연 · 마이그레이션한 원장이다(위 증거). (a) 가 추가로 받는
   「더 옛 원장」은 생산에 존재할 수 없고, 받아 주면 열은 있으나 트리거 · 제약이 그 버전 이전인 원장(예: v34 정책 레코드 트리거 이전)을
   읽는 **허용 방향 구멍**만 남는다.
2. **선례의 상황이 다르다 — 교차 바이너리 vs 동일 프로세스.** `journal.ReadOnly.checkSchema`(`readonly.go:229-`)는 **콘솔**용이다 — 다른
   바이너리가 다른 판본의 원장을 열 수 있는 자리라 범위 수락이 맞다. `strategyevidence/readonly.go:43` 은 **같은 빌드의 자기 저장소** —
   `version != SchemaVersion` 거절, 더 새면 `ErrSchemaTooNew`. a127 의 적재기는 후자의 상황이다.
3. **필요 표 · 열 목록은 읽기 집합의 둘째 철자다.** SQL 원문과 목록이 어긋나도 서로의 시험을 통과시킨다(「판정이 둘이면」). 표 · 열 존재는
   SQLite prepare 가 이미 강제한다(증거 · 대조 줄).
4. **결함은 동결 리터럴이다.** 상한을 `journal.SchemaVersion` 상수로 주입하면 마이그레이션마다 값이 함께 움직인다 — 같은 결함이 재발하는
   구조가 사라진다.

### D2. 상한은 호출자가 주입한다 — 0 이면 거절

riskbucket · strategyrouter 는 journal 이 import 한다(`journal/risk_bucket*.go` → riskbucket, `journal/strategy_first_leg_atomic.go` → strategyrouter).
두 패키지가 `journal.SchemaVersion` 을 읽으면 순환이다. 그래서:

- `ProductionRiskSnapshotConfig` · `ProductionRouteConfig` 에 원장 스키마 필드(이름은 구현 로트) 하나씩.
- engine 의 두 호출 자리(`strategy_risk_authority.go:211` 의 config 리터럴 · `strategy_route_authority.go:179` 의 config 리터럴)가 `journal.SchemaVersion`
  **상수**를 넣는다. 런타임에 파일에서 읽은 값을 넣으면 같은 파일을 자기와 비교하는 공허한 검사가 된다 — 상수여야 한다.
- 필드가 0 이하면 두 적재기는 원장을 열기 전에 거절한다(주입 누락 = 결함).

### D3. 거절 신원은 그대로 — 범위 국소가 아니다

`ErrProductionRiskScopeRefused` 주석(`production_snapshot_authority.go` `:41-45`)이 스키마 실패를 결함으로 분류한다. 바꾸지 않는다: 스키마
불일치는 `ErrProductionRiskSnapshotUnavailable`(`LoadProductionRiskSnapshotAuthority` 가 `%w` 로 감쌈) · `ErrProductionRouteUnavailable` 로 남고,
엔진은 그 범위를 건너뛰지 않고 그 시장 주기를 거절한다(오늘과 같다). 문구만 방향을 말한다.

### D4. 수락은 실제 원장으로 — 축소 픽스처는 단독 수락 근거가 아니다

- 양성: 두 적재기를 `journal.Open` 이 연 원장 경로로 부르는 시험. 엔진 패키지(서명 정책 · route 매니페스트 픽스처가 있는 곳)에서 a112 트립와이어를
  양성 시험으로 뒤집는다 — 트립와이어 자신이 「핀이 고쳐지면 실패한다」로 설계돼 있다(`a112_owner_scope_trading_test.go:420-`).
- a112 시험 다리 `a112MirrorLedgerIntoRiskStub` 를 지우고 a112 거래 픽스처의 위험 적재기를 실제 원장으로 단일화(a112 소유 파일 — Manager 통지).
- 음성: 주입 0 거절, 더 새 `user_version` 거절, 더 옛 `user_version` 거절, 읽기 집합의 열 하나를 지운 원장 거절(prepare 실패 경로 고정).
- 구조: engine 의 두 호출 자리가 `journal.SchemaVersion` 상수를 넘긴다는 단언(리터럴 · 런타임 읽기 금지).
- 축소 픽스처(세 곳)는 주입 값과 같은 `user_version` 을 쓰게 바꾼다 — 단언 · 행은 바꾸지 않는다.

### D5. 바꾸지 않는 것

판독 SQL, 사용량 판정(`aggregateProductionRiskUsage`), owner 재구성(`loadProductionRouteOwnersFrom`), 원장 스키마 · 마이그레이션, 엔진의 진입 판정 순서.

### D6. §0 안전 불변식 대조

- LIVE 주문 side effect: 없음 — 적재기는 읽기 전용(`mode=ro` · `query_only`). 시험은 원장 픽스처만.
- 토글 OFF 동작: 서명 활성화 0 이면 오늘과 같다(권한이 ready 여도 활성화 없이는 진입 0 — a112 4-가족 관문).
- 손절 즉시성: 무관(진입 권한만).
- **진입을 여는 방향의 근거**(High-risk 보수 방향 원칙): 이 수리는 권한을 **넓히지** 않는다 — 원장 내용 판단은 그대로이고, 엔진이 이미
  admission · 재검증에서 같은 원장을 같은 함수로 읽고 있다. 열리는 것은 「서명 활성화 뒤에도 영원히 거절」이라는 결함 상태뿐이다.

## 반증 설계 (구현 로트가 세울 것)

| id | 변이 | 잡아야 할 시험 |
|---|---|---|
| S1 | risk 비교를 `version != 27` 로 되돌림 | 실제 원장 양성(risk) |
| S2 | route 비교를 `version != 27` 로 되돌림 | 실제 원장 양성(route) |
| S3 | 버전 검사 삭제(risk) | 더 새 원장 거절(risk) |
| S4 | 버전 검사 삭제(route) | 더 새 원장 거절(route) |
| S5 | `!=` → `>` (더 옛 원장 수락 — (a) 로의 후퇴) | 더 옛 원장 거절 |
| S6 | 주입 0 수락 | 주입 0 거절 |
| S7 | engine 이 상수 대신 런타임 `user_version` 을 넘김 | 구조 단언 + 더 새 원장 거절(engine 경유) |
| S8 | 읽기 집합 열 하나를 `COALESCE(…,'')` 등으로 우회 | 열 삭제 원장 거절 |

## 롤백

이전 바이너리로 돌아가면 두 권한은 다시 거절한다(오늘 상태 — 진입 0). 원장 · 스키마 변경이 없으므로 롤백은 바이너리 교체뿐.

## 잔여

- 스냅숏 일관성: risk 적재기는 버전 확인과 판독을 한 트랜잭션에 묶지 않는다(route 는 묶음). 엔진만 원장을 쓰고 마이그레이션은 기동 때뿐이라
  오늘 위험은 없다 — 기록만(구현 로트가 같은 자리를 편집하므로 묶을지는 freeze 리뷰가 판단).
- 레인 활성화 · 서명 정책 · route 매니페스트 발급은 a112 · 사람 승인 항목.
</content>
</invoke>

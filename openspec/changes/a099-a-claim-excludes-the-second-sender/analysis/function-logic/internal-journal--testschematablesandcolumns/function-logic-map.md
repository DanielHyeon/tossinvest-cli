# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이 테스트는 어떤 계획에도 없었다.**
>
> a099의 §0~§7 어디에도 이 함수가 안 나온다. schemaV31이 `alert_outbox`에 열
> 넷을 더하는 순간 **이 테스트가 깨졌고**, 그 사실은 `internal/journal` 패키지를
> 통째로 돌리기 전까지 안 보였다 — 열 목록을 **통으로 고정**하고 있기 때문이다.
>
> **그것이 이 테스트의 설계이고 옳다.** 스키마에 열이 조용히 생기면 안 된다.
> a099의 편집은 `wantColumns["alert_outbox"]`에 네 이름을 더한 것이고,
> 주석 두 줄로 **어느 change의 어느 스키마 버전이 더했는지**를 적었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 저널 | 새로 연 임시 저널 | `openTestJournal` | 열리면 최신 스키마다 |
| `wantTables` | **테이블 이름 전체 목록** | `:82-…` 리터럴 | 하나만 달라도 `t.Fatalf` |
| `wantColumns` | **일부 테이블의 열 전체 목록** | 리터럴 맵 | 하나만 달라도 `t.Errorf` |
| `alert_outbox`의 열 | **19개** (a099가 4개 더함) | `wantColumns` | 이름 정렬 후 문자열 비교 |

**불변식**: *"스키마 표면은 선언된 것과 정확히 같다."*
빠진 것도, **더해진 것도** 실패다.

## Branches and early returns

AST 열거 — 분기 7 · 이탈 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:161` | 테이블 목록 질의 실패 | 없음 | `t.Fatal` | 이 함수 자신 |
| B2 `:165` | `rows.Next()` — 테이블마다 | `gotTables`에 append | — | 같음 |
| B3 `:167` | `rows.Scan` 실패 | 없음 | `t.Fatal` | 같음 |
| B4 `:172` | `rows.Err()` | 없음 | `t.Fatal` | 같음 |
| B5 `:176` | **테이블 목록이 다르다** | 없음 | `t.Fatalf` | 같음 |
| B6 `:386` | `wantColumns`의 테이블마다 | `tableColumns` 조회 | — | 같음 |
| B7 `:389` | **열 목록이 다르다** | 없음 | `t.Errorf` | 같음 |

**이탈이 0이다.** `t.Fatal`이 흐름을 끊는다.

**a099가 바꾼 것은 분기가 아니라 B7이 비교하는 리터럴이다.**
`alert_outbox`의 want 목록에 `claim_expires_at`·`claim_token`·`claimed_at`·`claimed_by`
넷이 들어갔다. **분기 수도 이탈 수도 그대로다.**

## Calls and live bindings

- `openTestJournal` — 저널 배치
- `j.db.QueryContext` — `sqlite_master`에서 테이블 이름
- `rows.Next` / `rows.Scan` / `rows.Err` / `rows.Close` — 커서
- `strings.Join` — **정렬된 이름을 한 문자열로 비교한다**
- `tableColumns` — 테이블 하나의 열 이름 (헬퍼)
- `sort.Strings` — 열 이름 정렬

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.

## State mutations and fallbacks

- **임시 저널 하나.** 읽기만 한다.
- **폴백 없음.**

## Safety conclusion

- **Safe edit boundary**: **`wantColumns`의 리터럴만 바꿨다.** B1~B7의 조건은 그대로다.
  이 목록에서 열을 **빼는** 편집이 금지선이다 — 그러면 스키마가 조용히 커질 수 있다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 지키는 계약은 **원장 스키마의
  표면**이고, 그것은 High-risk다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **열의 타입·NOT NULL·기본값을 안 본다.** 이름만 본다.
    schemaV31의 `NOT NULL DEFAULT ''` 넷 중 어느 것도 여기서 확인되지 않는다.
    **`TestMigrationV30ToV31LeavesExistingAlertsClaimable` `migration_v31_test.go:18`이
    기본값의 *효과*를 본다** — v30에서 올라온 행이 곧바로 청구 가능한지.
  - **인덱스를 안 본다.** a099는 인덱스를 안 더했지만, 더했더라도 이 테스트는 조용했다.
  - **`wantColumns`에 없는 테이블은 열 검사를 아예 안 받는다.**
    맵에 있는 것만 본다 — `alert_outbox`는 있다.
  - **이 테스트가 계획에 없었다는 사실 자체가 발견이다.**
    스키마를 건드리는 change는 이 함수를 §0에서 먼저 찾아야 한다.
    §5.1이 그것을 기록한다.

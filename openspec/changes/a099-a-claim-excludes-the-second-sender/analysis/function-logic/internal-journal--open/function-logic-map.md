# Function Logic Map: `Open`

- Source: `internal/journal/journal.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099는 이 함수를 편집한다** — task 4.1b. `Options.AlertLease`가 비면
> `DefaultAlertLease`를 채운다(design C4). 편집은 **한 자리, 세 줄**이고
> 그 자리는 이미 존재하는 두 선례 바로 옆이다.
>
> **이 산출물은 3라운드에 만들어졌다.** 2·3판은 *"`journal`이 `Options.AlertLease`로
> 받는다"*라고 적으면서 **그 값을 채우는 함수의 산출물을 안 만들었다.**
> 산출물 없이 쓴 「기본값을 쓴다」는 미검증이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Path` | 비면 `DefaultPath()` | B1 `:110` | 해결 실패면 이탈 `:113` |
| `opts.Clock` | nil이면 `clock.System()` | **B6 `:136-138`** | 없음 — 항상 채워진다 |
| `opts.BusyTimeout` | `<= 0`이면 `defaultBusyTimeout` | **B7 `:139-142`** | 없음 — 항상 채워진다 |
| **`opts.AlertLease`** | **오늘 없다** | — | **a099가 더한다. B7과 같은 모양** (C4) |
| `opts.FSProber` | nil이면 `SystemFSProber()` | B3 `:121-123` | 파일시스템 거절이면 이탈 `:128` |
| `opts.migrationOverride` | nil이면 `defaultMigrationPlan()` | B11 `:170-172` | — |

**불변식 — a099가 지켜야 하는 것**: *"주입 안 된 옵션은 이 함수에서 **한 번** 채워지고,
그 뒤로 `Journal`은 자기 값만 쓴다."*

이 불변식이 사용자 결정 6-1의 절반을 진다. **lease는 claim UPDATE 한 자리에서만
쓰이고**(design C2) 거기서 계산한 만료가 **행에 저장된다.** 그래서 다른 핸들이 다른
`AlertLease`로 열려 있어도 **남의 유효한 임차를 재해석하지 못한다.**

> **⛔ 3판은 만료를 저장하지 않았고, 그래서 이 함수가 채우는 값이 「남의 임차를
> 판정하는 자」가 됐다 — 3라운드 A-P5 = B-R12.** 1라운드 B-P4가 인자에서 쫓아낸
> 것이 `Options`를 통해 되돌아온 자리가 정확히 여기다.

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 14 · 이탈 8 · 호출 24 · defer 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:110` | `opts.Path`가 비었다 | `DefaultPath()` `:111` | (B2로) | 기존 |
| B2 `:112` | 경로 해결 실패 | 없음 | 이탈 `:113` 오류 | **없음** |
| B3 `:121` | `opts.FSProber == nil` | `SystemFSProber()` `:122` | — | 기존 |
| B4 `:127` | 파일시스템 거절 | 없음 | 이탈 `:128` 오류 | 기존 (FS 검사 테스트) |
| B5 `:131` | `os.MkdirAll` 실패 | 없음 | 이탈 `:132` 오류 | **없음** |
| **B6 `:136`** | **`clk == nil`** | **`clk = clock.System()`** `:137` | — | 기존 (모든 테스트가 주입한다) |
| **B7 `:140`** | **`busy <= 0`** | **`busy = defaultBusyTimeout`** `:141` | — | 기존 — `durability_test.go:902` `openTestJournalWithBusy`가 **주입 쪽만** 덮는다 |
| B8 `:145` | `sql.Open` 실패 | 없음 | 이탈 `:146` 오류 | **없음** |
| B9 `:155` | `db.PingContext` 실패 | `db.Close()` `:156` | 이탈 `:157` 오류 | **없음** |
| B10 `:165` | `checkIntegrity` 실패 | `db.Close()` `:166` | 이탈 `:167` 오류 | 기존 (손상 DB 테스트) |
| B11 `:170` | `migrationOverride != nil` | `plan` 교체 `:171` | — | 기존 (migration 테스트) |
| B12 `:173` | `migrate` 실패 | `db.Close()` `:174` | 이탈 `:175` 오류 | 기존 |
| B13 `:179` | `range` 세 경로 | — | — | 기존 |
| B14 `:180` | `os.Stat` 성공 | `os.Chmod(0600)` `:181` | — | 기존 |
| — 이탈 `:184` | 정상 | — | `j, nil` | 기존 |

**a099가 더하는 것은 B7 바로 뒤의 분기 하나다** — `lease := opts.AlertLease;
if lease <= 0 { lease = DefaultAlertLease }`. **분기가 14 → 15가 된다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clock.System` `:137` | 시계 기본값 | 오류 없음 | AST |
| `dsn` `:144` | DSN 조립 — `busy`가 여기로 들어간다 | 오류 없음 | AST |
| `db.SetMaxOpenConns(1)` `:151` | **단일 쓰기 연결** | — | AST · `:147-150` 주석 |
| `j.checkIntegrity` `:165` | 스키마 건드리기 **전에** 손상 검사 | 실패면 기동 거부 | AST |
| `j.migrate` `:173` | v30 → **v31** (a099) | 실패면 `db.Close()` + 오류 | AST |

**`SetMaxOpenConns(1)`가 §3.5 측정의 조건이다** — 임차 UPDATE 하나가 같은 연결을
쓰는 다른 쓰기 뒤에 줄 선다.

## State mutations and fallbacks

- 이 함수는 **`Journal` 값을 만드는 자리**다 (`:160`). `AlertLease`는 그 구조체에
  들어가고, 그 뒤로 **호출자가 못 바꾼다.**
- 기본값 채우기의 방향은 **항상 「값이 있는 쪽」**이다 — nil/0을 그대로 두지 않는다.
  a099의 lease도 같다. **0을 그대로 두면 모든 임차가 즉시 만료된다** — 그것은
  임차가 없는 것과 같고, a099의 요구 자체가 무효가 된다.
- **migration 실패는 `db.Close()`로 끝난다** (B12). v31이 실패하면 열리지 않는다.

## Safety conclusion

- **Safe edit boundary**: **B7 바로 뒤에 분기 하나와 `Journal` 필드 하나.**
  B1~B14의 기존 조건과 여덟 이탈의 의미는 안 바꾼다. 편집 후 AST의 branches가
  **15**이고 새로 는 하나가 `:14x`의 lease 기본값이면 다른 제어 흐름은 무변화다.
- **High-risk impact**: **yes** — 원장 개방 경로(불변식 5). 이 함수가 실패하면
  엔진이 안 뜬다. **다만 a099의 편집은 오류 경로를 하나도 안 더한다** — 값 채우기뿐이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B7의 「기본값이 쓰인다」쪽을 직접 단언하는 테스트가 없다.**
    `openTestJournalWithBusy`(`durability_test.go:902`)는 **주입하는 쪽**만 덮는다.
    a099는 lease에 대해 그 구멍을 **R20으로 메운다** — `DefaultAlertLease > bound`.
    `BusyTimeout` 쪽 구멍은 a099 밖이다 — **`not-applicable`**.
  - **B2·B5·B8·B9에 테스트가 없다.** 경로 해결·mkdir·sql.Open·Ping 실패.
    a099가 안 건드리므로 **`not-applicable`**이지만 이름은 적어 둔다.

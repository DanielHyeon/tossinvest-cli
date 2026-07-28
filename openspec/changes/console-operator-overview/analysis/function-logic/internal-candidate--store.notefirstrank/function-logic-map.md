# Function Logic Map: `Store.NoteFirstRank`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. `NoteFirstPrice`의 쌍둥이에 정체성 창 검사가 하나 더 붙었다. D20이 실행으로 재현한 결함 때문이다 — D8이 관측과 승격을 일부러 분리하므로 냉각·만료 중에도 순위 행이 들어오고, 그 중 하나가 ±TTL 창 안에 떨어져 새 삶의 '최초 목격'이 된다. 4위로 승격한 후보가 148위로 읽히고 `seen_late`가 clear된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rank` / `total` | 둘 다 양수, `rank <= total` | 호출자(`recordFirsts`) | 하나라도 어기면 에러. 부재는 rank 0이 아니고, 목록 길이 없는 순위는 정규화할 수 없다 |
| `at` | 0이 아닌 관측 순간 | **관측 자신의 `ObservedAt`**, 스캔의 `at`이 아니다 | 0이면 에러 — 늦음을 판정할 수 없는 최초 목격은 최초 목격이 아니다 |
| `source` | trim·소문자 후 비어 있지 않아야 함 | 호출자 | 빈 id는 에러. 읽기·쓰기 양쪽 정규화가 어긋나면 조회가 0행을 돌려주고 그것은 WARMING_UP으로 읽힌다 |
| `first_seen_at` | 후보의 이번 삶 시작 | `candidates` | 읽을 수 없으면 에러 |
| 정체성 창 | `nearFirstSighting` — `|at - first_seen_at| < DefaultStalenessTTL` | `veto.go:736` | **저장소 로컬 `StalenessTTL` 오버라이드가 아니라 상수**다. 읽기 쪽 backstop이 상수로 쓰여 있어, 로컬 값을 쓰면 이쪽이 저장하고 저쪽이 거부하는 방향으로 어긋난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 인자 검증 switch | 없음 | — | `TestARankOfZeroIsNotAFirstSighting` |
| B2 | `rank <= 0 || total <= 0` | 없음 | `FirstRank{}, error` | `TestARankOfZeroIsNotAFirstSighting` |
| B3 | `rank > total` | 없음 | `FirstRank{}, error` | `TestARankOfZeroIsNotAFirstSighting` |
| B4 | `at.IsZero()` | 없음 | `FirstRank{}, error` | `TestARankOfZeroIsNotAFirstSighting` |
| B5 | `normaliseSource` 에러 | 없음 | `FirstRank{}, wrap` | `TestARankOfZeroIsNotAFirstSighting` |
| B6 | `BeginTx` 에러 | 없음 | `FirstRank{}, wrap` | 직접 테스트 없음 |
| B7 | `row.Scan` 결과 switch | — | — | `TestARankOfZeroIsNotAFirstSighting` |
| B8 | `sql.ErrNoRows` | 롤백 | `FirstRank{}, ErrNoCandidate` — 대상 없는 provenance write가 조용히 성공하면 실패한 스캔보다 더 완전히 감춘다 | `TestARankOfZeroIsNotAFirstSighting` |
| B9 | 그 밖의 읽기 에러 | 롤백 | `FirstRank{}, wrap` | 직접 테스트 없음 |
| B10 | `storedRank.Valid && > 0` — 이미 기록됨 | **쓰지 않는다** | 기존 값 반환(idempotent) | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` |
| B11 | `parseStamp(firstSeen)` 에러 | 롤백 | `FirstRank{}, wrap` | 직접 테스트 없음 |
| B12 | `!nearFirstSighting(at, seen)` | **쓰지 않는다** | `FirstRank{}, nil` — 에러가 아니다 | `TestARankFromOutsideTheIdentityWindowIsNotStored` |
| B13 | `tx.ExecContext` 에러 | 롤백 | `FirstRank{}, wrap` | 직접 테스트 없음 |
| B14 | `tx.Commit` 에러 | 롤백 | `FirstRank{}, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `normaliseSource` | trim + 소문자 | 빈 id는 거부 | `TestAPaddedSourceIdIsNotASilentWarmUp`가 읽기 쪽의 같은 규칙을 고정한다 |
| `tx.QueryRowContext` | `first_seen_at` + 네 컬럼을 한 번에 | `_txlock=immediate` 안 | ast.json calls |
| `nearFirstSighting` | 이 관측이 이번 삶의 최초 목격일 수 있는가 | 순수, 대칭 창 | `veto.go:736` |
| `tx.ExecContext` | `... WHERE ... AND first_rank IS NULL` | 술어가 읽기를 반복한다 — 두 writer가 둘 다 NULL을 보면 나중 것이 최초 목격을 덮는다 | ast.json calls |
| `decodeFirstRank` | 이미 있는 값을 그대로 돌려준다 | 읽을 수 없는 `first_rank_at`은 에러 | 같은 파일 |

## State mutations and fallbacks

- `candidates.first_rank`/`first_rank_total`/`first_rank_at`/`first_rank_source` 네 컬럼. 각 삶에 한 번만.
- 1회성이 계약 전부다. 스캔 루프는 tick마다 순위를 제안하고, 재시작·재진입·순위를 나르지 않는 원천이 처음 올린 후보에서 첫 제안은 최초 목격이 아니다.
- 창 밖 관측은 **저장되지 않고 실패도 아니다.** 가격이나 캔들로 올라온 뒤 한참 뒤에야 순위 목록에 잡힌 후보의 평범한 상태이고, 순위가 없는 편이 `seen_late`를 **틀리게** 만드는 것보다 낫다 — 미측정은 D10 덕에 통과가 아니지만 그럴듯한 숫자는 통과다.
- 부분 실패: 트랜잭션이라 부분 상태가 없다. `recordFirsts`가 실패를 `ScanResult.Rejected`에 이름으로 남기고 나머지 심볼은 계속 간다.

## Safety conclusion

- Safe edit boundary: 창 검사를 없애기, 창을 저장소 로컬 `StalenessTTL`로 바꾸기, 창 밖을 에러로 만들기, `first_rank IS NULL` 술어 제거는 모두 D20이 실행으로 재현한 결함의 복원이라 금지
- High-risk impact: no — 발굴 저장소 write다. 주문 경로 무접촉. 다만 여기 저장되는 값이 `seen_late` veto의 내구성 있는 절반이고, D20의 결함은 그 veto를 **안전한 척 끄는** 방향이었다. 이 함수의 두 가드(1회성·정체성 창)가 그 방향을 막는 전부다.

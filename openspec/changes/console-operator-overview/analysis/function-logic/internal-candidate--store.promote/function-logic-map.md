# Function Logic Map: `Store.Promote`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

본문이 바뀐 기존 함수다. 이 change가 넣은 것은 upsert의 `DO UPDATE SET`에 `first_rank`/`first_rank_total`/`first_rank_at`/`first_rank_source` 네 컬럼의 만료 초기화와 그에 맞춘 positional 인자 4개다. 규칙 자체는 `first_price`와 동일하다 — 냉각·재진입에서 보존, 만료에서 초기화.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` / `symbol` | 대문자화·trim 후 비어 있지 않아야 함 | 호출자 | 공백이면 에러 |
| `at` | 0이 아니고 **마지막 관측보다 이르지 않은** 순간 | 호출자 인자 | 역행하면 에러. `first_seen_at`은 이 패키지가 지키는 유일한 필드라 검사 없는 인자로 닿을 수 없어야 한다(§1 2-b) |
| 기존 행 | `readCandidate` | `candidates` | 없으면 새 삶, 있으면 `stateAt`으로 만료 여부 판정 |
| `reset` | `boolToInt(expired)` | 계산값 | 모든 placeholder가 positional이다. `?1`과 맨 `?`를 섞으면 주변 번호가 밀려 잘못된 컬럼에 바인딩된다 — 첫 판이 market="0", symbol="KR"을 썼다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `market == "" || symbol == ""` | 없음 | `Candidate{}, error` | 직접 테스트 없음 |
| B2 | `at.IsZero()` | 없음 | `Candidate{}, error` | 직접 테스트 없음 |
| B3 | `BeginTx` 에러 | 없음 | `Candidate{}, wrap` | 직접 테스트 없음 |
| B4 | `readCandidate` 에러 | 롤백 | `Candidate{}, err` | 직접 테스트 없음 |
| B5 | `found && at.Before(existing.LastSeenAt)` | 롤백 | `Candidate{}, error` — 수명은 앞으로만 간다 | `TestAPromotionCannotRunBackwards` |
| B6 | `found && !expired` | `firstSeen = existing.FirstSeenAt` | — | `TestAReEntryWithinTheCoolingTTLKeepsTheOriginalFirstSeenAt` |
| B7 | upsert 에러 | 롤백 | `Candidate{}, wrap` | 직접 테스트 없음 |
| B8 | `tx.Commit` 에러 | 롤백 | `Candidate{}, wrap` | 직접 테스트 없음 |
| B9 | `found && !expired` (반환 조립) | 이전 삶의 출처·완전성을 이어받는다 | `Candidate` | `TestANewCandidateDoesNotInheritTheDeadOnesSources` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readCandidate` | 트랜잭션 안에서 기존 행 | `_txlock=immediate`가 read-then-write를 감싼다. DEFERRED면 두 번째 프로세스가 `SQLITE_BUSY_SNAPSHOT`을 받고 busy_timeout은 그것을 재시도하지 않는다(§1 7절) | `store.go:dsn` |
| `stateAt` | 저장된 타임스탬프에서 상태를 **유도** | 순수. 암묵 냉각(`last_seen_at + staleness`) 포함 | `store.go:1669` |
| `tx.ExecContext` | upsert | 실패는 롤백 | ast.json calls |
| `stamp` | 고정폭 직렬화 | — | `stamp`는 고정폭(`2006-01-02T15:04:05.000000000Z07:00`)이다. TEXT 컬럼이라 모든 `<`·`>=`·`ORDER BY`가 BINARY 문자열 비교이고, `time.RFC3339Nano`는 소수부 뒤 0을 떼어내 텍스트 순서가 시간 순서가 아니게 만든다 — §1 리뷰가 실행으로 재현한 P0다(`TestTheStoredInstantOrdersLikeTime`). |

## State mutations and fallbacks

- `candidates` upsert 1건. 컬럼별 규칙: `first_seen_at`은 만료가 아니면 보존, `last_seen_at`은 항상 `at`, `cooled_at`은 항상 NULL(재활성).
- 만료면 `sources`·`sources_attempted`·`sources_responded`·`degraded`·`first_price{,_at,_source}`·`first_rank{,_total,_at,_source}` 전부 초기화. 만료가 아니면 전부 보존.
- 만료 초기화가 필요한 이유: 죽은 삶의 기준선으로 새 삶의 확장률을 재면 두 삶을 가로질러 재는 것이고, 죽은 삶의 순위로 새 삶의 `seen_late`를 답하면 D20이 실행으로 재현한 '4위로 승격한 후보가 148위로 읽힘'이 된다.
- 재진입 보존이 필요한 이유: 한 스캔 동안 목록을 떠났다 돌아온 이미 두 배 오른 종목이 새 기준선·새 순위를 받으면 D1이 앞문에서 막은 우회가 옆 필드로 들어온다.
- 부분 실패: 트랜잭션이라 부분 상태가 남지 않는다. 실패한 심볼은 `Collect`의 `Rejected`에 이름만 남고 나머지 시장은 계속 간다.

## Safety conclusion

- Safe edit boundary: 만료 판정에서 `first_rank`/`first_price` 초기화를 빼거나 재진입에서 초기화하는 것, 역행 시각 거부를 없애는 것, positional placeholder에 `?1` 형식을 섞는 것 모두 금지
- High-risk impact: no — 발굴 저장소의 후보 수명 갱신이다. 주문 경로 무접촉이고 의존 폐포는 `{internal/clock}`. 다만 이 함수가 `first_seen_at`을 쓰는 유일한 자리이고, D1은 그것이 추격 veto가 우회되는 경로라고 명시한다. 안전 불변식에는 닿지 않지만 이 패키지의 주장 전체가 여기에 있다.

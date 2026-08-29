# Function Logic Map: `decodeFirstRank`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. `decodeBaseline`의 쌍둥이 — 저장된 NULL·비양수를 미기록으로 접는 한 자리.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rank` / `total` sql.NullInt64 | NULL이거나 비양수면 미기록 | `candidates` | 제로값 `FirstRank{}`로 나가고 에러가 아니다 |
| `at` sql.NullString | RFC3339 | `candidates.first_rank_at` | 읽을 수 없으면 **에러** — 나이를 판정할 수 없는 최초 목격은 조용히 통과시키지 않는다 |
| `source` sql.NullString | 원천 id | `candidates.first_rank_source` | NULL은 빈 문자열 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!rank.Valid || !total.Valid || rank <= 0 || total <= 0` | 없음 | `FirstRank{}, nil` — 미기록 | `TestARankOfZeroIsNotAFirstSighting` |
| B2 | `at.Valid` | `f.At` 설정 | — | `TestPruningRawObservationsLeavesTheFirstRankToo` |
| B3 | `parseStamp` 에러 | 없음 | `FirstRank{}, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseStamp` | 저장된 순간 해석. `time.RFC3339Nano`를 받아 옛 빌드가 남긴 소수부 폭도 읽는다 | 실패는 에러 | `stamp`는 고정폭(`2006-01-02T15:04:05.000000000Z07:00`)이다. TEXT 컬럼이라 모든 `<`·`>=`·`ORDER BY`가 BINARY 문자열 비교이고, `time.RFC3339Nano`는 소수부 뒤 0을 떼어내 텍스트 순서가 시간 순서가 아니게 만든다 — §1 리뷰가 실행으로 재현한 P0다(`TestTheStoredInstantOrdersLikeTime`). |
| `SourceID` / `int` | 변환 | — | ast.json calls |

## State mutations and fallbacks

- 상태 변경 없음 — 순수 디코더다.
- 비양수를 **에러가 아니라 미기록**으로 접는 것이 계약이다. 마이그레이션 백필이 창 밖 행을 남기지 않아 NULL이 정상 상태이기 때문이다.
- 반대로 읽을 수 없는 타임스탬프는 에러다. 값은 있는데 나이를 모르는 최초 목격은 D17이 `first_price_at`에서 이미 거부한 형태다.

## Safety conclusion

- Safe edit boundary: 읽을 수 없는 `first_rank_at`을 제로 시각으로 흡수하는 것은 금지 — 그러면 정체성 검사가 모든 후보에서 통과한다
- High-risk impact: no — 순수 디코더.

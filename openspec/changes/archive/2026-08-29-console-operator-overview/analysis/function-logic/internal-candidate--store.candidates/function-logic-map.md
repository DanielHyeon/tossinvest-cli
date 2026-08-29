# Function Logic Map: `Store.Candidates`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: base, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**. 뒤에 삽입된 `Summary`/`Summaries`/`Checkpoint`/`FreeSpace`의 diff hunk가 이 함수와 교차해 evidence가 요구되었다. base L1093-1117의 본문은 현재 L1469-1493과 byte 동일하고, ast.json은 base revision이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `at` | 상태를 판정할 순간 | 호출자 인자 — **시계가 아니다** | 상태는 저장이 아니라 유도라, 프로세스 재시작 전에 냉각된 후보가 재시작 후에도 만료로 읽혀야 한다 |
| `candidates` 행 전량 | `ORDER BY market, symbol` | 저장소 | 만료 행도 숨기지 않고 돌려준다 — 우리가 보고 놓아준 후보 수가 이 시스템이 이르다는 증거의 일부다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 질의 에러 | 없음 | `nil, wrap` | 직접 테스트 없음 |
| B2 | 행 순회 | — | — | `TestExpiryIsReadFromTheClockAndNotFromASweeper` |
| B3 | `scanCandidate` 에러 | 없음 | `nil, err` | 직접 테스트 없음 |
| B4 | `rows.Err()` | 없음 | `nil, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryContext` | 후보 전량 | 단일 문장. §2 리뷰가 남긴 미해결: 시장·상태를 SQL에서 걸러야 한다 | `review.md` §2 '수용하지 않은 것' |
| `scanCandidate` | 행 → `Candidate`, 읽을 수 없는 타임스탬프는 에러 | — | 같은 파일 |
| `stateAt` | 저장된 순간들에서 상태 유도. 암묵 냉각 포함 | 순수 | `store.go:1669` |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기 전용이다.
- 상태를 스윕이 아니라 읽기에서 유도한다. 스윕이 돌아야만 전진하는 상태는 방금 올라온 프로세스에서 거짓말을 하고, 그것이 바로 물어보는 시점이다.
- 본문 무변경이므로 이 change가 만든 동작 변화는 없다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재
- High-risk impact: no — 읽기 전용이고 본문 byte 동일. 결과가 `coolAbsent`의 입력이므로 파괴 방향으로만 의미가 있다.

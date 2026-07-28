# Function Logic Map: `Store.FirstRank`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. `Store.Baseline`의 쌍둥이이고, 두 번째 반환이 '그런 후보 없음'과 '순위 원천이 아직 보고하지 않은 후보'를 가르는 것도 같다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` / `symbol` | 대문자화·trim | 호출자 | 읽기·쓰기 정규화가 같아야 패딩된 값이 0행을 조용히 고르지 않는다 |
| 네 컬럼 | INTEGER/TEXT NULL | `NoteFirstRank` 또는 v3 마이그레이션 백필 | NULL은 '이 삶에 기록된 순위 관측 없음'이고 rank 0이 아니다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `row.Scan` 결과 switch | — | — | `TestARankOfZeroIsNotAFirstSighting` |
| B2 | `sql.ErrNoRows` | 없음 | `FirstRank{}, false, nil` | `TestARankOfZeroIsNotAFirstSighting` |
| B3 | 그 밖의 에러 | 없음 | `FirstRank{}, false, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryRowContext` | 네 컬럼 단건 조회 | 단일 문장 | ast.json calls |
| `decodeFirstRank` | NULL·비양수를 미기록으로 | 읽을 수 없는 `first_rank_at`은 에러 | 같은 파일 |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기 전용이다.
- found=false와 미기록 `FirstRank{}`는 다른 사실이다. 둘 다 `Recorded()==false`지만 하나만 수명에 관한 질문이다.

## Safety conclusion

- Safe edit boundary: 두 번째 반환을 없애는 것, 미기록을 rank 0으로 접는 것은 금지
- High-risk impact: no — 읽기 전용. 반환값이 `seen_late`의 입력이므로 부재를 위치로 접는 방향만 위험하다.

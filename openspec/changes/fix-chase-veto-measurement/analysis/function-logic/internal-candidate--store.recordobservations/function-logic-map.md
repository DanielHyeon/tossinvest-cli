# Function Logic Map: `Store.RecordObservations`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L709–767, 분기 9개)
- Risk scan: `risk-pattern-report.md`

스캔의 읽기를 `observations`에 append한다. 이 change가 INSERT에 **칼럼 두 개**를 더했다 —
`newly_listed_state`(3-상태 TEXT, nullable)와 `rank_requested`(nullable INTEGER).

schema-1의 `newly_listed` boolean은 **그대로 쓰되 다시 읽지 않는다**. 전진 전용이라 drop도
rename도 하지 않으므로 스키마 4 이전 빌드가 이 파일을 열어도 기대하는 모양을 찾는다. 다만
그 칼럼의 0은 "아니오"와 "아무도 보지 않았다"를 구분할 수 없어서, 이 빌드는
`newly_listed_state`만 읽는다(`scanObservation`).

`positive(o.Reported.RankRequested)`가 0을 NULL로 만든다 — 이 저장소에서 0이 더 작은 수량이
**아닌** 유일한 정수이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in` | 0..N `Observation` | `Collect` | 비면 no-op |
| `o.validate()` | 모든 행이 통과해야 한다 | `candidate.go` | 한 행이라도 실패하면 **트랜잭션을 열기 전에** 전체 거부 |
| `o.Reported.RankRequested` | 0 또는 양수(음수는 validate가 거부) | 소스 어댑터 | 0은 NULL — 미기록 |
| `o.Reported.NewlyListed` | 3-상태 | 소스 어댑터 | unknown은 NULL |
| `o.Source` | 정규화 가능한 id | `normaliseSource` | 실패는 오류 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `len(in) == 0` | 없음 | `nil` | `TestAnEmptyReadingIsNotEvidenceOfAbsence`(빈 읽기 → 관측 0건) |
| B2 | `for i, o := range in` — 검증 pass | `rows` 구성 | — | `TestANegativeRequestedCountIsRefusedByTheObservationBoundary` |
| B3 | `err != nil` — validate 실패 | **아무것도 쓰지 않는다** | 행 번호를 실은 오류 | `TestANegativeRequestedCountIsRefusedByTheObservationBoundary` |
| B4 | `BeginTx` 실패 | 없음 | 오류 | 커버 없음 — I/O 실패 경로 |
| B5 | `PrepareContext` 실패 | 없음 | 오류 | 커버 없음 — 칼럼 수 불일치가 여기서 터진다 |
| B6 | `for _, o := range rows` — INSERT pass | 행마다 한 INSERT | — | `TestATruncatedReadingReachesTheVerdictAsTruncated`(두 새 칼럼의 왕복) |
| B7 | `normaliseSource` 실패 | 부분 쓰기는 rollback | 오류 | **도달 불가** — `validate`가 빈 source를 먼저 거부한다 |
| B8 | `stmt.ExecContext` 실패 | rollback | 오류 | **커버 없음** — 칼럼 CHECK는 raw handle로만 시험된다 |
| B9 | `tx.Commit()` 실패 | 없음 | 오류 | 커버 없음 — I/O 실패 경로 |

검증(B2·B3)이 **트랜잭션 밖**이라는 것이 이 change에서 중요해졌다. 음수 요청 행 수를 실은
행 하나가 배치를 통째로 거부시키고 부분 결과를 남기지 않는다는 것을
`TestANegativeRequestedCountIsRefusedByTheObservationBoundary`가 확인한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.validate()` | 경계 검증 | 오류는 그대로 상승, 쓰기 없음 | ast.json calls |
| `newlyListedToStore(o.Reported.NewlyListed)` | 3-상태 → NULL/'yes'/'no' | 오류 없음 | ast.json calls |
| `positive(o.Reported.RankRequested)` | 비양수 → NULL | 오류 없음 | ast.json calls |
| `boolToInt(o.Reported.NewlyListed.Yes())` | schema-1 boolean 유지 | 오류 없음 — 다시 읽히지 않는다 | ast.json calls |
| `tx.PrepareContext` / `stmt.ExecContext` / `tx.Commit` | 배치 INSERT | 실패는 defer rollback | ast.json calls/defers |

## State mutations and fallbacks

- `observations` 테이블에 N행 append. upsert 아님 — 같은 심볼의 두 instant는 두 행이다.
- 이 change가 더한 두 칼럼은 **가산·nullable**이고 기존 칼럼을 하나도 건드리지 않는다.
- fallback 없음. 실패는 rollback이며 부분 배치가 남지 않는다.

## Safety conclusion

- Safe edit boundary: INSERT 칼럼 2개 + 값 2개 가산, `NewlyListed` 대입이 `bool`에서 `.Yes()`로. 삭제 0.
- High-risk impact: no (원장 아님 — 파생 관측 테이블). D16의 볼륨 위험은 기존 그대로이고 이 change가 행 수를 늘리지 않는다.

# Function Logic Map: `TestAPaddedSourceIdIsNotASilentWarmUp`

- Source: `internal/candidate/store_test.go`
- AST evidence: `ast.json` (revision: base, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**. 뒤에 삽입된 §5 보존 테스트 세 개의 diff hunk가 이 함수와 교차해 evidence가 요구되었다. base L1289-1328의 본문은 현재 L1790-1829와 byte 동일하고, ast.json은 base revision이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 저장소 | `openStore` — 임시 디렉터리 + `FixedFSProber(ext4)` | 테스트 헬퍼 | 실 파일시스템 판정을 우회하되 프로덕션은 그럴 수 없다(`TestFixedFSProberIsTestOnly`) |
| 관측 1건 | `obs("005930", t0, SourceOfficialPrices, ...)` | 테스트 | 정규 철자로 기록 |
| 조회 철자 3종 | `" official_prices"`, `"official_prices "`, `"Official_Prices"` | 테스트 | 셋 다 같은 1행을 찾아야 한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 관측 기록 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B2 | 세 철자 순회 | — | — | (테스트 자체) |
| B3 | 조회 에러 | 없음 | `t.Fatalf` | (테스트 자체) |
| B4 | `len(got) != 1` | 없음 | `t.Errorf` — 0행 + 에러 없음은 하류에서 WARMING_UP이다 | (테스트 자체) |
| B5 | 패딩된 id로 쓰기 실패 | 없음 | `t.Fatalf` | (테스트 자체) |
| B6 | 정규 철자 조회 에러 | 없음 | `t.Fatalf` | (테스트 자체) |
| B7 | 패딩되어 쓰인 행이 정규 철자로 안 찾힘 | 없음 | `t.Errorf` | (테스트 자체) |
| B8 | 빈 source id가 `SourceObservations`에서 통과 | 없음 | `t.Error` | (테스트 자체) |
| B9 | 빈 source id가 `SourceSeries`에서 통과 | 없음 | `t.Error` | (테스트 자체) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.RecordObservations` | 쓰기 쪽 정규화 확인 | — | `store.go:normaliseSource` |
| `s.SourceObservations` / `s.SourceSeries` | 읽기 쪽 정규화 확인 | 빈 id는 에러여야 한다 | 같은 함수 |

## State mutations and fallbacks

- 임시 디렉터리의 `candidates.db`만 만진다. `t.TempDir`가 정리한다.
- 본문 무변경이므로 이 change가 만든 동작 변화는 없다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재
- High-risk impact: no — `_test.go`이므로 프로덕션 바이너리에 들어가지 않고 주문 경로에 닿지 않는다. 다만 이 테스트가 지키는 대상(후보 수명·`first_seen_at`·두 저장된 사실)이 이 change의 유일한 주장이라, 단언을 느슨하게 하는 방향은 결함을 조용히 통과시킨다.

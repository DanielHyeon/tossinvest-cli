# Function Logic Map: `StrategyRuntimeSnapshotFunc`

- Source: `internal/httpapi/strategy_runtime.go` (52-67)
- AST evidence: `ast.json` — AST 분기 3 · return 5 · defer 0
  (source_sha256 `0338efa58b203e8be67620e10f262d0e2f39794d759f431dd50e6ba579f82366`,
  a109 §2.3 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **a109 편집 대상: B1 의 조건** — `reader == nil` 을 `StrategyRuntimeAbsent(reader)` 로
  바꾼다. 설계 D4 가 열거한 두 자리 중 하나다(freeze P1-4).
- **정직한 기록**: 이 함수는 **오늘 production 호출자가 없다**
  (`StrategyRuntimeSnapshotFunc` 참조는 정의 1 + 계약 테스트 1). 그래서 이 자리를 고쳐도
  운영 동작은 바뀌지 않는다 — 진짜 REST 경로는 `router.read` B14 이고, 설계가 그것을
  빠뜨렸다(issues.md T2-1). 여기를 함께 고치는 이유는 **판정을 한 벌로 유지**하기
  위해서다: 남겨 두면 다음에 이 helper 를 배선하는 사람이 옛 판정을 되살린다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | nil 또는 부재를 말하는 wrapper 이면 **부재** | `StrategyRuntimeAbsent` | 부재면 오류 (B1) — stream 은 부재를 스냅샷으로 그리지 않는다 |
| `now` | non-nil 이어야 한다 | 호출자 | nil 이면 같은 오류 (B1) |
| 스냅샷 계약 | `strategyprojection.Validate` 통과 | 모델 | 위반은 오류 (B3) |

**불변식**: 부재의 뜻은 nil 검사 시절과 **같아야 한다**. wrapper 는 정의상 non-nil 이라
`reader == nil` 만으로는 부재를 물을 수 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| **B1 (:54)** | **`StrategyRuntimeAbsent(reader)` 또는 `now == nil`** | 없음 | 오류 | `TestTheStreamHelperRefusesAnUnconfiguredWrapper` |
| B2 (:58) | 읽기 실패 | 없음 | 원인 그대로 | `TestStrategyRuntimeRESTDormantHealthAndSSEFullSnapshotParity` |
| B3 (:61) | 스냅샷 계약 위반 | 없음 | 원인 그대로 | 기존 계약 테스트 |
| — (:64) | 셋을 지났다 | 없음 | 봉투 JSON | `TestTheStreamHelperRefusesAnUnconfiguredWrapper`(붙은 쪽) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrategyRuntimeAbsent` | 부재 판정 한 벌 | 순수 함수 | AST · strategy_runtime.go:54 |
| `reader.Read` | 실제 읽기 | 실패는 그대로 올린다 | AST · :56 |
| `strategyprojection.Validate` / `Clone` | 계약 검사·복사 | 위반은 오류 | AST · :60·64 |
| `json.Marshal` | 봉투 직렬화 | 오류를 그대로 올린다 | AST · :63 |

## State mutations and fallbacks

- 상태 변경 없음. fallback 없음 — 부재도 실패도 오류로 답한다(그것이 stream 의 계약이다).

## Safety conclusion

- Safe edit boundary: **B1 의 조건**. 나머지 둘은 계약 검사이고 건드리지 않는다.
- High-risk impact: **no** — 조회 전용 stream helper 이고 오늘은 호출자도 없다.
- 금지: 부재 판정의 두 번째 사본을 여기에 쓰는 것.

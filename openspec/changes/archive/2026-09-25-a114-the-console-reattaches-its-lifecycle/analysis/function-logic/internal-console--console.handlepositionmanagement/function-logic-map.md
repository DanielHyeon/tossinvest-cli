# Function Logic Map: `Console.handlePositionManagement`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json` (base `634cf3c5`, 분기 25)
- Risk scan: `risk-pattern-report.md`

**a114 는 이 함수를 편집하지 않는다.** design 이 이 함수의 분기를 근거로 쓰므로(FLM-before-claiming)
AST 를 먼저 만들었다. 근거: `Wired` 는 `PositionPolicies != nil`(:225)이고, nil 이면 B5(:236)에서 「배선되지 않아 조회만 가능」으로 끝난다. non-nil 이면 Runtime·List 를 부르고 List 실패는 B9(:251) → `LoadErr`(부재를 상태로 표시).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.PositionPolicies` | nil 또는 commander | runConsole | nil 이면 미배선 화면 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c.opts.Settings == nil {` (:228) | — | — | 편집 없음 |
| B2 | `} else if desired, verdict, err := c.opts.Settings.Load(); err != nil {` (:230) | — | — | 편집 없음 |
| B3 | `} else if desired, verdict, err := c.opts.Settings.Load(); err != nil {` (:230) | — | — | 편집 없음 |
| B4 | `} else {` (:232) | — | — | 편집 없음 |
| B5 | `if c.opts.PositionPolicies == nil {` (:236) | — | — | **근거**: nil commander → 미배선 화면으로 조기 반환 |
| B6 | `if runtimeErr != nil {` (:241) | — | — | **근거**: Runtime 실패 → RuntimeErr |
| B7 | `} else {` (:243) | — | — | 편집 없음 |
| B8 | `for _, block := range runtime.Blocks {` (:246) | — | — | 편집 없음 |
| B9 | `if err != nil {` (:251) | — | — | **근거**: List 실패 → LoadErr 로 조기 반환 |
| B10 | `if exitJournal.Readable() {` (:259) | — | — | 편집 없음 |
| B11 | `for _, row := range exits {` (:260) | — | — | 편집 없음 |
| B12 | `if commander, ok := c.exitQuarantines(); ok {` (:269) | — | — | **근거**: 격리 해제 표면 발견(`exitQuarantines`) |
| B13 | `if quarantineErr != nil {` (:271) | — | — | 편집 없음 |
| B14 | `for _, row := range rows {` (:274) | — | — | 편집 없음 |
| B15 | `for _, state := range states {` (:285) | — | — | 편집 없음 |
| B16 | `if !exitJournal.Readable() {` (:293) | — | — | 편집 없음 |
| B17 | `if stored, ok := exitByPosition[state.PositionID]; ok && stored.HasExit &&` (:299) | — | — | 편집 없음 |
| B18 | `if view.Snapshot != nil {` (:304) | — | — | 편집 없음 |
| B19 | `if management.Block != nil {` (:310) | — | — | 편집 없음 |
| B20 | `if quarantine, ok := quarantines[state.PositionID]; ok {` (:320) | — | — | 편집 없음 |
| B21 | `if state.Status == positionpolicy.StatusManaged {` (:323) | — | — | 편집 없음 |
| B22 | `} else if state.Status == positionpolicy.StatusReleased && state.ExternalLifecyc` (:332) | — | — | 편집 없음 |
| B23 | `for _, policy := range exitpolicy.RegisteredCommonPolicies() {` (:325) | — | — | 편집 없음 |
| B24 | `if state.ExternalLifecycleEligible() {` (:329) | — | — | 편집 없음 |
| B25 | `} else if state.Status == positionpolicy.StatusReleased && state.ExternalLifecyc` (:332) | — | — | 편집 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.opts.PositionPolicies` 의 메서드 | 엔진 lifecycle 읽기/명령 | 오류는 화면 값으로 | AST calls |

## State mutations and fallbacks

- 편집 없음. a114 이후 이 함수가 받는 commander 는 engineDir 가 있으면 non-nil wrapper 이고, 부착 전
  호출은 연결 없는 detached 오류를 돌려준다.

## Safety conclusion

- Safe edit boundary: 편집 없음(0줄).
- High-risk impact: no.

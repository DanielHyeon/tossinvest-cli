# Function Logic Map: `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes`

- Source: `internal/execgw/riskguardian_account_base_testseam_test.go` (102-187)
- Qualified: `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes`
- AST evidence: `ast.json`(현재) — 현재 AST branches 14 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` e23432f073a3c6a7…) — 편집 전 갈래 수 14
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — 여덟 레인 첫 다리의 Guardian 계보 투영.

**a139 편집.** 같은 이유로 rig 예산 명시.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if len(descriptors) != 8 {` (if 104) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B2 | `for _, descriptor := range descriptors {` (range 108) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B3 | `if err != nil {` (if 127) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B4 | `if err != nil {` (if 134) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B5 | `if err != nil {` (if 147) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B6 | `if precheck.QCandidate() != 20 \|\| precheck.QFinal() != 10 \|\| rig.collections != 0 {` (if 150) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B7 | `if err != nil {` (if 154) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B8 | `if receipt.DecisionID != decisionID \|\| receipt.QFinal != 10 \|\| receipt.Market != market \|\|` (if 157) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B9 | `if err != nil {` (if 162) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B10 | `if err != nil {` (if 166) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B11 | `if err != nil {` (if 170) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B12 | `if !ok \|\| intent.Quantity != "10" \|\| intent.PolicyVersion == result.ExecutionTerms.Policy().Identity() \|\|` (if 174) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B13 | `if plans, err := rig.journal.PendingStrategyPlans(context.Background(), "acct-7"); err != nil \|\| len(plans) != 0 {` (if 178) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |
| B14 | `if markets["KR"] != 4 \|\| markets["US"] != 4 {` (if 184) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run AllEightLanes` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 29종 — 대표: `Identity`, `bindPairedAccountBaseFXForTest`, `context.Background`, `errors.Unwrap`, `fixedIDs`, `guardianPolicy`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a139 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a139 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)

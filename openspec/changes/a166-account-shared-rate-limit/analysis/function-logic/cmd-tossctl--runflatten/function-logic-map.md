# Function Logic Map: `runFlatten(params=3, results=1)`

- 원문: `cmd/tossctl/flatten.go`. pre-AST는 priority-origin-lot, post-AST는 ast.json.
- 입력·상태·호출·오류·불변식: bootstrap 전부터 critical context. 취소·청산·기존 사람 확인과 오류·deadline 유지. 시험 TestA166FlattenBootstrapPriority.
- CodeGraph 미초기화; source 및 Go AST 근거. High-risk.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L125: `if ctx == nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | if L130: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | if L135: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | if L147: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L152: `if opts.cancelOnly {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | if L158: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L161: `if len(targets) == 0 {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B8 | if L166: `if !opts.dryRun {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B9 | if L169: `if err := flatten.Confirm(cmd.InOrStdin(), out, confirmation, interactive, time.Now()); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B10 | if L176: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B11 | if L181: `if !liquidateReport.Flat() && !opts.dryRun {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |

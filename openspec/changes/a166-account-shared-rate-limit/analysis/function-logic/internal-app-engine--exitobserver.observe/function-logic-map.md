# Function Logic Map: `ExitObserver.observe(params=3, results=2)`

- 원문: `internal/app/engine/exitloop.go`. pre-AST는 priority-origin-lot, post-AST는 ast.json.
- 입력·상태·호출·오류·불변식: Retrier.Query 전에 critical을 지정하고 조회·기존 검증·오류·fallback을 유지함. 시험 TestA166ExitObservePricesCritical; retry context는 실제 401 시험에서 확인.
- CodeGraph 미초기화; source 및 Go AST 근거. High-risk.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L871: `if qerr != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | range L880: `for _, symbol := range symbols {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | range L884: `for _, q := range quotes {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | if L886: `if _, needed := wanted[symbol]; !needed \|\| q.Last <= 0 \|\| math.IsNaN(q.Last) \|\| math.IsInf(q.Last, 0) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L890: `if q.FetchedAt.IsZero() {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | else L893: `} else if effectiveAt.After(validatedAt) \|\|` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L893: `} else if effectiveAt.After(validatedAt) \|\|` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B8 | if L900: `if len(next) == 0 {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B9 | if L906: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B10 | if L909: `if needsFallback {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B11 | if L915: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B12 | for L918: `for {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B13 | if L920: `if current >= maximum \|\| o.observationSequence.CompareAndSwap(current, maximum) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B14 | range L926: `for symbol, quote := range out {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B15 | if L927: `if quote.Source != "" {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |

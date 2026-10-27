# Branch Test Map: `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | overview marker 반복 | `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions` | not-applicable | 대상 GREEN 통과 |
| B2 | overview lifecycle/evidence marker 누락 | 같은 시험 | `green-console-output.txt`의 기존 exit marker 요구 실패를 overview 검증으로 정정 | 대상 GREEN 통과 |
| B3 | exit-protection HTML parse 실패 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B4 | DOM node walk 반복 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B5 | DOM child walk 반복 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B6 | non-element node 건너뜀 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B7 | DOM element switch | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B8 | `textarea` 발견 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B9 | `input` 발견 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B10 | hidden/checkbox 외 input type | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B11 | owner registry 밖 `option_id` | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B12 | `contenteditable` 발견 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B13 | text affordance 반복 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B14 | text affordance 발견 | 같은 시험 | not-applicable | 대상 GREEN 통과 |

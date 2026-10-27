# Branch Test Map: `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | preset HTML parse 실패 | `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft` | not-applicable | 대상 GREEN 통과 |
| B2 | form/button 3개 계약 위반 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B3 | `select`·`textarea`·`script` 생성 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B4 | 금지 `select`·`textarea`·`script` 반복 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B5 | 금지 control이 실제로 발견됨 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B6 | preset 흐름 문구·중계 링크 반복 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B7 | 흐름 문구·중계 링크가 누락됨 | 같은 시험 | `green-console-output.txt`에서 기존 비개요 marker 요구 실패 | 대상 GREEN 통과 |
| B8 | invented client draft marker 반복 | 같은 시험 | not-applicable | 대상 GREEN 통과 |
| B9 | invented client draft marker가 발견됨 | 같은 시험 | not-applicable | 대상 GREEN 통과 |

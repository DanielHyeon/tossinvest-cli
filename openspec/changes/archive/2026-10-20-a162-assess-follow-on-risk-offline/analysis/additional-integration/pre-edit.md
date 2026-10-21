# a162 추가 통합검증 pre-edit / post-edit 범위

- 변경 함수: test-tagged `strategyflowEvaluation`, `strategyflowAdvanceLeg`, 그리고 exported test-only negative-input helpers.
- FLM/BTM: [strategyflowEvaluation](../function-logic/internal-weeklyvaluelane--strategyflowevaluation/function-logic-map.md), [strategyflowAdvanceLeg](../function-logic/internal-weeklyvaluelane--strategyflowadvanceleg/function-logic-map.md).
- CodeGraph: [strategyflowAdvanceLeg-final.json](codegraph/strategyflowAdvanceLeg-final.json). 기존 index만 read-only query했고 sync/rebuild는 하지 않았다.
- 기본 build 제외: [weekly-default-build-gofiles.txt](verification-logs/weekly-default-build-gofiles.txt)는 `strategyflow_testseam.go`가 production GoFiles에 없음을 보인다.
- production evaluator, journal, broker, runtime, API, 설정은 수정하지 않았다.

# Function Logic Map tools

기존 Go 함수의 내부 로직을 바꾸기 전에 다음 증거 묶음을 만든다.

```bash
python3 tools/logic-map/scaffold_analysis.py \
  --change <change-id> --file internal/pkg/file.go --func Receiver.Method

go run ./tools/logic-map \
  --file internal/pkg/file.go --func Receiver.Method \
  > openspec/changes/<change-id>/analysis/function-logic/<target>/ast.json

python3 tools/logic-map/risk_pattern_report.py internal/pkg/file.go \
  --output openspec/changes/<change-id>/analysis/function-logic/<target>/risk-pattern-report.md

python3 tools/logic-map/check_analysis.py --change <change-id>
```

`ast.json`은 현재 source SHA-256을 포함한 구조 증거다. Markdown scaffold는 사람이 CodeGraph 호출 관계,
불변식, 실패 경로와 테스트를 채우는 분석 문서다. `TODO`, 빈/오래된 AST, source·함수 불일치,
수정된 기존 함수의 누락은 gate에서 거절된다. CI의 `SDD_BASE_REF`는 persisted
`base-commit.txt`와 같은 commit으로 resolve되어야 한다.
ast-grep 발견은 자동 결함 판정이 아니다.

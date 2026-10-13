# a134 CodeGraphContext supporting context

- 도구: CodeGraphContext 0.5.1. 이 change 는 기존 함수 내부를 바꾸지 않고(편집하는 기존 함수 = `newEngineCmd` 의 `AddCommand` 한 줄)
  새 leaf 패키지 · 새 파일만 더하므로 CGC 질의로 얻을 후보 문맥이 hard evidence 이상을 주지 않음.
- **not-applicable** 사유: 새 leaf(`internal/enginewatch`) · 새 명령 파일 · 새 내보내기 입구는 base 에 없는 기호라 CGC 색인에 대조할 대상이 없고,
  편집하는 기존 함수 하나는 분기 0 의 등록 함수임. 후보 확장 대신 CodeGraph `callers` · `impact` 와 HEAD 직접 읽기로 표면을 닫음(codegraph-baseline.md).

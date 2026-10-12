# CodeGraphContext / GBrain — a129

`not-applicable`(advisory): 대상은 세 함수(`ExitObserver.submit` · `ExitObserver.ObserveOnce` · `engineRuntime`)와 새 파일 하나로 좁고,
구조 질의는 CodeGraph(`codegraph-baseline.md`)·HEAD 직접 읽기·AST 로 충분했다. 과거 결정 recall 은 `scripts/memory-recall.sh`
("exit order log engine.log account", "unobserved log dedicated logger" — 결과 0)와 a090 · a091 아카이브(전용 로거 선례
`UnobservedLog` · `ZeroFloorLog`, `exitloop.go:190-216`, `exitwiring.go:356-357`)를 직접 읽었다.

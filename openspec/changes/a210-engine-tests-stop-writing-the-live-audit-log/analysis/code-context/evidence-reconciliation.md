# a210 증거 조정

- 기준: base `8db0463b7`. 현재 HEAD 사실은 `facts.md`, CodeGraph 원문은 `codegraph-callers.txt` · `codegraph-affected.txt`.
- 변경 표면: 시험 파일만 — 새 파일 `internal/app/engine/a210_isolated_home_test.go`(새 `TestMain` · 새 함수),
  새 파일 `cmd/tossctl/a210_isolated_home_test.go`(새 함수), 기존 `cmd/tossctl/help_convention_test.go` `TestMain` 한 줄.
- CodeGraphContext: `not-applicable` — 생산 심볼을 바꾸지 않고, 누수 판정은 정적 문맥이 아니라 가짜 HOME 실측(`../red` · `../green`)이 정본.
- CodeGraph 와 HEAD 의 차이: 없음. `openAuditLog` 호출자 목록의 `cmd/tossctl` 8 개는 동명 다른 함수(이름 단위 병합)로 HEAD 에서 확인.
  `affected` 는 1,346 시험 파일로 변별력이 없어 후보 선정은 직접 호출 + 직접 import 로 좁히고 실측으로 판정(facts.md).
- design 과 실측의 차이: design 은 엔진 시험 패키지를 누수원으로 보았으나 base 실측은 엔진 0 · `cmd/tossctl` 1(audit.log). facts.md 「차이」.
- 함수 내부 증거(FLM/BTM): 기존 함수 1 개(`cmd/tossctl` `TestMain`)를 고쳐 번들 1 개(`../function-logic/cmd-tossctl--testmain/`).
  `tools/logic-map/check_analysis.py --change a210-…` rc 0(required 1, evidence complete). 엔진 쪽은 새 함수만이라 대상 아님.

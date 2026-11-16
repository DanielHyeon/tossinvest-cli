# CodeGraph baseline

- 도구: codegraph 1.6.0
- 기준 HEAD: `4d7d15571` (origin/main, a190 worktree 생성 시점)
- 색인: 공유 체크아웃(`/mnt/D/Axipient/workspace/TossOS`, 같은 HEAD·작업 트리 깨끗함)의 색인을 **읽기만** 했다.
  a190 worktree 에는 `.codegraph/` 가 없고, `make sdd-sync` 는 GBrain 단계가 HEAD 를 옮긴 전례가 있어 돌리지 않았다
  (완료 gate ⑥ 전에 Manager 가 sync 한다).

## 질의

| 질의 | 결과 |
|---|---|
| MCP `codegraph_explore "tools/gate.sh make validate openspec validate gate step 11"` | `No indexed file uniquely matches tools/gate.sh` — 셸 스크립트·Makefile 은 색인 밖. 돌아온 `validate` 심볼은 전부 무관한 Go 메서드 |
| `codegraph query gate.sh` | 변수 `GATE`(`tools/sdd/test_gate_resolves_archived_changes.py:36`)만 실재 관련 |
| `codegraph affected tools/gate.sh --filter 'test_*'` | `No test files affected` — 셸 파일의 소비자를 모델링하지 않음(그래프가 빈 것이 아니라 셸 import 엣지가 없음) |

## 한계와 대체 증거

CodeGraph 는 `bash tools/gate.sh`·`make validate` 호출 관계를 모델링하지 않는다. 결속은 `git grep`(HEAD) 으로 확인했다.

| 위치 | 내용 |
|---|---|
| `Makefile:130-131` | `validate:` → `openspec validate --all --strict --no-interactive` |
| `Makefile:185-188` | `gate:` → `bash tools/gate.sh $(CHANGE)` |
| `tools/gate.sh:19,58,335-345` | 머리 주석 ⑪, `TOTAL_STEPS=11`, 6~11단계 순회(`validate` 가 마지막 타깃) |
| `tools/sdd/test_race_detector_actually_runs.py:58-83` | 순회 줄 정규식으로 `test-race` 포함과 `TOTAL_STEPS == STEP_NO-1+len(targets)` 를 단언 — 순회에서 `validate` 를 빼면 이 시험을 같이 고쳐야 한다 |
| `tools/sdd/test_gate_resolves_archived_changes.py` | gate.sh 를 1~4단계까지만 돌림(⑪ 무관) |
| `tools/logic-map/test_check_analysis.py:9448-9507` | gate.sh 1단계 해소기만 돌림(⑪ 무관) |
| `.github/workflows/ci.yml` | `make validate`·`openspec` 없음. CI 는 ⑪ 을 돌지 않는다. `make sdd-check-ci` → `make sdd-test` 는 돈다 |
| `docs/WORKFLOW.md:277-280,296` | ⑪ 서술 두 곳 |
| `openspec/specs/sdd-workflow/spec.md:48-57` | 「자동화된 완료 게이트」가 `validate 통과` 를 문장으로 말함 |

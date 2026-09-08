# a122 — The Function Logic Map gate outlives a merge

## Why

완료 게이트 5단계는 change 가 트리의 끝일 때만 답할 수 있는 질문을 한다.

`tools/logic-map/check_analysis.py:111` 의 `changed_existing_functions` 는
`git diff <base> -- '*.go'` 를 **워킹트리**에 대고 돈다(`target` 매개변수가 있으나
CLI 는 그것을 노출하지 않는다 — `main()` 의 인자는 `--change` 와 `--root` 뿐이다).
그래서 base 와 워킹트리 사이에 **다른 change 의 작업이 들어오면** 그 함수들이 전부
"이 change 가 증거를 내지 않은 함수"로 집계된다.

2026-09-08 HEAD `62b35779` 에서 잰 값이다.

| change | `check_analysis.py --change <id>` |
|---|---|
| a074-critical-events-reach-the-operator | FAIL — missing evidence 316건 |
| a077-screens-show-what-they-already-know | FAIL — 318건 |
| a079-operator-can-lift-a-quarantine | FAIL — 317건 |
| a075 · a076 (2026-09-08 아카이브) | 같은 실패 |

지적된 함수는 `internal/verifylive/*` · `internal/scheduler/*` 처럼 이 change 들이
이름조차 언급한 적 없는 패키지의 것이다. 즉 게이트는 "a074 가 a112 의 함수에 증거를
냈는가"를 묻고 있고, **어떤 change 도 그 질문에 답할 수 없다.**

이것은 우연한 사고가 아니라 구조다. `배포 후 실측` 태스크는 **정의상** 배포 뒤에
닫히고, 배포는 다른 작업이 착지한 뒤에 온다. 그래서 그런 태스크를 가진 change 는
전부 이 상태로 끝난다. 지금 다섯이 걸려 있고 앞으로도 같다.

기존 우회로 둘은 답이 아니다.

- **재기준화**(a074~a079 의 review §병합 후 재기준화, 커밋 `840b3377`)는 이제 spec 과
  충돌한다 — `sdd-workflow` 는 "일반 변경의 함수 분석 비교 기준은 **불변**
  `base-commit.txt`여야 한다(SHALL)"고 적는다. 게다가 그 절차는 `revision: current`
  AST 재추출과 분기 ID 재번호를 포함해 조용히 틀릴 여지가 크다.
- **`execution_baseline`** 은 `CHANGE = "a063-…"` 로 하드핀된 일회성 예외이고 일반
  경로가 아니다.

## What changes

- 비교의 **base 쪽은 손대지 않는다.** 불변 `base-commit.txt` 요구를 그대로 지킨다.
  고치는 것은 비교의 **반대쪽 끝**이다.
- change 의 작업이 착지한 지점을 기록하고, 그 지점을 5단계 비교의 target 으로 쓴다.
  그러면 질문이 "이 change 가 고친 함수에 증거가 있는가"로 되돌아오고, 병합 뒤에도
  같은 답을 유지한다.
- 착지 지점이 기록되지 않은 change 는 지금과 똑같이 워킹트리와 비교한다 — 작업 중인
  change 의 판정은 바뀌지 않아야 한다.
- 착지 지점의 기록은 위조 가능하면 안 된다. 그 값이 5단계를 통과시키는 유일한 손잡이가
  되므로, 어떤 값이 유효한지와 그것을 누가 언제 쓰는지를 설계 단계에서 정한다.

## Non-goals

- `base-commit.txt` 를 옮기거나 재기준화 절차를 되살리는 것.
- a063 의 고정 실행 기준선 예외를 일반화하는 것.
- 5단계가 요구하는 증거의 **내용**(function-logic-map · branch-test-map ·
  risk-pattern-report)을 바꾸는 것. 바뀌는 것은 어떤 함수 집합에 그것을 요구하는가다.
- 이미 아카이브된 change 를 소급해 다시 판정하는 것.

## Impact

- `tools/logic-map/check_analysis.py` 와 그 테스트.
- `tools/gate.sh` 5단계 호출부(필요하면).
- `sdd-workflow` 스펙만. 런타임 거래·위험·원장·엔진 기동 의미는 건드리지 않는다.

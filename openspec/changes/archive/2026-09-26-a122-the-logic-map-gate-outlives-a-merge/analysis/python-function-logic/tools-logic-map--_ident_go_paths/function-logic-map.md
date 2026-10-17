# Function Logic Map: `_ident_go_paths` (Python, a122 task 7.5.28, 새 함수)

`ast.after-7.5.28.json` 이 범위를 적는다.

## 답하는 질문

**"`ident` 속성이 켜진 추적 `*.go` 가 있는가."** 워킹트리 대상에서만 불린다. `ident` 는 `$Id: …$` 를 `$Id$` 로
접은 **뒤에** 비교하므로 그 안에 넣은 논리 변경이 두 시야 모두에서 사라진다(재리뷰 재현). 속성이라 `-c` 로 못
끄고, 인덱스 플래그 검사의 해시도 같은 변환을 거친다 — 그래서 켜져 있으면 호출자가 이름 대고 거절한다.

## 갈래

| 갈래 | 무엇 |
|---|---|
| `git ls-files -z -- '*.go'` rc ≠ 0 | 결함 |
| 추적 `*.go` 가 없음 | 빈 목록 (check-attr 를 안 띄운다) |
| `git check-attr -z --stdin ident` rc ≠ 0 | 결함 |
| 출력 칸 수가 3 의 배수가 아님 | 결함 — 지어내지 않는다 |
| 값이 `unspecified`·`unset` 이 아님 | 켜진 것으로 센다 |

**비용**: 워킹트리 대상 판정마다 프로세스 둘. 오늘 저장소의 `.gitattributes` 는 0 개다.
시험 `…the_ident_attribute_is_refused_by_name`(`AG3`·`AG5`) · 경계 `…an_explicitly_unset_ident_is_not_refused`(`AG4` —
`-ident` 는 **끈** 것이다; 그 갈래를 뒤집는 변이가 첫 판에서 살아남아 더한 시험이다).

## task 7.5.25 — 지웠다 (2026-09-24)

이 함수는 git 의 워킹트리 투영에서 **알려진 문 하나**를 닫았다. 7.5.25 가 투영 자체를 안 쓰게 되면서(`_worktree_snapshot`)
닫을 문이 판정에 닿을 자리가 없어졌다 — 호출자도 시험도 없다. 이 디렉터리의 `ast.*.json` 은 그 시점의 기록으로 둔다.

# Function Logic Map: `_header_name` (Python, a122 task 7.5.28, 새 함수)

`ast.after-7.5.28.json` 이 범위를 적는다.

`--- a/x.go` · `+++ b/x.go` 머리 줄에서 이름을 읽는다. 갈래는 하나다 — 줄 끝이 탭이면 떼고, 아니면 그대로.
git 은 이름에 **공백이 있으면** 머리 줄 끝에 탭을 붙인다(GNU diff 의 관례). 떼지 않으면 `my file.go\t` 를 base 에서
찾다가 `cannot load existing base file` 로 거짓 차단했다 — 이 change 이전부터 있던 결함이다. 이름 가드가 `\t` 든
이름을 이미 거절하므로 끝의 탭 하나를 떼는 것은 모호하지 않다.

시험 `…a_name_with_a_space_is_not_refused`(git 이 정말 탭을 붙이는지 먼저 단언한다) ·
`…a_unicode_line_separator_in_a_path…`(그 이름에도 공백이 있다). 변이 `AG6`.

# Function Logic Map: `section_name` (Python, a122 task 7.5.13 — `_changed_existing_functions` 의 안쪽 함수)

## task 7.5.13 — 새 안쪽 함수 (2026-09-27)

편집 후 `tools/logic-map/check_analysis.py:654-669` · 분기 3 · 반환 1 · raise 2 · 호출 6 (`ast.after-7513.json`, source sha `c9a58a69af03`)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 658 | BoolOp | `index < 0 or index >= len(named)` |
| B2 | 658 | If | `if index < 0 or index >= len(named):` |
| B3 | 664 | If | `if header != _git_header_path(prefix + name):` |

`_changed_existing_functions` 안쪽. 구역 번호(`len(bodied) - 1`)로 레코드를 고르고(B1 · B2 — 레코드보다 구역이 많으면 이름 댄 결함), 머리 줄이 그 이름의 렌더링과
다르면 이름 댄 결함(B3). 두 결함 모두 `the two views of the same diff disagree` 를 말한다 — 7.5.23 의 수 대조와 같은 문장이다.

## 마감 수리 — 정상 입력 typechange (적대 리뷰 P2-1, 2026-09-27)

B1 · B2 는 "정상 git 은 안 내는 모양" 의 방어가 **아니다**. typechange(일반 `x.go` → 심링크)에서 `--numstat -z` 는 레코드 **하나**(`1\t2\tx.go\0`),
판정 diff 는 구역 **둘**(`--- a/x.go`/`+++ /dev/null` · `--- /dev/null`/`+++ b/x.go`)을 낸다(리뷰어 실측 · 이 세션이 스크래치에서 다시 쟀다 · 시험이
단언한다). 그 입력은 편집 전(`26e5bb3f`)에도 이름 댄 거절이었다 — 끝의 수 대조(`git listed … but the judged diff has …`). 지금은 typechange 만이면 B1 · B2
(`… but the judged diff has more sections`), 뒤에 파일이 더 있으면 B3(`… names 'b/x.go' where git listed 'y.go'`)이 말한다. **거부하는 정상 입력**: `*.go` 를
심링크로 바꾸는 커밋 — 저장소 노출 0(`git ls-files -s | grep '^120000'` 의 `*.go` 0, 26e5bb3f).

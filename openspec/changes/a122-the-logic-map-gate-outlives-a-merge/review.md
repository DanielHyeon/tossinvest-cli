# a122 · Review

## Function Logic Map: not-applicable (proposal 시점)

**사유.** 이 change 가 편집하려는 함수는 Python 이다 —
`tools/logic-map/check_analysis.py` 의 `changed_existing_functions` 와 `resolve_base`,
그리고 그 CLI. 이 저장소의 FLM 산출물은 `tools/logic-map/extract_go_ast.go` 가 **Go**
함수에서 뽑고, `docs/WORKFLOW.md` 는 "Python AST 분석기는 Go 코드에 재사용하지 않으며
`tools/logic-map/extract_go_ast.go` 가 그 역할을 맡는다"고 적는다. Go 함수를 하나도
고치지 않는 change 에 Go AST 산출물을 요구할 자리가 없다.

**이 면제는 스스로 좁다.** 검사는 이 표식을 `required` 가 비었을 때만 받아들인다
(`check_analysis.py:618`). 즉 이 change 가 Go 파일을 건드리는 순간 표식은 아무것도
통과시키지 못하고 실제 FLM 을 요구한다. 그때 tasks 1.2 의 단서대로 해당 함수의 FLM 을
먼저 만든다.

**손으로 읽은 증거의 한계를 적어 둔다.** proposal 이 인용한 `changed_existing_functions`
의 동작(대상이 워킹트리라는 것, `target` 매개변수가 CLI 에 노출되지 않는다는 것)은
`check_analysis.py:111-140` 과 `main()` 을 읽어서 얻었다. Python 에는 이 저장소가 쓰는
AST 열거 도구가 없으므로 이것이 얻을 수 있는 증거이며, **볼 곳을 골랐을 수 있다.**
proposal 의 수치 주장은 그렇지 않다 — 아래는 실행한 명령이다.

## 문제의 측정 (2026-09-08, HEAD 62b35779)

```
for id in a074-critical-events-reach-the-operator \
          a077-screens-show-what-they-already-know \
          a079-operator-can-lift-a-quarantine; do
  python3 tools/logic-map/check_analysis.py --change "$id"
done
```

| change | 결과 |
|---|---|
| a074-critical-events-reach-the-operator | FAIL — missing evidence 316건 |
| a077-screens-show-what-they-already-know | FAIL — 318건 |
| a079-operator-can-lift-a-quarantine | FAIL — 317건 |

셋 다 이 세션이 손대지 않은 change 다. 같은 날 a075 · a076 도 같은 이유로 실패해
5단계를 사유를 적고 면제한 뒤 아카이브했다(각 review 의 §완료 게이트 5단계 면제).

## 아직 결정하지 않은 것

**착지 지점을 무엇으로 볼지가 이 change 의 전부다.** tasks 1.3 이 그것이고, 후보를
고르기 전에 각 후보가 **위조되는 방식**을 먼저 적기로 했다. 그 값이 5단계를 통과시키는
유일한 손잡이가 되므로, "있으면 통과"로 구현되면 이 change 는 게이트를 고치는 것이
아니라 여는 것이 된다. tasks 2.4 가 그 변이를 심는 자리다.

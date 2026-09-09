# 5단계 호출 사슬 (task 1.2)

CodeGraph 로 열거했다. 2026-09-09, HEAD `1687baac`.

## 사슬

| 자리 | 위치 | 하는 일 |
|---|---|---|
| 진입 | `tools/gate.sh:251` | `python3 tools/logic-map/check_analysis.py --change "$CHANGE_ID"` |
| CLI | `check_analysis.py:660` `main` | 인자는 `--change` 와 `--root` **둘뿐**이다 |
| 판정 | `check_analysis.py:587` `check` | 아래 둘을 부른다 |
| base 해소 | `check_analysis.py:259` `resolve_base` | `base-commit.txt` → `rev-parse` → `execution_baseline` 채택 |
| 요구 집합 | `check_analysis.py:111` `changed_existing_functions` | `git diff <base> [target] -- '*.go'` |

`codegraph_callers` 로 확인한 호출자는 각각 하나다 — `changed_existing_functions` 는
`check` 만, `resolve_base` 도 `check` 만, `check`(이 파일의 것)는 `main` 만.
파일 밖 사용자는 하나 더 있다: `tools/logic-map/execution_baseline.py:271` 이
`changed_existing_functions` 를 지연 import 한다(a063 전용 예외 경로).

## `check` 이 target 을 안 넘긴다

```python
required = changed_existing_functions(root, base)   # check_analysis.py:597
```

`changed_existing_functions` 의 세 번째 매개변수 `target` 은 기본값 `""` 이고,
`""` 이면 `git diff <base> -- '*.go'` 가 되어 **워킹트리**와 비교한다. `main` 이
그 값을 받을 인자를 만들지 않으므로 CLI 에서 바꿀 방법이 없다.

## `check` 이 아카이브를 안 본다

```python
change_dir = root / "openspec" / "changes" / change   # check_analysis.py:590
```

접미사도 아카이브도 보지 않는다. 반면 **빌린 증거** 경로는 본다 —
`resolve_referenced_change`(`check_analysis.py:231`)가 `archive/<YYYY-MM-DD>-<id>`
를 날짜 접두사를 벗겨 전부 일치시키고, 둘이면 fail-closed 한다. 커밋 `f6965ebb`
이 고친 것이 이쪽이다. `check` 의 `change_dir` 은 그 수리를 못 받았다.

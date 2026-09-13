# Function Logic Map: `_change_analysis_path`

`tools/logic-map/execution_baseline.py:74-82` · Python · **task 7.6 (리뷰 I7)** ·
분기 3 → 3 · 반환 1 → 1 · raise 2 → 2

## Inputs and invariants

기록에 **적힌** 증거 경로가 이 change 의 분석 디렉터리(`openspec/changes/<CHANGE>/analysis/`) 안인지
본다. 6.2 뒤로 `validate` 는 이 함수에 **옮기기 전** 자리를 넘긴다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L75 | If | `if not isinstance(value, str):` |
| B2 | L78 | BoolOp | `candidate.is_absolute() or '..' in candidate.parts or value == analysis.rstrip('/') or (not value.st` |
| B3 | L78 | If | `if candidate.is_absolute() or '..' in candidate.parts or value == analysis.rstrip('/') or (not value` |

## 편집 (task 7.6) — 분기·반환·raise 수 불변, 문장 하나

| | raise 문장 |
|---|---|
| 전 | `raise AdoptionError('adoption evidence path is outside current change analysis')` |
| 후 | `raise AdoptionError("adoption evidence path is outside this change's recorded analysis directory")` |

"current" 는 아카이브된 change 에서 **지금 자리**를 본 것처럼 읽혔다. 판정은 그대로다.

## Calls and live bindings

`isinstance` · `Path` · `str.startswith` · `str.rstrip`. git 없음.

## State mutations and fallbacks

없다.

## Safety conclusion

조건식(`candidate.is_absolute() or '..' in … or not value.startswith(analysis)`)은 열거에서 전후가 같다.

# Function Logic Map: `resolve_base`

`tools/logic-map/check_analysis.py:282-322`(편집 전 `52d5cb2c`) → `:282-326`(편집 후) · Python

> **이 표는 손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.before-1.12.json` ·
> `ast.after-1.12.json` 을 `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다.

## Inputs and invariants

`change_dir` · `root` · 선택적 `context` 를 받아 **유효한 비교 기준** 하나를 돌려준다.
a063 이관 예외면 `E`, 아니면 `base-commit.txt` 가 해소된 값이다.

불변식: 돌려주는 값은 **한 개**이고, `SDD_BASE_REF` 는 그 값을 확인만 할 수 있을 뿐
바꾸지 못한다(`B10·B11`).

## Branches and early returns — 편집 전 (분기 10 · 반환 2 · raise 4)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1·B2 | 286·288 | Try·ExceptHandler | `base-commit.txt` 읽기 실패 |
| B3 | 304 | If | `if process.returncode:` — rev-parse 실패 |
| B4 | 306 | BoolOp | 오류 문구 선택 |
| B5·B6 | 311·313 | Try·ExceptHandler | `AdoptionError` → `invalid execution-baseline adoption` |
| B7 | 315 | IfExp | `str(adoption['effective_base']) if adoption else persisted` |
| B8 | 316 | If | `if context is not None:` |
| B9·B10 | 320 | BoolOp·If | `SDD_BASE_REF` 확인 |

반환 둘: `308 process.stdout.strip()`(내부 `resolve`) · `322 effective`.

## 결함 (task 1.12)

`validate_execution_baseline` 은 `{"effective_base", "source", "ledger"}` 를 돌려준다.
**`B7` 은 그중 `effective_base` 하나만 읽는다.** 감사된 창의 **끝**인 `source` 는
그대로 버려진다 — 저장소 전수로 `adoption["source"]` 를 읽는 곳이 **0곳**이었다.

그래서 호출자 `check` 는 창의 끝을 `landed-commit.txt` 에서 찾고, 없으니 워킹트리를
대상으로 삼는다. 그 답이 오늘 맞는 이유는 `validate` 의 **다른** 판정
(`source-to-evidence drift`)이 source..head 에 Go 를 못 들어오게 하기 때문이다.
실측으로 그 우연은 a063 에서 required **9 대 36** 짜리다.

## 편집 — 분기 10 → 11, 반환 2 → 2, raise 4 → 4

| 새 ID | 줄 | 소스 |
|---|---|---|
| B9 | 319 | `if adoption:` → `context["adoption_source"] = str(adoption["source"])` |

경로는 하나도 안 지운다. 반환도 raise 도 그대로다 — 이 편집은 **이미 계산된 사실을
버리지 않는 것**뿐이다.

## Calls and live bindings

`validate_execution_baseline`(= `execution_baseline.validate`) · `subprocess.run(git
rev-parse)` · `os.environ.get`. `execution_baseline.py` 는 **한 줄도 안 바꾼다**.

## State mutations and fallbacks

`context` 사전에만 쓴다. fallback 없다 — `AdoptionError` 는 `ValueError` 로 감싸
호출자에게 올라가고, 조용히 계획 기준으로 되돌아가지 않는다(그 회귀 시험이
`test_real_adoption_sdd_base_ref_accepts_only_e_and_invalid_record_never_falls_back`).

## Safety conclusion

주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. 읽기
전용이고 실패 방향은 게이트가 **안 열리는** 쪽이다. Go 파일 변경 0.

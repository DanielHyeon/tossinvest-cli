# Function Logic Map: `test_index` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1559-1567` · 분기 3 · 반환 1 · raise 0 (편집 전 L1317-1325 · 분기 3 · 반환 1 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

트리 순회를 `_globbed` 로 — `*_test.go` 집합이 원장에 남는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1562 | For | `for path in _globbed(root, '*_test.go'):` |
| B2 | 1563 | If | `if '.git' in path.parts:` |
| B3 | 1565 | For | `for name, (start, end) in test_spans(path).items():` |

## task 7.5.9 — 시험 인용은 추적 파일로만 충족된다 (7.5.2.3 재리뷰 정확성, 2026-09-26)

> **이 표는 손으로 읽어서 만들지 않았다.** `enumerate.py` 로 편집 **전**(`ast.before-759.json`, revision `1d1e5ca7`)과
> 편집 **후**(`ast.after-759.json`)를 뽑고 `analysis/harness/759_flm_rows.py` 가 `difflib` 으로 정렬해 찍었다.
> 중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서는 `:2212-2223` 이다(뒤 task 의 편집이 앞에 줄을 더했다 — 이 함수는 다시 안 바뀌었다).

편집 후 `tools/logic-map/check_analysis.py:2177-2188` · 분기 3 · 반환 1 · raise 0 · 호출 9 (`ast.after-759.json`, source sha `194ec1a29941`)
편집 전 `tools/logic-map/check_analysis.py:2146-2154` · 분기 3 · 반환 1 · raise 0 · 호출 5 (`ast.before-759.json`, revision `1d1e5ca7`, source sha `b28398e26f3d`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | — | — | For | `for path in _globbed(root, '*_test.go'):` (옛) | **빠짐** |
| B2 | — | — | If | `if '.git' in path.parts:` (옛) | **빠짐** |
| — | B1 | 2184 | comprehension | `for path in _tracked(root) if path.name.endswith('_test.go')` | **새** |
| — | B2 | 2185 | For | `for path in sorted(index.files):` | **새** |
| B3 | B3 | 2186 | For | `for name, (start, end) in test_spans(path).items():` | 같음 |

**바뀐 것.** 순회가 디스크 전수(`_globbed(root, "*_test.go")` = `rglob`)에서 **추적 목록**(`_tracked(root)` = `git ls-files -z`)으로.
`.git` 거름(옛 B2)은 없어졌다 — git 은 `.git` 구성 요소가 든 경로를 추적하지 않는다. 반환은 `TestIndex`(사전 + `files`):
인용 해소가 **같은 목록**으로 고른다(`resolve_test_file` 의 `files`). 색인 순서는 `sorted(index.files)` 로 결정적이다.

**거부하는 정상 입력(편집 전에 셌다).** `analysis/harness/759_census.py`(HEAD `1d1e5ca7`): 추적 `*_test.go` **959** · 디스크 **971** ·
디스크에만 **12**(전부 `analysis/harness/_work/**/extract_go_ast_test.go`(10 은 한 단계, 2 는 `75_mut_work.*/logic-map/` 아래) — gitignore 된 하네스 사본) · 추적에만 0. 저장소 번들 **3,084** 의
인용 **6,121** 을 편집 전 코드 · 추적 거름을 입힌 편집 전 코드 · 편집 뒤 코드 셋으로 돌려 **다른 번들 0**(문장 꼬리 정규화). 시험 픽스처에서는
**다섯**이 걸렸다: `TestNamedTestsAreOpened` 의 `…exists_is_accepted` · `…past_the_end…` · `…doc_comment…` · `…shared_harness…` ·
`…named_in_one_file…` — `run_check` 가 시험 파일을 **인덱스에 안 올린** 저장소를 썼다. `run_check` 가 `git add -A` 하도록 고쳤다
(`untracked=` 로 뺄 경로를 받는다).

**남는 것.** `test_spans` 는 못 읽는 추적 파일을 `{}` 로 건너뛴다(편집 전부터) — 그 파일의 시험은 "추적 파일 어디에도 없다" 로
**거절**된다(보수적이지만 사유가 틀린다). 이 로트 범위 밖이라 안 건드렸다.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/check_analysis.py:2251-2264` · 분기 3 · 반환 1 · raise 0 · 호출 9 · source sha `ff590b79db6e` (비교 기준 `ast.after-759.json`, 그 판 `194ec1a29941` `:2177-2188`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.

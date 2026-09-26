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


---

## 편집 계획 — task 6.2 (코드보다 먼저, `ast.before-6.2.json` = HEAD `fb4e8f92`)

분기 11 · 반환 2 · raise 4 는 **그대로** 여야 한다. 바뀌는 것은 호출 하나의 인자다.

- `resolve_base(change_dir, root, context=None, *, change_id)` — 게이트가 요청받은 id 를
  받아 `validate_execution_baseline(change_dir, root, persisted, change_id)` 에 넘긴다.
  `validate` 가 신원을 디렉터리 이름이 아니라 이 id 로 가르게 하려는 것이다(아카이브는
  이름을 `<YYYY-MM-DD>-<id>` 로 바꾼다).
- 호출자 셋(`check` 의 게이트 대상 · `check` 의 빌린 증거 · `record_landing`)은 각자가
  해소한 id 를 넘긴다. 셋 다 그 id 로 `resolve_referenced_change` 를 부른 직후다.
- 키워드 전용 필수로 둔다. 기본값을 두면 id 를 안 넘기는 호출자가 조용히 옛 판정(이름)
  으로 떨어진다.


## 편집 결과 — task 6.2 (`ast.after-6.2.json`)

분기 11 → 11 · 반환 2 → 2 · raise 4 → 4, 열거 diff 에 갈래 변화 **0**. 바뀐 것은 시그니처
(키워드 전용 필수 `change_id`)와 `validate_execution_baseline` 에 넘기는 인자 하나다.
변이 M6(인자를 `change_dir.name` 으로 되돌림) CAUGHT — 아카이브 시험이 빨개진다.

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:304-356` · 분기 12 · 반환 2 · raise 5 (편집 전 L293-341 · 분기 11 · 반환 2 · raise 4, `ast.before-7.5.2.3.json` = revision `1d12520c`).

창의 **시작**을 정하는 글자를 깔때기로 읽는다(옛 `read_text` 는 FIFO 에 멎었다). 없는 것(`missing base-commit.txt` + 복구 명령)과 못 읽는 것을 가른다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 309 | Try | `try:` |
| B2 | 313 | ExceptHandler | `except FileNotFoundError as exc:` |
| B3 | 319 | ExceptHandler | `except (OSError, UnicodeDecodeError) as exc:` |
| B4 | 331 | If | `if process.returncode:` |
| B5 | 333 | BoolOp | `process.stderr.strip() or f'invalid Function Logic Map base: {value}'` |
| B6 | 338 | Try | `try:` |
| B7 | 343 | ExceptHandler | `except AdoptionError as exc:` |
| B8 | 345 | IfExp | `str(adoption['effective_base']) if adoption else persisted` |
| B9 | 346 | If | `if context is not None:` |
| B10 | 349 | If | `if adoption:` |
| B11 | 354 | BoolOp | `override and resolve(override) != effective` |
| B12 | 354 | If | `if override and resolve(override) != effective:` |

## task 6.4(b) — 창의 시작도 끝처럼 잠근다 (2026-09-26)

`tools/logic-map/check_analysis.py:817-897` · 분기 16 · 반환 2 · raise 8 (편집 전 L813-865 · 분기 12 · 반환 2 · raise 5,
`ast.before-6.4.json` = revision `0023fd12`; 편집 후 `ast.after-6.4.json` = 워킹트리). 열거 diff 에서 **더해진 것만** 있고
지워진 갈래 · 반환 · raise 는 0 이다.

| 새 id | 줄 | 종류 | 소스 | 뜻 |
|---|---|---|---|---|
| B4 | 845 | If | `if not FULL_SHA.fullmatch(candidate):` | 이름 · 짧은 id · 대문자 id 거절 |
| B5·B6 | 851 | BoolOp·If | `if committed is not None and committed.decode(...).strip() != candidate:` | HEAD 에 있는 base 와 디스크가 다르면 거절, HEAD 에 없으면 디스크 그대로 |
| B9 | 876 | If | `if persisted != candidate:` | `^{commit}` 이 벗긴 값이 적힌 값이 아니면(태그 객체 id) 거절 |

옛 B4~B12 는 번호만 밀렸다(B7~B8 · B10~B16). 새 raise 셋(855 · 848 · 878)은 전부 `ValueError` 라 호출자 셋의 `GATE_FAULTS`
경계 안이다 — 판정 줄 없이 끝나는 경로가 없다.

**새 입력 `head`** — 키워드 전용 필수. 호출자 셋(`_judged` 의 대상 · `_judged` 의 빌린 증거 · `_recording_refusal`)이 각자 한 번
푼 sha 를 넘긴다. `_committed_bytes(root, head, relative)` 하나가 새 git 호출이다(`cat-file --batch -Z` 한 프로세스).

**거절하는 정상 입력(편집 전에 셌다).** 저장소의 `base-commit.txt` 119 개(활성 · 아카이브) 전부가 40자리 소문자 + LF,
자기 자신이 커밋 id, HEAD 의 것과 바이트가 같다 → 거절 0(측정 순간을 안 적었다 — 보수가 HEAD `5a54f78d` 에서 재측정, 같은 답). 거절하는 **모양**은 넷(이름 · 짧은 id/대문자 · 태그 객체 id ·
커밋된 값과 다른 커밋 안 한 편집). 시험 픽스처에서 셌더니 **둘**이 모양에 걸렸다: `test_real_archived_reference_still_requires_the_same_base`
(커밋된 base 를 커밋 없이 바꿔 "공유 base" 거절을 재던 것 — 편집을 커밋하게 고쳤다) · `test_environment_base_cannot_override_persisted_change_base`
(`persisted` 라는 글자 — 모양 검사에서 먼저 멈춰 **다른 이유로** 통과하고 있었다, 40자리로 바꾸고 문장까지 본다).

## task 6.4 보수 — 이동이 대조를 벗지 못한다 (독립 리뷰 둘, 2026-09-26)

`tools/logic-map/check_analysis.py:845-942` · 분기 18 · 반환 2 · raise 9 (편집 전 `ast.before-64r.json` = 워킹트리 판
`a88c966b`(source sha) — 편집 전에 `ast.after-6.4.json` 과 sha 가 같음을 확인해 그대로 옮겼다 · 분기 16 · 반환 2 · raise 8;
편집 후 `ast.after-64r.json` = 워킹트리 `b28398e2`). 짝은 옛/새 열거를 `difflib.SequenceMatcher` 로 정렬해서 지었다(손 재번호 아님).

| 옛 id | 새 id | 줄 | 종류 | 소스 | 바뀐 것 |
|---|---|---|---|---|---|
| B1~B4 | B1~B4 | 869~881 | | 읽기 · 모양 | 같음 |
| B5 · B6 | — | | BoolOp · If | `if committed is not None and … != candidate:` | **대체** |
| — | B5 | 888 | IfExp | `[(relative, committed)] if committed is not None else _committed_elsewhere(…)` | 지금 경로가 HEAD 에 없으면 같은 id 의 다른 자리 |
| — | B6 | 889 | For | `for place, value in held:` | 커밋된 자리마다 대조 |
| — | B7 | 891 | If | `if shown == candidate: continue` | 같으면 받는다 |
| — | B8 | 895 | If | `if place == relative:` | 제자리 편집 문장 · 아니면 이동 문장 |
| B7~B16 | B9~B18 | 914~940 | | rev-parse · 태그 · 이관 · `SDD_BASE_REF` | 번호만 +2 |

raise 8 → 9: 제자리 편집(896, 문장을 좁혔다 — P1-C)과 이동(900, 새). 둘 다 `ValueError` 라 호출자 셋의 `GATE_FAULTS` 경계 안이다.
새 호출 `_committed_elsewhere`(새 잎 함수, 열거 `../tools-logic-map--_committed_elsewhere/ast.json` · 분기 5 · 반환 1 · raise 1)는
`git ls-tree -z -d --name-only <head> -- openspec/changes/archive/` 한 번 + `_committed_many` 한 번이다. 지금 경로가 HEAD 에 있으면 안 불린다.

**결함 (P0-A).** 6.4(b) 의 옛 B5 는 대조 경로를 **지금 디스크 경로**로 계산했다. 디렉터리를 옮기면 그 경로에 blob 이 없어
`committed is None` → "HEAD 에 없는 새 base" 로 받았다. 리뷰어 재현 셋(커밋 안 한 `mv` · staged `git mv` · 아카이브 날짜 변경)이
전부 `[]`. 이 세션의 RED(편집 전 코드, `AMovedChangeKeepsItsCommittedBase`): 옮긴 네 모양(셋 + 아카이브 → 활성) 전부 `[]`.

**거절하는 정상 입력(편집 전에 셌다).** `analysis/harness/64r_census.py`, HEAD `5a54f78d`: change 디렉터리 128 · base 있는 119
전부 자기 경로로 HEAD 에 있고 바이트가 같다 → 새 갈래(B5 의 else)에 닿는 것 **0**, 거절 0. 같은 id 가 HEAD 에 두 자리 이상 0.
시험 픽스처에서 하나가 닿았다: `test_environment_base_cannot_override_persisted_change_base` — `subprocess.run` 을 차례표 mock 으로
막는 시험이라 새 `ls-tree` 호출이 `persisted` 응답을 먹었다(ERROR). 그 시험이 재는 모양(HEAD 에 없는 새 base)에 맞게
`_committed_elsewhere` 도 "없음" 으로 막았다.

**보장 범위 (P1-C).** 막는 것은 커밋 안 한 편집과 이동이다. base 를 고쳐 **커밋**하는 재기록은 여전히 받는다 — 역사에 묶을지는
열린 사람 결정이다(tasks 6.4). 옛 docstring 의 "시작도 끝처럼 잠근다" · "spec 이 불변(SHALL)" 과 옛 오류 문장 "the comparison base is
frozen at proposal freeze" 는 그 이상을 약속하는 말이라 좁혔다. 끝(`landed-commit.txt`)은 도구가 계산하고 게이트가 같은 계산으로
대조하므로 정말 저자가 고를 수 없다 — 그 대비를 docstring 에 남겼다.

**남는 모양(고치지 않음, tasks 6.4 (i)(k)).** 심링크 change 디렉터리 · `root` ≠ git toplevel 이면 경로 기준이 어긋나 B5 의 두 읽기가
조용히 "없음" 이 된다. id 를 바꿔 옮기면(다른 id 의 새 change) 대조할 자리가 없다 — 그것은 새 change 와 같은 모양이다.

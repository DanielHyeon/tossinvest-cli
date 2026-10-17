# Function Logic Map: `_archived_change_id`

`tools/logic-map/check_analysis.py:232-241` · Python · **task 7.6 (리뷰 I6)** ·
분기 1 · 반환 1 (`ast.after-7.6.json`)

## Inputs and invariants

아카이브 디렉터리 **이름** 하나. `<YYYY-MM-DD>-<id>` 전체가 맞을 때만 `<id>` 를 준다
(`ARCHIVED_CHANGE.fullmatch`). 접미사로 고르지 않는다 — `2026-08-29-other-reference` 는
`other-reference` 이지 `reference` 가 아니다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L241 | IfExp | `matched.group('change') if matched else ''` |

## Calls and live bindings

`ARCHIVED_CHANGE.fullmatch` · `matched.group`. 부르는 쪽: `resolve_referenced_change`(id → 디렉터리)와
`_pre_archive_path`(아카이브 경로 → 옮기기 전 경로). 편집 전에는 둘이 각자 정규식을 불렀다.

## State mutations and fallbacks

없다.

## Safety conclusion

해독 규칙 자체는 안 바뀌었다 — 정규식이 같고 두 호출자가 그것을 부르는 방식(`fullmatch` +
`group("change")`)도 같다. 바뀐 것은 그 규칙이 **사는 자리의 수**다(2 → 1). shell 쪽 사본
(`tools/gate.sh`)은 그대로다.

## task 6.4(c) — 날짜 숫자는 ASCII 뿐 (2026-09-26)

함수 본문은 한 글자도 안 바뀌었다 — `ast.before-6.4.json`(revision `0023fd12`) 과 `ast.after-6.4.json` 의 분기 1 · 반환 1 · 호출 2 가
같다(줄만 L754-763 → L757-767). 바뀐 것은 이 함수가 부르는 모듈 상수 `ARCHIVED_CHANGE` 의 깃발(`re.ASCII`)이다. 글자 패턴의 `\d` 는
유니코드 숫자(전각 `２` · 아라비아-인도 `٢` · 수학 굵은 `𝟐`)를 먹었다. 편집 전에는 월 `０９` 인 이름의 아카이브를 Python 은 찾고 shell 은
못 찾았다(RED 기록: 표 두 행 · 해독기 세 행). 저장소의 아카이브 이름 중 ASCII 가 아닌 숫자를 가진 것은 0 이다 — 거절하는 정상 입력 0.
shell 사본과 같이 묶는 것은 이제 문장이 아니라 `TheTwoChangeResolversAgree` 다(6.4(f)).

> **정정 (6.4 보수, 주장정확성 리뷰 P0-F1).** 첫 판은 여기 "`tools/gate.sh` 의 `[0-9]` 는 안 먹는다(`en_US.UTF-8`/`C.UTF-8` 둘 다
> 실측)" 라고 적었다. **거짓이었다.** bash 5.2.21 실측(이 세션, 한 글자씩): `en_US.UTF-8` 에서 범위 `[0-9]` 는 콜레이션으로
> `０…８` · `٠…٨` · `𝟎…𝟖` 를 먹고 각 벌의 `９` 만 안 먹는다. `C.UTF-8` · `ko_KR.utf8` 에서는 셋 다 안 먹는다. `[[:digit:]]` 와 나열
> `[0123456789]` 는 세 로케일 모두 안 먹는다. 첫 판 표의 유니코드 행 둘은 월 `０９` 의 `９` 덕에 **우연히** 떨어졌다 — 9 없는 날짜
> (`２０２６-０８-１１`)면 `en_US.UTF-8` 에서 Python=missing · shell=found 로 갈렸다(RED, 아래). `gate.sh` 도 `Makefile` 도 로케일을
> 고정하지 않으므로 사람의 셸 로케일이 게이트의 로케일이다. 수리: `gate.sh` 를 나열 `[0123456789]` 로(로케일과 무관), 표에 9 없는
> 날짜 세 행, shell 을 `C.UTF-8` 과 `en_US.UTF-8` 둘에서 돈다. RED(수리 전 `gate.sh`): 새 세 행이 `en_US.UTF-8` 에서만 빨갛다
> (`('missing', 'found:…')`). 이 함수의 본문과 AST 는 이 정정으로 안 바뀌었다.

변이 AL1(`re.ASCII` 삭제) CAUGHT — `test_one_table_through_both_resolvers` · `test_the_archive_name_reader_takes_ascii_digits_only`.
편집 전 스위트(377)는 이 변이와 같은 코드에서 초록이었다 — 유니코드 숫자를 재는 시험이 0 이었다.

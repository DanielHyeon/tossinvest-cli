# Function Logic Map: `_committed_many`

`tools/logic-map/check_analysis.py:379-463` · Python · **task 7.5 (리뷰 I1 · 2026-09-18 독립 리뷰 P0)** ·
분기 12 · 반환 3 · raise 3 (`ast.after-7.5.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** `enumerate.py` 의 열거를 그대로 옮긴 것이다.
> 7.5 가 새로 만든 함수라 편집 전 판본이 없다.

## Inputs and invariants

`(root, ref, relatives)` — 한 커밋에서 여러 경로의 blob 을 **한 프로세스**로 읽어
`{경로: bytes | None}` 로 돌려준다. 키 집합은 중복을 없앤 `relatives` 와 **정확히 같다**.

존재 이유는 알고리즘이 아니라 **fetch 단위**다 (2026-09-18 실측: 아무 후보도 안 받는
walk 에서 spawn 의 **97.1~97.3%** 가 blob fetch). 후보당 83.4ms 가 8.5ms 가 된다.

`-Z` 는 입력과 출력을 **둘 다** NUL 로 끊는다. **git 2.42+ 필요**(`GIT_BATCH_MINIMUM`).

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 401 | comprehension | `for relative in wanted` (전부 `None` 으로 시작 — 사전 채움) |
| B2 | 402 | If | `if not wanted:` |
| B3 | 404 | For | `for relative in wanted:` (요청 검사) |
| B4 | 405 | If | `if '\x00' in relative:` |
| B5 | 415 | comprehension | `for relative in wanted` (NUL 로 끊은 입력) |
| B6 | 419 | If | `if process.returncode:` |
| B7 | 432 | For | `for relative in wanted:` (응답 파싱) |
| B8 | 434 | If | `if end < 0:` — 응답이 요청보다 짧다 |
| B9 | 449 | BoolOp | `len(fields) != 3 or not fields[2].isdigit()` |
| B10 | 449 | If | 같은 줄의 `if` — `missing`·`ambiguous` 는 내용이 안 따라온다 |
| B11 | 452 | If | `if fields[1] == b'blob':` |
| B12 | 455 | If | `if position != len(data):` — 남은 바이트 |

| 반환 줄 | 값 |
|---|---|
| 403 | `found` — 요청이 비었다 (git 을 안 부른다) |
| 431 | `found` — 프로세스 실패, 전부 `None` |
| 463 | `found` — 읽은 것 |

| raise 줄 | 무엇을 못 쟀나 |
|---|---|
| 408 | 경로에 NUL — 프레이밍 문자 자체다. 물어보면 레코드가 쪼개진다 |
| 440 | 응답이 잘렸다 — **부분 답을 쓰지 않는다** |
| 458 | 남은 바이트 — 프레이밍을 잘못 읽었다 |

## Calls and live bindings

`subprocess.run` ×1 (`git cat-file --batch -Z`) · `dict.fromkeys` · `bytes.join` ·
`data.find` · `header.rsplit` · `int`. 워킹트리·시각·환경을 안 읽는다.

## State mutations and fallbacks

`answered` 를 다 채운 **뒤에** `found` 에 한 번 합친다 — 중간에 쓰면 부분 답이 새어 나간다.
fallback 은 **하나**: 프로세스가 실패하면 전부 `None` (옛 `git show` 의 rc≠0 과 같은 방향).

## Safety conclusion

생산 Go 코드 변경 0. 판정을 바꾸지 않는다 — 옛 `git show` 판본과 **같은 바이트**를 낸다
(실물 전수 A/B 116/116 SAME, 두 번).

엄해지는 자리가 셋이다. **디렉터리**를 물으면 `None`(옛 `git show` 는 트리 목록 바이트를
줬다 — 파일 내용이 아닌데 판정에 들어오면 내용인 척한다). **경로에 NUL** 이면 거절.
**응답을 끝까지 못 읽으면** 결함이다. 셋 다 오늘 실물 입력에서 도달 0 건이다
([[fail-closed-must-name-what-it-rejects]]).

> **`end < 0` 은 비용 선택이 아니라 판정이다 (2026-09-18 독립 리뷰 P0).** 첫 판은 여기서
> `break` 하고 "나머지는 `None` 이라 답이 같다"고 적었다. **거짓이었다**: 마지막 머리가
> 멀쩡하고 내용만 잘린 응답에서 그 자리는 `b""` 가 되고, `None` 과 `b""` 는
> `_unheld_bundles` 의 아카이브 대체 갈래를 여닫아 **판정을 바꾼다**(재현으로 확인).
> 변이가 살아남은 것은 동등해서가 아니라 **이 갈래에 닿는 시험이 없어서**였다
> ([[mutation-must-reach-the-thing-under-test]]).

`rc≠0` 을 결함으로 올리려다 되돌렸다 — 거부할 정상 입력을 세어 보니 **저장소가 아닌 루트**가
이 함수의 정상 호출 모양이었다(시험 21개). 최소 git 버전은 `GIT_BATCH_MINIMUM` 에 적고
README·WORKFLOW 가 그 수를 인용한다.

## task 7.5.1 — "못 물었다" 는 `None` 이 아니다 (gstack 리뷰 2026-09-19)

`tools/logic-map/check_analysis.py:387-496` · 분기 17 · 반환 2 · raise 7 (편집 전 분기 12 · 반환 3 · raise 3, `ast.before-7.5.1.json` = revision `b29e1f4e`).

**위 Safety conclusion 의 "프로세스가 실패하면 전부 `None` — 부르는 쪽은 `None` 을 불일치로 센다"
는 거짓이었다.** 두 출처의 적대 리뷰가 두 자리에서 깼다 — 가드 7(한쪽 실패 → `None != bytes` →
"바뀌었다" → 편집 전 커밋 통과, **재현**)과 `_landing_record`("기록 없음" → 착지 검증 생략).
이제 `None` 은 딱 하나 — git 이 **물은 spec 그대로** `missing` 이라고 답했다(또는 blob 이 아니다).

git 이 실제로 내는 머리를 먼저 쟀다(blob · tree · 심링크 · **gitlink** · 없는 경로 · HEAD · 짧은
ref) — 정확히 두 모양 `<hex oid> <type> <size>` 와 `<물은 spec> missing` 이다. 엄격한 파서가
거부할 정상 입력은 0 이고, `check()` 전체 A/B **126/126 SAME** 이 그것을 확인한다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 415 | comprehension | ` for relative in wanted` |
| B2 | 416 | If | `if not wanted:` |
| B3 | 418 | comprehension | ` for relative in wanted` |
| B4 | 419 | For | `for spec in specs:` |
| B5 | 420 | If | `if '\x00' in spec:` |
| B6 | 427 | comprehension | ` for spec in specs` |
| B7 | 430 | comprehension | ` for spec in asked` |
| B8 | 433 | If | `if process.returncode:` |
| B9 | 446 | IfExp | `f': {said[0][:160]}' if said else ''` |
| B10 | 451 | For | `for index, (relative, spec) in enumerate(zip(wanted, asked)):` |
| B11 | 453 | If | `if end < 0:` |
| B12 | 462 | If | `if header == spec + b' missing':` |
| B13 | 465 | If | `if match is None:` |
| B14 | 473 | If | `if position + size >= len(data):` |
| B15 | 480 | If | `if data[position + size] != 0:` |
| B16 | 486 | If | `if match.group('type') == b'blob':` |
| B17 | 489 | If | `if position != len(data):` |

| raise 줄 | 소스 |
|---|---|
| 423 | `raise RuntimeError(f'request contains a NUL byte, cannot be asked for: {spec!r}')` |
| 444 | `raise RuntimeError(f'cannot read blobs at {ref[:12]}: `git cat-file --batch -Z` failed (rc` |
| 456 | `raise RuntimeError(f'`git cat-file --batch -Z` answered {index} of {len(wanted)} request(s` |
| 468 | `raise RuntimeError(f'`git cat-file --batch -Z` gave an unrecognised answer to request {ind` |
| 476 | `raise RuntimeError(f'`git cat-file --batch -Z` answered {index} of {len(wanted)} request(s` |
| 482 | `raise RuntimeError(f'`git cat-file --batch -Z` payload {index + 1} of {len(wanted)} at {re` |
| 491 | `raise RuntimeError(f'`git cat-file --batch -Z` left {len(data) - position} byte(s) unread ` |

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:408-517` · 분기 19 · 반환 2 · raise 7 (편집 전 L387-496 · 분기 17 · 반환 2 · raise 7, `ast.before-7.5.2.json` = revision `e9f905bd`).

새 git 의 `<oid> submodule`(커밋 없는 gitlink)을 blob 아님(`None`)으로 받는다(영수증: `builtin/cat-file.c` 의 `report_object_status(opt, NULL, &data->oid, "submodule")`). git 버전 조언은 rc 129(git 의 사용법 오류 — 이 기계에서 `unknown switch` 로 실측)일 때만 붙인다 — 저장소가 아닌 루트(rc 128)에 붙이던 것은 git 의 말 옆에 추측한 진단이었다. 파서 오류 둘이 물은 경로를 이름으로 댄다. 잘림 문장 두 벌을 `_TRUNCATED` 한 벌로. 파싱 루프의 변수를 `spec`→`request`(인코딩된 요청)로. docstring 의 `None` 계약을 고쳤다: 트리 안인데 못 읽는 객체도 git 은 rc 0 `missing` 이라 답한다(재리뷰 보안 전문가 실측) · 못 물었다 = `GATE_FAULTS` 안의 예외 · `_landing_record` 는 permissive 가 아니라 가장 넓은 창으로 갔다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 440 | comprehension | ` for relative in wanted` |
| B2 | 441 | If | `if not wanted:` |
| B3 | 443 | comprehension | ` for relative in wanted` |
| B4 | 444 | For | `for spec in specs:` |
| B5 | 445 | If | `if '\x00' in spec:` |
| B6 | 452 | comprehension | ` for spec in specs` |
| B7 | 455 | comprehension | ` for spec in asked` |
| B8 | 458 | If | `if process.returncode:` |
| B9 | 473 | IfExp | `f': {said[0][:160]}' if said else ''` |
| B10 | 474 | IfExp | `f' — `-Z` needs git {GIT_BATCH_MINIMUM} or newer' if process.returncode == 129 else ''` |
| B11 | 478 | For | `for index, (relative, request) in enumerate(zip(wanted, asked)):` |
| B12 | 480 | If | `if end < 0:` |
| B13 | 485 | BoolOp | `header == request + b' missing' or _SUBMODULE_HEADER.fullmatch(header)` |
| B14 | 485 | If | `if header == request + b' missing' or _SUBMODULE_HEADER.fullmatch(header):` |
| B15 | 488 | If | `if match is None:` |
| B16 | 496 | If | `if position + size >= len(data):` |
| B17 | 501 | If | `if data[position + size] != 0:` |
| B18 | 507 | If | `if match.group('type') == b'blob':` |
| B19 | 510 | If | `if position != len(data):` |

| raise 줄 | 소스 |
|---|---|
| 448 | `raise RuntimeError(f'request contains a NUL byte, cannot be asked for: {spec!r}')` |
| 471 | `raise RuntimeError(f'cannot read blobs at {ref[:12]}: `git cat-file --batch -Z` failed (rc` |
| 482 | `raise RuntimeError(_TRUNCATED.format(count=index, total=len(wanted), ref=ref[:12]))` |
| 491 | `raise RuntimeError(f'`git cat-file --batch -Z` gave an unrecognised answer to request {ind` |
| 500 | `raise RuntimeError(_TRUNCATED.format(count=index, total=len(wanted), ref=ref[:12]))` |
| 503 | `raise RuntimeError(f'`git cat-file --batch -Z` payload {index + 1} of {len(wanted)} ({rela` |
| 512 | `raise RuntimeError(f'`git cat-file --batch -Z` left {len(data) - position} byte(s) unread ` |

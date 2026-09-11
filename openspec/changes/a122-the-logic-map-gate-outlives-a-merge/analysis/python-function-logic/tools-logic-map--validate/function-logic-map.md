# Function Logic Map: `validate`

`tools/logic-map/execution_baseline.py:393-487`(편집 전) · Python ·
기존 함수 (task 6.2). `ast.before-6.2.json` = HEAD `fb4e8f92` 의 blob,
`source_sha256` `e9d8a55c718f…`.

> **이 표는 손으로 읽어서 만들지 않았다.** `enumerate.py` 가 기계로 열거한 것을 스크립트가
> 옮겼다. 그리고 이번에는 **편집 전에** 뽑았다 — 6.1.2 가 어긴 순서다.

## Inputs and invariants

`change_dir` · `root` · `persisted`(=`base-commit.txt` 를 해소한 값). a063 하나를 위한
고정 이관 예외를 감사한다. 반환은 `None`(기록 없음) 또는
`{"effective_base": E, "source": S, "ledger": …}`, 그 밖은 전부 `AdoptionError`.

불변식: 판정에 들어가는 입력은 **빠짐없이 열거되고 digest 로 묶인다**. 이관 예외의
정당성이 이것 하나다.

## Branches and early returns (ast 열거 그대로, 분기 42)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 395 | Try | `try:` |
| B2 | 397 | ExceptHandler | `except FileNotFoundError:` |
| B3 | 399 | If | `if not stat.S_ISREG(record_mode):` |
| B4 | 401 | BoolOp | `change_dir.name != CHANGE or persisted != P` |
| B5 | 401 | If | `if change_dir.name != CHANGE or persisted != P:` |
| B6 | 403 | If | `if subprocess.run(['git', 'symbolic-ref', '-q', 'HEAD'], cwd=root, capture_output=True, ch` |
| B7 | 407 | If | `if _regular_committed(root, base_relative, head) != (P + '\n').encode('ascii'):` |
| B8 | 411 | If | `if set(record) != required:` |
| B9 | 415 | If | `if not all((isinstance(record[key], str) for key in required - {'schema'})):` |
| B10 | 415 | comprehension | ` for key in required - {'schema'}` |
| B11 | 417 | BoolOp | `record['change'] != CHANGE or record['planning_base'] != P or record['execution_base'] != ` |
| B12 | 417 | If | `if record['change'] != CHANGE or record['planning_base'] != P or record['execution_base'] ` |
| B13 | 419 | If | `if record['inherited_history_disposition'] != 'committed historical work; missing original` |
| B14 | 421 | For | `for field in ('ledger_path', 'adversarial_review_path', 'gstack_review_path'):` |
| B15 | 423 | For | `for field in ('ledger_sha256', 'adversarial_review_sha256', 'gstack_review_sha256'):` |
| B16 | 425 | If | `if record['adversarial_review_path'] == record['gstack_review_path']:` |
| B17 | 428 | BoolOp | `not isinstance(record['source_tree'], str) or len(record['source_tree']) != 40 or any((cha` |
| B18 | 428 | If | `if not isinstance(record['source_tree'], str) or len(record['source_tree']) != 40 or any((` |
| B19 | 428 | comprehension | ` for char in record['source_tree']` |
| B20 | 433 | If | `if _git(root, 'rev-parse', source + '^{tree}').stdout.strip() != record['source_tree']:` |
| B21 | 436 | For | `for args in (('diff', '--quiet'), ('diff', '--cached', '--quiet')):` |
| B22 | 437 | If | `if subprocess.run(['git', *args], cwd=root, capture_output=True, check=False).returncode:` |
| B23 | 439 | For | `for path in nul_paths(root, 'diff', '--name-only', source, head):` |
| B24 | 440 | BoolOp | `path.startswith('openspec/') or path.startswith('docs/pm/')` |
| B25 | 440 | If | `if not (path.startswith('openspec/') or path.startswith('docs/pm/')):` |
| B26 | 443 | For | `for path in sorted(candidates):` |
| B27 | 445 | BoolOp | `not _allowed_metadata(path) or not stat.S_ISREG(item.lstat().st_mode) or os.access(item, o` |
| B28 | 445 | If | `if not _allowed_metadata(path) or not stat.S_ISREG(item.lstat().st_mode) or os.access(item` |
| B29 | 450 | If | `if overlap:` |
| B30 | 453 | If | `if hashlib.sha256(ledger_raw).hexdigest() != record['ledger_sha256']:` |
| B31 | 455 | For | `for name in ('adversarial_review', 'gstack_review'):` |
| B32 | 456 | If | `if hashlib.sha256(_regular_committed(root, record[name + '_path'], head)).hexdigest() != r` |
| B33 | 460 | If | `if set(ledger) != ledger_required:` |
| B34 | 463 | If | `if ledger['inherited_history_disposition'] != 'committed historical work; missing original` |
| B35 | 465 | BoolOp | `not isinstance(ledger['planning_to_execution'], dict) or not isinstance(ledger['execution_` |
| B36 | 465 | If | `if not isinstance(ledger['planning_to_execution'], dict) or not isinstance(ledger['executi` |
| B37 | 468 | If | `if ledger['planning_to_execution'] != expected_planning:` |
| B38 | 471 | comprehension | ` for path in nul_paths(root, 'diff', '--name-only', E, source) if path.endswith('.go')` |
| B39 | 472 | If | `if recorded_paths != expected_paths:` |
| B40 | 474 | If | `if any((path not in {'cmd/tossctl/soak.go', 'cmd/tossctl/soak_test.go', 'internal/soak/att` |
| B41 | 474 | comprehension | ` for path in expected_paths` |
| B42 | 483 | If | `if ledger.get('execution_to_source') != expected_execution:` |

반환 2: `None`(기록 없음) · 감사된 사전. raise 24:

| 줄 | 소스 |
|---|---|
| 400 | `raise AdoptionError('execution-baseline record is not a regular file')` |
| 402 | `raise AdoptionError('execution-baseline adoption is not allowed for this change/base')` |
| 404 | `raise AdoptionError('adoption requires detached HEAD')` |
| 408 | `raise AdoptionError('planning base file is substituted')` |
| 412 | `raise AdoptionError('invalid execution-baseline record')` |
| 416 | `raise AdoptionError('execution-baseline record field types are invalid')` |
| 418 | `raise AdoptionError('invalid execution-baseline record')` |
| 420 | `raise AdoptionError('inherited history disposition is invalid')` |
| 426 | `raise AdoptionError('adoption review paths must be distinct')` |
| 429 | `raise AdoptionError('source tree must be a full lowercase SHA-1')` |
| 434 | `raise AdoptionError('source tree mismatch')` |
| 438 | `raise AdoptionError('adoption requires a clean worktree')` |
| 441 | `raise AdoptionError(f'source-to-evidence drift: {path}')` |
| 446 | `raise AdoptionError(f'untracked/ignored input is not allowed: {path}')` |
| 451 | `raise AdoptionError(f'untracked/ignored Go build input: {overlap[0]}')` |
| 454 | `raise AdoptionError('ledger digest mismatch')` |
| 457 | `raise AdoptionError(f'{name} digest mismatch')` |
| 461 | `raise AdoptionError('invalid execution-baseline ledger')` |
| 464 | `raise AdoptionError('ledger inherited history disposition is invalid')` |
| 466 | `raise AdoptionError('execution-baseline ledger range types are invalid')` |
| 469 | `raise AdoptionError('planning history mismatch')` |
| 473 | `raise AdoptionError('execution-to-source Go path inventory mismatch')` |
| 481 | `raise AdoptionError('execution-to-source Go path is outside the fixed allowlist')` |
| 484 | `raise AdoptionError('execution-to-source function inventory mismatch')` |

## 이 편집이 닿는 자리 — 아카이브 뒤 막히는 넷 (task 6.2, 실측)

`62_chain.py` 가 저장소 픽스처(`_adoption_with_complete_bundle`)를 `git mv` 로
`archive/2026-09-11-<id>/` 에 옮기고, `validate` 의 **메모리 사본**에서 가드를 하나씩
풀어 다음 실패를 쟀다. 생산 코드는 안 건드렸다.

| 푼 것 | 다음 실패 |
|---|---|
| (없음) | `execution-baseline adoption is not allowed for this change/base` — **B5** |
| 이름 | `adoption evidence path is outside current change analysis` — **B14** 의 `_change_analysis_path` |
| + 접두사 | `missing evidence: openspec/changes/a063-…/analysis/execution-baseline-ledger.json` — `L452` 원장 읽기 |
| + 원장 읽기 | `missing evidence: openspec/changes/a063-…/analysis/adversary.md` — **B32** 리뷰 읽기 |
| + 리뷰 읽기 | **통과**, `effective_base == E` |

**다섯째는 없다.** 기록은 옮기기 **전** 경로를 적는다 — `draft` 가
`ledger_path.relative_to(root)` 로 쓰고 a063 의 실물 기록도
`openspec/changes/a063-align-attestation-renewal-profile/analysis/…` 다. 아카이브는
내용을 안 바꾸고 자리만 옮기므로 digest 셋·ancestry·tree·drift(`openspec/` 아래)는
그대로 통과한다.

4.4 의 처방("정본 신원을 basename 과 분리하고, 증거 경로·digest 검사는 **그대로 둔다**")은
앞 절반만 맞다. 경로 검사를 그대로 두면 이름을 고쳐도 둘째 줄에서 막힌다.

## 편집 계획 (코드보다 먼저 적는다)

1. **신원** — `change_dir.name != CHANGE` 를 `change_id != CHANGE` 로. `change_id` 는
   **필수 인자**다. 호출자(`check_analysis.resolve_base`)는 게이트가 요청받은 id 를 넘기고,
   그 id 로 디렉터리를 찾은 것은 아카이브 문법을 아는 해소기(`resolve_referenced_change`)다.
   `validate` 는 아카이브 문법을 **새로 배우지 않는다** — 그 문법이 사는 집이 이미 둘이고
   (6.4(f)), 셋째를 만들지 않는다. 기본값을 두지 않는 이유: 두면 id 를 잊은 호출자가
   **조용히 옛 판정(이름)으로 떨어지고**, 그것이 아카이브된 a063 을 막던 바로 그 판정이다.
   필수면 잊는 순간 `TypeError` 다.

   > **정정 (같은 날, 편집 전).** 처음에 이 자리에 "기본값 갈래를 부르는 시험 호출자는
   > `SDD_PYTHON` 이 없으면 건너뛰는 하나뿐"이라고 적었다. 호출자를 세어 보니
   > `test_execution_baseline.py` 가 **20 번** 직접 부른다 — 틀린 사실이었다. 결론(필수
   > 인자)은 그대로지만 사유가 달라서 고쳐 적는다. 그 20 자리는 **디렉터리 이름**을 id 로
   > 넘기게 바꾼다 — 각 시험이 재던 뜻(이름 = 신원)을 한 글자도 안 바꾸려는 것이다.
2. **접두사** — `local_analysis` 를 지금 자리가 아니라 **적힌 자리**
   (`openspec/changes/{CHANGE}/analysis/`)로. B14 의 순회가 그 자리에서 판정하고,
   같은 순회에서 **지금 자리**로 옮긴 읽기 경로를 만든다.
3. **읽기 둘** — 원장(`L452`)과 리뷰(B32)를 옮긴 경로에서 읽는다. digest 비교는 한 글자도
   안 바꾼다.

분기 수는 그대로 42 여야 한다(갈래를 더하지 않고 조건의 피연산자와 읽는 경로만 바뀐다).
편집 뒤 같은 열거기로 다시 뽑아 대조한다.

## 이 편집이 없애는 것 — 우연히 서 있던 둘째 벽

지금은 a063 디렉터리를 다른 이름으로 **통째로 복사**하면 이름(B5)에서 막히고, 이름을
지워도 접두사(B14)에서 한 번 더 막힌다 — 복사본의 `local_analysis` 는 복사본 자리인데
기록의 경로는 a063 자리니까. 2번 편집 뒤로 접두사는 복사를 막지 **못한다**. 복사를 막는
것은 신원 하나가 되고, 그래서 그것을 못 박는 시험을 같이 쓴다
(`test_a_copied_adoption_record_does_not_make_another_change_a063`, 활성·아카이브 이름 둘).
[[surviving-mutant-may-mean-accidental-safety]]

## Calls and live bindings

`_regular_committed`(워킹트리 바이트 = HEAD blob 대조) · `_git` · `strict_json` ·
`_change_analysis_path` · `ancestry` · `verify_source_go_lock` · `nul_paths` ·
`untracked_filesystem_entries` · `go_inputs` · `range_payload` · `exact_inventory`.

## State mutations and fallbacks

없다. 읽기 전용 감사다. fallback 없다 — 실패는 `AdoptionError` 이고 `resolve_base` 가
`ValueError` 로 감싸 올리며 계획 기준으로 조용히 되돌아가지 않는다.

## Safety conclusion

주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. 게이트 도구의
감사 함수다. 편집의 실패 방향: 신원을 id 로 받으면 **아카이브된 a063 하나**가 새로
통과한다(spec 이 SHALL 로 요구하는 것). 그 밖의 change 가 새로 통과하는 경로는 없어야
하고, 복사 시험이 그것을 잰다. Go 파일 변경 0.


---

## 편집 결과 — task 6.2 (`ast.after-6.2.json`, 같은 열거기)

분기 42 → 42 · 반환 2 → 2 · raise 24 → 24. **계획대로다** — 갈래는 늘지도 줄지도 않았다.

| 자리 | 전 | 후 |
|---|---|---|
| B4·B5 | `change_dir.name != CHANGE or persisted != P` | `change_id != CHANGE or persisted != P` |
| B14 순회 | `_change_analysis_path(…, local_analysis, …)` — `local_analysis` = 지금 자리 | 같은 판정을 **적힌 자리**(`canonical + "/analysis/"`)로 하고 `located[field]`(지금 자리)를 채운다 |
| L452 (갈래 아님) | `_regular_committed(root, record["ledger_path"], head)` | `… located["ledger_path"] …` |
| B32 | `_regular_committed(root, record[name + '_path'], head)` | `… located[name + '_path'] …` |

활성 디렉터리에서는 적힌 자리 = 지금 자리라서 `located[field] == record[field]` 다 — 활성
a063 의 판정은 편집 전과 같다(기존 이관 시험 초록, 실물 a063 출력 동일).

### 변이 (실측 `62_mutations.py`, 원복 sha256 동일, 대조군 16 시험 초록)

| 변이 | 결과 | 빨개진 시험 |
|---|---|---|
| M1 신원을 다시 디렉터리 이름으로 | CAUGHT | `test_an_archived_adoption_is_rechecked_by_its_id` |
| M2 신원 판정 삭제 | CAUGHT | `test_a_copied_adoption_record_does_not_make_another_change_a063` |
| M3 접두사를 지금 자리로 판정 | CAUGHT | 아카이브 시험 |
| M4 원장을 적힌 자리에서 읽기 | CAUGHT | 아카이브 시험 |
| M5 리뷰를 적힌 자리에서 읽기 | CAUGHT | 아카이브 시험 |
| M6 `resolve_base` 가 id 대신 이름을 넘김 | CAUGHT | 아카이브 시험 |

### "복사를 막는 벽"은 변이가 아니라 직접 쟀다 (`62_copywall.py`)

M2 의 CAUGHT 는 이 질문에 답하지 못한다 — 복사 시험의 바늘이 신원 판정의 문장이라서 다른
검사가 복사를 막아도 빨개진다. 그래서 복사본 둘을 `validate` 에 직접 넣었다.

| | `a130-copy`(활성 이름) | `2026-09-11-a131-copy`(아카이브 이름) |
|---|---|---|
| 편집 **전** 사본에서 이름 판정만 지움 | `adoption evidence path is outside current change analysis` | 같음 |
| 편집 **후** 사본에서 신원 판정을 지움(M2) | **통과**, `effective_base == E` | **통과** |

계획의 주장 그대로다. 편집 전에는 접두사가 복사를 한 번 더 막는 **우연한 둘째 벽**이었고,
편집 뒤에는 신원 하나다. 그 하나를 복사 시험이 못 박는다.

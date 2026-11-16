# a190 리뷰 기록

- 날짜: 2026-11-16
- 작성자: Opus 팀메이트(구현). 설계 결정 D1~D6: Manager(Fable). 검증: Manager.
- 위험 등급: 게이트 **판정**을 바꾸는 변경 — `docs/WORKFLOW.md` 「비례 원칙」의 무거운 규율 대상(변이 원장, 다른 계열 적대 리뷰 포함 다중 리뷰).
  High-risk 트레이딩 경로(주문·손절·사이징·Guardian·원장·인증·체결) 영향 없음.

## 생략과 면제 (침묵한 생략 금지)

- Function Logic Map: not-applicable — base→작업 트리에 수정된 Go 함수가 0개다. 바뀌는 것은 `tools/gate.sh`(bash),
  새 `tools/sdd/check_openspec_validate.py`, Python 시험 2개, 문서뿐이다. Go FLM 은 bash·python 에 적용되지 않는다.
  대신 편집 전 gate.sh 구간을 `analysis/pre-edit/gate-before.txt` 에 원문으로 보존하고, 판정 갈래를 `design.md` D3 표로 열거했다.
  기존 Python 함수 본문 편집은 `test_race_detector_actually_runs.py` 의 `test_the_gate_step_count_matches_the_targets_it_runs` 하나다(아래 Pre-Edit).
- Pre-Edit Gate: 아래.
- CodeGraphContext: not-applicable — `analysis/code-context/codegraphcontext-context.md`.
- 기억 회고: `analysis/code-context/evidence-reconciliation.md` 「기억 회고」.

## Pre-Edit Gate

- change id / task id: a190 / 2.2
- 대상: `tools/gate.sh` 6~11단계(335-345행), 머리 주석(1-20행); `tools/sdd/test_race_detector_actually_runs.py::test_the_gate_step_count_matches_the_targets_it_runs`
- CodeGraph: 셸 색인 밖 — `analysis/code-context/codegraph-baseline.md` 의 `git grep` 표
- 기존 동작 근거: HEAD `4d7d15571` 원문(`analysis/pre-edit/gate-before.txt`), 실측 `analysis/strict-debt/`
- Function Logic Map / Branch Test Map: not-applicable(위). 판정 갈래표는 `design.md` D3
- upstream 상속 테스트 영향: no — 게이트 도구만
- 실패 테스트 선행 작성: yes (tasks 2.1)
- 설정·DB·journal 변경과 rollback: 없음. rollback 은 이 브랜치 커밋 되돌림
- 안전 불변식 §0: 통과 — 주문·토글·live 경로를 건드리지 않음

## proposal-freeze 리뷰 (2026-11-16)

- 보이스: 작성자와 분리된 Claude 서브에이전트 1개(적대 Eng 관점, 읽기 전용, gstack `autoplan` 에 준하는 1회).
  openspec 1.16.0 `dist/` 소스(`validate.js`·`validator.js`·`root-selection.js`·`archive.js`)를 읽고
  `$HOME/tossos-work/a190-review-scratch/` 픽스처로 CLI 를 실측했다. 다른 계열(Codex) 적대 보이스는 최종 diff 리뷰(3.2)에 붙인다.
- 판정: **D1~D6 변경 필요 없음.** 리뷰어 문장: 「코드나 CLI 동작으로 봤을 때 실행할 수 없는 결정은 없습니다」.

| # | 심각도 | 발견 | 처리 |
|---|---|---|---|
| 1 | P2 | strict 에서도 INFO(「Archive would refuse this delta」·「Could not check archive merge conflicts」)는 valid·rc 0 → (가) 통과, archive 가 거절 | **후속.** 옛 `--all --strict` 에도 있던 구멍이고 archive 가 따로 거절한다. D1 이 rc 0 을 통과로 정했으므로 새 가드를 만들지 않음(지시: 범위 밖 새 가드는 후속). design D3 끝에 명시 |
| 2 | P2 | spec 「합계가 남아야 한다」 vs D2 「죽으면 그 사실만」 불일치, FAIL 시 advisory 실행 여부 미정 | 수용. spec 을 「합계 또는 합계를 얻지 못한 사유」로, D2 에 「결과와 상관없이 항상」 |
| 3 | P2 | (나) 표본 검사가 0 만 막음 — 부분 맹목(spec 탐색 실패 등) | 부분 수용. 최소 적용 V10(닫는 change 가 (나) 항목에 있어야 함). spec 수·디스크 수 대조는 **후속** |
| 4 | P2 | 배선 시험이 진짜 CLI 뿐이면 CI 에서 skip, 정적 단언만으로는 ⑪ 삭제+`TOTAL_STEPS=10` 이 초록 | 수용. 배선 시험에 PATH 스텁 판 추가, 단계 수 단언을 순회 뒤 step 줄까지 셈 |
| 5 | P3 | root 비교 방식 미정(openspec 은 실경로), `root` 없음 처리 미정 | 수용. V6: 양쪽 실경로, 없으면 FAIL |
| 6 | P3 | `valid` 엄격도(`"false"` 문자열), `items`+error `status` 동시 | 수용. V5·V7 |
| 7 | P3 | 같은 이름 change+spec 은 spec 해소가 아니라 `ambiguous_item` | 수용. D3 부연 정정, 사유에 이름 변경 안내 |
| 8 | P3 | 아카이브 대상 FAIL 주장은 맞음, 사유가 「Did you mean」뿐 | 수용. unknown_item 사유에 「아카이브된 change 는 openspec 이 검증하지 않는다」 덧붙임(판정 갈래 추가 없음) |
| 9 | P3 | 시험 갈래 번호가 D3 표와 안 이어짐, 시한 주입 수단 미정 | 수용. D3 행 번호 V1~V13 + 시험 열, 시한은 함수 인자로 주입 |
| 10 | P3 | telemetry: 1.16.0 은 기본 전송, gate 한 번에 1회→3회 | **Manager 결정 항목.** 스크립트는 환경을 바꾸지 않음(기존 `make validate` 와 같음). 시험·하네스만 `OPENSPEC_TELEMETRY=0` |
| 11 | P3 | (나)는 남의 **오류**(delta 없는 새 change)로 여전히 막힘 — proposal 이 이 잔여를 말하지 않음 | 수용. proposal·design 「여전히 거부함」에 예시 |
| 12 | P3 | D5 누락: `gate.sh:6` 「10개 조건」(이미 낡음), `Makefile:129` 「(SDD 게이트)」, 요구 문장 괄호에 sdd-check·test-race 없음 | 수용. D5 에 더하고 요구 문장 보강(475자, 500자 경고선 아래) |

- 리뷰어가 확인한 영향: 순회에서 `validate` 를 빼면 깨지는 시험은 `test_the_gate_step_count_matches_the_targets_it_runs` 하나.
  `test_gate_resolves_archived_changes.py`·`TheTwoChangeResolversAgree` 는 4단계 전에 멈춰 무관. CI·`.githooks`·`scripts` 소비자 0.

## 최종 diff 리뷰 1회차 (2026-11-16, HEAD `4a29b3e53`)

- 보이스: gstack `review`(Claude) — 전문가 testing·maintainability·simplification 3개(scope 밖: security·performance·data-migration·
  api-contract·design·state-machines), Claude 적대 서브에이전트 1개(「attacker and chaos engineer」, 테스트 파일은 요약 모드),
  **다른 계열 적대 보이스 Codex CLI 0.165.0**(`codex exec -s read-only … </dev/null`, 프롬프트 `analysis/review/codex-prompt-r1.md`,
  원문 `analysis/review/codex-r1.txt`). Codex 판정: **CHANGES REQUESTED**. Claude 적대: 「merge, optionally fix 2·3·6」.
- gstack gate scan: `GATE_SUMMARY: listed=2 eligible=1 inspected=1 unread=0 tagged={}` — 게이트 편집 1건
  `test_race_detector_actually_runs.py @@ -76` (단계 수 단언이 순회 뒤 step 줄까지 셈). 처리: **유지 — 정당**(완화가 아니라 셈을 넓힘.
  11단계 블록 삭제 + `TOTAL_STEPS=10`(M9)은 이 단언이 아니라 배선 시험이 잡는다 — 변이 원장 확인).
- maintainability·simplification: NO FINDINGS.

| # | 출처 | 심각도 | 발견 | 처리 |
|---|---|---|---|---|
| 1 | Codex #2 | P2 | **fail-open**: `root.path=""`/`"."` 를 `Path("").resolve()` 가 cwd(=저장소 루트)로 바꿔 V6 통과, (가)(나) OK rc 0 | **고침.** 절대경로만 받음. 시험 `test_v6_an_empty_relative_or_unreadable_root_is_not_this_repository`(편집 전 스크립트에서 RED 3/3 — `analysis/tdd/red-review-r1.txt`), 변이 M14 |
| 2 | Codex #1 · 적대 #3 | P2 | UTF-8 아닌 출력이 `UnicodeDecodeError` traceback — advisory 에서 나면 통과 판정이 FAIL 로 바뀜(D2 위반) | **고침.** `encoding="utf-8", errors="replace"` → V3, advisory 는 `except Exception` 으로 사유 한 줄. 시험 2개, 변이 M16·M17 |
| 3 | Codex #3 · 적대 #2 | P3 | `issues: null`·`root.path` NUL 이 TypeError/ValueError 로 (나)·advisory 를 끊음(D1·D2) | **고침.** `issue_lines` 가 비목록을 사유로, resolve 예외를 V6 로. 시험 2개, 변이 M15 |
| 4 | testing 전문가 | INFO(8) | (나)의 V11 갈래가 스텁 시험에 없음 — CI 에서 `repository_verdict` 끝을 `return True, []` 로 바꿔도 초록(전문가 실측) | **고침.** `test_v11_an_invalid_item_anywhere_fails_the_repository_check`, `test_v12` 에 (나) 경우, 변이 M13 |
| 5 | Codex #4 | P3 | 하네스 docstring 이 「blob 기록」을 약속하지만 HEAD·status 만 기록 | **고침.** 파일별 sha256 출력. 옛 원장은 `history-run2-ledger-before-review-r1.txt` 로 보존, 현 원장은 새 실행 |
| 6 | 적대 #6 | Low | stdout 버퍼링으로 파이프 출력에서 FAIL 줄이 OK 줄보다 먼저 보임 | **고침.** `print(..., flush=True)` |
| 7 | 적대 #1 | Medium(INVESTIGATE) | delta 가 아직 없는 남의 새 change 가 (나) ERROR 로 모든 gate 를 막음 | **근거 있는 기각(설계대로).** D4 「여전히 거부함」에 이미 명시된 사용자 결정. 완화(예: `skip_specs`, (나) 범위 축소)는 Manager 결정 항목 |
| 8 | 적대 #4 · freeze #1 | Low | strict 도 INFO(「Archive would refuse」)는 valid | **후속**(아래) |
| 9 | 적대 #5 | Low | `valid: true` 인데 issues 에 ERROR 가 있는 자기모순 보고 | **후속.** openspec 1.16.0 으로 재현 불가, 새 가드라 범위 밖 |
| 10 | 적대 #7 | Low | 아카이브된 대상은 ⑪ 에서 항상 FAIL | **설계대로.** D3 「새로 거부하는 것」, 사유 문장 있음. Manager 확인 항목 |
| 11 | 적대 #8 · freeze #10 | Low | telemetry 1→3회 | **Manager 결정 항목**(아래 후속) |
| 12 | 적대 #9 | Low | 첫 `make sdd-test` 기록이 `AF_UNIX path too long`(긴 TMPDIR)로 tools/logic-map 에서 멈춰 tools/sdd 가 안 돎 | **고침.** 옛 기록은 `analysis/verify/history-sdd-test-run1-af-unix-path-too-long.txt`, 짧은 TMPDIR 로 재실행 |

- 변이 원장(수리 뒤): 19개 전부 CAUGHT, 무변이 대조군 앞뒤 GREEN, 원본 바이트 불변 — `analysis/mutation/ledger.txt`(파일별 sha256 포함).
  M10 의 찾을 문자열이 수리로 사라져 한 판이 멈췄다(`history-run3-stale-m10-anchor.txt`) — 하네스가 「안 닿는 변이」를 거부한 결과.

## 후속

- strict 부채 51 항목(경고 262: 요구 문장 500자 초과 235 · Purpose 자리표시 22 · 첫 `##` 위 본문 5) — `analysis/strict-debt/debt-list.md`. 이 change 는 고치지 않는다.
- (가)의 INFO 구멍(proposal-freeze #1): strict 가 「Archive would refuse this delta」 INFO 를 valid 로 친다. archive 가 따로 거절한다.
- (나)의 더 강한 표본 검사(proposal-freeze #3): spec·change 수를 디스크 수와 대조.
- telemetry 기본값(proposal-freeze #10, 최종 R1 #11): gate 의 openspec 호출에 `OPENSPEC_TELEMETRY=0` 을 넣을지 — Manager 결정.
- 자기모순 보고(`valid: true` + ERROR issue) 거부(최종 R1 #9).
- 남의 delta 없는 새 change 가 (나)로 모든 gate 를 막는 결합(최종 R1 #7) — 사용자 결정 범위, 완화 여부는 Manager.

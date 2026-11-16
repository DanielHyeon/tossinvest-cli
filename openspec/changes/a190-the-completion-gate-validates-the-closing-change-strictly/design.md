## Context

`tools/gate.sh` 6~11단계는 한 순회(`for target in sdd-check test test-seams test-race vet validate`)로 `make <target>` 을 부른다.
11단계 `make validate` 는 저장소 전체를 strict 로 본다. openspec 1.16.0 에서 그 결과는 남의 경고 51 항목 때문에 rc 1 이다.
편집 전 원문은 `analysis/pre-edit/gate-before.txt` 에 있다. 결정 D1~D6 은 Manager 가 정했고(2026-11-16) 여기서 다시 설계하지 않는다.

## Goals / Non-Goals

- Goal: 닫는 change 는 strict 로, 저장소 전체는 오류만 판정한다. 못 본 것은 통과가 아니다. 판정에서 빠진 부채는 출력에 남는다.
- Non-goal: `proposal.md` 「Non-goals」 그대로.

## Decisions

### D1 판정

11단계 통과 = (가) `openspec validate <CHANGE_ID> --strict --no-interactive --json` rc 0 **그리고**
(나) `openspec validate --all --no-interactive --json` rc 0. 둘 다 평가하고 사유를 전부 낸 뒤 하나라도 FAIL 이면 FAIL.
`--json` 은 판정을 사람이 읽는 문구가 아니라 구조로 읽기 위해서만 더한다(D3).

### D2 부채 가시성

스크립트가 (가)(나) 뒤에 — 그 결과와 **상관없이 항상** — `openspec validate --all --strict --no-interactive --json` 을 한 번 더 돌려
`advisory — 판정 아님: 저장소 전체 strict … passed, … failed (… items)` 한 줄을 출력한다.
명령이 죽거나 JSON 이 아니면 합계 대신 **합계를 얻지 못한 사유**를 같은 advisory 줄로 쓰고 종료 코드는 바꾸지 않는다.
목록은 `make validate`(그대로 둠)로 본다. openspec 실행 자체가 불가(행 V1)면 advisory 도 그 사유를 적는다.

### D3 fail-closed

행 번호는 시험 이름과 변이 원장이 가리키는 좌표다. 「시험」 열: R = 진짜 openspec, S = 스텁(CI 에서도 돎), U = 함수 직접 호출.

| 행 | 입력 | (가) | (나) | 시험 |
|---|---|---|---|---|
| V1 | `openspec` 실행 불가(`OSError`, PATH 부재 포함) | FAIL | FAIL | S |
| V2 | 시한(기본 300초) 초과 `TimeoutExpired` | FAIL | FAIL | U(시한을 인자로 주입, 잠자는 스텁) |
| V3 | stdout 이 엄격한 UTF-8 이 아니거나(문자열 속 깨진 바이트 포함), JSON 객체가 아니거나, 파싱이 RecursionError | FAIL | FAIL | S |
| V4 | `items` 가 목록이 아니거나 없음(예: `status[].code = unknown_item` · `ambiguous_item` · `no_openspec_root`) | FAIL, status 문장을 그대로 전함 | FAIL | R |
| V5 | severity `error` 인 `status` 가 있음(`items` 와 함께 와도) | FAIL | FAIL | S |
| V6 | `root.path` 문자열이 이 저장소 루트의 실경로 문자열과 **정확히** 같지 않음(없음·빈 문자열·상대경로·`/proc/self/cwd`·다른 저장소 포함) | FAIL | FAIL | S |
| V7 | 항목이 dict 가 아니거나 `valid` 가 bool `True` 가 아님(`"true"` 문자열 포함) | FAIL | FAIL | S |
| V8 | (가) `items` 가 정확히 1개가 아니거나 그 `id≠CHANGE_ID` 또는 `type≠change` | FAIL | — | S |
| V9 | (나) `items` 0개 (표본 0) | — | FAIL | R |
| V10 | (나) `items` 에 `id=CHANGE_ID, type=change` 가 없음 (닫는 change 를 못 본 저장소 검증) | — | FAIL | S |
| V11 | rc ≠ 0 | FAIL(issue 나열) | FAIL(invalid 항목·issue 나열) | R |
| V12 | rc = 0 인데 JSON 이 invalid 항목을 말함, 또는 rc ≠ 0 인데 전부 valid | FAIL(불일치) | FAIL(불일치) | S |
| V13 | rc = 0, 위 어느 것도 아님 | OK | OK | R |

- `--type change` 를 붙이지 않는다. 2026-11-16 실측으로 `--type change` 는 없는 id 를 「delta 없음」 ERROR 로 돌려주어
  「못 찾음」과 「delta 없음」을 가를 수 없다. 붙이지 않으면 `unknown_item` 이 나온다. 같은 이름이 spec 에만 있으면 `type≠change`(V8),
  change 와 spec 에 둘 다 있으면 `ambiguous_item`(V4)으로 FAIL 이다 — 그 change 는 이름을 바꿔야 닫을 수 있다(사유 문장에 적는다).
  2026-11-16 활성·아카이브 change 중 spec 과 이름이 겹치는 것은 0건(proposal-freeze 리뷰 실측).
- `root.path` 대조(V6)는 openspec 이 nearest 탐색·등록 store 로 **다른 저장소**를 검증하는 경우를 「못 본 것」으로 친다.
  openspec 1.16.0 은 실경로 문자열을 낸다. 보고된 문자열은 **해석하지 않고** `str(ROOT.resolve())` 와 정확히 비교한다 —
  해석하면 빈 문자열·상대경로·`/proc/self/cwd` 가 판정 프로세스의 cwd(= 저장소 루트)로 둔갑한다(최종 리뷰 R1 Codex #2, R2 적대 F3 이 재현한 fail-open).
- 비정상 필드(`issues: null` 등)·advisory 의 예상 못 한 예외는 traceback 이 아니라 사유 한 줄이 된다 — (가)(나)를 둘 다 평가하고
  advisory 를 항상 출력한다는 D1·D2 를 지키기 위함(최종 리뷰 R1).
- V10 은 D3 「표본 수를 확인할 것」의 최소 적용이다: 0 이 아니어도 닫는 change 가 빠진 저장소 검증은 그 change 를 본 적이 없다.
  spec 수·디스크 수 대조 같은 더 강한 표본 검사는 만들지 않고 「후속」에 적는다.
- **새로 거부하는 것:** 아카이브된 change 를 대상으로 한 gate 는 (가)가 `unknown_item`(V4)이라 FAIL 이다. openspec 은 활성 change 만
  찾는다. 닫는 순서가 gate → archive 라 정상 절차에는 영향이 없다. 아카이브 뒤 재검증은 ⑪ 에서 멈춘다. 그 사유 문장에
  「아카이브된 change 는 openspec 이 검증하지 않는다」를 덧붙여 원인이 드러나게 한다.
- **(가)의 INFO 는 보지 않는다.** openspec 1.16.0 strict 는 INFO(「Archive would refuse this delta」 등)를 valid 로 친다. 이 구멍은
  옛 `--all --strict` 에도 있었고 archive 가 따로 거절한다. D1 은 rc 0 을 통과로 정했으므로 여기서 새 가드를 만들지 않고 「후속」에 적는다.

### D4 무엇을 더는 거부하지 않고, 무엇을 여전히 거부하는가

- 더는 거부하지 않음: 닫는 change 가 아닌 change·main spec 의 strict **경고**(2026-11-16: 51 항목, 경고 262 — `analysis/strict-debt/debt-list.md`).
- 여전히 거부함: 저장소 어디서든 **오류**(예: delta 가 아직 없는 남의 새 change — strict 없이도 ERROR 라 남의 gate 를 막는다),
  닫는 change 의 경고(strict), 위 표의 판정 불가 전부(V1~V12).
- 이 두 목록은 `tools/gate.sh` 머리 주석에도 적는다.

### D5 문서

`docs/WORKFLOW.md` ⑪ 서술(277-280행)과 「단계별 강제 지점」 표의 「strict validate는 gate ⑪」(296행), `tools/gate.sh` 머리 주석(6행 「10개 조건」 — 이미 낡음, 19행),
`Makefile:129` 주석 「(SDD 게이트)」(이 change 뒤 거짓),
main spec `sdd-workflow` 「자동화된 완료 게이트」(validate 를 문장으로 말함 → MODIFIED 델타).
「OpenSpec 기반 변경 관리」의 `openspec validate <change> --strict`(구현 전 조건)는 이 change 와 같은 뜻이라 고치지 않는다.

### D6 부채 목록

`analysis/strict-debt/` 에 원문(`make-validate-strict.txt`, `strict-all.json`, `lenient-all.txt`)과 그로부터 만든 목록
(`debt-list.md`, 생성 스크립트 `summarize.py`)을 둔다. 고치지 않고 `review.md` 「후속」에 올린다.

### 구현 형태

새 Python 스크립트 `tools/sdd/check_openspec_validate.py <change-id>` 를 gate.sh 11단계가 직접 부른다(5단계가 `check_analysis.py` 를 부르는 방식).
이유: 판정이 JSON 을 읽어야 하고(셸에 파서가 없음), 1~10단계를 돌리지 않고 11단계만 시험할 단위가 필요하다. 새 make 타깃은 두지 않는다
— gate 말고 부를 곳이 없고 `make validate` 는 사람용으로 남는다.
6~10단계 순회에서 `validate` 를 빼므로 `test_race_detector_actually_runs.py` 의 단계 수 단언을 「순회 뒤 단계」까지 세도록 고친다.

## 시험 전략

- 단위: 스크립트를 임시 저장소에 복사해 직접 부른다. Manager 갈래 ①(남의 경고만 → 통과) ②(닫는 change 경고 → FAIL)
  ③(남의 오류 → FAIL) ⑤(id 못 찾음 → FAIL) ⑥(표본 0 → FAIL) ⑦(advisory 줄, 실패해도 판정 불변)은 **진짜 openspec** 으로
  (없으면 사유를 대고 skip), ④(openspec 부재)와 D3 의 S 행은 스텁으로(CI 에서도 돈다). 스텁은 받은 argv 를 기록해 (가)(나)(advisory)
  명령 모양(`--strict` 유무, id)을 단언하게 한다.
- 배선: 실제 `tools/gate.sh` 를 임시 저장소에서 돌린다. 5단계·6~10단계는 스텁(`check_analysis.py`·`Makefile`)으로 통과시키고
  11단계는 진짜 openspec 판과 **PATH 스텁 openspec 판**(CI 에서도 돎) 둘로 돌려 「11/11 에서, 그 사유로」 멈추거나 `GATE PASS` 에 닿는지 본다.
- 정적: `test_race_detector_actually_runs.py` 의 단계 수 단언은 순회 뒤 `step "$STEP_NO/$TOTAL_STEPS` 줄 수까지 센다.
- 변이 원장: 사본(`$HOME/tossos-work/`)에서, 무변이 대조군 GREEN 확인 뒤, 기대한 시험 이름이 실패할 때만 CAUGHT.

## Risks / Trade-offs

- [CI 에 openspec 없음] → 진짜 CLI 시험은 CI 에서 skip. 스텁 시험이 판정 갈래를 CI 에서도 잰다. 로컬 gate 는 openspec 없이는 ⑪ 에서 FAIL 이다.
- [openspec JSON 모양이 판마다 바뀜] → 모양이 다르면 FAIL(fail-closed). 1.16.0 실측 모양만 받는다.
- [부채 누적] → advisory 줄과 `make validate` 로 보인다. 판정은 아니다 — 사용자 결정이다.
- [telemetry] openspec 1.16.0 은 기본으로 사용 통계를 보낸다. Manager 결정(2026-11-16)으로 판정 단위의 openspec 호출 세 번은
  `OPENSPEC_TELEMETRY=0` 으로 돈다(부모 환경은 그대로 물려주고 이 하나만 덮음). `make validate` 타깃은 손대지 않았다.

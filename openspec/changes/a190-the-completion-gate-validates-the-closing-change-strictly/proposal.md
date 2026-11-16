## Why

완료 게이트 `tools/gate.sh` 11단계는 `make validate` = `openspec validate --all --strict --no-interactive`(`Makefile:130-131`)다.
2026-11-16 main `4d7d15571` 에서 이 명령은 rc 1, `Totals: 8 passed, 51 failed (59 items)` 다
(`analysis/strict-debt/make-validate-strict.txt`). 실패 51 항목의 issue 262건은 **전부 WARNING** 이고
(요구 문장 500자 초과 235 · main spec Purpose 자리표시 22 · 첫 `##` 위 본문 5), strict 가 경고를 실패로 친다.
`--strict` 를 빼면 59/59 통과, rc 0, 오류 0 이다(`analysis/strict-debt/lenient-all.txt`).

닫으려는 a189·a187 은 단독 `openspec validate <id> --strict` 가 valid 인데, 남의 진행 중 change 와 main spec 의 경고 때문에
11단계에서 막힌다. 이 상태면 **어떤 change 도 archive 할 수 없다.** 같은 모양의 막힘은 전에도 있었다
(`a063 issues.md:56`: 무관한 a119 때문에 `make validate` exit 2).

사용자 결정(2026-11-16, 선택지 1): 11단계를 「닫는 change 는 strict 로, 저장소 전체는 오류만」 보도록 고치는 작은 change 를 먼저 낸다.

## What Changes

- 11단계 판정을 둘의 논리곱으로 바꾼다.
  (가) `openspec validate <CHANGE_ID> --strict --no-interactive` rc 0, 그리고
  (나) `openspec validate --all --no-interactive`(strict 없음) rc 0.
- 판정은 새 스크립트 `tools/sdd/check_openspec_validate.py <change-id>` 가 `--json` 출력을 읽어 내린다.
  openspec 부재·실행 실패·시한 초과·JSON 아닌 출력·`<CHANGE_ID>` 미발견·(나) 표본 0 은 FAIL 이다(fail-closed).
- 저장소 전체 strict 결과는 판정에서 빠지되, gate 가 그 합계(또는 합계를 얻지 못한 사유)를 「advisory — 판정 아님」 한 줄로 항상 출력한다. 그 명령이 실패해도 gate 는 계속한다.
- `make validate` 타깃은 그대로 둔다(사람이 부채 목록을 보는 명령).
- `docs/WORKFLOW.md` 의 ⑪ 서술, `tools/gate.sh` 머리 주석, main spec `sdd-workflow` 「자동화된 완료 게이트」를 맞춘다.
- 오늘의 strict 실패 51 항목 목록을 `analysis/strict-debt/` 에 원문으로 남긴다(고치지 않는다).

## 더는 거부하지 않는 것 / 여전히 거부하는 것

- **더는 거부하지 않음:** 닫는 change 가 아닌 다른 change 와 main spec 의 strict **경고**(오늘 51 항목, 262건).
- **여전히 거부함:** 저장소 어디서든 **오류**(non-strict 에서 invalid — 예: delta 가 아직 없는 남의 새 change 도 남의 gate 를 막는다),
  닫는 change 의 경고, 판정할 수 없는 validate.
- **새로 거부함:** 아카이브된 change 를 대상으로 돈 gate(openspec 은 활성 change 만 찾으므로 (가)가 `unknown_item`).
  전에는 ⑪ 이 대상 id 를 보지 않아 통과했다. 닫는 순서는 gate → archive 라 정상 절차에는 영향이 없다(`design.md` D3).

## Non-goals

- 51 항목의 문서 수정, openspec 판 고정·재설치, 다른 gate 단계 변경.
- `internal/performance` p95 시험의 부하 민감성.
- a189·a187 archive (a190 뒤 별도 지시).
- 주문·위험·원장·인증 등 트레이딩 코드, Go 코드 변경 없음.

## Capabilities

### Modified Capabilities

- `sdd-workflow`: 「자동화된 완료 게이트」 요구의 validate 조건을 닫는 change strict + 저장소 전체 오류 0 으로 바꾸고,
  저장소 전체 strict 합계를 advisory 로 남기는 것과 fail-closed 시나리오를 더한다.

## Impact

- 위험 등급: 게이트 **판정**을 바꾼다. `docs/WORKFLOW.md` 「비례 원칙」상 무거운 규율(변이 원장 · 다른 계열 적대 리뷰 포함 다중 리뷰)을 적용한다.
  High-risk 트레이딩 경로 영향은 없다.
- 영향 파일: `tools/gate.sh`, `tools/sdd/check_openspec_validate.py`(새), `tools/sdd/test_gate_validate_step.py`(새),
  `tools/sdd/test_race_detector_actually_runs.py`(단계 수 단언), `docs/WORKFLOW.md`, `openspec/changes/<id>/specs/sdd-workflow/spec.md`.
- CI(`.github/workflows/ci.yml`)에는 openspec 이 없다. 새 시험 중 진짜 CLI 를 쓰는 것은 그곳에서 이유를 밝히고 건너뛰고,
  스텁 시험은 그곳에서도 돈다.

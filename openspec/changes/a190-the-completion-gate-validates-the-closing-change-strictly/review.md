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

## proposal-freeze 리뷰

(tasks 1.4 에서 기록)

## 후속

- strict 부채 51 항목(경고 262: 요구 문장 500자 초과 235 · Purpose 자리표시 22 · 첫 `##` 위 본문 5) — `analysis/strict-debt/debt-list.md`. 이 change 는 고치지 않는다.

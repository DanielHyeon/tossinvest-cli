# a208 review

## proposal-freeze

- 상태: **Manager 몫(tasks 0.7, 미완)**. 사람 결정 Q1~Q3(2026-11-28)은 Manager 전달로 design 10 에 기록됨.

## Pre-Edit Gate — a208 tasks 1.2(2026-11-28, Opus 팀메이트)

```text
Pre-Edit Gate:
- change id / task id: a208-adoption-floor-only-exit / 2.2 · 2.3 · 2.4
- 대상 심볼(패키지.함수):
  engine.ReconcileDriver.adoptOne (High-risk — 편입 익절·보호선 정책 선택)
  engine.recordGateSettings (audit 표)
  config.mergeAdoption · config.Service.LoadRawEngineAdoption · config.Service.SaveEngineAdoption (설정)
  console.Console.handleSettingsSave (콘솔 저장 이월)
  engine.TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten (시험 단언 첨자)
  새 파일: internal/app/engine/a208_adoption_floor_only.go (adoptionExitPolicyID · adoptionAlertBody)
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md 「구현 착수 재확인」
  (1.6.0, affected 는 --filter '*_test.go'; adoptOne←adopt 1, handleSettingsSave←route 1, recordGateSettings←NewContext+시험 5)
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/ — 불일치 1(Save/LoadRaw 호출자)은 HEAD 로 해소, 편집 차단 없음
- 기존 동작 파악 근거: 현재 HEAD 원문 + tools/logic-map AST 7 번들(analysis/function-logic/), 기존 시험
  reconcileloop_test(편입군) · adoption_io_test · adoption_test · console settings_test · a074/a163/a197 audit 시험
- Function Logic Map / Branch Test Map: analysis/function-logic/<7 번들>, check_analysis TODO 0
- upstream 상속 테스트 영향: no — 편입은 TossOS 전용 경로. 토글 OFF 는 기록 정책 · 알림 · 설정 바이트가 a208 이전과 같음(T3 · T7 바이트 단언)
- 실패 테스트 선행 작성: yes — 810aec6f7, analysis/tdd/red.txt(T1·T2·T4·T6·T7·T8·T9·T10 실패, T3·T5 는 기대대로 통과)
- 설정·DB·journal 변경과 rollback: 설정 키 1 개 additive(기본 false, omitempty). 스키마 · journal · 저장 스냅숏 변경 0.
  롤백 순서 제약 없음 — design 7(옛 이미지는 키를 무시, RUNNER 편입 사본은 a633ecff3 부터 존재)
- 안전 불변식 §0 위반 여부 검토: 통과 — §0.2 OFF 동일(T3·T7), §0.3 손절 경로 · 합성 손절 무편집(T1 이 손절가 단언),
  §0.4 브로커 호출 추가 0, §0.5 기동 audit 행(T10), §0.6 스키마 무변, §0.7 토글은 사람이 설정 파일로만 켬(콘솔은 이월만),
  §0.9 익절을 없애고 보호선 잠금만 남김 — 손절선을 낮추는 갈래 없음(보수 방향), 유예 없음(Q1)
```

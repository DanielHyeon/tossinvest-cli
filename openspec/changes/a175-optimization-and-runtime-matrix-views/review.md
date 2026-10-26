# a175 review

## proposal-freeze (2026-10-26, 경량: validate + Manager 셀프리뷰, docs/WORKFLOW.md 「위험 등급 가중」 UI change)

- 보이스: Manager(Opus 5.5) 단독. 적대적 Eng 보이스는 해당 없음. 주문·위험관리·원장·reconciliation 은 건드리지 않고, 프리셋 POST 는 preview 단계로 가는 기존 폼을 바이트 그대로 옮긴다.
- base: `b7d34acad`(a174 archive, origin/main 과 같음).
- 입력: Fable 5.1 정의(proposal·design·tasks.draft·spec 델타), HEAD `b7d34acad` 원문 grep(`templates_optimization.go`·`templates_signals.go`·`templates.go`·`optimization_view.go`·`internal/candidate/threshold_descriptors.go`·`internal/exitpolicy/common_policy.go`·`display_primitives_test.go`).
- 사람 결정 승계: worktree 형제 디렉터리, 구현자 GPT-5.6 Terra, WORKFLOW 비례 원칙 문구(기존 함수 편집 = 번들).
- 발견과 결정(정의본 대비 정정, design §1 ★):
  1. 익절 프리셋 컨테이너는 `.tier-card` 가 아니라 `.choice-tile` 이고, 폼은 preview 로 가는 1단계다. hover 금지 시험은 `display_primitives_test.go:565-581` 이다(`TestAmbientMotionContract` 아님). → 시험 확장 대상과 함수 이름을 정정했다.
  2. runner 의 T4 `TargetPct="999.0"` sentinel 이 지금 표에 `999.0%` 로 찍혀, 같은 타일의 "sentinel 을 노출하지 않는다" 문장과 모순된다(기존 결함). → 행렬은 "고정 목표 없음" 으로 표시하고 시험으로 고정한다(2.3).
  3. 중계 카테고리의 링크 수는 1·2·1 이다(전략·실행은 2). 그리고 성과·이력은 설정 변경 감사 이력 표를 가진다. → "링크 1" 단언을 "링크 수 불변" 으로 바꾸고, 감사 표는 유지한다(정보 손실 방지). 700px 기준에서 성과·이력은 뺐다.
  4. 후보 필터의 "공통 행" 은 값이 같다는 가정이다. → KR·US 값이 바이트 단위로 같을 때만 합치고, 다르면 시장별로 남긴다. 다른 값 fixture 로 시험한다(2.4).
  5. "기록만 하고 판정하지 않는다 6회" 는 렌더 기준 값이다(소스는 Go 문자열 1개) — 1.1 에서 렌더로 다시 센다.
  6. FLM 면제 문장을 거뒀다: 기존 함수(시험 함수 포함)를 고치면 함수마다 번들이 필요하다.
- 남은 위험: 신호 화면 정찰 값(3006px·공통 문장 수)은 Manager 가 다시 재지 않았다(구현자 1.2 기준선이 대신 잰다). `strategy-runtime` 두 열 배치가 한 시장 실패를 다른 시장 표시와 독립으로 유지하는지는 2.5 에 KR 실패 fixture 단언을 넣어 확인한다.
- 결론: proposal-freeze 승인. 착수 가능. 위치 `/mnt/D/Axipient/workspace/TossOS.worktrees/a175-optimization-and-runtime-matrix-views`. 정정본을 커밋한 **뒤에** worktree 를 만든다.

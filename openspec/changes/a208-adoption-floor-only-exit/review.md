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
  §0.9 손절 기준선(합성 손절)은 무변 · 손절선을 낮추는 갈래 없음 · 유예 없음(Q1). 단 공통 정책 대비 이익 구간 보호선은
  늦다(본전 잠금 +2.5%, runner trail 없음 — 리뷰 F3 에서 발견해 이 줄을 정정) — 사용자 결정(부분익절 없음 · Q2 RUNNER 표)으로 수용
```

## 구현 리뷰 1회차(2026-11-28, HEAD `a547fa67f`) — 2 보이스

- 보이스: (A) **codex exec 0.165.0, read-only sandbox — 작성자와 다른 계열**(적대적, 결함만) · (B) gstack `review` 의 native 적대 패스
  (Claude Sonnet subagent — **작성자와 같은 계열**, "Think like an attacker and a chaos engineer"). 생산 diff 46+/7−(6 파일) → gstack 소형 diff 규칙으로
  specialist 생략. 두 보이스 모두 P0/P1 · Blocks:yes 0.
- gstack gate scan: `GATE_SUMMARY: listed=5 eligible=1 inspected=1 unread=0 tagged={}` — 읽은 hunk 1(`a163_source_age_audit_test.go` 첨자 이동, 아래 G1).

| # | 출처 | 발견 | 처분 |
|---|---|---|---|
| F1 | codex P3 | T7 「키 없음」이 키 문자열 부재만 보고 저장 바이트 동일을 재지 않음 | **수정**(시험만): 저장 뒤 파일 바이트 == 원문 바이트로 강화. M8(omitempty 제거) 재변이 CAUGHT |
| F2 | codex P2 · native #4 | 콘솔 저장이 잠금 밖에서 읽은 토글을 이월 → 화면을 연 뒤 저장 사이의 파일 편집을 덮음 | **미수정 · 위험 기록**(design R8). include/exclude 와 같은 기존 seam 구조, 새 가드 금지 지시. 사람 판단 대상 |
| F3 | native #1(High, Blocks:no) | 켜면 본전 잠금이 BALANCED +1.5% · HYBRID_50 +1.8% → RUNNER +2.5% 로 늦어지고 runner trail 이 없음 | **사람 결정 Q2 의 귀결로 수용** · design R7 · operations.md 에 명시. Manager 에게 재확인 요청 |
| F4 | native #2 | 콘솔 · positionpolicy 런타임 스냅숏이 토글을 표시하지 않음(재시작 대기 표시 없음) | 사람 결정 Q3(콘솔 체크박스 없음)의 범위 — 수용, 기동 audit 행이 유일한 흔적(operations.md) |
| F5 | native #3 · codex | 롤백 뒤 옛 이미지의 콘솔 저장이 키를 지움 | 이미 design 7 · R6 에 있음 — operations.md 에 확인 절차 |
| F6 | native #5 | include 지정 편입도 토글을 따름 | 의도된 범위(design R9 · operations.md) |
| F7 | native #6 | a163 시험 첨자가 이후 행 추가에 또 깨짐 · 첫 기동 false 행 | 수용 — 「끝에 덧붙임」 규칙을 시험이 의도대로 지킴, false 행은 spec 이 허용 |
| G1 | gate scan | `TestA163…` 의 `len-1`/`len-2` → `len-2`/`len-3` | 권고: **Keep — justified(verified)**: 같은 change 의 `TestA208FloorOnlyIsAuditedAtStartup` 이 끝 행 = a208, 바로 앞 = a163 를 단언하므로 per_trade → a163 → a208 순서가 두 시험으로 전부 고정됨(완화 아님). 사람 처분 대기(open) |

- 2회차: **돌리지 않음** — 1회차 뒤 편집은 시험 강화 1 건(F1)과 문서뿐이고 생산 코드 무변. 강화는 변이 M8 재실행으로 확인.
- 리뷰 발견에서 새 가드 0.

## F3 재작업 — 사용자 결정 「F3 2」(2026-11-28, Manager 전달)

F3 처분 = **사용자 선택 2**: 토글이 참이면 새 편입은 「설정된 공통 정책」의 보호선 표(단계 트리거 · 보호선 · 러너 트레일)를 그대로 쓰고
부분익절 비율만 0 인 사본으로 판정된다. 1회차의 「Q2 귀결로 수용」 처분은 이 결정으로 대체됨(design 2 · 3 · 7 · 11 개정 2).

### Pre-Edit Gate — F3(2026-11-28, Opus 팀메이트)

```text
Pre-Edit Gate:
- change id / task id: a208-adoption-floor-only-exit / 2.5 · 2.6(F3)
- 대상 심볼(패키지.함수):
  exitpolicy.CommonLadderForPosition (High-risk — 판정 표 해소: 사본 ID 갈래를 B1 앞에 추가, 기존 갈래 무편집)
  exitpolicy.LegacyLadderPolicyIdentity (High-risk — 새 exit state 정체성: 사본 ID case 추가, adopted 아니면 거부)
  journal.AdoptionRequest.record (High-risk — 편입 기록 정책 ID 문: 사본 ID 하나만 더 받음)
  journal.Journal.OpenExitState (High-risk — 보호 상태 개시: 사본 ID 하나만 더 받음, 정체성은 seedPolicyIdentity 가 adopted 로 결정)
  config.mergeEngine (설정 — 무시 사유 파생 한 줄, json:"-" 필드)
  engine.ReconcileDriver.adoptOne (알림 본문 호출 인자 한 줄)
  engine.recordGateSettings (a208 행 값 · detail 을 새 파일 함수로)
  새 파일: internal/exitpolicy/a208_adopted_floor_only.go(사본 ID · 고정 정체성 · 표 유도 · 공통 정책→사본 대응),
          internal/config/a208_floor_only_ignored.go(무시 사유), internal/app/engine/a208_adoption_floor_only.go 함수 수정(이 change 의 새 함수)
- CodeGraph(1.6.0, 주 체크아웃 5ae7c7511 — worktree 미색인): affected --filter '*_test.go' common_policy.go 1421 · identity.go 1452 ·
  journal/adoption.go 1410 · exit_state.go 1440 · config/engine.go 1367(전이 폐포 = 거의 전 저장소). 직접 호출자: CommonLadderForPosition ← ladderFor 1;
  LegacyLadderPolicyIdentity ← seedPolicyIdentity · legacyPolicyIdentity · managedPolicyIdentity; record ← AdoptPosition 1; OpenExitState ← OpenAdoptedExitState ·
  openState; mergeEngine ← Service 로더 1. 정책 ID 대조 자리 전수(rg): exit_state.go:139 · :215, adoption.go:380, exit_snapshot.go:645 · :718, exitloop.go:383 · :768 · :1324,
  reconcileloop.go:338, config/engine.go:81, positionpolicy/model.go:164 · :179, console optimization.go:108 · position_policy.go:325(RegisteredCommonPolicies).
- a166 census(shapes.json): 편집 함수 7 개 모두 census 밖(rg — NewContext 는 census 안이라 호출 줄 무변경, recordGateSettings 시그니처 무변경)
- 기존 동작 파악 근거: tools/logic-map AST 5 새 번들(analysis/function-logic/internal-exitpolicy--* · internal-journal--* · internal-config--mergeengine) + 기존 2 번들
- 실패 테스트 선행 작성: yes — analysis/tdd/red-f3.txt(FAIL 10: 사본 표 · 정체성 · 원장 수용 · 비편입 거부 · 무시 사유 · 무시 audit · 엔진 T1 · T2 · T4 · T5 · T6,
  PASS 기대 6: 편입 HYBRID_50 행 무변 · 비편입 거부(지금도 미지 ID 라 거부) · 선택 불가 · 설정 거부 · OFF 무변 · RUNNER 기존 사본)
- 설정·DB·journal 변경과 rollback: 스키마 무변. 새 정책 ID 1 개가 position_adoptions.exit_policy_id · exit_states 에 적힐 수 있음 →
  **롤백 순서 제약 생김**(design 7 개정 2): 사본 ID 를 모르는 옛 이미지는 그 포지션의 판정(손절 포함)을 거부함. 새 가드 없음, 절차만.
- 안전 불변식 §0 위반 여부 검토: §0.2 OFF 동일(T3 — BALANCED · HYBRID_50 둘 다 기록 · 알림 바이트), §0.3/§0.9 보호선은 공통 정책과 같은 값(느슨해지지 않음),
  손절 경로 · 합성 손절 무편집, 이미 편입된 HYBRID_50 행 의미 무변(기존 갈래 무편집 · 시험 고정), 지원 안 하는 공통 정책은 토글 무시 = OFF(fail-closed) + 기동 audit
```

# issues — align-full-sdd-pm-contract

## task 4.1 의 산출물이 같은 날 삭제됐고 6주 동안 아무도 몰랐다 (2026-09-09 발견)

### 무슨 일이 있었나

task 4.1(`docs/WORKFLOW.md`를 StockOS Full SDD 순서·READY·증거 조정·Pre-Edit·완료 계약에
맞춘다)은 `c0619279`(2026-07-31 06:38)에서 실제로 랜딩했다. `docs/WORKFLOW.md` +122/-33.

9시간 22분 뒤, **무관한 콘솔 change**의 커밋이 그 산출물을 지웠다.

```
6d61e988  2026-07-31 16:00:32 +0900
fix(console): restore truthful adoption controls [STORY-TOS-038]
154 files changed, 4558 insertions(+), 211 deletions(-)
  docs/WORKFLOW.md | 32 +++--- 138 -----------
```

`c0619279`는 `6d61e988`의 직계 조상이고 그 사이에 `docs/WORKFLOW.md`를 만진 커밋은 없다.
병합 커밋도 아니다. 콘솔 change의 diff에 워크플로 계약이 섞여 들어가 옛 판으로 덮인 것이다.
삭제분 138줄 중 SDD 거버넌스는 아래 전부이며, 콘솔 change의 scope에 포함될 이유가 없다.

- `### SDD 4계층 앵커`
- `## SDD 사이클`의 Story 선등록·READY, **5단계 증거 조정**, 7단계 Pre-Edit Gate
- `### READY 판정` 8항목 체크리스트
- 코드 증거 절차의 `analysis/code-context/` 3파일 규칙
- `### PM 계층`의 fail-closed 계약 6항목과 파생 상태 문단 (코드펜스도 함께 깨졌다)
- `## Pre-Edit 선언`이 **비자명 production 편집 전체 → High-risk 전용**으로 축소
- `## 완료 보고 금지 조건` 3항목 (archive·Story 1:1 / 독립 리뷰 판정 / merge·deploy·승격 근거)
- `## 에이전트 실행 순서` 13단계 (섹션 통째로)

`tasks.md`의 4.1은 `[x]`인 채 남아서, 산출물이 없는 동안에도 완료로 보였다.

### 측정된 결과

승인된 delta spec은 `docs/WORKFLOW.md`가 각 단계의 입력·산출물·차단 조건을 명시할 것과,
증거 조정 없이 편집을 시작하면 Pre-Edit Gate가 차단할 것을 MUST로 요구한다. 6주 동안 둘 다
문서에 없었고, 결과는 산출물 수로 나타난다.

| 산출물 | StockOS | TossOS |
| --- | --- | --- |
| `analysis/code-context/` 보유 change | 80 (활성 190) · 2026-08-17까지 계속 생산 | **8** · 마지막 실제 생산 **2026-08-02** |
| `analysis/function-logic/index.md` (recall 영수증) | 183 | **0** (`function-logic/` 보유 96 change 중) |

즉 이 change가 이식한 단계 중 증거 조정은 문서에서 사라진 직후 실무에서도 멈췄다.
반대로 gate가 검사하는 5단계 Function Logic Map은 96개 change에서 계속 생산됐다.
**차이는 규율이 아니라 강제 지점의 유무다.**

### 조치 (2026-09-09)

1. 삭제된 8개 블록을 `docs/WORKFLOW.md`에 복원했다. 통째 revert가 아니라 블록 단위 복원이며,
   이후 6주간의 정당한 추가(a063 execution-baseline 예외, `sdd-check-ci`, 이관 worktree
   인터프리터, GBrain busy 복구, aNNN 명명 규칙, gate 11단계)는 전부 보존했다.
2. 복원 전에 각 문장이 **지금도 참인지** 코드로 확인했다 — `generate_master_tracker.py`는
   `bootstrap_change_allowlist is forbidden`, `manual status is forbidden`,
   forward/reverse 링크, change당 Story 정확히 1개, 고아 활성 change를 실제로 거절한다.
3. `## 완료 게이트 (자동화)`의 설명이 7항목으로 낡아 있어 실제 `tools/gate.sh` 11단계로 고쳤다.
4. **새로 추가**: 「단계별 강제 지점」표. 각 SDD 단계가 어느 스크립트에서 막히는지, 그리고
   어디가 `없음`인지 적는다. 이 사고가 6주간 안 보인 이유가 그 목록이 없어서였다.

### 남은 구멍 (이 change 밖)

- 5단계 증거 조정에 강제 지점이 없다. 지금 켜면 활성 28개 중 27개가 즉시 빨갛다.
  `check_analysis.py`를 저장소 전체로 켰을 때 31개 중 1개만 통과한 전례와 같은 모양이므로,
  기준선 이후 생성 change에만 요구하는 식의 설계가 먼저 필요하다.
- 0단계 recall 영수증(`analysis/function-logic/index.md`)은 TossOS가 이식한 적이 없다.
  StockOS 183건 대 TossOS 0건. 새 필수 산출물이므로 별도 change로 다룬다.
- `docs/WORKFLOW.md` 자체의 무결성을 검사하는 것이 없다. `check_agent_config_sync.py`가
  `.claude/CLAUDE.md`↔`.codex/agents.md` 공유 블록에 대해 하는 일과 같은 종류가 없다.

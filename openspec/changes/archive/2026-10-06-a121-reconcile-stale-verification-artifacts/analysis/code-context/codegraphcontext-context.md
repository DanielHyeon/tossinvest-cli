# a121 CodeGraphContext supporting context — task 1.2

- 도구: `cgc analyze callers` (FalkorDB 백엔드, `~/.codegraphcontext/.env`) — 읽기 전용 질의만. `cgc update`·`index` 는
  공유 그래프 저장소에 쓰므로 돌리지 않았다.
- 한계: 이 색인은 저장소 하나가 아니라 여러 체크아웃(`/mnt/D/project/axipient/TossOS*`, 옛 워크트리, 스크래치 경로)을
  함께 담는다. 결과가 경로별로 중복되고 어느 것이 base `de147cc2` 인지 구분되지 않는다. 두 질의 모두 정확히
  **20 행**("Total: 20 caller(s)")에서 끝났다 — 출력 상한으로 보이며, 그 뒤가 잘렸을 수 있다. 그래서 **함수 이름
  집합만**, 그리고 "CGC 가 CodeGraph 에 없는 것을 냈는가" 방향으로만 대조에 쓴다.

| 질의 | CGC 20 행 (경로 중복 제거 후) | CodeGraph 1.6.0 | 판정 |
|---|---|---|---|
| callers `PendingCleanup` | `readVerify`(3 경로) + 시험 함수 16 (`TestAReDeclaredHoldOutlivesAnOlderVerdict` 는 2 행 — 호출 자리 둘) | `readVerify` + 시험 18 | CGC 에 없는 2 개(`TestM0TriggeredWithoutChildIDIsManualReconcileOnly`·`TestM0RestartWithDurableParentOrChildOwnerNeverReentersTriggerMutation`)는 20 행 상한에서 잘린 것으로 본다. CGC 만의 호출자 0 |
| callers `M0Unsettled` | `validateM0TriggerMode` · `m0PendingCheckpoint` · `New` (여러 경로 반복) | 같은 셋 | 일치 |

누락 후보: 없음 — CGC 가 CodeGraph·HEAD grep 에 없는 호출자를 내지 않았다. 정본은 codegraph-baseline.md 의 HEAD 대조다.

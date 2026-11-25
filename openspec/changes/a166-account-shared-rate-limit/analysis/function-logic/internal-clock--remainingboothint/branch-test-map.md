# Branch Test Map: `RemainingBootHint`

RED/GREEN 원문: `../../time-namespace-lot/red-before-edit.log`, `../../time-namespace-lot/green-after-edit.log`.
변이: `../../time-namespace-lot/mutation-ledger.md`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 무효 stamp(ns 0 포함) · duration 범위 | `TestBootRemainingInvalidStamps`, `TestBootCrossNamespaceStillRejectsInvalid` | 회귀 핀(편집 전에도 GREEN) | yes |
| B2 | 같은 boot anchor 가 stamp 보다 앞섬(ns 다름) → 거절 / 순서 맞는 다른 ns anchor → 수락 | `TestBootCrossNamespaceOrderingStillRejected`, `TestBootCrossNamespaceAnchorRestart`, `TestBootRemainingRejectsAnchorBeforeSameBootProvenance` | 수락 판 yes | yes |
| B3 | 같은 boot 다른 ns 차감 / uptime 역행 거절 | `TestBootCrossNamespaceSameBootSubtractsElapsed`, `TestBootCrossNamespaceOrderingStillRejected`, `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor` | 수락 판 yes | yes |
| B4 | 다른 boot anchor 뒤 다른 ns 재시작 차감 / anchor 보다 역행 거절 | `TestBootCrossNamespaceAnchorRestart`, `TestBootCrossNamespaceOrderingStillRejected` | 수락 판 yes | yes |
| B5 | duration 0 은 ns 차이를 거절 사유로 삼지 않고, uptime 역행은 숨기지 않음 | `TestBootZeroDurationDoesNotHideInvalidClock` | 편집 뒤 단언 | yes |
| B6 · B7 | 다른 boot → NeedsAnchor(ns 무관) | `TestBootCrossNamespaceOtherBootNeedsAnchor`, `TestBootRemainingBoundariesAndRestarts` | 회귀 핀 | yes |
| B8 | 도달 불가(B3 · B4 가 먼저 거절) | 없음 — 변이 M6 | n/a | n/a |
| B9 · 끝 | 만료 0 / 남은 시간 | `TestBootCrossNamespaceExpiredHintClears`, `TestBootCrossNamespaceSameBootSubtractsElapsed` | yes | yes |
| 저장소 경로 | 다른 ns 의 hint 복원 · 429 갱신 성공 · poison 없음 · 만료 삭제 · 가득 찬 slot prune | `TestA166ObservationCrossNamespaceHintRestore`, `TestA166ObservationCrossNamespacePrune` | yes | yes |

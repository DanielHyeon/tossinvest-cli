# Branch Test Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (393-452); file SHA-256 `4eede127fbbec4233391d783660d1bca000e8d85ba61b02d394b5776840f4e50`. AST branch positions are authoritative.
- 이 태스크는 B2 만 편집했다. 나머지 행은 측정만 갱신했다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 399:2 — 배선 미완/nil | `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` | 아니오 — 편집 없음 | 예 |
| B2 | if at 413:2 — **handoff 거절 포함**, 권한 준비 미완 | `TestARefusedHandoffLeavesTheWorkerDormant` | 아니오 — 컴파일 실패로서의 RED 는 있었으나(`dispatchHandoff` 미존재) 동작 RED 는 없다. 이 태스크는 동작을 보존했다 | 예 |
| B3 | if at 417:2 — 봉인 깨진 제안 | 없음 | 아니오 | 아니오 — **진입 0** |
| B4 | if at 421:2 — 보호 관측 실패 | `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` | 아니오 — 편집 없음 | 예 |
| B5 | if at 424:2 — 진입 게이트 관측 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B6 | if at 437:2 — **검증된 4-가족 활성화가 있는가** (8.7.1) | `TestAVerifiedActivationWhoseRiskOrProtectionDigestDisagreesLeavesTheMarketDormant` | 예 — 편집 전에는 이 분기가 없었다(활성화를 아무도 읽지 않았다) | 예 |
| B7 | if at 438:3 — 활성화가 이름 부른 위험·ProtectionReady digest 가 살아 있는 값과 어긋난다 | `TestAVerifiedActivationWhoseRiskOrProtectionDigestDisagreesLeavesTheMarketDormant` | 예 — 반증 E13~E15 로 각각 확인(둘 다 · 위험만 · 보호만) | 예 |
| B8 | if at 446:2 — digest/revision/만료 (옛 B6) | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M5: B2 에서 관문을 지우고 `_ = handedOff` 로 컴파일러를 달랜다 | KILLED — `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 실패. 5.5 판본에서는 같은 모양이 **SURVIVED** 했다: 거절된 값이 영값이라 B3 의 `ValidProposal()` 이 대신 걸러 주었고, 그 자리의 안전은 경계가 아니라 우연이 지키고 있었다. |
| M7: 이 본문에 `gateway.PlaceClaimedStrategy(...)` 를 넣는다 | KILLED — `TestTheWorkerBuilderOnlyObservesThroughTheGateway` 실패 |
| M10: 어댑터가 첫 항목만 싣는다(`break`) | KILLED — `TestARefusedHandoffLeavesTheWorkerDormant` 실패. 상한 초과가 승인으로 바뀌면 이 worker 가 올라가 버린다. |

원복은 12개 전부 sha256 대조로 확인했고 `unrestored_files` 는 비어 있다.

행은 측정한 것을 말한다. 진입 0 인 arm 은 커버리지 공백이지 통과가 아니다.

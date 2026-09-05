# Branch Test Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (411-471); file SHA-256 `22855de0f27de05c60c2b5ff8cf2d5c7e3ed50e78a9fa6f67fb81ec38decdbfa`. AST branch positions are authoritative.
- 이 태스크는 B2 만 편집했다. 나머지 행은 측정만 갱신했다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 417:2 — 배선 미완/nil | `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` | 아니오 — 편집 없음 | 예 |
| B2 | if at 431:2 — **handoff 거절 포함**, 권한 준비 미완 | `TestARefusedHandoffLeavesTheWorkerDormant` | 아니오 — 컴파일 실패로서의 RED 는 있었으나(`dispatchHandoff` 미존재) 동작 RED 는 없다. 이 태스크는 동작을 보존했다 | 예 |
| B3 | if at 435:2 — 봉인 깨진 제안 | 없음 | 아니오 | 아니오 — **진입 0** |
| B4 | if at 441:2 — 보호 관측 실패 | `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` | 아니오 — 편집 없음 | 예 |
| B5 | if at 444:2 — 진입 게이트 관측 실패 | 없음 | 아니오 | 아니오 — **진입 0** |
| B6 | if at 465:2 — digest/revision/만료 (옛 B6) | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M5: B2 에서 관문을 지우고 `_ = handedOff` 로 컴파일러를 달랜다 | KILLED — `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 실패. 5.5 판본에서는 같은 모양이 **SURVIVED** 했다: 거절된 값이 영값이라 B3 의 `ValidProposal()` 이 대신 걸러 주었고, 그 자리의 안전은 경계가 아니라 우연이 지키고 있었다. |
| M7: 이 본문에 `gateway.PlaceClaimedStrategy(...)` 를 넣는다 | KILLED — `TestTheWorkerBuilderOnlyObservesThroughTheGateway` 실패 |
| M10: 어댑터가 첫 항목만 싣는다(`break`) | KILLED — `TestARefusedHandoffLeavesTheWorkerDormant` 실패. 상한 초과가 승인으로 바뀌면 이 worker 가 올라가 버린다. |

원복은 12개 전부 sha256 대조로 확인했고 `unrestored_files` 는 비어 있다.

행은 측정한 것을 말한다. 진입 0 인 arm 은 커버리지 공백이지 통과가 아니다.

## 2026-09-04 — 태스크 8.8.2 가 들어낸 두 분기

옛 B6(`activation.Verified()`)·B7(위험·ProtectionReady digest 대조)이 코드에서
사라졌으므로 행도 지웠다. 옛 B8 은 새 B6 이다.

**왜 들어냈나.** 8.5 적대 리뷰가 그 결속이 두 가지 이유로 아무것도 막지 못한다는
것을 값으로 보였다. (1) 두 값이 per-cycle 스냅샷 봉인이라 사람이 서명한 상수가
어떤 정상 입력으로도 같아질 수 없었다 — 매니페스트를 배포하면 두 시장이 영원히
dormant 가 된다. (2) 이 함수가 만드는 `Effective` 는 화면과 승격만 움직이고,
주문은 refresh worker 의 사이클이 `dispatchHandoff().Deliver` 로 내보내며 그
경로는 이 서술자를 읽지 않는다.

결속은 두 자리로 옮겼다: 넷은 제안 수집 단계(`loadFamilyActivation`, 존재하고
변하지 않는 사실), ProtectionReady 하한은 `strategyDispatchCycle.dispatch`
(보호 세대가 실제로 존재하고 주문을 거절할 수 있는 유일한 자리).

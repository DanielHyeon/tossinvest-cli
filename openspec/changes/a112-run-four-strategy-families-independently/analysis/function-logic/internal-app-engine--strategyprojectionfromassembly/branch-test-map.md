# Branch Test Map: `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go` (97-156); file SHA-256 `14ec90c888e64ccb7e45d5823f415cbf53a1b97b4a62adf9b476db478892f80a`. AST branch positions are authoritative.
- 이 태스크가 B7 을 편집했다. **어떤 시험도 이 함수에 닿지 않는다.**
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 100:2 — KR·US 순회 | 없음 | 아니오 | 아니오 — **진입 0** |
| B2 | if at 107:3 — worker 부재/비Effective | 없음 | 아니오 | 아니오 — **진입 0** |
| B3 | switch at 109:4 — 거절 코드 선택 | 없음 | 아니오 | 아니오 — **진입 0** |
| B4 | case at 110:4 — 활성화 부재 | 없음 | 아니오 | 아니오 — **진입 0** |
| B5 | case at 112:4 — 증거 묵음 | 없음 | 아니오 | 아니오 — **진입 0** |
| B6 | case at 114:4 — 보호 미배선 | 없음 | 아니오 | 아니오 — **진입 0** |
| B7 | if at 123:3 — **handoff 거절** 또는 봉인 깨진 제안 | 없음 | 아니오 | 아니오 — **진입 0** |
| B8 | if at 132:3 — 레인 증거 다이제스트 부재 | 없음 | 아니오 | 아니오 — **진입 0** |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M9: 이 함수에 `len(...entries) != 1` 사본을 되살린다 | KILLED — `TestTheSingleProposalAssumptionLivesOnlyWhereTheCensusSaysItDoes` 실패 |
| M12: `!handoff.Admitted()` 를 지워 거절 여부를 묻지 않고 result 를 읽는다 | KILLED — `TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted` 실패 |

동작 커버리지는 0 이다. 위 두 뮤테이션은 소스 구조에 대한 반증이지 동작에 대한 반증이 아니다.

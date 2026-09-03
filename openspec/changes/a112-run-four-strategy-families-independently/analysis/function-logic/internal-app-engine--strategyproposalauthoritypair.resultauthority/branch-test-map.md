# Branch Test Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go` (183-195); file SHA-256 `913050cc0cc0763295af577e49fbb4ccb7d4e838fbfc3408f0f33057fbbe2418`. AST branch positions are authoritative.
- 이 태스크가 이 함수를 편집했다. 아래 행은 실제로 돌린 시험과 뮤테이션 결과다.
| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 189:3 — handoff 가 거절했거나 봉인이 깨진 제안 | `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` (태그 스위트 1회 진입) | 아니오 — 이 태스크는 이 분기의 **동작**을 바꾸지 않았고 조건의 출처만 옮겼다 | 예 |

## 반증 실측

| 뮤테이션 | 결과 |
|---|---|
| M7: 관문을 지우고 `_ = handedOff` 로 컴파일러를 달랜다 | KILLED — `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 실패. 원복은 SHA-256 대조로 확인했다. |

행은 측정한 것을 말하고 의도한 것을 말하지 않는다.

# Branch Test Map: `TestTheRequestedCountIsCappedAtTheDocumentedMaximum`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 생성 실패 | 자체 실행 | yes (컴파일) | yes |
| B2 | 읽기 실패 | 자체 실행 | yes | yes |
| B3 | 500이 100으로 잘린다 | 자체 실행 | — (기존 동작) | yes |

이 테스트는 **엔드포인트에 간 수**만 확인한다. 그 수가 `Row.RankRequested`로도 나온다는
것 — 이 change의 새 계약 — 은 `TestTheRequestedCountIsTheCappedOneRatherThanTheOneTheCallerAsked`가
확인한다.

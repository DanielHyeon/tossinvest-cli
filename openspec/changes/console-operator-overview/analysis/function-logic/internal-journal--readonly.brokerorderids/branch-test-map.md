# Branch Test Map: `ReadOnly.BrokerOrderIDs`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 질의 실패는 빈 결과가 아니라 오류다 — 그리고 실제로는 open 단계에서 먼저 막힌다 | `TestALedgerWithoutTheAttemptTableIsRefusedAtOpenRatherThanPerQuery` | yes (테이블 미등록 시 open 성공 + 질의만 실패) | yes |
| B2 | 여러 attempt의 id를 정렬·중복 제거해 돌려준다 | `TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued` | yes (메서드 부재로 컴파일 실패) | yes |
| B3 | Scan 실패를 삼키지 않는다 | 동 테스트의 컬럼 계약(문자열 1컬럼) | — | yes |
| B4 | `rows.Err()`를 무시하지 않는다 | 동상 | — | yes |

추가 가드(분기 아님): `TestTheReadOnlyHandleHasNoWriteMethods`가 `*ReadOnly`의 메서드 집합을
열거하므로, 이 메서드는 allowlist에 명시되지 않으면 테스트가 실패한다.

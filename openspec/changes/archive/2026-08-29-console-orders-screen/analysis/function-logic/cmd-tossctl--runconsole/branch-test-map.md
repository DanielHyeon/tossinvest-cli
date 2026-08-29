# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | context 없는 실행 경로에서도 서버가 뜬다 | cobra 실행 경로(테스트는 항상 context 주입 — 방어적 분기) | no (기존 방어) | yes |
| B2 | KR 검증 기록 경로 해석 실패 → 콘솔 미기동 | verify record 해석 테스트 | no (기존 동작) | yes |
| B3 | US 검증 기록 경로 해석 실패 → 콘솔 미기동 | 동일(US) | no (기존 동작) | yes |
| B4 | soak 기록 경로 실패 → 콘솔 미기동 | soak record 해석 테스트 | no (기존 동작) | yes |
| B5 | attestation 경로 실패 → 콘솔 미기동 | 동일 | no (기존 동작) | yes |
| B6 | 원장 경로 미해석 → stderr 안내 + 화면은 "배선되지 않았다" | `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes` | no (기존 동작) | yes |
| B7 | 엔진 마커 해석 성공 → 엔진 패널이 상태를 읽는다 | 엔진 상태 패널 렌더 테스트 | no (기존 동작) | yes |
| B8 | 엔진 마커 해석 실패 → 안내 후 계속 | 동일 | no (기존 동작) | yes |

이 change 묶음이 실제로 바꾼 것은 `console.Options` 리터럴의 필드 집합과 그 앞줄의 공유 계좌 해석기 1개이며, 그 검사는 `TestTheConsoleIsHandedOneCapabilityAndNotABroker`(Orders 필수·금지 동사 목록), `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem`(GateLimits 필수), `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`(공유 resolver 1개 + 모든 읽기 seam이 그것을 받는다)가 소유한다.

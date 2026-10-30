# 변경 후 AST 분기 교차표

현재 source SHA-256 `f745d62993a7a8add2f69182eb39ba39cfc4b19becb204c43d41a6791c16dd56`, `ast-current.json`. 원본 41개 AST는 `ast.json`에 별도 보존했다. 아래 ID는 현재 AST 전용이며 원본 ID와 일치한다고 가정하지 않는다. named test는 관련 시나리오이지 개별 분기 커버리지 증명이 아니다. 명시된 미검증·fixture 제한은 유지한다. 전체 tests/gate 미실행.

| 현재 AST ID | kind | 행 | 흐름 | 관련 선택 시험 또는 미검증 |
|---|---|---:|---|---|
| B1 | if | 194 | G01 | TestAttemptClassificationAfterDrain |
| B2 | if | 199 | F01 | TestA169FenceWrongBudgetAndDirty / TestA169FenceEarlyBodyCloseFailureSticky |
| B3 | if | 200 | F01 | TestA169FenceWrongBudgetAndDirty / TestA169FenceEarlyBodyCloseFailureSticky |
| B4 | if | 201 | F01 | TestA169FenceWrongBudgetAndDirty / TestA169FenceEarlyBodyCloseFailureSticky |
| B5 | if | 203 | F01 | TestA169FenceWrongBudgetAndDirty / TestA169FenceEarlyBodyCloseFailureSticky |
| B6 | if | 211 | F01 | TestA169FenceWrongBudgetAndDirty / TestA169FenceEarlyBodyCloseFailureSticky |
| B7 | if | 217 | G02 | TestObservedTransportRateLimitAndPoison / TestA169FenceEarlyBodyCloseFailureSticky |
| B8 | if | 219 | G02 | TestObservedTransportRateLimitAndPoison / TestA169FenceEarlyBodyCloseFailureSticky |
| B9 | if | 221 | G02 | TestObservedTransportRateLimitAndPoison / TestA169FenceEarlyBodyCloseFailureSticky |
| B10 | if | 225 | G02 | TestObservedTransportRateLimitAndPoison / TestA169FenceEarlyBodyCloseFailureSticky |
| B11 | if | 227 | G02 | TestObservedTransportRateLimitAndPoison / TestA169FenceEarlyBodyCloseFailureSticky |
| B12 | if | 236 | G03 | TestTransportPreAdmissionIsNotSent / TestA169FenceEarlyBodyCloseFailureSticky |
| B13 | if | 237 | G03 | TestTransportPreAdmissionIsNotSent / TestA169FenceEarlyBodyCloseFailureSticky |
| B14 | else | 244 | G03 | TestTransportPreAdmissionIsNotSent / TestA169FenceEarlyBodyCloseFailureSticky |
| B15 | if | 239 | G03 | TestTransportPreAdmissionIsNotSent / TestA169FenceEarlyBodyCloseFailureSticky |
| B16 | if | 249 | G04 | TestProxySelectionRefusedBeforeAdmission |
| B17 | if | 251 | G04 | TestProxySelectionRefusedBeforeAdmission |
| B18 | if | 253 | G04 | TestProxySelectionRefusedBeforeAdmission |
| B19 | if | 255 | G04 | TestProxySelectionRefusedBeforeAdmission |
| B20 | if | 265 | G05 | TestA169FenceAuxBusyAndCancellationNotSticky |
| B21 | if | 266 | G05 | TestA169FenceAuxBusyAndCancellationNotSticky |
| B22 | if | 268 | G05 | TestA169FenceAuxBusyAndCancellationNotSticky |
| B23 | if | 274 | G05 | TestA169FenceAuxBusyAndCancellationNotSticky |
| B24 | if | 279 | F02 | TestA169FenceReleaseFailureConcurrentWaiter / TestA169FenceWrongBudgetAndDirty |
| B25 | if | 280 | F02 | TestA169FenceReleaseFailureConcurrentWaiter / TestA169FenceWrongBudgetAndDirty |
| B26 | if | 283 | F02 | TestA169FenceReleaseFailureConcurrentWaiter / TestA169FenceWrongBudgetAndDirty |
| B27 | if | 284 | F02 | TestA169FenceReleaseFailureConcurrentWaiter / TestA169FenceWrongBudgetAndDirty |
| B28 | if | 297 | G06 | TestObservedTransportPollRejectsPost / TestA169FenceBeginFailureSticky |
| B29 | if | 299 | G06 | TestObservedTransportPollRejectsPost / TestA169FenceBeginFailureSticky |
| B30 | else | 301 | G06 | TestObservedTransportPollRejectsPost / TestA169FenceBeginFailureSticky |
| B31 | if | 304 | G07/F03 | TestObservedTransportRateLimitAndPoison / TestA169FenceAuxBusyAndCancellationNotSticky |
| B32 | if | 307 | G07/F03 | TestObservedTransportRateLimitAndPoison / TestA169FenceAuxBusyAndCancellationNotSticky |
| B33 | if | 310 | G07/F03 | TestObservedTransportRateLimitAndPoison / TestA169FenceAuxBusyAndCancellationNotSticky |
| B34 | if | 313 | G07/F03 | TestObservedTransportRateLimitAndPoison / TestA169FenceAuxBusyAndCancellationNotSticky |
| B35 | if | 318 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B36 | if | 320 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B37 | if | 322 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B38 | if | 327 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B39 | if | 332 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B40 | if | 333 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B41 | if | 341 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B42 | else | 343 | G08/F03 | TestObservedTransportBeginFaultClosesBody / TestA169FenceBeginFailureSticky |
| B43 | if | 359 | G10 | 미검증: SplitHostPort 오류 |
| B44 | if | 364 | G11 | TestTransportOwnsFullBody; nil TLS config 개별 미검증 |
| B45 | else | 366 | G11 | TestTransportOwnsFullBody; nil TLS config 개별 미검증 |
| B46 | if | 372 | G12 | TestA169FenceAuxBusyAndCancellationNotSticky; default dial 미검증 |
| B47 | if | 376 | G12 | TestA169FenceAuxBusyAndCancellationNotSticky; default dial 미검증 |
| B48 | if | 382 | G13 | TestCancelBlockedTLSWriteDrains / TestAttemptClassificationAfterDrain |
| B49 | if | 386 | G13 | TestCancelBlockedTLSWriteDrains / TestAttemptClassificationAfterDrain |
| B50 | if | 390 | G13 | TestCancelBlockedTLSWriteDrains / TestAttemptClassificationAfterDrain |
| B51 | if | 396 | G14/F04 | TestA169FenceNilResponseCleanupSticky / TestObservedTransportNoResponseAfterSent |
| B52 | if | 397 | G14/F04 | TestA169FenceNilResponseCleanupSticky / TestObservedTransportNoResponseAfterSent |
| B53 | if | 404 | G14/F04 | TestA169FenceNilResponseCleanupSticky / TestObservedTransportNoResponseAfterSent |
| B54 | if | 409 | G14/F04 | TestA169FenceNilResponseCleanupSticky / TestObservedTransportNoResponseAfterSent |
| B55 | if | 414 | G14/F04 | TestA169FenceNilResponseCleanupSticky / TestObservedTransportNoResponseAfterSent |
| B56 | else | 417 | G14/F04 | TestA169FenceNilResponseCleanupSticky / TestObservedTransportNoResponseAfterSent |
| B57 | if | 425 | G15/F05 | TestA169FencePostcommitFinishConcurrentWaiter / TestA169FenceTransportAfterEventAndOwn429 |
| B58 | if | 428 | G15/F05 | TestA169FencePostcommitFinishConcurrentWaiter / TestA169FenceTransportAfterEventAndOwn429 |
| B59 | if | 430 | G15/F05 | TestA169FencePostcommitFinishConcurrentWaiter / TestA169FenceTransportAfterEventAndOwn429 |
| B60 | if | 434 | G15/F05 | TestA169FencePostcommitFinishConcurrentWaiter / TestA169FenceTransportAfterEventAndOwn429 |
| B61 | if | 441 | G15/F05 | TestA169FencePostcommitFinishConcurrentWaiter / TestA169FenceTransportAfterEventAndOwn429 |
| B62 | if | 442 | G15/F05 | TestA169FencePostcommitFinishConcurrentWaiter / TestA169FenceTransportAfterEventAndOwn429 |
| B63 | if | 450 | G16 | TestTransportOwnsFullBody; TLS 보완 개별 미검증 |
| B64 | if | 456 | G17/F05 | TestA169FenceBodyFailureConcurrentWaiter / TestCancelUnreadResponseReleasesOwnership |
| B65 | if | 462 | G17/F05 | TestA169FenceBodyFailureConcurrentWaiter / TestCancelUnreadResponseReleasesOwnership |
| B66 | else | 465 | G17/F05 | TestA169FenceBodyFailureConcurrentWaiter / TestCancelUnreadResponseReleasesOwnership |
| B67 | select | 473 | G17/F05 | TestA169FenceBodyFailureConcurrentWaiter / TestCancelUnreadResponseReleasesOwnership |
| B68 | if | 475 | G17/F05 | TestA169FenceBodyFailureConcurrentWaiter / TestCancelUnreadResponseReleasesOwnership |

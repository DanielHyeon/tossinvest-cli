# Branch Test Map: `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | verifier가 잘못 nil 반환 | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` | 기존 방어, 이번 RED 대상 아님 | 패키지 시험 PASS (156.072s) |
| B2 | verifier가 digest와 다른 오류 문구 반환 | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` | 기존 동작이 digest 문구여서 통과 | 패키지 시험 PASS (156.072s) |

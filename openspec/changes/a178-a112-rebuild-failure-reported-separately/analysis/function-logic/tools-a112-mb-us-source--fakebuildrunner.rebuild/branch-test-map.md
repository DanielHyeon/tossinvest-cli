# Branch Test Map: `fakeBuildRunner.rebuild`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | command 생략 시 prescribed command를 유지하고, 주입 시 그 값을 반환하며 injected error도 그대로 반환 | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch`; `TestVerifyPrescribedBinaryReportsUnexpectedCommandSeparately`; `TestVerifyPrescribedBinaryReportsRebuildFailureSeparately` | zero-value fixture는 통과, command/error 주입 consumer는 RED에서 digest mismatch를 받아 실패 | 패키지 시험 PASS (156.072s) |

# Branch Test Map: `verifyPrescribedBinary`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 입력 또는 builder 누락 | 기존 직접 시험 없음 | 해당 없음 | 해당 없음 |
| B2 | 메모리 fake builder의 zero Go identity | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` | 통과 — fake path 유지 | 패키지 시험 PASS (156.072s) |
| B3 | Go execution snapshot 결속 불일치 | 기존 직접 시험 없음 | 해당 없음 | 해당 없음 |
| B4 | private Go distribution 결속 불일치 | 기존 직접 시험 없음 | 해당 없음 | 해당 없음 |
| B5 | rebuild 전 Go binary drift | 기존 직접 시험 없음 | 해당 없음 | 해당 없음 |
| B6 | builder가 `context.DeadlineExceeded` 반환 | `TestVerifyPrescribedBinaryReportsRebuildFailureSeparately` | 실패 — digest mismatch로 잘못 보고됨 (`analysis/red-output.txt`) | 패키지 시험 PASS (156.072s) |
| B7 | builder가 예상 밖 명령 반환 | `TestVerifyPrescribedBinaryReportsUnexpectedCommandSeparately` | 실패 — digest mismatch로 잘못 보고됨 (`analysis/red-output.txt`) | 패키지 시험 PASS (156.072s) |
| B8 | rebuilt digest 불일치 | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` | 통과 — 기존 digest 문구 유지 | 패키지 시험 PASS (156.072s) |
| B9 | rebuild 후 Go binary drift | 기존 직접 시험 없음 | 해당 없음 | 해당 없음 |

# Branch Test Map: `Console.readAttestation`

Source: `internal/console/data.go` lines 266-306

| Branch | 현재 AST의 조건/역할 | 연결 테스트 | 실제 증거와 한계 |
|---|---|---|---|
| B1 | 설정된 경로가 공백 | TestDashboardMismatchedRenewalStatusAndBlankPathAreUnknownWithoutZeroTimestamp | unknown/사용 불가 통과 |
| B2 | attestation 로드 오류 | TestDashboardRedactsMalformedAttestationContents; TestTheDashboardReportsAnUnstartedMachineWithoutFailing | 손상/누락 통과 |
| B3 | 로드 오류가 단순 누락이 아님 | TestDashboardRedactsMalformedAttestationContents | 고정 문구/원문 비노출 통과 |
| B4 | attestation 내용의 계좌/endpoint 검사 switch | TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability | 정상 유효성 보존 통과 |
| B5 | 계좌 참조 공백 | TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability | 정상 값 통과; 공백 계좌 직접 주입 없음 |
| B6 | endpoint가 하나도 없음 | TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability | 정상 값 통과; 빈 목록 직접 주입 없음 |
| B7 | attestation 만료 | TestRenewalWarningBoundariesAreInclusiveAt72HoursAndStaleOnlyAfter12Hours | 경고 경계 통과; 정확한 만료 분기는 별도 계측 없음 |
| B8 | 만료 시각이 유효하고 현재 이후여서 남은 시간 계산 | TestRenewalWarningBoundariesAreInclusiveAt72HoursAndStaleOnlyAfter12Hours | 72시간 경계/시간 표시 통과 |
| B9 | 필수 endpoint 누락 사유 반복 | TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability | 누락 없는 경로 통과; 개별 누락 분기 계측 없음 |

이 표는 현재 AST의 모든 분기 ID를 연결한다. 테스트가 통과했다는 사실과 개별 true/false 분기의 실행 여부는 다르다. 직접 주입/계측하지 않은 경로는 위에 명시했으며 검증된 것으로 소급하지 않는다. `analysis/regression-proof.md`의 warning 변이 실패는 회귀 민감도 증거다.

# Branch Test Map: `BuildAttestation`

Source: `internal/soak/attest.go` lines 220-273

| Branch | 현재 AST의 조건/역할 | 연결 테스트 | 실제 증거와 한계 |
|---|---|---|---|
| B1 | 단일 평가의 issue 목록이 비어 있지 않음 | TestBuildAttestationRefusesAnIncompleteSoak | typed 오류/ErrIncomplete 호환 통과 |
| B2 | 성공 endpoint를 반복 검사 | TestBuildAttestationPassesTheEngineInterlock | 통과 |
| B3 | 성공 endpoint가 GET이 아님 | TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise | 기존 회귀 통과; 개별 비-GET 주입 여부 별도 계측 없음 |
| B4 | 감독 증거 수용 오류 | TestSupervisedEvidenceFromAnotherAccountRefusesTheIssue; TestSupervisedEvidenceFromTheFutureIsNotEvidence | 거부 통과 |
| B5 | 수용된 감독 증거 endpoint 결합 | TestSupervisedEvidenceCompletesTheEnginesRequiredSet | 통과 |
| B6 | 추가 operator 메모가 존재 | TestBuildAttestationCarriesTheMeasuredRate | 회귀 통과; 별도 메모 분기 계측 없음 |

이 표는 현재 AST의 모든 분기 ID를 연결한다. 테스트가 통과했다는 사실과 개별 true/false 분기의 실행 여부는 다르다. 직접 주입/계측하지 않은 경로는 위에 명시했으며 검증된 것으로 소급하지 않는다. `analysis/regression-proof.md`의 warning 변이 실패는 회귀 민감도 증거다.

# Branch Test Map: `runSoakAttest`

Source: `cmd/tossctl/soak.go` lines 476-582

| Branch | 현재 AST의 조건/역할 | 연결 테스트 | 실제 증거와 한계 |
|---|---|---|---|
| B1 | 진단 기록 옵션 ON/OFF | TestSoakAttestDefaultDoesNotWriteRenewalStatus; TestSoakAttestRecordsBoundedRefusalStatus | ON/OFF 통과 |
| B2 | 기록 또는 출력 경로 override가 있음 | TestSoakAttestRenewalStatusRejectsPathOverrides | 거부 경로 통과 |
| B3 | 초기 프로필 경로 해석 오류 | TestSoakAttestRecordingUsesOneResolvedAttestationPathSnapshot | 정상 해석 통과; 오류 주입은 직접 검증하지 않음 |
| B4 | defer 시 발급 결과가 성공임 | TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation; TestSoakAttestRecordsBoundedRefusalStatus | 성공/거부 통과 |
| B5 | 발급한 attestation 재읽기 오류 | TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation | 정상 읽기 통과; 재읽기 오류 직접 주입 없음 |
| B6 | 발급 파일 읽기 성공 후 만료 시각 복사 | TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation | 통과 |
| B7 | 진단 저장 오류이며 기존 결과는 성공 | TestSoakAttestStatusWriteFailureKeepsIssuedAttestationAndFailsCommand | nonzero와 발급 파일 보존 통과 |
| B8 | soak 요약 로드 오류 | TestSoakAttestRecordsBoundedRefusalStatus | 정상 로드 후 거부 통과; 로드 오류 직접 주입 없음 |
| B9 | 사용자 지정 유효 기간이 양수 | TestSoakAttestWritesAVerifiableAttestation | 기존 발급 테스트 통과; 양수 override 분기 계측 없음 |
| B10 | surveyed base가 존재해 메모에 결합 | TestSoakAttestWritesAVerifiableAttestation | 회귀 통과; 이 분기 별도 계측 없음 |
| B11 | 감독 검증 증거 로드/검증 오류 | TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount | 거부 통과 |
| B12 | 자격 판정 또는 증거 결합 거부 | TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing; TestSoakAttestRecordsBoundedRefusalStatus | 거부와 bounded 코드 통과 |
| B13 | 진단 OFF일 때만 기존 시점에서 경로 해석 | TestSoakAttestDefaultDoesNotWriteRenewalStatus; TestSoakAttestRecordingUsesOneResolvedAttestationPathSnapshot | OFF 보존/ON 단일 해석 통과 |
| B14 | 최종 출력 경로 해석 오류 | TestSoakAttestWritesAVerifiableAttestation | 정상 경로 통과; 오류 직접 주입 없음 |
| B15 | 출력 디렉터리 생성 오류 | TestSoakAttestWritesAVerifiableAttestation | 정상 생성 통과; 생성 오류 직접 주입 없음 |
| B16 | attestation 저장 오류 | TestSoakAttestWritesAVerifiableAttestation | 정상 저장 통과; attestation 저장 오류 직접 주입 없음 |
| B17 | 감독 증거 출력 반복 | TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun | 회귀 통과 |
| B18 | 출력할 감독 증거 market이 비어 있음 | TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun | 출력 회귀 통과; 빈 market 분기 계측 없음 |
| B19 | 미검증 live endpoint가 남아 있음 | TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn | 안내 후 정상 발급 반환 통과 |
| B20 | 미검증 endpoint 안내 반복 | TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn | 회귀 통과 |

이 표는 현재 AST의 모든 분기 ID를 연결한다. 테스트가 통과했다는 사실과 개별 true/false 분기의 실행 여부는 다르다. 직접 주입/계측하지 않은 경로는 위에 명시했으며 검증된 것으로 소급하지 않는다. `analysis/regression-proof.md`의 warning 변이 실패는 회귀 민감도 증거다.

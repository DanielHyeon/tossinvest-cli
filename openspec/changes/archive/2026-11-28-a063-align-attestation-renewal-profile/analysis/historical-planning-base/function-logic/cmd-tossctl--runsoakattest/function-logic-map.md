# Function Logic Map: `runSoakAttest`

- Source: `cmd/tossctl/soak.go` lines 476-582
- AST evidence: `ast.json` (20 branches)
- 2026-09-06 Manager 사후 문서 검토. 현재 AST와 테스트 결과를 정렬한 문서이며 새 pre-edit 증거로 소급하지 않는다.

## Inputs and invariants

기록 ON은 처음 해석한 attestation 경로 하나를 발급과 상태 저장에 재사용한다. OFF는 기존 늦은 경로 해석과 반환을 보존한다. 자격 판정은 한 번이며, 상태 저장 오류는 성공을 실패로 바꿀 수 있어도 발급 파일을 지우지 않는다. broker 호출이나 운영 토글 변경은 없다.

## Branches and early returns

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

## Calls and live bindings

| AST line | Callee | Evidence |
|---|---|---|
| 480 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 480 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 481 | `fmt.Errorf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 483 | `resolveAttestationPath` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 488 | `soak.RenewalStatusPath` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 489 | `UTC` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 489 | `time.Now` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 490 | `익명 defer 호출` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 491 | `renewalStatusForResult` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 495 | `attest.Load` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 496 | `fmt.Errorf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 501 | `saveRenewalStatus` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 502 | `fmt.Errorf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 506 | `loadSoakSummary` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 515 | `soakSurveyedBase` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 516 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 519 | `UTC` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 519 | `time.Now` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 520 | `supervisedProofs` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 525 | `soak.BuildAttestation` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 534 | `resolveAttestationPath` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 539 | `os.MkdirAll` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 539 | `filepath.Dir` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 540 | `fmt.Errorf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 540 | `filepath.Dir` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 542 | `attest.Save` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 546 | `cmd.OutOrStdout` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 547 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 548 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 549 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 549 | `attest.Mask` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 550 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 551 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 551 | `attestation.ExpiresAt.Format` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 552 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 552 | `strings.Join` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 553 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 560 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 561 | `Format` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 561 | `p.At.UTC` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 564 | `attestation.MissingEndpoints` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 564 | `soak.LiveOnlyEndpoints` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 565 | `len` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 566 | `fmt.Fprintln` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 567 | `fmt.Fprintln` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 569 | `fmt.Fprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 577 | `fmt.Fprintln` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 578 | `fmt.Fprintln` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 579 | `fmt.Fprintln` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 580 | `fmt.Fprintln` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |

## State mutations and fallbacks

기록 ON은 처음 해석한 attestation 경로 하나를 발급과 상태 저장에 재사용한다. OFF는 기존 늦은 경로 해석과 반환을 보존한다. 자격 판정은 한 번이며, 상태 저장 오류는 성공을 실패로 바꿀 수 있어도 발급 파일을 지우지 않는다. broker 호출이나 운영 토글 변경은 없다.

실제 오류 반환/로컬 view/파일 저장 경계는 현재 AST와 연결 테스트를 기준으로 검토했다. 표의 직접 주입 없음/계측 없음은 실행하지 않은 분기를 뜻한다. 관련 테스트의 성공을 모든 분기의 실행 증거로 바꾸지 않는다.

## Safety conclusion

라이브 주문·설정 변경·엔진 재시작은 수행하지 않았다. a063 독립 적대적 리뷰와 gstack 코드 리뷰는 통과했지만 원래 기준 커밋의 전역 FLM 및 운영 수용은 별도 차단 상태다. `analysis/verification.md`와 `review.md`의 실행 범위 및 한계를 따른다.

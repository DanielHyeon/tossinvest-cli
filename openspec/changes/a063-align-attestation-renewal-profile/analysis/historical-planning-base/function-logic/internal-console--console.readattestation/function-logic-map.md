# Function Logic Map: `Console.readAttestation`

- Source: `internal/console/data.go` lines 266-306
- AST evidence: `ast.json` (9 branches)
- 2026-09-06 Manager 사후 문서 검토. 현재 AST와 테스트 결과를 정렬한 문서이며 새 pre-edit 증거로 소급하지 않는다.

## Inputs and invariants

현재 attestation만 Usable의 근거다. 갱신 상태와 72시간 경고는 별도 advisory이며 엔진 동작이나 인터록을 바꾸지 않는다. 파일 오류는 고정 문구로 표시한다. 경로가 없으면 unknown으로 시작한다.

## Branches and early returns

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

## Calls and live bindings

| AST line | Callee | Evidence |
|---|---|---|
| 268 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 271 | `attest.Load` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 273 | `errors.Is` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 274 | `append` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 276 | `v.readRenewalStatus` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 281 | `attest.Mask` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 285 | `a.MissingEndpoints` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 288 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 289 | `append` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 290 | `len` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 291 | `append` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 293 | `a.Expired` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 294 | `append` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 296 | `a.ExpiresAt.IsZero` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 296 | `a.ExpiresAt.Before` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 296 | `a.ExpiresAt.Sub` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 297 | `a.ExpiresAt.IsZero` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 297 | `a.ExpiresAt.After` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 298 | `humanHorizon` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 298 | `a.ExpiresAt.Sub` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 301 | `append` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 303 | `len` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 304 | `v.readRenewalStatus` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |

## State mutations and fallbacks

현재 attestation만 Usable의 근거다. 갱신 상태와 72시간 경고는 별도 advisory이며 엔진 동작이나 인터록을 바꾸지 않는다. 파일 오류는 고정 문구로 표시한다. 경로가 없으면 unknown으로 시작한다.

실제 오류 반환/로컬 view/파일 저장 경계는 현재 AST와 연결 테스트를 기준으로 검토했다. 표의 직접 주입 없음/계측 없음은 실행하지 않은 분기를 뜻한다. 관련 테스트의 성공을 모든 분기의 실행 증거로 바꾸지 않는다.

## Safety conclusion

라이브 주문·설정 변경·엔진 재시작은 수행하지 않았다. a063 독립 적대적 리뷰와 gstack 코드 리뷰는 통과했지만 원래 기준 커밋의 전역 FLM 및 운영 수용은 별도 차단 상태다. `analysis/verification.md`와 `review.md`의 실행 범위 및 한계를 따른다.

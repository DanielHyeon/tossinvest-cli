# Function Logic Map: `BuildAttestation`

- Source: `internal/soak/attest.go` lines 220-273
- AST evidence: `ast.json` (6 branches)
- 2026-09-06 Manager 사후 문서 검토. 현재 AST와 테스트 결과를 정렬한 문서이며 새 pre-edit 증거로 소급하지 않는다.

## Inputs and invariants

기존 자격 조건과 감독 증거 규칙을 유지하며 한 번의 evaluateIssues 결과를 사용한다. 불완전하면 IncompleteError가 ErrIncomplete와 기존 문구를 보존한다. 이 함수는 메모리상의 attestation만 반환하며 파일을 쓰지 않는다.

## Branches and early returns

| Branch | 현재 AST의 조건/역할 | 연결 테스트 | 실제 증거와 한계 |
|---|---|---|---|
| B1 | 단일 평가의 issue 목록이 비어 있지 않음 | TestBuildAttestationRefusesAnIncompleteSoak | typed 오류/ErrIncomplete 호환 통과 |
| B2 | 성공 endpoint를 반복 검사 | TestBuildAttestationPassesTheEngineInterlock | 통과 |
| B3 | 성공 endpoint가 GET이 아님 | TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise | 기존 회귀 통과; 개별 비-GET 주입 여부 별도 계측 없음 |
| B4 | 감독 증거 수용 오류 | TestSupervisedEvidenceFromAnotherAccountRefusesTheIssue; TestSupervisedEvidenceFromTheFutureIsNotEvidence | 거부 통과 |
| B5 | 수용된 감독 증거 endpoint 결합 | TestSupervisedEvidenceCompletesTheEnginesRequiredSet | 통과 |
| B6 | 추가 operator 메모가 존재 | TestBuildAttestationCarriesTheMeasuredRate | 회귀 통과; 별도 메모 분기 계측 없음 |

## Calls and live bindings

| AST line | Callee | Evidence |
|---|---|---|
| 222 | `c.withDefaults` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 224 | `s.evaluateIssues` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 225 | `len` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 229 | `s.Window.SuccessfulEndpoints` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 235 | `strings.HasPrefix` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 235 | `strings.ToUpper` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 235 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 236 | `fmt.Errorf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 242 | `acceptSupervised` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 247 | `append` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 250 | `fmt.Sprintf` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 254 | `Format` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 254 | `s.WindowStart.UTC` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 254 | `Format` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 254 | `s.LastAt.UTC` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 256 | `mutationNote` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 257 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 258 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 264 | `now.UTC` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 265 | `Add` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 265 | `now.UTC` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 269 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |

## State mutations and fallbacks

기존 자격 조건과 감독 증거 규칙을 유지하며 한 번의 evaluateIssues 결과를 사용한다. 불완전하면 IncompleteError가 ErrIncomplete와 기존 문구를 보존한다. 이 함수는 메모리상의 attestation만 반환하며 파일을 쓰지 않는다.

실제 오류 반환/로컬 view/파일 저장 경계는 현재 AST와 연결 테스트를 기준으로 검토했다. 표의 직접 주입 없음/계측 없음은 실행하지 않은 분기를 뜻한다. 관련 테스트의 성공을 모든 분기의 실행 증거로 바꾸지 않는다.

## Safety conclusion

라이브 주문·설정 변경·엔진 재시작은 수행하지 않았다. a063 독립 적대적 리뷰와 gstack 코드 리뷰는 통과했지만 원래 기준 커밋의 전역 FLM 및 운영 수용은 별도 차단 상태다. `analysis/verification.md`와 `review.md`의 실행 범위 및 한계를 따른다.

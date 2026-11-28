# Function Logic Map: `Summary.Evaluate`

- Source: `internal/soak/attest.go` lines 123-133
- AST evidence: `ast.json` (2 branches)
- 2026-09-06 Manager 사후 문서 검토. 현재 AST와 테스트 결과를 정렬한 문서이며 새 pre-edit 증거로 소급하지 않는다.

## Inputs and invariants

evaluateIssues가 유일한 자격 판정이다. 공개 Evaluate는 같은 순서의 메시지를 반환하고 성공 시 기존처럼 true,nil이다. 새 코드를 얻기 위해 판정을 다시 실행하지 않는다.

## Branches and early returns

| Branch | 현재 AST의 조건/역할 | 연결 테스트 | 실제 증거와 한계 |
|---|---|---|---|
| B1 | 단일 issue 평가의 메시지 순서대로 복사 | TestEvaluateRefusesAnEmptyRecord; TestEvaluateRefusesAShortStreak; TestEvaluateRefusesWhenARequiredEndpointNeverSucceeded | 기존 거부 메시지 경로 통과 |
| B2 | 이유가 없으면 true와 nil 반환 | TestEvaluatePreservesNilReasonsForAQualifyingSoak; TestEvaluateAcceptsACompletedSoak | 정상 nil 호환 통과 |

## Calls and live bindings

| AST line | Callee | Evidence |
|---|---|---|
| 124 | `s.evaluateIssues` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 125 | `make` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 125 | `len` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 127 | `append` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 129 | `len` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |

## State mutations and fallbacks

evaluateIssues가 유일한 자격 판정이다. 공개 Evaluate는 같은 순서의 메시지를 반환하고 성공 시 기존처럼 true,nil이다. 새 코드를 얻기 위해 판정을 다시 실행하지 않는다.

실제 오류 반환/로컬 view/파일 저장 경계는 현재 AST와 연결 테스트를 기준으로 검토했다. 표의 직접 주입 없음/계측 없음은 실행하지 않은 분기를 뜻한다. 관련 테스트의 성공을 모든 분기의 실행 증거로 바꾸지 않는다.

## Safety conclusion

라이브 주문·설정 변경·엔진 재시작은 수행하지 않았다. a063 독립 적대적 리뷰와 gstack 코드 리뷰는 통과했지만 원래 기준 커밋의 전역 FLM 및 운영 수용은 별도 차단 상태다. `analysis/verification.md`와 `review.md`의 실행 범위 및 한계를 따른다.

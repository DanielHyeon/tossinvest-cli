# Branch Test Map: `Summary.Evaluate`

Source: `internal/soak/attest.go` lines 123-133

| Branch | 현재 AST의 조건/역할 | 연결 테스트 | 실제 증거와 한계 |
|---|---|---|---|
| B1 | 단일 issue 평가의 메시지 순서대로 복사 | TestEvaluateRefusesAnEmptyRecord; TestEvaluateRefusesAShortStreak; TestEvaluateRefusesWhenARequiredEndpointNeverSucceeded | 기존 거부 메시지 경로 통과 |
| B2 | 이유가 없으면 true와 nil 반환 | TestEvaluatePreservesNilReasonsForAQualifyingSoak; TestEvaluateAcceptsACompletedSoak | 정상 nil 호환 통과 |

이 표는 현재 AST의 모든 분기 ID를 연결한다. 테스트가 통과했다는 사실과 개별 true/false 분기의 실행 여부는 다르다. 직접 주입/계측하지 않은 경로는 위에 명시했으며 검증된 것으로 소급하지 않는다. `analysis/regression-proof.md`의 warning 변이 실패는 회귀 민감도 증거다.

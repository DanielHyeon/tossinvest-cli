# 코드 영향 기준선

## 원문 확인

- `internal/console/optimization.go:66`의 `Console.handleOptimization`은 `/optimization` 라우트에서 한 번만 호출됨.
- `internal/console/optimization_view.go`의 `optimizationPage`는 템플릿 입력이며, 새 파일의 새 메서드로 파생 읽기 모델을 제공할 수 있음.
- `internal/console/templates_optimization.go`는 최적화·runtime 템플릿을, `templates_signals.go`는 신호 템플릿을 정의함.
- `internal/console/templates.go`의 `style` define 끝은 a174 CSS 블록 뒤이며, a175 CSS는 이 위치에 한 블록만 추가해야 함.

## 변경 경계

| 대상 | 변경 | 기존 함수 본문 |
|---|---|---|
| `optimization_matrix_view.go` | 프리셋·후보 행렬의 새 표시 모델 | 없음 — 새 파일·새 메서드만 |
| `templates_optimization.go` | 최적화·runtime HTML 재배치 | 해당 없음 — const 템플릿 |
| `templates_signals.go` | 공통 문장과 설명 접기 | 해당 없음 — const 템플릿 |
| `templates.go` | a175 CSS 한 블록 추가 | 해당 없음 — const 템플릿 |
| 기존 후보/모션 시험 | 바뀐 DOM 역할 검증 | `TestCandidateFilterCardsContainExactMarketMetricMatrix`, `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`, `TestAmbientMotionContract`만 사전 FLM 번들 작성 |

핸들러·레지스트리·값·POST form 계약은 변경 대상이 아님. 예상 밖의 기존 production 함수 편집이 필요하면 중단하고 Manager에게 보고함.

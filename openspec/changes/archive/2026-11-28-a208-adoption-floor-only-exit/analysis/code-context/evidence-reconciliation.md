# Evidence reconciliation — a208

기준 commit `baeddfe31`. 세 출처: CodeGraph 1.6.0(hard) · CodeGraphContext(supporting) · 현재 HEAD 원문 + `tools/logic-map` AST(`../ast-evidence/`).

| 질문 | CodeGraph | CGC | HEAD(rg · 원문) | 결론 |
|---|---|---|---|---|
| `SaveEngineAdoption` · `LoadRawEngineAdoption` 호출자 | 없음 | `Save` | `cmd/tossctl/adoptionsettings.go:59` · `:63` 가 부름 | HEAD 가 정본: 콘솔 저장 seam 이 두 함수를 부름. CodeGraph 가 `*config.Service` 메서드 호출 엣지를 못 맺은 것(빈 그래프 아님). 편집 표면에 포함 |
| `EvaluateLadderSnapshot` 생산 호출자 | judgeLadder · ValidateRecoveryDerivation | — | `rg 'EvaluateLadderSnapshot\('` 생산 2 곳 | 일치. 손절 · 청산 일괄(internal/flatten) · 브로커 보호(internal/protection) 는 사다리 평가를 부르지 않음 |
| 편입 정책 선택 자리 | adoptOne ← adopt | adopt | adoption.go:392 한 자리 | 일치(claims C1) |
| 판정 시 사다리 표 | ladderFor → CommonLadderForPosition | — | exitloop.go:1324 | 일치(claims C4) |
| 저장 스냅숏 재검증 | ValidateRecoveryDerivation ← StoredExitSnapshot.validate | — | exit_snapshot.go:182 · decode 는 `DisallowUnknownFields`(:253) | 일치. 새 입력 필드의 롤백 위험 근거(claims C11) |
| a166 배선 census 에 편집 예정 함수가 있는가 | — | — | `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 에 `Context.ExitObserver` · `Context.ReconcileDriver` 없음(engine.go `NewContext` · `NewOrderPath` 만) | 배선 한 줄 편집이 a166 census 를 깨지 않음 |

불일치 1 건(SaveEngineAdoption 호출자)은 HEAD 로 해소됨 — 편집 차단 사유 없음.

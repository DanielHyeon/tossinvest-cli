# Evidence reconciliation — a197 (tasks 1.2 · 1.3)

- 기준 커밋 `064f59fd9`. hard evidence = CodeGraph 1.6.0 + 현재 HEAD + logic-map AST(`analysis/cited-ast/`). supporting = CodeGraphContext 0.5.1(색인 0) · GBrain(낡음).

| 항목 | CodeGraph | AST · 현재 HEAD | 결론 |
|---|---|---|---|
| `alertProposalRefused` 호출 | 함수 1(`submit`) | 자리 **2**(`:1530` · `:1607`) | 자리 수는 AST. D9 편집 표면 = 2 자리 |
| `IssueReduction` | 이름 하나로 4 호출자 | **두 함수**: `execgw.Issuer.IssueReduction(ReductionRequest)`(flatten) · `RiskGuardian.IssueReduction(ReductionIssuance)`(exit 루프 2 자리) | CodeGraph 가 이름으로 합친 것을 AST · 타입으로 가름. **이 구분이 a197 D6 의 성립 여부를 정한다**(probe) |
| `submit` 호출자 | 4(동명 `Gateway.submit` 3 포함) | `ExitObserver.submit` 의 호출 1(`record:1453`) | 이름 충돌 제거 |
| `LiveOrdersForSymbol` 정의 | 호출자 3 | SQL: CONFIRMED · 종결 증거 없음 · 소유 유일 | 귀속 판정에는 못 씀 — (c) |
| 길 치우기의 생산 취소 | (CodeGraph 는 결합 동작을 말하지 않음) | probe: `RiskGuardian` 결정으로 `Gateway.Cancel` = `Prepare` 거절, 브로커 0 | **불일치가 아니라 새 사실** — 기존 a094 경로가 생산에서 취소를 못 냄. 문서·시험이 그것을 잰 적 없음(가짜 제출자) |
| GBrain `LiveOrdersForSymbol` 좌표 | — | HEAD `fills.go:1854-1919` vs GBrain `1849-1927` | GBrain 낡음. HEAD 채택 |
| CodeGraphContext | 색인 0 | — | 보조 문맥 없음, 판정 영향 없음 |

## 편집 차단 판정

WORKFLOW §5 「불일치는 현재 HEAD 와 시험으로 해소하기 전까지 편집을 차단한다」 — 도구 간 불일치는 전부 HEAD 로 해소됐다. 남은 것은 도구 불일치가 아니라 **계약(D2 · D6 · D8)과 현재 코드의 불일치**이며 `analysis/cited-ast/README.md` 의 「Manager 결정 필요」 항목이다. 그 결정 전에는 Pre-Edit(2.1)로 가지 않는다.

# a168 CodeGraph baseline (tasks 0.3)

- 도구: codegraph **1.6.0**(`codegraph --version`), worktree `~/tossos-work/a168-impl`, HEAD = base `e57d06cd6`. 색인은 이 worktree 의
  `make sdd-sync`(2026-11-25)가 새로 만든 것 — 질의 시점 `codegraph status .` 의 미반영은 이 change 가 새로 둔 분석 파일 4개뿐(Go 소스 아님).
- 원문: `codegraph-queries.txt`(callers·callees·impact × 9 심볼), `codegraph-affected-tests.txt`(`codegraph affected
  internal/official/reconcile_reads.go internal/official/reconcile_detail_reads.go --filter '*_test.go'`).

## 결과 요약

| 심볼 | 생산 호출자(CodeGraph) | 시험 호출자 |
| --- | --- | --- |
| `DecodeReconcileConditionalPage` | `Client.ReconcileConditionalOrdersPage` 1 | 4 |
| `DecodeReconcileOrderPage` | `Client.ReconcileOpenOrdersPage` 1 | 2 |
| `decodeReconcileEnvelope` | 위 둘 | — |
| `reconcileObject` | `DecodeReconcileConditionalDetail` 1 | — |
| `unwrapAndDecode` | `Client.send` 1 (impact: `deleteAcct`·`getWithHeaders`·`postWithHeaders` → 모든 공식 호출) | a166 시험 |
| `Client.ReconcileConditionalOrdersPage` · `ReconcileOpenOrdersPage` · `ReconcileInstrument` · `ReconcileConditionalOrder` | cmd `reconcileReads` 위임 1씩(+ verifylive `ReconcileReader` 인터페이스 선언) | seam·fake·HTTP 시험 |

- 생산 사슬은 `official.Client.Reconcile*` ← `cmd/tossctl reconcileReads.*` ← `verifylive.ReconcileReader` 호출 자리(`readReconcile*Group` ·
  `checkReconcileInstrument` · `readReconcileByID`) ← `verifylive.Reconcile` 하나다. 엔진·`engine reconcile-resolve`·주문 경로 호출 0.
- `codegraph affected --filter '*_test.go'`: 시험 파일 **817**(app/engine 174 · journal 107 · cmd/tossctl 95 · execgw 47 · verifylive 42 ·
  official 42 …). 이는 파일 단위 전이 의존(`internal/official` 패키지를 import 하는 모든 패키지)이고 심볼 단위 영향은 위 사슬뿐이다 —
  선택 시험은 official · verifylive · cmd/tossctl(design §8). 기본 판별식(필터 없음)은 Go 시험을 못 찾으므로 쓰지 않았다(.claude/CLAUDE.md 하네스 주석).

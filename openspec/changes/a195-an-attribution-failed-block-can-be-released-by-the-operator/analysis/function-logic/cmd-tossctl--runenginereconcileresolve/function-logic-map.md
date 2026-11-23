# Function Logic Map: `runEngineReconcileResolve` (편집 뒤 판)

- Source: `cmd/tossctl/engine_reconcile.go` (109-216), 작업 트리 sha `d5b36622a940…`. 편집 전: `../../pre-edit-ast/cmd-tossctl--runenginereconcileresolve.ast.json`(107-214, 분기 24). 재번호 `renumber.txt` — B1~B24 전부 유지(줄 +2, 도움말 2줄 증가 때문).
- AST evidence: `ast.json`(분기 24, return 21, defer 2). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file cmd/tossctl/engine_reconcile.go --func runEngineReconcileResolve`.
- Risk scan: `risk-pattern-report.md`(일치 0). 커버리지: `../../coverage/pre-edit-runEngineReconcileResolve-and-cmd-blocks.txt`.
- 편집(Manager 결정 Q5·N2): 성공 출력 문장(옛 :212 → 새 :214)을 원인 중립으로, 원인 목록 출력 없음. 분기·return·호출 순서는 바꾸지 않는다. a166 census 키 없음(shapes.json 의 engine_reconcile.go 키는 `buildReconcileResolveRuntime` 하나).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.confirm/operator/note` | true / 비어 있지 않음 | cobra 플래그 | B1~B3 거절, 잠금 전 |
| `deps` | 네 필드 non-nil | `productionReconcileResolveDeps` | B4 거절 |
| 불변식 1 | 잠금(B7) → 조립(B8) → 스냅숏 3(B11~B19) → 장부 재구성(B20) → 비교(B21) → 원인 판정(B22) → 해제(B23) 순서 | 현재 HEAD | 순서가 바뀌면 엔진 실행 중 해제 가능 |
| 불변식 2 | 해제 성공 전에는 성공 문장을 쓰지 않음 | `TestEngineReconcileResolveDoesNotReportSuccessWhenPersistenceFails`(:316 `released` 부재 단언) | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@112 | opts nil 또는 confirm 거짓 | 없음 | return@113 | 편집 무관 |
| B2 if@116 | operator 빈 값 | 없음 | return@117 | 편집 무관 |
| B3 if@120 | note 빈 값 | 없음 | return@121 | 편집 무관 |
| B4 if@123 | deps 결선 누락 | 없음 | return@124 | 편집 무관 |
| B5 if@128 | ctx nil | `context.Background()` | — | 편집 무관 |
| B6 if@132 | journalDir 오류 | 없음 | return@133 | 편집 무관 |
| B7 if@136 | 엔진 잠금 실패 | 없음 | return@137; 이후 defer@139 `lock.Release` | 편집 무관 |
| B8 if@142 | build 오류 | 없음 | return@143 | 편집 무관 |
| B9 if@147 (defer@146 안) | 미닫힘 + close 있음 | `runtime.close()` | — | 편집 무관 |
| B10 if@151 | runtime 결선 검증 실패 | 없음 | return@152 | 편집 무관 |
| B11 if@160 | 1번 수집 실패 | 브로커 읽기 | return@161 | 편집 무관 |
| B12 if@164 | 대기 실패 | 없음 | return@165 | 편집 무관 |
| B13 if@168 | 2번 수집 실패 | 브로커 읽기 | return@169 | 편집 무관 |
| B14 if@172 | 2번 스냅숏 불안정 | 없음 | return@177 | 편집 무관 |
| B15 if@174 | B14 안, Why 빈 값 | 기본 문장 | — | 편집 무관 |
| B16 if@179 | 대기 실패 | 없음 | return@180 | 편집 무관 |
| B17 if@183 | 3번 수집 실패 | 브로커 읽기 | return@184 | 편집 무관 |
| B18 if@187 | 3번 스냅숏 변함 | 없음 | return@192 | 편집 무관 |
| B19 if@189 | B18 안, Why 빈 값 | 기본 문장 | — | 편집 무관 |
| B20 if@196 | 장부 재구성 실패 | 없음 | return@197 | 편집 무관 |
| B21 if@200 | `diff.BlocksEntry()` 참(수량 불일치·원장에만 있는 주문) — 계좌 전체, 외부 보유·외부 주문·평단은 세지 않음 | 없음 | return@201 `blocking reconciliation diff remains` | 편집 무관(D2: 이 판정을 그대로 씀) |
| B22 if@203 | `runtime.validate()` 오류 | 없음 | return@204 | 편집 무관(판정은 build 의 클로저) |
| B23 if@206 | `runtime.resolve` 오류 | 원장 원자 해제 실패(롤백) | return@207 | 편집 무관 |
| B24 if@209 | close 오류 | 해제는 이미 커밋 | return@210 | 편집 무관 |
| (return@215) | 성공 | :214 성공 문장(원인 중립 「active RECONCILE blocks released …」) 출력, `closed = true` | nil | 편집함 — 문장만(D7) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `deps.acquire` | 엔진 배제 | 실패 즉시 반환 | AST B7 |
| `runtime.collect` ×3 · `deps.sleep` ×2 · `reconcile.Stabiliser.Offer` ×3 | 안정 스냅숏(Required 2, 간격 2s) | 재시도 없음 | AST · snapshot.go:376·380 |
| `runtime.localState` · `runtime.compare` · `diff.BlocksEntry` | 장부 대 계좌 비교 | — | AST B20·B21 · compare.go:312 |
| `runtime.validate` · `runtime.resolve` · `runtime.close` | 원인 판정·해제·닫기 | — | AST B22~B24 |
| `fmt.Fprintf(cmd.OutOrStdout(), …)` | 성공 문장 | — | :214 |

## State mutations and fallbacks

- 쓰기는 `runtime.resolve`(원장 해제) 하나. fallback: B5(ctx), B15·B19(빈 Why 기본 문장).

## Safety conclusion

- Edit boundary(실제): :214 의 출력 문자열만. 「released」 단어는 유지(시험 :291·:394 가 포함, :316 이 실패 시 부재를 단언).
- High-risk impact: yes(대사 해제 명령) — 단 편집은 출력 문장이라 판정 경로 무변경.

# Function Logic Map: `runEngineReconcileResolve` (pre-edit 판)

- Source: `cmd/tossctl/engine_reconcile.go` (107-214), base `064f59fd9`, source sha256 `7192c69246ed…`.
- AST evidence: `ast.json`(분기 24, return 21, defer 2). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file cmd/tossctl/engine_reconcile.go --func runEngineReconcileResolve`.
- Risk scan: `risk-pattern-report.md`(일치 0). 커버리지: `../../coverage/pre-edit-runEngineReconcileResolve-and-cmd-blocks.txt`.
- 편집 범위(Manager 결정 Q5): 성공 출력 문장(:212)을 원인 중립으로. 분기·return·호출 순서는 바꾸지 않는다. a166 census 키 없음(shapes.json 의 engine_reconcile.go 키는 `buildReconcileResolveRuntime` 하나).

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
| B1 if@110 | opts nil 또는 confirm 거짓 | 없음 | return@111 | 편집 무관 |
| B2 if@114 | operator 빈 값 | 없음 | return@115 | 편집 무관 |
| B3 if@118 | note 빈 값 | 없음 | return@119 | 편집 무관 |
| B4 if@121 | deps 결선 누락 | 없음 | return@122 | 편집 무관 |
| B5 if@126 | ctx nil | `context.Background()` | — | 편집 무관 |
| B6 if@130 | journalDir 오류 | 없음 | return@131 | 편집 무관 |
| B7 if@134 | 엔진 잠금 실패 | 없음 | return@135; 이후 defer@137 `lock.Release` | 편집 무관 |
| B8 if@140 | build 오류 | 없음 | return@141 | 편집 무관 |
| B9 if@145 (defer@144 안) | 미닫힘 + close 있음 | `runtime.close()` | — | 편집 무관 |
| B10 if@149 | runtime 결선 검증 실패 | 없음 | return@150 | 편집 무관 |
| B11 if@158 | 1번 수집 실패 | 브로커 읽기 | return@159 | 편집 무관 |
| B12 if@162 | 대기 실패 | 없음 | return@163 | 편집 무관 |
| B13 if@166 | 2번 수집 실패 | 브로커 읽기 | return@167 | 편집 무관 |
| B14 if@170 | 2번 스냅숏 불안정 | 없음 | return@175 | 편집 무관 |
| B15 if@172 | B14 안, Why 빈 값 | 기본 문장 | — | 편집 무관 |
| B16 if@177 | 대기 실패 | 없음 | return@178 | 편집 무관 |
| B17 if@181 | 3번 수집 실패 | 브로커 읽기 | return@182 | 편집 무관 |
| B18 if@185 | 3번 스냅숏 변함 | 없음 | return@190 | 편집 무관 |
| B19 if@187 | B18 안, Why 빈 값 | 기본 문장 | — | 편집 무관 |
| B20 if@194 | 장부 재구성 실패 | 없음 | return@195 | 편집 무관 |
| B21 if@198 | `diff.BlocksEntry()` 참(수량 불일치·원장에만 있는 주문) — 계좌 전체, 외부 보유·외부 주문·평단은 세지 않음 | 없음 | return@199 `blocking reconciliation diff remains` | 편집 무관(D2: 이 판정을 그대로 씀) |
| B22 if@201 | `runtime.validate()` 오류 | 없음 | return@202 | 편집 무관(판정은 build 의 클로저) |
| B23 if@204 | `runtime.resolve` 오류 | 원장 원자 해제 실패(롤백) | return@205 | 편집 무관 |
| B24 if@207 | close 오류 | 해제는 이미 커밋 | return@208 | 편집 무관 |
| (return@213) | 성공 | :212 성공 문장 출력, `closed = true` | nil | **편집 대상** — 문장만 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `deps.acquire` | 엔진 배제 | 실패 즉시 반환 | AST B7 |
| `runtime.collect` ×3 · `deps.sleep` ×2 · `reconcile.Stabiliser.Offer` ×3 | 안정 스냅숏(Required 2, 간격 2s) | 재시도 없음 | AST · snapshot.go:376·380 |
| `runtime.localState` · `runtime.compare` · `diff.BlocksEntry` | 장부 대 계좌 비교 | — | AST B20·B21 · compare.go:312 |
| `runtime.validate` · `runtime.resolve` · `runtime.close` | 원인 판정·해제·닫기 | — | AST B22~B24 |
| `fmt.Fprintf(cmd.OutOrStdout(), …)` | 성공 문장 | — | :212 |

## State mutations and fallbacks

- 쓰기는 `runtime.resolve`(원장 해제) 하나. fallback: B5(ctx), B15·B19(빈 Why 기본 문장).

## Safety conclusion

- Safe edit boundary: :212 의 출력 문자열만. 「released」 단어는 유지(시험 :291·:394 가 포함, :316 이 실패 시 부재를 단언).
- High-risk impact: yes(대사 해제 명령) — 단 편집은 출력 문장이라 판정 경로 무변경.

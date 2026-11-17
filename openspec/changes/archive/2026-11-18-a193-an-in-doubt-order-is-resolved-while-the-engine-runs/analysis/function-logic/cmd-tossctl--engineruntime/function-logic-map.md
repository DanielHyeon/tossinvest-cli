# Function Logic Map: `engineRuntime`

> **편집 후 지도 — 2단계 GREEN 뒤 갱신(2026-11-18).** 편집 전 원본 `../../pre-edit-ast/cmd-tossctl--engineruntime.ast.json`(655-752, sha `2365d757c29d`, 분기 6).
> 재번호 `renumber.txt` — 옛/새 분기를 (종류, 줄)로 대조, 분기 6개 줄 그대로(편집은 B6 뒤 L748 근처 주석 1줄 + 슬라이스 원소 1).

- Source: `cmd/tossctl/engine.go` (655-753)
- AST evidence: `ast.json` — branches **6** / returns **7** / calls **17**, source SHA-256 `27513e682d3bee3ea304ba884e8d66345577b51c3ff7160424742f27fd0cd7cd`(편집 후).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file cmd/tossctl/engine.go --func engineRuntime`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- a166 census: `shapes.json` 키 50개에 이 함수 없음(키 목록 대조, 손으로 읽음).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx *engine.Context` | 조립 완료된 엔진 문맥(Journal·Resolver·Entry 등) | `assembleEngine` | 각 생성자가 거절하면 그 자리에서 `nil, err` |
| `clk` | 엔진 시계 | 호출자 | — |
| `ready` | 준비 신호(시험은 nil) | 부팅 6단계 | `recoverThenReady` 가 nil 을 허용 |
| 불변식 | 감독 루프 집합 = 넷(reconcile·exit·filldetect·전략 진입) — 보조 실행자는 Loops 에 들어가지 않는다 | engine-safety 「엔진 런타임 수명주기」 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@658) | `engineFillDetector`(call@657) 오류 | 없음 | return@659 `nil, err` | 기존 표 시험(`TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess`)에 **이 갈래 행 없음**(그 표는 B2·B3·B4 셋 — engine_runtime_branch_test.go:57·64·71). 커버는 2.3 에서 잼 |
| B2 (if@667) | `ectx.ReconcileDriver`(call@662) 오류 | 없음 | return@668 | 같은 시험 「B2 reconcile construction」 |
| B3 (if@684) | `ectx.ExitObserver`(call@671) 오류 | 없음 | return@685 | 같은 시험 「B3 exit observer construction」 |
| B4 (if@689) | `ectx.Recovery`(call@688) 오류 | 없음(성공 시 진입 게이트에 복구 미완 래치 — `reconcile.New`) | return@690 | 같은 시험 「B4 recovery construction」 |
| B5 (if@693) | 전략 진입 감독자(call@692) 오류 | 없음 | return@694 | 표 시험에 행 없음 — 2.3 에서 잼 |
| B6 (if@705) | `ectx.AlertDeliverer`(call@704) 오류 | 없음 | return@706 | 표 시험에 행 없음 — 2.3 에서 잼 |
| (끝) | 전부 성공 | `engine.NewRuntime`(call@709) — **Auxiliary 목록(L749~750)에 a193 이 원소 `ectx.InDoubtSettler(clk)` 하나를 더함(편집 완료)** | return@709 | 기존 성공 갈래 + `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(이름 집합) + `TestA193TheProductionRuntimeStartsTheSettlerAsAuxiliary` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.Recovery`(call@688) | 기동 복구 — `c.Resolver` 를 씀 | nil 배선이면 `ErrRecoveryUnavailable`(`Context.Recovery` B1 if@176) | AST(cited-ast) |
| `engine.NewRuntime`(call@709) | 감독자 | 이름 중복·빈 Run 거절 — 새 이름 `in-doubt-settle` 은 기존 9개 이름과 다름 | AST + runtime.go:227-240(손으로 읽음) |
| `ectx.InDoubtSettler(clk)`(call@750) | 세션 중 IN_DOUBT 해소 보조 실행자 | 오류를 돌려주지 않음(값만) — 배선 누락은 실행자의 Run 이 정지로 드러냄 | design D2 |
| `ectx.NormalAlertRelayExecutor`·`AttestationWatch`·`ManualBuyRefresher`·`ExitLowerLimitRefresher`(call@748·@749) | 기존 보조 실행자 | 무변경 | AST |

## State mutations and fallbacks

- 편집: 복합 리터럴 `engine.RuntimeOptions{… Auxiliary: []engine.AuxiliaryExecutor{…}}` 의 슬라이스 원소 하나 + 주석 1줄. **분기·return 추가 0 — 편집 후 AST 로 확인**(분기 6 · return 7 그대로, calls 17 → 18).
- 감독 루프(Loops) 무변경 — 실행자를 Loops 에 넣지 않는다(변이 m12 가 시험으로 잼).

## Safety conclusion

- Safe edit boundary: `Auxiliary` 슬라이스 원소 1. 분기 6개·return 7개 문장 불변.
- High-risk impact: yes(배선) — 이 원소가 운영에서 세션 중 해소를 켠다(브로커 조회 증가, park 시 계정 전역 진입 래치 — design D5 공시).

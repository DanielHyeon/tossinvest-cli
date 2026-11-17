# Function Logic Map: `engineRuntime`

> **편집 전(pre-edit) 지도 — 1단계.** 2단계 GREEN 뒤 `ast.json` 을 다시 뽑고 옛/새 분기를 (종류, 그 줄의 문장)으로 정렬해 재번호한다(손 재번호 금지 —
> 기억 「분기 번호는 위치다」).

- Source: `cmd/tossctl/engine.go` (655-752)
- AST evidence: `ast.json` — branches **6** / returns **7** / calls **17**, source SHA-256 `2365d757c29dfa08a0ce1db9bd6862c4bfc2d450114562479bc34f18f7d4b0eb`(HEAD `791fb6ef1`).
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
| (끝) | 전부 성공 | `engine.NewRuntime`(call@709) — **Auxiliary 목록(L748~749)에 a193 이 원소 `ectx.InDoubtSettler(clk)` 하나를 더함** | return@709 | 기존 성공 갈래 + `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(이름 집합) + 2단계에 생길 a193 배선 시험(tasks 3.10 — 이름은 아직 없는 함수라 여기 적지 않음; check_analysis 가 존재하지 않는 시험 이름을 거절함) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.Recovery`(call@688) | 기동 복구 — `c.Resolver` 를 씀 | nil 배선이면 `ErrRecoveryUnavailable`(`Context.Recovery` B1 if@176) | AST(cited-ast) |
| `engine.NewRuntime`(call@709) | 감독자 | 이름 중복·빈 Run 거절 — 새 이름 `in-doubt-settle` 은 기존 9개 이름과 다름 | AST + runtime.go:227-240(손으로 읽음) |
| (2단계 추가) `ectx.InDoubtSettler(clk)` | 세션 중 IN_DOUBT 해소 보조 실행자 | 오류를 돌려주지 않음(값만) — 배선 누락은 실행자의 Run 이 정지로 드러냄 | design D2 |
| `ectx.NormalAlertRelayExecutor`·`AttestationWatch`·`ManualBuyRefresher`·`ExitLowerLimitRefresher`(call@748·@749) | 기존 보조 실행자 | 무변경 | AST |

## State mutations and fallbacks

- 편집: 복합 리터럴 `engine.RuntimeOptions{… Auxiliary: []engine.AuxiliaryExecutor{…}}` 의 슬라이스 원소 하나 추가. **분기·return 추가 0**(예상 — GREEN 뒤 AST 로 확인).
- 감독 루프(Loops) 무변경 — 실행자를 Loops 에 넣지 않는다(변이 m12 가 시험으로 잼).

## Safety conclusion

- Safe edit boundary: `Auxiliary` 슬라이스 원소 1. 분기 6개·return 7개 문장 불변.
- High-risk impact: yes(배선) — 이 원소가 운영에서 세션 중 해소를 켠다(브로커 조회 증가, park 시 계정 전역 진입 래치 — design D5 공시).

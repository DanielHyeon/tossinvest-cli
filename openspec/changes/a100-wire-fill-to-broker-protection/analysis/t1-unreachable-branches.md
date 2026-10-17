# a100 T1 결함 기록 — 도달 불가 거부 분기 3 + 부수 관찰 1

- 로트: T1(tasks 1.1~1.3), 2026-10-05 사용자 승인 예외(review.md 「사용자 결정 — 2026-10-05」 1항)
- 대상 소스: `internal/protectionlifecycle/lifecycle.go` sha256 `de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5`
  (a100 번들 ast.json 과 동일), `state.go` sha256 `df5c5459c6d2…`. 작업 base `973dd696`, go1.26.5.
- 분기 ID 정본: `analysis/function-logic/internal-protectionlifecycle--applyfill/ast.json`,
  `--prepareregister/ast.json`. 보조 함수 AST(이 기록의 분기 주장 근거): `analysis/t1-ast/`
  (`mutablePosition` · `validState` · `validPositionTruth` · `validPending`, `go run ./tools/logic-map` 산출).
- **core 수정은 하지 않았다.** 예외의 한계(GREEN 이 core 수정을 요구하면 RED+결함 기록만 남기고 반환)에 따름.

## 측정 요약 (tasks 1.3.1)

`go test -covermode=set -coverprofile` 을 시험별(`-run ^<Test>`)로 떠 true 결과 **본문 블록**의 실행을 쟀다.

| 함수 | 분기 | 본문 블록 | 실행 | 담당 시험 |
|---|---|---|---|---|
| applyFill | B1 | 237.63–239.3 | yes | `TestA100T1ApplyFillB1UnknownPositionRefusedWithoutCreatingState` |
| applyFill | B2 | 240.24–242.3 | **NO (도달 불가, D1)** | 시나리오는 `TestA100T1ApplyFillB2BrokenSealRefusedWithoutReseal` 가 B1 경로로 고정 |
| applyFill | B3 | 244.154–246.3 | yes | `TestA100T1ApplyFillB3ForeignOrInvalidFillIdentityRefused` |
| applyFill | B6 | 254.107–256.3 | yes | `TestA100T1ApplyFillB6ZeroOrExcessQuantityRefused` |
| applyFill | B7 | 264.37–266.3 | yes | `TestA100T1ApplyFillB7FullFillClosesTerminalThenRefusesFurtherFills` |
| prepareRegister | B2 | 8.134–10.3 | yes | `TestA100T1PrepareRegisterB2ClosedEntryOrWrongPhaseRefused` |
| prepareRegister | B3 | 11.25–13.3 | **NO (도달 불가, D2)** | 시나리오는 `TestA100T1PrepareRegisterB3PendingOperationNeverSubmitsTwice` 가 B2 경로로 고정 |
| prepareRegister | B4 | 14.46–16.3 | **NO (도달 불가, D2)** | 시나리오는 `TestA100T1PrepareRegisterB4ActiveProtectionNeverSubmitsSecond` 가 B2 경로로 고정 |
| prepareRegister | B5 | 17.38–19.3 | yes | `TestA100T1PrepareRegisterB5MissingExactOperationLookupRefused` |

**9개 중 6개 yes, 3개는 입력으로 실행시킬 수 없다.** tasks 1.3.2 의 「9개 모두 yes 가 아니면 2절로 가지 않는다」가 걸린다.

## D1 — `applyFill` B2(L240) 본문 도달 불가

- 근거(AST `mutablePosition.ast.json`): `mutablePosition` B1(L297)이 `validState(state)` 거짓이면
  `refuse(RefusalInvalidState, "state seal invalid")` 를 돌려준다. B2(L300)는 `capability.seal != 0` 일 때만
  봉인 검사를 하는데, applyFill 은 `brokerCapability{}`(봉인 0)를 넘긴다(L236). 따라서 applyFill 안에서
  `mutablePosition` 이 낼 수 있는 오류는 `invalid_state`(L298)·`invalid_identity`(L305) 둘뿐이고
  `RefusalInvalidObservation` 은 나오지 않는다.
- 결과: applyFill B1 의 예외절 `errorCode(err) != RefusalInvalidObservation` 은 오류가 있을 때 **언제나 참**이고,
  무효 봉인은 항상 B1 본문에서 끝난다. B2 조건 `!validState(state)` 가 참이 되려면 직전의 같은 순수 호출이
  거짓이었어야 하므로 B2 본문은 도달 불가. B1 과 B2 의 반환값(`state, FillResult{PreserveExit:true},
  invalid_state: state seal invalid`)은 바이트 단위로 같아 **관측으로 가를 수 없고, 따라서 RED 시험도 쓸 수 없다**
  (RED 는 커버리지 0 이라는 측정으로만 존재).
- 안전 영향: 없음. tasks 1.1.2 의 요구(거부 + 어떤 필드도 복구·재봉인 안 됨)는 B1 경로에서 충족되고 시험으로 고정됨.
  다만 예외절 `!= RefusalInvalidObservation` 의 의도(InvalidObservation 을 통과시켜 B2·B3 로 넘기려던 것으로 보임)는
  현 코드에서 죽어 있다 — 설계 의도와 코드의 불일치.
- 변이 실측: B2 조건을 `false && …` 로 바꾸면 패키지 전 시험 PASS(SHADOWED).
- 구조 고정: `TestA100T1UnreachableApplyFillB2IsShadowedByB1` — zero capability 로 `InvalidObservation` 이 나오는
  날이 오면 실패해 재측정을 요구함.

## D2 — `prepareRegister` B3(L11)·B4(L14) 본문 도달 불가

- 근거(AST `validPositionTruth.ast.json` B2/B3 L178–179, B13/B14 L200–201; `validState.ast.json` B5 L112):
  `validState` 는 모든 포지션에 `validPositionTruth` 를 적용하고, 그 진리표는
  UNPROTECTED 에서 `HasPending || Observed.Status != ""` 를, TERMINAL 에서
  `HasPending || Observed.Status ∉ {CANCELED, FILLED}` 를 거부한다.
- prepareRegister B2(L8)를 통과하려면 phase 가 UNPROTECTED·TERMINAL 이어야 하므로, B1(`mutablePosition` 의
  `validState`)을 통과한 상태에서는 `HasPending=false` 이고 `Observed.Status != ACTIVE` 다. 따라서 B3·B4 조건은
  거짓으로 고정되어 본문이 도달 불가.
- 실제 입력의 실패 지점: SUBMIT_PENDING·SUBMIT_UNKNOWN·REPLACE_PENDING·CANCEL_PENDING·ACTIVE 는 전부
  B2 의 `entry_latched: entry is closed` 에서 막히고 명령은 만들어지지 않는다(시험 고정). **안전 요구
  (1.2.2 두 번째 제출 없음, 1.2.3 두 번째 보호주문 없음)는 충족**된다. 다만 오류 코드가 `operation_pending` 이
  아니라 `entry_latched` 라 호출자가 사유를 코드로 가르면 "pending/active" 와 "latch" 가 구별되지 않는다.
- 변이 실측: B3·B4 조건을 각각 `false` 로 바꾸면 패키지 전 시험 PASS(SHADOWED). B2 를 끄면 pending 은
  `operation already pending`(B3), active 는 `protection already active`(B4) 로 떨어진다 — B3·B4 는 B2 의
  예비 방어선으로만 존재한다.
- RED 증인: `analysis/harness/a100_t1_red_unreachable_test.go.txt`(기본 스위트 밖, overlay 주입) — B3·B4 의 자기
  문구를 단언. 현 코드 FAIL(둘 다 `got=entry_latched: entry is closed`), B2 의 포지션 절(EntryLatch·phase)을 뺀
  가상 수정에서 PASS(양성 대조 — 시험이 B3·B4 본문에 닿을 수 있음을 보임).
- 구조 고정: `TestA100T1UnreachablePrepareRegisterB3B4AreShadowedByStateTruthTable` — KR·US 두 시장의 UNPROTECTED·TERMINAL 에
  `HasPending=true` / `Observed=ACTIVE` 를 심고 재봉인해도 `validState` 가 거부하고 prepareRegister 가
  `invalid_state` 로 끝남을 고정. 진리표가 완화되면 실패해 재측정을 요구함.

## D3 (부수 관찰, 9개 분기 밖) — REPLACE_PENDING·REPLACE_UNKNOWN 중 부분 체결이 상태 전체를 영구 정지시킨다

- **패키지 안 실측 확인**(`analysis/harness/t1_probe_fill_during_replace_pending_test.go.txt`, overlay 주입;
  A1 리뷰어가 귀속 가능한 체결로 독립 재확인):
  KR·US 모두 ACTIVE(KR Holdings 10·claim 8·other 2) → KR `prepareReplace(8,101)` → 정확한 KR broker id 를 실은
  `applyFill(qty 3)` 은 `{Applied:true}, nil` 을 돌려주지만 그 상태는 `validState=false` 다.
  `applyReplaceResult(UNKNOWN)` 로 만든 **REPLACE_UNKNOWN 에서도 같다.** CANCEL_PENDING 대조군은 유효 상태를 남긴다.
- **영구 정지:** 그 뒤 KR 다음 체결, **US 포지션의 정확한 id 체결**, `recoverReplace`, `applyReplaceResult`,
  US `prepareCancel`, US `recoverSubmit` 이 전부 `invalid_state: state seal invalid`, `discoverOrphan` 은
  `invalid_exact_observation: invalid orphan observation`(같은 조건식 안에서 validState 실패를 접음)으로 거부된다.
  lifecycle.go 의 전이 함수 11개는 모두 첫 단계에서 `validState`(직접 또는 `mutablePosition`·`pendingPosition`)를
  요구하므로 **무효 상태를 받아 주는 함수가 패키지에 없다** — 새 상태(`newState`)로 다시 시작하는 것 말고는 복구 경로가 없다.
- 기전(AST `validPending.ast.json` B2 L217–218): REPLACE/SUBMIT pending 은
  `command.Quantity + OtherSellClaims == Holdings` 를 요구하는데, applyFill 은 `Holdings` 와 `Observed.Quantity` 만
  줄이고 `Pending.Quantity` 는 그대로 두고 재봉인한다(L257–269). 결과 상태가 진리표를 어긴 채 봉인된다.
- 영향: 보호 교체 중 상주 보호주문이 부분 체결되는 것은 실제로 일어날 수 있는 순서다. 배선 후 이 경로가 열리면
  lifecycle 상태 하나가 **두 시장의 모든 전이**를 막는다 — 시장 간 고장 격리(`TestKRFailureDoesNotMutateUSProtection`
  이 지키려는 성질)가 깨진다. `PreserveExit:true` 라 함수 계약상 청산을 막지는 않지만, 호출자 쪽에서 지켜지는지는
  배선 전이라 미확인.
- 처분: T1 범위 밖. core 수정은 M-A 이후 — Manager 가 T2/T3 의 어느 task 로 편입할지 결정 필요.

## A1 적대 리뷰(P0=0·P1=2·P2=6) 처분 — 2026-10-05

- **P1-1 수리:** `requireSameState` 가 호출 **뒤** 상태끼리 비교해 공유 맵의 제자리 변경을 못 잡았다. 호출 **전**
  스냅숏(`snapshotState`: 저장 봉인·내용 해시·포지션 수)을 떠서 입력·반환 상태 양쪽을 비교하도록 바꿨다.
  하네스 `ALIAS_PR_B2`(B2 본문에서 `state.positions[key].Holdings++`) — 리뷰어 실측 생존 → **CAUGHT**.
- **P1-2 수리:** B2 의 포지션 latch 절만 참인 사례(UNPROTECTED + EntryLatch, 재봉인·`validState` 통과 전제 단언)를
  추가. 하네스 `PR_B2_no_poslatch`(`|| position.EntryLatch != ""` 제거) — 생존 → **CAUGHT**(변이 코드에서는 등록 성공 nil).
  공개 전이로 만든 반례는 B6 이 우연히 막는 상태라 직접 심었다.
- **P2-1 (기록):** D1 구조 고정 시험(`TestA100T1UnreachableApplyFillB2IsShadowedByB1`)은 `mutablePosition` 전제만 본다.
  applyFill B1 예외절 자체의 완화는 이 시험이 못 잡고, check_analysis 의 lifecycle.go `source_sha256` stale 검출에 의존한다.
- **P2-2 수리:** D2 구조 고정 시험을 KR·US 두 시장으로 확장. 하네스 `TRUTH_SKIP_US`(state.go 에서 US 만
  `validPositionTruth` 건너뜀, 리뷰어 M8) → **CAUGHT**.
- **P2-3 수리:** 위 D3 를 "추정, 미검증" 에서 패키지 안 실측 확인으로 상향하고 REPLACE_UNKNOWN·영구 정지를 추가.
- **P2-4 (기록):** applyFill B6 의 `fill.Quantity > position.Holdings` 절은 동치 변이다 — 도달 가능한 유효 상태에서는
  `Observed.Quantity ≤ Holdings − OtherSellClaims` 불변식 때문에 앞 절(`> Observed.Quantity`)이 항상 먼저 참이 된다.
- **P2-5 일부 수리:** AF_B2 SHADOWED 의 양성 대조 `AF_B1_B2_off`(B1·B2 동시 off) → **CAUGHT**(B2 시나리오 시험이
  `실제 nil` 로 실패). 나머지(기록): 하네스는 주로 가드 **전체** 단위 변이라 절 단위 생존은 명시한 변이
  (`AF_B3_mismatch_only`·`PR_B2_no_poslatch`)만 센다. `AF_B3` 는 실제로 broker-id 두 절만 끄므로
  `AF_B3_brokerid_clauses` 로 이름을 바꿨고, 하네스 파일 모드는 755.
- **P2-6 (기록):** go1.26.5 에서 `-cover` 와 `-overlay` 를 함께 쓰면 커버리지는 overlay 사본이 아니라 디스크 원본을
  계측한다. 이번 로트의 커버리지 측정은 overlay 없이 떴으므로 무해하나, 변이 사본의 커버리지를 재려는 향후 하네스의 함정.

## 반환 사유와 남은 결정

1. D1·D2 의 3개 분기는 core 수정 없이는 본문 yes 가 될 수 없다. 선택지(사람·Manager 결정):
   (가) 도달 불가 분기를 제거하거나 순서를 바꾸는 core 수정(M-A 이후, FLM 재생성 필요),
   (나) 1.3.2 의 판정을 「본문 yes 또는 도달 불가 증명(구조 시험 + SHADOWED 변이)」으로 개정,
   (다) 현 상태를 `unsupported` 로 명시하고 2절 진입 조건에서 제외.
2. D3 은 9개 분기와 별개인 실제 런타임 결함 후보다 — WORKFLOW 「비례 원칙」의 첫째 분류(실제 런타임 오류)에 해당할 수 있다.

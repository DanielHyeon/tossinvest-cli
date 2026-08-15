# a100 High-risk Pre-Edit 선언 (tasks 0.7)

`docs/WORKFLOW.md:387-401`의 형식을 따른다. **선언은 편집 직전에 유효해야 하므로**, 아래는
편집 대상별로 지금 확정할 수 있는 항목을 채운 것이고 「실패 테스트 선행 작성」은 각 task
착수 시점에 `yes`로 갱신하고 그 커밋을 여기 적는다. 채워지지 않은 항목이 남은 대상은
편집하지 않는다.

이 change가 닿는 경로는 **보호주문·체결 경로·대사·인프로세스 청산**이므로 전부 High-risk다
(`.claude/CLAUDE.md` §0-5).

> **2026-08-15 재동결:** base는 current main `882a0b49`다. 1~7의 2026-08-11 line/coverage는
> historical evidence이고, current 권위는 `analysis/current-main-evidence.md`와 재생성 bundle이다.
> 새 runtime/child/API 경계는 8~14다. 각 구현 로트의 RED가 준비되면 해당 항목의
> 「실패 테스트 선행 작성」을 실제 테스트명·명령으로 갱신하기 전에는 production을 편집하지 않는다.

---

## 1. `engine.buildGateway` — 수렴 워커 조립

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 3절 (수렴 워커 조립)
- 대상 심볼(패키지.함수): internal/app/engine.buildGateway
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-app-engine--buildgateway/
    측정: engine 패키지 62.1% + cmd/tossctl(coverpkg=engine) 8.9% 합산 —
          분기 4개 전부 미실행, 함수 블록 11개 중 7개 실행
    호출부: cmd/tossctl/engine_assembly_test.go:52 외
- upstream 상속 테스트 영향: no (TossOS 고유 조립)
- 실패 테스트 선행 작성: (착수 시 갱신)
- 안전 불변식 §0 위반 여부 검토: 통과 — 워커 생성만 추가하고 기동은 호출자,
    새 return을 만들지 않으면 분기 수가 늘지 않는다
```

## 2. `journal.scanExitStateResult` — 보호 컬럼 스캔

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 2절 (journal 스키마)
- 대상 심볼(패키지.함수): internal/journal.scanExitStateResult
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-journal--scanexitstateresult/
    측정: internal/journal 75.0% — 분기 22개 중 7개 미실행(B1·B2·B4·B11·B13·B16·B22)
    호출부: scanExitState → Journal.ExitState, Journal.OpenExitStates (전 경로)
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: (착수 시 갱신)
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 보호 컬럼을 v10Evidence·full 판정·
    평탄화 비교 중 하나라도에 넣으면 멀쩡한 행이 부패로 판정되고 **exit 정책이 멈춘다**
    (= 손절 정지, §0-4 위반). 세 리스트 불변이 이 편집의 통과 조건이다.
```

## 3. `journal.Journal.OpenExitStates` — 워커의 대상 집합

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 2·3절
- 대상 심볼(패키지.함수): internal/journal.Journal.OpenExitStates
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-journal--journal.openexitstates/
    측정: internal/journal 75.0% — 분기 4개 중 3개 미실행, 정상 순회만 실행
    호출부: exit 관측 루프의 working set + 크래시 복원 (주석 L614-620)
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: (착수 시 갱신)
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** WHERE 절을 바꾸면 exit 관측 루프의
    working set이 함께 바뀐다. SELECT 컬럼과 Scan 인자만 같은 순서로 갱신한다.
```

## 4. `engine.ExitObserver.record` — 청산 전 상주 주문 취소

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 4절 (이중 매도 권한)
- 대상 심볼(패키지.함수): internal/app/engine.ExitObserver.record
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-app-engine--exitobserver.record/
    측정: engine 62.1% — 분기 14개 중 B6(정리 에러)·B10 미실행
    호출부: exit 관측 사이클
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: (착수 시 갱신)
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 상주 주문 취소를 clearTheSymbol의
    err/cleared에 반영하면 취소 실패가 손절 매도를 막거나 보류시킨다 (§0-4 위반).
    별도 시도·비차단이 이 편집의 통과 조건이다.
```

## 5. `engine.ExitObserver.submit` — 「이미 보호됨」 구별

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 4절
- 대상 심볼(패키지.함수): internal/app/engine.ExitObserver.submit
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-app-engine--exitobserver.submit/
    측정: engine 62.1% — 판정 가능한 10개 중 6개 미실행
          (발행을 막는 네 경로 B1·B3·B4·B5가 전부 미검증)
    호출부: ExitObserver.record L1195 (유일)
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: (착수 시 갱신)
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 제출을 막는 새 조건을 넣지 않는다.
    「이미 flat」은 기존 수량 0 분기가 이미 잡으므로 바꾸는 것은 결과의 이름뿐이다.
```

## 6. `protection.TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport` — 봉인 개방

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 4.0 (다섯 번째 봉인 포함)
- 대상 심볼(패키지.함수): internal/protection.TestProtectionRemainsUnwired…
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-protection--testprotectionremains…/
    측정: 커버리지 not-applicable(테스트 파일은 계측 대상이 아니다).
          측정한 것은 `go test ./internal/protection` exit 0, 패키지 70.8%
- upstream 상속 테스트 영향: no (a071이 만든 TossOS 고유 가드)
- 실패 테스트 선행 작성: (착수 시 갱신)
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 금지 4개 중 `protectionofficial.New`
    **하나만** 뺀다. 허용 목록(L61)과 필수 목록(L78)은 건드리지 않는다. `ProfileProtection`은
    UNWIRED로 남으므로 a071의 보장이 유지된다.
```

## 7. `soak.Runner.RunCycle` — 조건주문 read probe (tasks 0.10 (a))

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.10 (a)
- 대상 심볼(패키지.함수): internal/soak.Runner.RunCycle
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-soak--runner.runcycle/
    측정: internal/soak 84.9%, RunCycle 95.5% — 분기 3개 중 1개 미실행(B1 취소 경로)
    호출부: Runner.Run(soak.go:369) 하나. 산출물은 Recorder.Append → Summarize →
          Evaluate → BuildAttestation → capability-attestation.json → 엔진 automation gate
- upstream 상속 테스트 영향: no (TossOS 고유 도구)
- 실패 테스트 선행 작성: **yes** — internal/soak/protection_probe_test.go.
    RED은 컴파일 실패(`undefined: soak.EndpointConditionalOrders`)로 확인했고,
    통과 조건 4개가 각각 하나의 테스트다. 편집 후 재측정: 분기 3개 그대로, 새 probe 3개 100.0%
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.**
    §0-5 인증 경로이므로 High-risk다 — 앞선 판단("High-risk 경로가 아니다")은 틀렸다.
    통과 조건 4개:
      (1) cycle.Credential은 accountsResult에서만 유도된다(soak.go:513).
          조건주문 read 실패가 streak를 끊으면 a100과 무관한 2026-08-29 재발급까지 막힌다.
      (2) completenessOf의 인자 목록을 바꾸지 않는다.
      (3) 새 probe는 probePrices 뒤에 둔다(M8의 429 penalty window).
      (4) soak.RequiredEndpoints()·LiveOnlyEndpoints()·engine.RequiredEndpoints()를
          이 단계에서 건드리지 않는다. 셋 다 **거부 기준**이고, 지금 넓히면
          거부가 먼저 오고 증거가 나중에 온다.
```

---

## 8. `cmd/tossctl.engineRuntime` — recovery 이후 auxiliary 기동

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 3.9, 4.4
- 대상 심볼: cmd/tossctl.engineRuntime
- 기존 동작 파악 근거: FLM analysis/function-logic/cmd-tossctl--engineruntime/
    current-main 6 branches/84.2%; 기존 분기는 컴포넌트 조립 실패의 early return
- 실패 테스트 선행 작성: (T4-A RED 후 갱신)
- 안전 불변식 검토: 조건부 통과. interlock/recovery 전 worker start 0건. worker는 Loops가 아니라
    Auxiliary에 들어가며 같은 context로 drain한다. stop은 다른 loop/nonzero exit/entry gate를
    건드리지 않고 전용 event+durable reconcile/alert만 남겨야 한다.
```

## 9. `protectionofficial.Gateway.adapt` — raw status/child id 보존

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.11, 2.1~2.2, 4.5.7
- 대상 심볼: internal/protectionofficial.Gateway.adapt
- 기존 동작 파악 근거: FLM analysis/function-logic/internal-protectionofficial--gateway.adapt/
    WATCHING/PAUSED/ORDERING/ORDERED가 bool로 접히고 TriggeredOrderID 원문이 소실됨
- 실패 테스트 선행 작성: (T2-A RED 후 갱신)
- 안전 불변식 검토: 조건부 통과. raw status와 opaque child id를 손실 없이 보존한다. unknown/PAUSED를
    ACTIVE로 승격하거나 quantity/trigger/terminal 판정을 느슨하게 만들면 실패.
```

## 10. `journal.Journal.TrackedFillOrders` — child tracked set

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 3.10
- 대상 심볼: internal/journal.Journal.TrackedFillOrders
- 기존 동작 파악 근거: FLM analysis/function-logic/internal-journal--journal.trackedfillorders/
    current-main 84.0%; confirmed attempts와 replace lineage가 현 권위
- 실패 테스트 선행 작성: (T2-A RED 후 갱신)
- 안전 불변식 검토: 조건부 통과. complete canonical SELL scope와 pre-fill registered_at를 가진 child만
    UNION한다. terminal/applied/conflicting owner는 넣지 않고 detector interface/order는 불변.
```

## 11. `journal.confirmedFillOwners` — causal owner 시각

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 3.10
- 대상 심볼: internal/journal.confirmedFillOwners
- 기존 동작 파악 근거: FLM analysis/function-logic/internal-journal--confirmedfillowners/
    current-main 88.9%; earliest ownership이 snapshot보다 strictly earlier인지 판정
- 실패 테스트 선행 작성: (T2-A RED 후 갱신)
- 안전 불변식 검토: 조건부 통과. child owner가 fill receipt/snapshot보다 늦으면 owner로 세지 않고
    ATTRIBUTION_FAILED로 거부한다. runtime backfill/delta replay 금지.
```

## 12. `journal.resolveFillOrigin` — exact owner union

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 3.10, 4.5.7
- 대상 심볼: internal/journal.resolveFillOrigin
- 기존 동작 파악 근거: FLM analysis/function-logic/internal-journal--resolvefillorigin/
    current-main 88.5%; confirmed attempt만 canonical scope/intent conflict 판정
- 실패 테스트 선행 작성: (T2-A RED 후 갱신)
- 안전 불변식 검토: 조건부 통과. ordinary+protection 후보 전체에서 정확히 한 owner만 허용한다.
    symbol/time 추론 또는 protection 우선순위 금지; 모호하면 durable conflict.
```

## 13. `engine.Runtime.runAuxiliary`/`runAuxiliaryBody` — 전용 stop event와 panic 격리

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 3.5, 3.9
- 대상 심볼: internal/app/engine.Runtime.runAuxiliary, runAuxiliaryBody
- 기존 동작 파악 근거: current-main FLM bundle(재동결 R0). Auxiliary는 stops channel에 쓰지 않고
    panic을 error로 바꾸지만 현재 log event가 alert-delivery 전용으로 고정됨
- 실패 테스트 선행 작성: (T4-A RED 후 갱신)
- 안전 불변식 검토: 조건부 통과. closed executor-specific stop event만 추가한다. 기존
    graceful-cancel 판정, panic recover, wait/drain, stops-channel 격리를 바꾸지 않는다.
```

## 14. `TestProductionAPIExportsNoAuthorityMintingFunction` — lifecycle exact allowlist

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 4.0 (lot T2-B)
- 대상 심볼: protectionlifecycle_test.TestProductionAPIExportsNoAuthorityMintingFunction
- 기존 동작 파악 근거: current-main FLM bundle(재동결 R0). 현재 exported package function 0을 강제
- 실패 테스트 선행 작성: (T2-B RED 후 exact function names로 갱신)
- 안전 불변식 검토: 조건부 통과. planned worker call graph가 요구한 pure transition/evidence API만
    exact allowlist로 허용한다. transport/approval/toggle 타입 또는 public bool/scalar authority
    constructor는 0건이어야 하며 dependency_test의 package ban은 불변.
```

## T1에서 조건부로 추가 선언할 대상

- `applyFill`/`prepareRegister`는 T1 RED가 현 core 결함을 드러내 실제 production을 편집할 때만
  현재 bundle과 named RED를 인용한 선언을 편집 전에 추가한다.

## 15. `verifylive.Runner.pollTrigger` — M0 parent fsync → child read 순서

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.(*Runner).pollTrigger
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: parent child-id response receipt sync hook이 끝나기 전 child raw read 0건;
    sync 실패·critical-window 401/429/read/decode/identity gap은 후속 retry 성공에도 childFilledAt/PASS를
    만들지 않는 named RED
- 안전 불변식 검토: 조건부 통과. existing create/cancel/holding race behavior는 바꾸지 않고,
    child read의 유일한 새 선행조건은 same-run parent receipt durable barrier다.
```

## 16. `verifylive.Runner.readConditional` — lossy adapter 이전 parent raw receipt

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.(*Runner).readConditional
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: request/response parent tag, pre-create pending client tag와 approved-field digest,
    symbol/market/type/orderType/quantity/first side·trigger/expiry/root·leg status/child tag, exact decimal
    strings와 raw-result-bytes-v1 digest; every 401/429/error attempt receipt 및 no-loss named RED
- 안전 불변식 검토: 조건부 통과. official-only raw by-id read를 쓰고 WTS/domain float mapping으로
    fallback하지 않는다. account/token/opaque id 원문을 receipt에 쓰지 않는다.
```

## 17. `verifylive.Runner.readOrder` — official child raw first-observed-fill receipt

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.(*Runner).readOrder
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: official request/response child tag, requested market scope,
    symbol/side/status/quantity/filled quantity/execution price·quantity, raw-result digest, matching triggered
    child tag, transport request-start/body-read-complete monotonic receipt, malformed/error gap named RED
- 안전 불변식 검토: 조건부 통과. 기존 OrderRawByID official endpoint를 유지하고 child receipt는
    synced parent receipt 없이는 호출될 수 없다.
```

## 18. `verifylive.Runner.finishTrigger` — receipt-complete PASS gate

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.(*Runner).finishTrigger
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: filled child라도 seq/fsync/order/identity 또는 critical-window completeness가
    없으면 PASS가 아니라 INCONCLUSIVE/HOLD이고, child fill receipt fsync 뒤 parent 404는 추가 terminal
    증거로 요구하지 않는 named RED
- 안전 불변식 검토: 조건부 통과. broker fill 사실을 되돌리거나 cleanup을 건너뛰지 않고 verdict만
    fail closed한다. server/wall time으로 monotonic proof를 대체하지 않는다.
```

## 19. `cmd/tossctl.newVerifyRunCmd` — explicit receipt/confirm-each surface

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: main.newVerifyRunCmd
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: include-trigger에 receipt flag와 confirm-each가 노출되며 M0 help는 exact
    `--resume --redo conditional-trigger`, ttl-edge/다른 redo 금지를 말하고 다른 verify mode와 mutation
    annotation/method set은 불변인 structural RED
- 안전 불변식 검토: 조건부 통과. 새 flag는 authority가 아니며 human confirmation을 줄이지 않는다.
```

## 20. `cmd/tossctl.runVerifyRun` — unsafe receipt는 mutation 전 거부

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: main.runVerifyRun
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: missing/existing/unwritable/wrong-owner/wrong-mode/symlink path, receipt
    header/file+directory sync/mixed-run, no-confirm-each, no-resume, ttl-edge·다른-redo·미완료 mutating step이
    confirmer·broker factory·Create/Modify/Cancel보다 먼저 0 side effect로 거부되는 named RED. prior
    Outstanding가 있으면 cleanup prologue와 broker factory가 모두 0건인 RED. 기존 verify record
    resume/redo는 fresh O_EXCL receipt일 때만 허용하는 보존 RED
- 안전 불변식 검토: 조건부 통과. receipt writer가 준비된 뒤에만 기존 verify execution/rate/auth path로
    들어가며 runtime/config/attestation/journal을 변경하지 않는다.
```

## 21. `verifylive.Runner.stepConditionalTrigger` — pending/create/child durable checkpoints

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.(*Runner).stepConditionalTrigger
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: pre-create pending intent fsync failure면 create 0건; create response 직후 kill은
    next resume all-page unique client match만 checkpoint하고 HOLD; zero/multiple/mismatch no-mutation;
    child-id 직후 kill은 child checkpoint가 causal receipt보다 먼저 exact child를 노출하는 named RED
- 안전 불변식 검토: 조건부 통과. checkpoint는 causal receipt에 raw ID를 유출하지 않고 기존 verify
    record에만 exact ID를 보존한다. triggered-but-child-unobserved child는 자동 취소하지 않는다.
```

## 22. `verifylive.Runner.createConditional` — create보다 앞선 pending cleanup owner

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.(*Runner).createConditional
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: gate 승인→pending intent append/fsync→CreateConditionalOrder 순서, pending sync
    failure에서 create 0건, create response 뒤 parent checkpoint 전 kill recovery named RED
- 안전 불변식 검토: 조건부 통과. pre-create owner intent가 unavoidable response→checkpoint crash window를
    복구한다. 신규 trading journal/API는 만들지 않고 이 동작은 M0 mode에만 한정한다.
```

## 23. `verifylive.readRetry` — every-attempt evidence가 성공 retry에 지워지지 않음

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.readRetry
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: critical window에서 429→success, 401 refresh→success, transport error가 모두
    durable attempt receipt와 irreversible gap을 남기는 named RED
- 안전 불변식 검토: 조건부 통과. 기존 일반 verify read retry/backoff는 보존하고 M0 observer가 실패
    이력을 잃지 않는다. 함수 본문을 편집하지 않으면 citation-only map으로 남긴다.
```

## 24. `official.Client.doRequest` — transport body-read-complete authority

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: official.(*Client).doRequest
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: request-start, numeric status/no-response, io.ReadAll 완료 시각과 exact response
    bytes가 helper-return delay 전 observer에 전달되는 named RED; success result bytes와 non-2xx raw body
    digest가 각각 versioned algorithm을 쓰고 body read error도 attempt로 보존
- 안전 불변식 검토: 조건부 통과. 기존 return/error/rate-budget semantics는 byte-for-byte 보존하며 nil
    observer가 기존 모든 caller의 동작이다.
```

## 25. `official.Client.send` — 401 포함 attempt observer와 raw result

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: official.(*Client).send
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: first 401/adopted-token retry/minted-token retry 각 attempt가 순서대로 기록되고,
    account header·token refresh·classifyStatus·unwrapAndDecode 결과가 기존 send와 동일한 named RED
- 안전 불변식 검토: 조건부 통과. 새 private traced send가 기존 send의 단일 구현을 공유하고 public raw
    readers의 계약을 바꾸지 않는다. receipt에는 token/account/raw path ID를 쓰지 않는다.
```

## 26. `verifylive.New` — direct caller도 M0 flag 조합을 우회하지 못함

```text
Pre-Edit Gate:
- change id / task id: a100 / tasks 0.2a (lot M0)
- 대상 심볼: verifylive.New
- 기존 동작 파악 근거: M0 current-main AST·FLM·BTM을 코드 편집 전에 생성한다
- 실패 테스트 선행 작성: IncludeTrigger가 true일 때 ConfirmEach+resume/redo trigger-only receipt authority가
    없거나 IncludeTTLEdge가 true면 direct constructor가 runner 생성 전에 거부하는 named RED
- 안전 불변식 검토: 조건부 통과. CLI와 library caller가 같은 fail-closed 조합 계약을 공유한다.
```

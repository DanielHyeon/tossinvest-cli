# a100 High-risk Pre-Edit 선언 (tasks 0.7)

`docs/WORKFLOW.md:387-401`의 형식을 따른다. **선언은 편집 직전에 유효해야 하므로**, 아래는
편집 대상별로 지금 확정할 수 있는 항목을 채운 것이고 「실패 테스트 선행 작성」은 각 task
착수 시점에 `yes`로 갱신하고 그 커밋을 여기 적는다. 채워지지 않은 항목이 남은 대상은
편집하지 않는다.

이 change가 닿는 경로는 **보호주문·체결 경로·대사·인프로세스 청산**이므로 전부 High-risk다
(`.claude/CLAUDE.md` §0-5).

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

## 아직 선언하지 않은 대상

- `internal/protectionlifecycle`의 전이 함수 노출(design D1, 다섯 번째 봉인) — 어떤 함수를
  어떤 이름으로 노출할지 확정되지 않았다. 확정 시 이 문서에 항목 8로 추가한다.

# Function Logic Map: `Runner.RunCycle`

- Source: `internal/soak/soak.go`
- AST evidence: `ast.json` (source SHA-256 `84fa92e8716d3a6b…`, 분기 3 / return 3 — **편집 후 재생성**)
- Risk scan: `risk-pattern-report.md`
- 작성 사유: tasks 0.10 (a) — 조건주문 read probe 3개를 이 함수에 추가한다. **기존 함수의 내부를
  편집하므로 편집 전에 만든다.** 대상은 인증·attestation 경로이므로 High-risk다(안전 불변식 §0-5).

## 이 함수가 하는 일

한 사이클에서 조사 대상 endpoint를 전부 한 번씩 읽고 `Cycle` 하나를 만든다. 반환된 `Cycle`은
`Recorder.Append`로 파일에 붙고, 그 파일이 `Summarize` → `Evaluate` → `BuildAttestation`을 거쳐
**엔진의 automation gate가 걸려 있는 attestation**이 된다. 즉 이 함수가 append하는 항목이
attestation의 endpoint 집합이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소되지 않음 | 호출자(`Run`) | 취소면 빈 `Cycle`과 `ctx.Err()`를 반환하고 **아무것도 기록하지 않는다**(B1) |
| `r.opts.Reads` | non-nil, GET 전용 9메서드(편집으로 6→9) | `New`가 nil을 거부(`soak.go:348`) | 메서드 실패는 에러가 아니라 **측정값**이다 — `EndpointResult`에 담기고 사이클은 계속된다 |
| `r.opts.Clock` | non-nil | `New`가 기본값 주입 | — |
| `r.opts.Symbols` | 비어 있지 않음 | `New`가 빈 목록을 거부(`soak.go:360`) | — |
| `r.startedWith` | 0값 허용 | `New`가 `Binary()`로 1회 취득 | 지문 불가는 0값이고 독자는 "모름"으로 읽는다 |
| `cycle.AccountRef` | 첫 비어있지 않은 account | `probeAccounts` (B2·B3) | 전부 공백이면 빈 값 → `Evaluate`가 "no account was resolved"로 거부 |
| `cycle.Credential` | `accountsResult`에서만 유도 | `observeCredential` (`:528`) | **다른 어떤 endpoint의 실패도 credential 판정에 들어가지 않는다** |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`:506`) | `ctx.Err() != nil` | 없음 | `Cycle{}, err` — 기록 없음 | 취소된 ctx로 호출하면 빈 사이클 |
| B2 (`:522`) | `range accounts` | 없음(순회) | — | 계정 목록이 여럿일 때 첫 항목이 선택된다 |
| B3 (`:523`) | `strings.TrimSpace(a) != ""` | `cycle.AccountRef = a`, `break` | — | 앞이 공백인 목록에서 뒤의 유효 항목이 선택된다 |

**분기가 셋뿐이고 전부 조립이 아니다.** 이 함수는 조건부 조립이 없는 **고정 순서 절차**다 —
어떤 endpoint를 읽을지는 실행 시점에 결정되지 않는다. 그래서 probe를 추가하는 편집은
분기를 바꾸는 편집이 아니라 **순서 목록에 한 항목을 더하는 편집**이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ctx.Err` (`:506`) | 취소 확인 | — | ast.json |
| `r.opts.Clock.Now` (`:513`, `:565`) | 사이클 시작·종료 시각 | — | ast.json |
| `r.probeAccounts` (`:520`) | **자격증명 probe이자 계정 해석.** 반드시 첫 번째 | 재시도 없음(`probe`) | ast.json + `soak.go:646` |
| `r.observeCredential` (`:528`) | streak의 입력이 되는 credential 레코드 | — | ast.json + `soak.go:808` |
| `r.probe(BuyingPower)` (`:531`) | 잔고 read | 재시도 없음 | ast.json |
| `r.probeHoldings` (`:535`) | 보유 개수 read | 재시도 없음 | ast.json |
| `r.probeOrders` (`:540`) | CLOSED+OPEN 두 walk | **429 재시도 있음**(`retryRateLimited`) | ast.json + `soak.go:742` |
| `r.probeOrderByID` (`:542`) | walk가 찾은 첫 id 1건 | 429 재시도 있음. **id가 없으면 Skipped** | ast.json + `soak.go:786` |
| `r.probePrices` (`:545`) | 시세 read | 재시도 없음 | ast.json |
| `completenessOf` (`:564`) | 페이지네이션·식별자·시세 개수 판정 | — | ast.json + `soak.go:851` |

live binding은 `cmd/tossctl/soak.go`의 `soakReads`가 유일하다(`buildSoakWiring`, `cmd/tossctl/soak.go:655`).
`static_test.go`가 그 파일 하나만 실제 client를 쥐도록 정적으로 고정한다.

## State mutations and fallbacks

- **`cycle.Endpoints`에 append하는 것 외에 프로세스 상태를 바꾸지 않는다.** 유일한 예외는
  `observeCredential`이 갱신하는 `r.lastTokenExpiry`/`r.haveTokenExpiry`이며, 그것은
  토큰 만료가 앞으로 움직였는지(=무인 갱신)를 다음 사이클과 비교하기 위한 것이다.
- 실패한 read는 **fallback이 없다.** 대체 transport도, 캐시된 값도 없다. 실패는 그대로
  `EndpointResult{OK:false, Class, Error}`로 기록된다 — 그것이 이 도구의 측정값이다.
- `probeOrderByID`만 **Skipped**라는 세 번째 결과를 갖는다. 읽을 id가 없을 때 성공으로도
  실패로도 기록하지 않으며, `Evaluate`가 `endpointReason`에서 이 경우를 따로 설명한다
  (`attest.go:152-158`).
- `completenessOf`의 입력은 order walk 2개, `positions`, `symbolsAsked`, `quotes`로 **고정**이다.
  endpoint를 추가해도 completeness 판정의 입력은 달라지지 않는다.

## 편집이 건드려선 안 되는 것 (a100 tasks 0.10 (a))

1. **`probeAccounts`가 첫 번째라는 것.** 계정 해석과 자격증명 판정이 여기서만 나온다.
2. **`cycle.Credential`이 `accountsResult`에서만 유도된다는 것**(`:528`). 조건주문 read가
   실패해도 **streak가 끊기면 안 된다.** streak는 하루 단위 credential 판정의 연속이고
   (`summary.go:105`), attestation의 3일 요건이 거기 걸려 있다.
3. **`completenessOf`의 인자 목록.** 새 endpoint는 completeness 판정에 참여하지 않는다.
4. **새 probe는 사이클의 맨 뒤에 둔다.** 이 계좌는 2026-07-27에 order walk가 연 429 penalty
   window에 뒤따르는 read가 걸린 전례가 있다(measurements.md M8, `soak.go:196-203`).
   요청 3개를 앞이나 중간에 넣으면 **기존 endpoint를 429로 밀어낼 수 있다.** 맨 뒤에 두면
   추가 요청이 만든 429는 추가 요청 자신에게 떨어진다.

## Safety conclusion

- **Safe edit boundary**: `cycle.Endpoints`에 append하는 새 단계를 `probePrices` **뒤**에
  추가한다. **편집 완료 — `:559-562`이 그 단계이고 분기는 3개 그대로다.** B1·B2·B3와 `observeCredential` 호출, `completenessOf` 호출은 손대지 않는다.
- **High-risk impact**: **yes.** attestation은 automation gate의 입력이고, 이 함수가 그
  attestation의 endpoint 집합을 만든다. 다만 이 편집은 집합을 **넓히기만** 하며
  `RequiredEndpoints()`(거부 기준)와 `engine.RequiredEndpoints()`(기동 기준)는 건드리지 않으므로,
  **잘못돼도 거부가 늘지 않고 증거가 안 늘 뿐이다.** 보수 방향이다.

# Function Logic Map: `Console.decoratePositionRows`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json` (`source_sha256` 8a110f16a288… (ast.json), lines 82–151 — **편집 후 revision**)
- Risk scan: `risk-pattern-report.md`
- 이 map은 **편집 전에** 작성했다 (base sha `16975cd71904…`, lines 82–137). 아래 표의
  줄 번호와 AST 증거는 gate가 요구하는 대로 편집 후 revision으로 갱신했고, 분기
  구조·조건·mutation은 편집 전과 동일하다 — 그것이 이 편집의 안전 근거다.

## 이 함수가 하는 일

두 보호선 화면(`/positions`·`/dashboard`)이 **공유하는 단 하나의 표시 경계**다.
`handlePositions`(line 69)와 `handleOverview`(`overview.go:652`)가 각각 부르며,
같은 보유 종목이 두 화면에서 다른 관리·보호 답을 내놓지 않게 하는 것이 이
함수의 존재 이유다(line 73–76 주석).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.PositionPolicies` | nil 허용 (미배선 빌드) | `console.go:321` | nil이면 B1 전체를 건너뛰고 `runtimeAttempted`가 false로 남아 B7도 건너뛴다 |
| `reading.Runtime` | 캐시가 서빙하는 한 시도의 결과 | `positionPolicyCache` (편집 전: 커맨더 직접 호출) | 그 시도가 실패했으면 zero value → `EffectiveKnown == false`. 에러 검사가 아니라 **zero value가** unknown 렌더를 만든다. 캐시가 실패를 실패로 서빙하는 이유 (design D3) |
| `reading.StatesErr` | B2가 검사 | `positionPolicyCache` (편집 전: 커맨더 직접 호출) | 실패면 `policyByID`가 nil로 남는다 → B10이 모든 행에서 참 → `journalKnown = false` |
| `c.opts.Settings` | nil 허용 | `console.go` settings seam | nil이면 B4를 건너뛰어 `Designated`·`Excluded`가 zero value로 남는다 |
| `rows` | 호출자 소유의 request-local 슬라이스 | `c.positions()` / overview | 빈 슬라이스면 모든 range가 0회 |
| `asOf` | 호출자의 `c.now()` | `portfolio_pages.go:69`, `overview.go:652` (무변경) | — |

**불변식 1 — 실효값을 desired로 위장하지 않는다.** `Runtime` 실패가
`EffectiveKnown = false`로 남는 것은 `operator-console`의 SHALL이다("runtime
unavailable인 non-managed 행은 desired를 effective로 위장하지 않고 `UNKNOWN`").
이 성질은 **에러 처리가 아니라 zero value**에 얹혀 있으므로, 이 경로를 만지는
어떤 편집도 실패 시 zero-value runtime을 그대로 통과시켜야 한다.

**불변식 2 (개정) — 시점은 셋이고, 불일치는 한 방향으로만 틀릴 수 있다.**
`c.positions()`가 렌더마다 journal을 읽으므로 **행은 언제나 reading보다 신선하다**.
"한 시점"은 이 함수에 존재한 적이 없다. 지켜야 하는 것은 동시성이 아니라 방향이다 —
목록이 모르는 행은 `관리 여부 불명`으로 렌더되고 보호선을 갖지 않는다. stale 목록은
판정을 보류할 수 있을 뿐 만들어낼 수 없다. 초안의 서술은 아래에 남긴다.

**불변식 2 (초안, 철회) — 한 렌더는 한 시점의 짝을 본다.** `runtime`(B1)과 `policyByID`(B2)는
B8 안에서 합쳐져 같은 행의 `ProjectManagement`(line 135)와, 그 결과가
`attachPositionExitLines`(line 144)의 `BuildExitLineReference`에 들어간다. 서로
다른 시점의 두 사실을 섞으면 어느 쪽도 틀리지 않았는데 조합만 틀린 판정이 나온다.
같은 논리가 line 103–104에 `Settings.Load`에 대해 이미 적혀 있다 — "One Load
stamps both lists".

**불변식 3 — mutation은 request-local 표시 행뿐이다.** 브로커 호출, journal 쓰기,
주문, config 저장 중 어느 것도 이 경로에서 도달 불가능하다(line 77–81 주석).

## Branches and early returns

early return이 없다. `returns: null` — 12개 분기가 전부 조건부 실행이고 함수는
끝까지 진행한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.opts.PositionPolicies != nil` (88) | `runtimeAttempted=true`, `runtime`·`policyByID` 채움. **엔진 읽기가 여기 있고, 이제 간격당 1벌이다** | 없음 | 2.2, 5.2 |
| B2 | `reading.StatesErr == nil` (108) | `policyByID` 생성 | err는 소비되고 전파 안 됨 | 5.2 |
| B3 | `range reading.States` (110) | `policyByID[TrimSpace(PositionID)] = state` | 없음 | 5.1 |
| B4 | `c.opts.Settings != nil` (115) | — | 없음 | 기존 |
| B5 | `Settings.Load()`의 `err == nil` (116) | — | err 소비 | 기존 |
| B6 | `range rows` (119) | `rows[i].Designated`·`.Excluded` | 없음 | 기존 |
| B7 | `runtimeAttempted` (125) | 아래 전부 | 없음 | 2.2, 5.2 |
| B8 | `range rows` (126) | 행별 lifecycle·management | 없음 | 5.1 |
| B9 | `row.InJournal` (130) | `LifecycleProofRequired=true` | 없음 | 기존 |
| B10 | `!ok` — `policyByID`에 이 PositionID가 없음 (133) | `journalKnown=false` | 없음 | 5.2 |
| B11 | else — 항목 있음 (135) | `LifecycleKnown`·`LifecycleStatus`·`LifecycleGeneration`·`released` | 없음 | 5.1, 4.4 |
| B12 | `row.Management.Block != nil` (145) | `ManagementBlock` view | 없음 | 기존 |

B10은 `policyByID`가 nil일 때도 참이다 (nil map 조회는 zero value + false). 즉
**엔진 `List` 실패의 렌더 결과는 "목록에 그 포지션이 없음"과 같다** — 둘 다
`journalKnown=false`로 수렴한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.enginePolicy.read` | 엔진의 실행 중 adoption 설정·대사 차단과 포지션별 lifecycle | 에러를 반환하지 않는다 — 두 err를 reading에 실어 보내고 위 표대로 소비된다. 두 절반은 **각자의 간격**으로 만료되며(design D2) 각 간격당 최대 1회가 엔진에 닿는다. 갱신은 요청 컨텍스트가 아니라 detach된 컨텍스트에서 돈다(design D4b) | ast `calls[0]` line 106 |
| `c.opts.Settings.Load` | desired include/exclude | err는 B5가 삼킨다. 콘솔 프로세스 안의 로컬 파일 읽기 — 엔진에 닿지 않는다 | ast `calls[4]` line 116 |
| `positionpolicy.ProjectManagement` | 순수 projection. **`Managed`를 `EffectiveKnown`보다 먼저 본다** — lifecycle 절반이 더 위험한 이유 | 에러 없음 | ast `calls[8]` line 141 |
| `attachPositionExitLines` | 저장된 exit 근거를 표시값으로 | 에러 없음. `runtime`을 받아 `EffectiveKnown`·`DefaultStopPct`를 참조 | ast `calls[10]` line 150 |
| `c.protectionLiveness` | 엔진 생존 신호 (freshness 판정용). **캐시를 거치지 않고 렌더마다 마커를 읽는다** — 엔진 정지 사실이 지연 없이 표시되는 이유 | 에러 없음 | ast `calls[11]` line 150 |

live config binding 없음. 런타임 토글을 읽지 않는다.

## State mutations and fallbacks

- mutation은 전부 `rows[i]` — 호출자가 방금 만든 request-local 표시 행이다.
  `go_statements: null`, `defers: null`.
- fallback은 셋 다 "모르는 것은 모른다고 남긴다"는 같은 형태다 — 커맨더 nil이면
  `runtimeAttempted=false`, `Runtime` 실패면 zero-value runtime, `List` 실패면
  nil map. 어느 것도 desired나 기본값으로 대체되지 않는다.

## a081이 바꾸는 것

**B1 안의 읽기 출처.** 커맨더 직접 호출 2건이 캐시 읽기 1건이 되었다.

```text
before   runtime, _ = c.opts.PositionPolicies.Runtime(ctx)
         states, err := c.opts.PositionPolicies.List(ctx)
after    reading := c.enginePolicy.read(ctx, asOf)   // 간격당 1벌
         runtime = reading.Runtime
         if reading.StatesErr == nil { ... }
```

바뀌지 않는 것:

- 분기 수와 구조. 편집 전후 모두 12갈래이고 조건도 같다. B1은 `c.opts.PositionPolicies`에
  대고 "커맨더가 배선되었는가"를 **계속 그대로** 묻는다 — 캐시 필드를 묻게 바꾸면
  생성 후 seam을 떼는 호출자가 배선된 시절의 읽기를 계속 보게 되고, 실제로
  `a053_exit_line_reference_test.go`가 그 경로를 쓴다. B2는 "목록을 얻었는가"를
  계속 묻는다. 조건의 **출처**만 캐시가 된다.
- 세 fallback의 형태. 캐시가 실패를 실패로 서빙하므로(design D3) zero-value
  runtime과 nil map이 그대로 도달한다 — 불변식 1이 유지되는 이유다.
- 불변식 2의 **방향**. 두 절반이 각자의 간격으로 만료되므로 한 시점의 짝은
  보장되지 않지만(design D4 철회), 목록이 모르는 행은 계속 `관리 여부 불명`으로
  수렴한다 — B10이 nil map과 미매칭을 같게 다루기 때문이다.
- `Settings.Load`(B4·B5), `attachPositionExitLines`, `protectionLiveness`는
  손대지 않는다.

## Safety conclusion

- **Safe edit boundary**: B1 **안쪽**의 읽기뿐. B1 조건과 B4 이하는 불변으로
  유지했고, 편집 후 AST가 그것을 확인한다 (12갈래·같은 종류·같은 순서).
- **High-risk impact**: **no.** 주문·손절·익절·사이징·Guardian·원장·대사·인증·
  체결 감지 어디에도 속하지 않는다. 표시 행만 mutate하며 브로커·journal 쓰기·
  주문·config 저장이 이 경로에서 도달 불가능하다.
- 안전 불변식 4에 대해서는 **개선 방향**이다. 이 함수가 엔진의 단일 쓰기
  커넥션에 거는 경합이 렌더 횟수에서 분리된다.
- 회귀 방지: 불변식 1은 `TestAFailedRuntimeReadingIsNotMaskedByThePreviousSuccess`와
  **`TestAFailedLifecycleReadingIsNotMaskedByThePreviousSuccess`**(뒤가 더 위험한
  절반이고 초안에 없었다), 불변식 2는
  `TestAStaleLifecycleListCanOnlyWithholdAVerdictNeverInventOne`, 상한은
  `TestRedrawingTheLineDoesNotAskTheEngineAgainWithinTheInterval`과
  `TestConcurrentRendersCostOneReading`이 고정한다. 전부 변이 검증에서 RED를
  관측했다 (8.12, 7종).

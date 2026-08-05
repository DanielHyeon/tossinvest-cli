# Function Logic Map: `New`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`source_sha256` 49e1a88df141… (ast.json), lines 435–479)
- Risk scan: `risk-pattern-report.md`
- a081이 추가하는 것은 **B7 한 갈래**다. 기존 7갈래 중 B1–B6은 손대지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.StartVerify` | non-nil 필수 | 호출자 (`cmd/tossctl`) | nil이면 `ErrNoVerifyWiring`으로 early return (B1) |
| `o.Now` | nil 허용 | 호출자 | nil이면 UTC 시스템 시계로 대체 (B2) |
| `o.Remote` | — | 호출자 | `newRemoteRuntime` 실패는 그대로 전파 (B3) |
| `o.PositionPolicies` | nil 허용 | 호출자 | **nil이면 캐시를 만들지 않는다 (B7)** — 미배선 빌드에서 `c.enginePolicy`는 nil로 남고 `read`는 아무 데도 닿지 않는다 |

**불변식**: 캐시는 생성자에서 한 번 만들어지고 `Console`의 수명 동안 같은 객체다.
`holdings`·`ordersCache`와 같은 자리, 같은 규칙이다. 런타임에 커맨더를 바꿔 끼우는
경로는 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `o.StartVerify == nil` (436) | 없음 | `nil, ErrNoVerifyWiring` (437) | 기존 |
| B2 | `now == nil` (440) | `now` 대체 | 없음 | 기존 |
| B3 | `newRemoteRuntime` err (444) | 없음 | `nil, err` (445) | 기존 |
| B4 | `c.out == nil` (457) | `c.out = io.Discard` | 없음 | 기존 |
| B5 | `c.opts.Binary == nil` (460) | `binstamp.Self` | 없음 | 기존 |
| B6 | boot note 비어 있지 않음 (463) | `engineNote`·`engineNoteAt` | 없음 | 기존 |
| B7 | `o.PositionPolicies != nil` (473) | **`c.enginePolicy` 생성** | 없음 | 2.2, 5.2 |

정상 종료는 line 478의 `c, nil` 하나다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newPositionPolicyCache` | 엔진 정책 읽기 캐시 생성 — **간격 둘**을 받는다 (lifecycle 30초, runtime 5초; design D2) | 에러 없음. reader가 nil이면 nil을 반환하므로 B7과 이중으로 안전하고, 간격이 0 이하이면 각자의 기본값으로 대체된다 | ast line 474 |
| `newHoldingsCache`·`newOrdersCache` | 기존 두 캐시 | 무변경 | ast lines 471–472 |

## State mutations and fallbacks

- `c.enginePolicy` 1건 추가. 나머지 대입은 무변경.
- fallback: 커맨더 미배선이면 필드가 nil로 남는다. `positionPolicyCache.read`는
  nil 수신자에서 빈 reading을 반환하고, 표시 경로는 그와 별개로
  `c.opts.PositionPolicies`로 배선 여부를 계속 묻는다.

## Safety conclusion

- Safe edit boundary: line 473–476. 다른 갈래·반환·필드는 불변.
- High-risk impact: **no.** 생성자이며 주문·손절·원장 경로가 없다.

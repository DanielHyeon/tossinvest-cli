# Function Logic Map: `Store.Now`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: base, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**. 바로 뒤에 삽입된 `Store.Clock`의 diff hunk가 이 한 줄 함수와 교차해 evidence가 요구되었다. base L477 = 현재 L577이고 본문은 byte 동일하다. ast.json은 base revision으로 붙였다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.clk` | `Options.Clock` 또는 `clock.System()` | `Open` | nil이 될 수 없다 — `Open`이 항상 채운다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | `s.clk.Now()` | `TestACycleEnforcesRetentionItselfRatherThanLeavingItToACaller` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.clk.Now` | 주입된 시각원 | — | 시각은 주입된 clock 또는 호출자 인자에서만 온다. `TestNothingInThisPackageAsksTheWallClockWhatTimeItIs`가 `time` import를 **경로로** 해석해 `Now/Since/Until/After/Tick/NewTimer/NewTicker/AfterFunc/Sleep` 9개 이름과 alias import·dot import 두 형태, 합쳐 11가지 철자를 막는다. |

## State mutations and fallbacks

- 상태 변경 없음. 순수 읽기이고 본문은 무변경이다.
- 이 접근자가 존재하는 이유가 곧 계약이다 — 보존 스윕과 스캔이 `time.Now()`가 아니라 여기서 순간을 가져가므로 테스트가 하루치 장중을 기다리지 않고 몰 수 있다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재
- High-risk impact: no — 시각 접근자. 주문 경로 무접촉이고 본문 byte 동일.

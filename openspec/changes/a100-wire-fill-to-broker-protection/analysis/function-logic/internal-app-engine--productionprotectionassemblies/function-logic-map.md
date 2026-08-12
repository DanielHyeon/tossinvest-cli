# Function Logic Map: `productionProtectionAssemblies`

- Source: `internal/app/engine/protection_wiring.go` (L39-44)
- AST evidence: `ast.json` — **분기 0**, return 1, 호출 2
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `buildDigest` | hex digest | `productionProtectionDigests()` (version + build info) | 없음 — 값을 그대로 digest 재료로 쓴다 |

## Branches and early returns

**분기가 없다.** AST의 `branches`는 `null`이고 return은 1개뿐이다.
이 함수는 조건 없이 KR·US 두 `SupervisorAssembly`를 만들어 돌려주는 직선 생성자이며,
두 assembly 모두 `Wired: false`가 **리터럴로** 박혀 있다. component digest 재료에는
문자열 `"fill-lifecycle-unwired"`가 들어간다.

| Branch | Condition | Mutation/side effect | Return/error |
|---|---|---|---|
| — | 없음 | 없음 | `[]SupervisorAssembly{KR{Wired:false}, US{Wired:false}}` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry 계약 | Evidence |
|---|---|---|---|
| `digestProtectionIdentity` | component digest 합성 | 순수, 길이 프리픽스 해시 | AST — 호출 2건 |

## State mutations and fallbacks

- 없음. 순수 생성자.

## Safety conclusion

- Safe edit boundary: **이 함수가 a100의 핵심 발견 지점이다.**
  AST가 분기 0을 보여줌으로써 확인된 사실 — **"언제 WIRED가 되는가"라는 판단 지점이
  코드에 아예 존재하지 않는다.** 따라서 a100은 이 함수에 조건을 *추가*하는 작업이 아니라
  **판단 주체를 새로 만들고 그 결과를 여기에 주입**하는 작업이다.
  `Wired`를 리터럴에서 파라미터로 바꾸면 이 함수는 다시 순수 생성자로 남고,
  판단은 호출부(=새 provider)가 진다. 그 형태를 권고한다.
- High-risk impact: **yes** — 이 리터럴이 곧 현재 빌드의 보호 상태다.

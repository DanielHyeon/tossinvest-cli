# Function Logic Map: `newNotifier`

- Source: `internal/app/engine/exitwiring.go` (71-81)
- AST evidence: `ast.json` — **branches 0, returns 1, calls 0, assignments 0,
  defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**a092가 편집하는 두 함수 중 하나.** 순수 구조체 리터럴이고 분기가 없다 —
그래서 여기서 *채우지 않은 필드*는 조용히 기본값이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `j *journal.Journal` | nil 허용 | `gateway.go:280` | nil이면 critical이 best-effort로 강등(`notifyCritical:154`) |
| `gate *execgw.EntryGate` | nil 허용 | 같은 위 | nil이면 래치 없음(`deliver` B10 `:283`) |
| `accountRef string` | 공백 허용 | 같은 위 | 공백이면 모드 승격 없음 |
| `log *obs.Logger` | nil 허용 | 같은 위 | nil이면 로그 없음 |
| `publisher obs.Publisher` | **nil 허용** | `engine.go:444` `resolveNotificationPublisher` | nil이면 모든 critical이 미배달 — 파일 헤더 `:60-70`이 의도라고 밝힘 |
| `clk clock.Clock` | 프로덕션 `clock.System()` | 같은 위 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 | **분기 없음** (AST branches 0) | 없음 | `return &obs.Notifier{...}` `:73` |

## Calls and live bindings

AST calls **0**. 이 함수는 아무것도 부르지 않는다.

### 채우는 필드와 비우는 필드

| `obs.Notifier` 필드 | 이 함수 | 안 채우면 |
|---|---|---|
| `Log` `:74` | 채움 | — |
| `Publisher` `:75` | 채움(nil일 수 있음) | — |
| `Journal` `:76` | 채움 | — |
| `Gate` `:77` | 채움 | — |
| `AccountRef` `:78` | 채움 | — |
| `Clock` `:79` | 채움 | — |
| **`Attempts`** | **안 채움** | `deliver` B1 `:245` → `DefaultCriticalAttempts` **3** |
| **`RetryDelay`** | **안 채움** | `wait` B1 `:292` → `DefaultRetryDelay` **2s** |

**34초 예산의 두 항이 여기서 비어 있다.** 세 번째 항(publish 1회 10초)은
`resolveNotificationPublisher`가 비운다.

## State mutations and fallbacks

- 상태 변경 없음. AST assignments 0.
- fallback은 **비운 필드가 피호출자의 기본값이 되는 것**이고, 그 기본값이 어디서
  오는지는 이 파일에 안 적혀 있다.

## Safety conclusion

- **Safe edit boundary**: 분기가 없으므로 필드를 채우는 편집은 **제어 흐름을 바꾸지
  않는다.** 바뀌는 것은 피호출자가 읽는 값뿐이다. AST branches가 편집 후에도 0이면
  구조가 그대로임이 증명된다.
- **High-risk impact**: **yes** — §0.5(알림·게이트 경로). 다만 이 함수는 판정도 주문도
  하지 않는다.
- **§0.3 방향**: `Attempts`·`RetryDelay`를 **줄이는** 방향은 exit 관측 루프의 체류를
  줄이므로 손절 즉시성에 유리하다. 늘리는 방향은 반대다.
- **토글 없음**: 이 change는 조건부가 아니다. "토글 OFF = upstream 동작"이 성립할
  토글이 없고, 없애려는 것이 바로 그 upstream 동작(34초)이다. §0.6의 "명확한 근거가
  있는 보수 방향"으로 정당화한다 — 근거는 `analysis/delivery-latency.md`.

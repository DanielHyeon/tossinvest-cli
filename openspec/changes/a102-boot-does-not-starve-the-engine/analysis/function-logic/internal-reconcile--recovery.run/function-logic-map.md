# Function Logic Map: `Recovery.Run`

- Source: `internal/reconcile/recovery.go` (233-324)
- AST evidence: `ast.json` — AST 기준 branches **12** / returns 8 / calls 24 / defers 0 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `e0d5690ff164c31e3f5cbb0086cdab84b206a0854a84110b120e87fb98aedfb1`

**a102 §1이 편집한 함수다 — 3단계 네 줄만.** `stableSnapshot`의 반환이 `int`에서
`snapshotProgress`로 바뀌었으므로 그 호출부가 함께 바뀐다. 절차의 순서·판정·게이트 조작은
**한 줄도 바뀌지 않았다.**

## 두 판 — 분기 수는 그대로다

| | 1판 (편집 전) | **2판 (편집 후, 이 문서)** |
|---|---|---|
| 위치 | `recovery.go:207-296` | **`:233-324`** |
| 분기 | 12 | **12** (동일) |
| 이탈 | 8 | **8** (동일) |
| 호출 | 24 | **24** (동일) |
| source SHA-256 | `80ee029c…` | `e0d5690f…` |

**분기·이탈·호출이 하나도 안 늘었다.** 편집은 대입 두 줄 추가(`:295`·`:296`)와
기존 두 줄의 우변 교체(`:293`·`:294`)뿐이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.opts.Journal` | non-nil (`New`가 강제) | `New`(`:161`) | B1 — 미완료 실패 |
| `r.opts.Resolver` | non-nil | `New`(`:164`) | B8 — 미완료 실패 |
| `r.opts.Replayer` | **nil 허용** | `Options` 주석 | `r.replay`가 nil이면 `false, nil` — 관측 절차로 넘어간다 |
| `r.opts.Collector` | non-nil | `New`(`:166`) | B10 — `stableSnapshot`의 오류를 그대로 올린다 |
| `r.opts.Gate` | non-nil | `New`(`:168`) | 해제는 `:317`에서 **한 번만** |
| `clk` | non-nil | `r.clock()` | 없음 |

> **불변식 1 — 순서가 인과다.** 저널 해소 → 계좌 읽기 → 상태 재구성. 저널 해소가 계좌의
> *의미*를 바꿀 수 있으므로 계좌를 먼저 읽으면 곧 움직일 그림에서 상태를 세우게 된다.
> **a102는 이 순서에 손대지 않았다.**
>
> **불변식 2 — 이탈 여덟 중 일곱이 fail-closed다.** 게이트 해제(`:317`)는 유일한 성공
> 경로(`:323`) 직전에만 있다. 429 대기가 길어져 3단계가 실패하면 B10이 `:298`로 나가고
> 게이트는 `New`가 잠근 채 남는다.

## Branches and early returns

`ast.json`의 열거를 그대로 옮긴다.

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:240` | `RecoverPending` 오류 | 없음 | `:241` 미완료 |
| B2 | `:253` | `PendingAttempts` 오류 | 없음 | `:254` 미완료 |
| B3 | `:256` | `range pending` | 반복 | — |
| B4 | `:257` | `rec.State != journal.StateInDoubt` | `report.StillPending` 추가 | `continue` |
| B5 | `:263` | `r.blockedSymbol` 오류 | 없음 | `:264` 미완료 |
| B6 | `:271` | `r.replay` 오류 | 없음 | `:272` 미완료 |
| B7 | `:275` | `settled` | 없음 | `continue` |
| B8 | `:280` | `Resolver.Resolve` 오류 | 없음 | `:281` 미완료 |
| B9 | `:285` | `res.State == StateUnresolvedInDoubt` | `report.Unresolved` 추가 | — |
| B10 | `:297` | `stableSnapshot` 오류 | **`report`에 3단계 실측이 이미 실려 있다** | `:298` 오류 그대로 |
| B11 | `:304` | `LocalStateFromJournal` 오류 | 없음 | `:305` 미완료 |
| B12 | `:319` | `report.Diff.BlocksEntry()` | `Gate.Block(ReasonReconcileMismatch)` | — |

Returns: `:241` · `:254` · `:264` · `:272` · `:281` · `:298` · `:305` · `:323` (AST 8개).

> **B10의 위치가 a102에서 의미를 하나 더 진다.** `SnapshotsTaken`·`RateLimitWaits`·
> `RateLimitWaited`의 대입이 오류 검사 **앞**에 있으므로(`:294-296`), 복구가 rate limit
> 예산을 다 쓰고 실패해도 **얼마나 기다렸는지가 report에 남는다.** 실패한 복구야말로
> 그 수가 필요한 경우다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.opts.Journal.RecoverPending` | 저널의 재시작 규칙 | 오류는 B1 | `:239` |
| `r.opts.Journal.PendingAttempts` | 해소 대상 목록 | 오류는 B2 | `:252` |
| `r.blockedSymbol` | ACKED 심볼 차단 정보 | 오류는 B5 | `:262` |
| `r.replay` | 멱등 재전송으로 정체 회수 | 오류는 B6. replayer nil이면 조용히 `false` | `:270` |
| `r.opts.Resolver.Resolve` | 관측으로 해소 | 오류는 B8 | `:279` |
| **`r.stableSnapshot`** | 안정 스냅샷 + **429 대기 실측** | 오류는 B10 — **재시도는 이 함수 안이 아니라 그 안에 있다** | `:293` (번들 `internal-reconcile--recovery.stablesnapshot`) |
| `LocalStateFromJournal` | 지역 믿음 재구성 | 오류는 B11 | `:303` |
| `r.opts.Comparer.Compare` | 대사 | 오류 없음 | `:308` |
| `r.opts.Gate.Clear` | 복구 latch 해제 | — | `:317` |
| `r.opts.Gate.Block` | 불일치 latch | — | `:320` |

**프로덕션 호출자**: `cmd/tossctl`의 엔진 기동 경로가 `opts.Recover` 클로저로 부른다
(`internal/app/engine/runtime.go:289-294`가 루프 시작 **전**에 호출). 겹2(T2)가 그 클로저를
`recoverThenReady`로 감싼다 — **이 함수는 겹2에서도 무편집이다.**

## State mutations and fallbacks

- `report`를 채우고, 성공 시에만 `r.complete = true`(`:316`)와 `Gate.Clear`(`:317`).
- a102가 더한 상태 변경은 **`report`의 두 필드 대입 두 줄**뿐이다. 게이트·저널·해소기에
  닿는 부분은 그대로다.
- fallback: replay가 정체를 못 회수하면 관측 절차가 받는다(B7의 반대편). a102와 무관.

## Safety conclusion

- Safe edit boundary: `:293-296` 네 줄. 그 밖은 무편집이며 AST 수치가 그것을 지지한다
  (분기 12·이탈 8·호출 24 — 편집 전과 동일).
- High-risk impact: **yes** — 재시작 복구의 절차 소유자.
- 보수 방향: 편집으로 **새로 열리는 경로가 없다.** 게이트 해제 지점은 여전히 하나이고
  그 앞의 조건도 그대로다.

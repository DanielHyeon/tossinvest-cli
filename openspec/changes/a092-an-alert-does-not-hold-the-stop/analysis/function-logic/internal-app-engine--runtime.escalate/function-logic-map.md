# Function Logic Map: `Runtime.escalate`

- Source: `internal/app/engine/runtime.go` (395-426)
- AST evidence: `ast.json` — branches 2, returns 2, calls 8, assignments 1,
  defers 0, go_statements 0
- Risk scan: `risk-pattern-report.md`

**H2가 반증된 자리다.** engine-safety 초안 ¶22는 감독자를 예산에서 면제하며
"감독자는 자기 알림에 이미 별도 기한을 갖고 있다"를 근거로 들었다.
**이 함수는 `Notify`에 두 번 도달하고, 두 번째에는 기한이 없다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 감독 루프 컨텍스트 | `superviseHealth:346-360` | **기한 없음** |
| `loop` | `LoopSpec` | `r.opts.Loops` | — |
| `consecutive` | ≥ `r.threshold` | 호출자 `CheckHealth:386` | — |
| `r.opts.Escalate` | nil 허용 | `RuntimeOptions` | nil이면 B1이 반환 |
| `r.opts.AccountRef` | 비어 있음 허용 | 같은 위 | 공백이면 B1이 반환 |
| `r.opts.Announcer` | nil 허용 | 프로덕션은 `ectx.Notifier` | nil이면 journal이 알리지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | `Notify` 도달 | 기한 |
|---|---|---|---|---|---|
| — `:396` | — | **`r.alert(...)` — `Notify` 도달 #1 (P3)** | — | ✅ | **있음** — `alertCtx`, `alertDeliveryBound = 30s` (`:461`) |
| B1 `:412` | `Escalate == nil \|\| AccountRef 공백` | 없음 | `return` `:413` | ❌ | — |
| — `:415` | — | **`EscalateOperatingMode(ctx, …, r.opts.Announcer)` — `Notify` 도달 #2 (P1b)** | — | ✅ | **없음** — 평범한 감독자 `ctx` |
| B2 `:416` | 승격이 오류 | `r.log(EventOperatingMode, true, …)` `:417` | `return` `:421` | — | — |
| — `:423` | — | `r.log(EventOperatingMode, false, …)` | 암묵 `:426` | — | — |

**`:396`과 `:415`의 `ctx`가 다르다.** AST calls가 `r.alert`(`:396`)과
`r.opts.Escalate.EscalateOperatingMode`(`:415`)를 둘 다 열거하고,
소스에서 앞은 `alertCtx`를, 뒤는 인자 `ctx`를 넘긴다.
**그래서 "감독자는 이미 기한을 갖고 있다"는 절반만 참이다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.alert` `:396` | 저하 알림 | **동기.** `context.WithoutCancel` + `alertDeliveryBound` 30s(`:444-456`) | `internal-app-engine--runtime.alert` FLM |
| `strings.TrimSpace` `:412` | 계정 검사 | 즉시 | AST calls |
| `r.opts.Escalate.EscalateOperatingMode` `:415` | 운영 모드 승격 | **동기·기한 없음.** 커밋(`operating_mode.go:468`) **뒤** `AnnounceOperatingMode`(`:479`) → `obs/mode.go:57` `Notify` | `internal-journal--journal.transitionoperatingmode` FLM |
| `r.log` `:417`·`:423` | 결과 기록 | 즉시 | AST calls |

### `alertDeliveryBound` 주석이 거짓이다

`runtime.go:461`의 주석은 30초 기한이 "전송이 죽어도 종료를 붙잡지 못하게 유계로
만든다"고 말한다. **오늘 1회 전송 상한은 10초·시도 3회·대기 2초로 34초이므로
30초 기한이 먼저 만료된다** — 즉 기한이 유계를 *만드는* 게 아니라 *자르는* 중이다.
a092가 예산을 4.2초로 내리면 그때 비로소 주석이 참이 된다.
**주석 자체의 수정은 a093으로 넘긴다**(a092는 `runtime.go`를 편집하지 않는다).

## State mutations and fallbacks

- 이 함수는 엔진 상태를 바꾸지 않는다. AST assignments 1은 `_, changed, err :=`다.
- `changed`는 `:423`의 로그에만 쓰인다.
- fallback 없음. 승격 실패는 로그 한 줄(B2)이고 루프는 계속 돈다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: **yes** — §0.5(Guardian 경로).
- **a092가 여기서 얻는 사실**: 감독자 면제의 근거는 **`takeLatch`의 래치 하나뿐**이고
  "이미 기한이 있다"는 근거는 `:415`에서 거짓이다. engine-safety ¶22는 그렇게 써야 한다.
- **그래도 면제가 성립하는 이유**: `CheckHealth` B5 `:383`의 래치가 두절당 1회로 묶으므로
  1초 루프가 매초 전송을 기다리지는 않는다. 근거는 둘이 아니라 하나다.

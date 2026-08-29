# Function Logic Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go` (**113-124**) — 편집 전 `113-120`
- AST evidence: `ast.json` — 편집 후 branches **0** / returns **1** / calls **4**
  (편집 전 0 / 1 / **3**)
- Risk scan: `risk-pattern-report.md`
- source SHA-256: **`f13e36b3…`**

## ⛔ 이 함수에는 분기가 없다 — 그래서 무엇을 적는가

`ast.json`의 열거가 정본이고 그것은 **비어 있다**(`"branches": null`). 등록 하나를
더한 편집이 판정을 하나도 안 만든다는 뜻이고, 그것이 이 편집의 성질이다.

**그래도 이 번들이 필요하다.** 규칙은 「분기를 바꿨을 때」가 아니라 **「기존 함수
내부를 편집했을 때」**이고, `check_analysis`도 같은 기준으로 잡는다. 그리고 분기가
없다는 것 자체가 재야 아는 사실이다 — 등록을 조건부로 짰다면(예: 플래그 뒤에
숨기면) 여기 분기가 생겼을 것이고, 그때는 **명령이 안 보이는 경우**가 하나 생긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | `*rootOptions` — nil 가능 | `newRootCmd` | 이 함수는 안 읽는다. 하위 명령이 실행 시점에 읽는다 |
| 하위 명령 집합 | 셋 — `run`·`reconcile-resolve`·**`alerts`** | 이 함수 | 없음. 조립만 한다 |

## Branches and early returns

**분기 0 · 이탈 1.** `ast.json`의 열거가 그대로 이 표다.

| Branch | 위치 | Condition | Return |
|---|---|---|---|
| — | — | **없다** | `:123` — 조립한 `*cobra.Command` |

**등록을 조건부로 하지 않은 것이 결정이다.** *"엔진이 돌 때만 `alerts`를 보여
준다"*는 판정을 여기 넣을 수 있었고 **안 넣었다**: 명령이 보이지 않는 것과 명령이
「엔진이 없다」고 말하는 것은 운영자에게 전혀 다르다. 전자는 자기 tossctl 버전을
의심하게 만들고, 후자는 **의심해야 할 것**을 말한다. 그 문장은 dial 실패 경로에 있다
(`engine_alerts_client_unix.go` `errEngineAlertsUnavailable`).

## Calls and live bindings

| Callee | Why | 오류 계약 |
|---|---|---|
| `cmd.AddCommand` | 하위 명령 등록 | 없음 |
| `newEngineRunCmd` | 기존 | 없음 — 생성자다 |
| `newEngineReconcileResolveCmd` | 기존 | 없음 |
| **`newEngineAlertsCmd` (a098 4.4b-2)** | **밀린 critical 알림의 읽기·승인 표면** | 없음 — 생성자다. 실패는 실행 시점의 dial 에 있다 |

**호출 넷 중 셋이 그대로다.** 3 → 4 가 이 편집의 전부다.

## State mutations and fallbacks

| Mutation | 무엇 | Fallback |
|---|---|---|
| `engine` 명령 트리에 `alerts` 가지 추가 | 하위 명령 둘(`list`·`ack`)이 그 아래 | 없음 — 조립뿐이고 실행하는 것이 없다 |

## Safety conclusion

- Safe edit boundary: **인자 하나**. 기존 둘의 순서와 생성자는 그대로다.
- High-risk impact: **no.** 이 함수는 주문·손절·사이징·인증 경로에 안 닿고, 판정을
  하나도 안 만든다. 위험은 **하위 명령이 실행될 때** 있고 그것은 이 함수 밖이다.
- 되돌리기: 인자 하나를 지우면 오늘로 돌아간다 — 그리고 오늘이 「운영자가 승인할
  수단이 없다」다.

# Function Logic Map: `Runtime.Run`

- Source: `internal/app/engine/runtime.go` (**286-380**) — 편집 전에는 `261-332`
- AST evidence: `ast.json` — 편집 후 branches **5** / returns **3** / calls 34 / defers 4
  (편집 전 4 / 3 / 30 / 3)
- Risk scan: `risk-pattern-report.md`
- source SHA-256: 편집 전 `4bcd9ee7…` → **편집 후 `306276bd…`**

## ✅ 편집 후 재측정 — 안전 성질 둘은 맞고 **id 예측이 틀렸다** (2026-08-12, task 4.1)

| | 편집 전 | 예측 | **편집 후 실측** | |
|---|---:|---:|---:|---|
| 분기 | 4 | **5** | **5** | 맞음 |
| **이탈** | 3 | **3 그대로** | **3** | **맞음 — 이것이 가장 중요한 칸이다** |
| 호출 | 30 | — | 34 | |
| defer | 3 | — | 4 | 새 `defer wg.Done()` |

**이탈 3이 그대로인 것이 「정지 정책을 안 건드렸다」의 실측이다.** 이탈이 4가 됐다면
보조 실행자가 종료 경로를 하나 더 만든 것이고, 그것은 결정 9-2가 안 고른 길이다.

> **⛔ 신설 갈래의 id 는 `B5`가 아니라 `B4`다 — 아래 branch-test-map 의 예측이 틀렸다.**
> 갈래를 **가운데** 넣었으므로 그 뒤의 id 가 **전부 하나씩 밀린다**:
>
> | | 편집 전 | **편집 후** |
> |---|---|---|
> | `Recover != nil` | B1 `:262` | B1 `:289` |
> | 회복 실패 | B2 `:267` | B2 `:294` |
> | `range Loops` | B3 `:277` | B3 `:304` |
> | **`range Auxiliary`** | — | **B4 `:322` (신설)** |
> | `gracefulStop` | **B4** `:304` | **B5** `:352` |
>
> 조건은 하나도 안 바뀌었고 **번호만 바뀌었다.** 그런데 branch-test-map 은 번호로
> 칸을 가리키므로, 그 표를 안 고치면 **「B4 = 취소 경로」라고 적힌 줄이 신설 갈래를
> 가리키게 된다.** buildGateway 번들이 적은 것과 같은 함정이고
> (*"B 번호로 소스 순서를 추론하면 안 된다"*), 이번에는 **삽입이 뒤의 id 를 민다**는
> 다른 얼굴이다. 표를 아래에서 고쳤다.

> **왜 이 산출물이 a098에 있는가** (19라운드 B-P7). a098의 GREEN은
> **`Runtime.Run`이 띄우는 goroutine 집합을 넓힌다**. `Runtime.Run`은 기존 High-risk
> 함수이므로 `.claude/CLAUDE.md`의 규칙상 **면제할 수 없다.**
>
> > **⛔ 5라운드 A-T5**: 이 문단은 *"`Runtime`에 **감독 정책 한 갈래**를 더한다"*라고
> > 적고 있었다. **사용자 결정 9-2가 그것을 죽였다** — 지금 설계는 감독 집합 밖에 두고
> > `stops`에 안 보내는 방식이라 **정책 갈래가 하나도 안 는다.**
> > 3판의 진실이 산출물 머리말에 남아 있었고, 그 머리말이 이 문서의 나머지가
> > 무엇을 예측하는지를 정한다 — **틀린 전제가 붙은 산출물은 예측도 틀린다.**
>
> 19판 4차까지 a098은 번들 셋(`gateway.parkalert`·`journal.pendingalerts`·
> `notifier.flush`)만 갖고 §2를 완료로 선언했고, `check_analysis`가 통과한 이유는
> **프로덕션 Go diff가 아직 비어서** 검사가 볼 것이 없었기 때문이다.
> **초록이 증거가 아니었던 자리다.**
>
> `ast.json`은 a092의 같은 함수 번들과 **바이트 단위로 같다**(같은 HEAD·같은 함수).
> 다른 것은 이 문서다 — a092는 *"편집하지 않는다"*를 적고, a098은
> **무엇을 편집하는지**를 적는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.opts.Recover` | nil 허용 | `NewRuntime`(`runtime.go:183`) | non-nil이고 오류면 **루프를 하나도 안 띄우고 반환**(B1·B2, `:268`) |
| `r.opts.Loops` | `Name` 유일 · `Run` non-nil · `Health` optional | `NewRuntime` B5·B8 검증 | 검증은 생성자에서 끝난다. 여기서는 **슬라이스를 그대로 돈다**(B3 `:277`) |
| `ctx` | 취소 가능 | 호출자 | 취소는 `loopCtx`를 타고 모든 루프에 전파된다 |
| `stops` 채널 | 버퍼 `len(Loops)` | `:275` | **버퍼가 루프 수와 같아서** 아무도 안 읽어도 goroutine이 안 막힌다 |
| **`wg`가 담는 것** | 루프 goroutine **전부** | `:276-283` | **`wg.Wait()`(`:300`)가 전부의 반환을 기다린다** — 이것이 a098의 제약이 걸리는 자리다 |

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — **편집 후** 분기 5 · 이탈 3 · 호출 34 · defer 4.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:289` | `r.opts.Recover != nil` | 없음 | — | 기존 회복 테스트 |
| B2 `:294` | `r.opts.Recover(ctx)`가 오류 | 없음 | **`:295` 반환** — 루프 0개 | 기존 |
| B3 `:304` | `range r.opts.Loops` | `wg.Add(1)` · goroutine 기동 · **`stops`에 보낸다** | — | **a098 R2 — 배달 실행자는 이 슬라이스에 안 들어간다** |
| **B4 `:322` (a098 신설)** | `range r.opts.Auxiliary` | `wg.Add(1)` · goroutine 기동 · **`stops`에 안 보낸다** | — | **a098 R2 · R14** |
| B5 `:352` | `r.gracefulStop(ctx, first.err)` | 로그 한 줄 | **`:356` `nil` 반환** | a098 **R7** — 취소 경로가 이 분기로 나가야 한다 |
| (분기 아님) `:379` | 위 전부 거짓 | `r.alert` + 로그 | **`:379` `ErrLoopFailed` 감싼 오류** | **a098 R2** — 배달 실행자는 **구조적으로** 여기 못 온다 |

**B3와 B4의 차이는 한 줄이고, 그 한 줄이 이 change 전부다.** 둘 다 `wg.Add(1)`을 하고
둘 다 `loopCtx`를 타고 둘 다 같은 `wg.Wait()`가 기다린다. B3만 `stops`에 보낸다.

> **⛔ 지금 B4의 goroutine 은 `aux.Run`의 반환값을 버린다** — `_ = aux.Run(loopCtx)`.
> 「기동하고 배수한다」까지만 4.1이 지고, **관측은 4.2가 진다.** 그래서 4.1 시점에는
> 배달 실행자를 **프로덕션 조립에 등록하지 않는다** — 감독 밖에 두면서 아무도 안 보면
> 승인된 정본이 금지하는 상태(*"「감독 밖」은 곧 「아무도 안 본다」"*)가 된다.

> **⛔⛔ 3판까지 이 표는 정반대를 적었다 — 사용자 결정 9-2가 뒤집었다.**
>
> | | 3판 (미결) | 4판 (결정 9-2) |
> |---|---|---|
> | 배달 실행자의 자리 | **`Loops`에 넣는다**(B3의 원소가 는다) | **`Loops`에 안 넣는다.** 감독 집합 밖의 보조 실행자다 |
> | `:331`을 어떻게 피하는가 | **정지 정책 한 갈래를 더해서** 판정으로 피한다 | **`stops` 채널에 아무것도 안 보내서** 구조적으로 못 온다 |
> | 승인된 정본과의 관계 | `Runtime.Run`의 방어적 종료 계약을 **바꾼다** → `MODIFIED` 필요 | ~~안 바꾼다~~ → **6판이 뒤집었다 (아래)** |
>
> 뒤엣것이 **더 적은 편집으로 더 강한 성질**을 준다. 판정으로 피하는 것은
> 판정이 틀리면 무너지고, 채널에 안 보내는 것은 **틀릴 판정이 없다.**
> **그 성질은 6판에도 그대로다** — 바뀐 것은 정본과의 관계뿐이다.

그 관계가 어떻게 바뀌었는지는 아래에 적는다.

> **⛔ 6판 — 위 표의 마지막 칸이 거짓이었다 (사용자 결정 10-1).**
>
> 4판은 *"정본이 말하는 「루프」는 `Loops`이므로 안 바꾼다"*로 적었다.
> 그 방어는 **정본의 집합이 코드의 집합과 같을 때만** 성립하는데,
> 정본은 **셋**(`openspec/specs/engine-safety/spec.md:170`)이고
> 프로덕션은 **넷**이다(`cmd/tossctl/engine.go:377-398` — 5라운드 A-P1·B-P2).
>
> 결정 10-1이 **`MODIFIED` 하나**를 쓰기로 했다:
> 열거를 **넷**으로 고치고 **보조 실행자** 개념을 더한다
> (`specs/engine-safety/spec.md`의 `## MODIFIED Requirements`).
>
> **이 함수의 예측은 안 바뀐다** — 분기 4→5 · 이탈 3 그대로.
> `MODIFIED`는 문서의 문장을 고치는 것이고 이 함수의 편집 계획을 안 건드린다.

**그래서 a098이 더하는 갈래는 B3 뒤 하나뿐이다.**

| 자리 | 무엇 | 왜 거기여야 하는가 |
|---|---|---|
| B3 **뒤** | `range r.opts.Auxiliary`로 보조 goroutine 기동 (**분기 하나 는다**) | 루프와 같은 `loopCtx`를 타야 취소가 함께 간다 |
| **그 goroutine 안** | `wg.Add(1)` / `defer wg.Done()` — **감독 루프와 똑같이.** 다른 것은 **`stops`에 안 보낸다**는 것뿐 | **아래 ⛔가 그 이유다** |

> **⛔ 보조 실행자를 안 기다리면 원장이 그 아래에서 닫힌다.**
> 이 함수의 doc(`:258-260`)이 계약을 적는다 — *"when this returns, every goroutine it
> started has returned too. That is what lets the caller close the journal immediately
> afterwards without racing a loop that is still writing."*
> 배달 실행자는 **원장에 쓴다**(claim·정산).
>
> **답은 새 WaitGroup이 아니라 기존 `wg`다.** 4판 첫 초안은 `auxWG`를 만들고
> `:300` 옆에 `auxWG.Wait()`를 더하라고 적었는데, 그것은 **잊을 수 있는 자리를 둘
> 만든다**(`Add`와 `Wait`). 보조 goroutine을 `wg`에 넣으면 `wg.Wait()`(`:300`)가
> **이미** 그들을 기다리고, 이 함수의 편집은 **range 하나로 끝난다.**
>
> **위험이 0이 되지는 않는다** — `wg.Add(1)`을 빼먹으면 여전히 깨진다.
> 줄어든 것은 자리의 수다. 관측은 **R14**가 진다.

**채널 산술이 그대로 맞는지 확인했다.** `stops` 버퍼는 `len(r.opts.Loops)`(`:275`)이고
보조 실행자가 거기 **안 보내므로 송신자 수가 여전히 `len(Loops)`와 정확히 같다.**
`first := <-stops`(`:297`)에 보조 실행자가 절대 안 나타나고, `drain(stops)`(`:302`)도 안 막힌다.
**이 셋이 다 참이어야 「감독 밖」이 성립한다** — 하나라도 어긋나면 보조 실행자가
`:331`의 엔진 정지 경로로 새어 들어간다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loop.Run(loopCtx)` `:280` | 루프 본체 | **기한 없음.** ctx 취소가 유일한 종료 신호다 | `ast.json` calls |
| `wg.Wait()` **`:300`** | 모든 루프 goroutine의 반환을 기다린다 | **상한이 없다** — 루프가 안 돌아오면 여기서 안 돌아온다 | `ast.json` calls · 소스 `:300` |
| `supervisorWG.Wait()` `:301` | 감독 goroutine | 같음 | `ast.json` calls |
| `r.gracefulStop(ctx, first.err)` `:304` | 취소인지 자발적 정지인지 판정 | — | `ast.json` calls |
| `r.alert(ctx, …)` `:315` | 루프 실패 알림 | `alertDeliveryBound` 30초(`Runtime.alert` 번들) | `ast.json` calls |
| `drain(stops)` `:302` | 나머지 정지 사유 수집 | 버퍼가 루프 수와 같아 막히지 않는다 | `ast.json` calls |

## State mutations and fallbacks

- `loopCtx`/`cancel` — `defer cancel()`(`:272`)와 `:298`의 명시 호출 둘 다 있다.
- `wg`·`supervisorWG` — goroutine 수명 소유. **`wg`가 담는 것이 는다** —
  배달 실행자는 `Loops`에 안 들어가지만 **같은 `wg`에는 들어간다.**
  **세 번째 WaitGroup을 만들지 않는다**(위 ⛔).
- `stops` 채널 — 첫 값이 정책을 정하고 나머지는 `drain`이 읽는다.
- **폴백 없음.** 이 함수는 재시작하지 않는다(`:258-260`의 doc comment가
  *"when this returns, every goroutine it started has returned too"*를 계약으로 적는다).

## Safety conclusion

- **Safe edit boundary**: a098은 **B3 뒤에 range 하나**를 더한다. `Wait` 호출은
  **안 더한다** — `wg.Wait()`(`:300`)를 그대로 쓴다.
  B1~B4의 조건과 **세 이탈 전부**의 반환은 안 바꾼다.
  편집 후 AST의 branches가 **5**, returns가 **3 그대로**면 의도한 편집이다 —
  **returns가 늘면 정지 정책을 건드린 것이고, 그것은 결정 9-2가 안 고른 길이다.**
- **High-risk impact**: **yes** — 엔진 생사의 자리다.
- **a098이 여기서 지는 제약 셋** (셋째는 결정 9-2가 만들었다):

  | 제약 | 무엇 | 관측 |
  |---|---|---|
  | 배달 실행자의 `Run`이 **반환해도** | `stops`에 안 보내므로 `:297`의 `first`가 안 깨어난다. 다른 루프는 계속 돈다 | **R2 · R3** |
  | 배달 실행자의 `Run`이 **안 돌아오면** | **`wg.Wait()`(`:300`)가 안 돌아온다.** 감독 루프와 같은 줄이지만 **그 줄에 걸리는 goroutine이 하나 는다** — 감독 루프 **넷**이 전부 정상 취소돼도 실행자 하나 때문에 엔진이 못 죽는다. `Runtime.Run` 자체에는 상한이 없고, **상한은 실행자가 ctx를 보는 것으로만 온다** | **R7** (상한 단언) |
  | 보조 goroutine을 `wg`에 **안 넣으면** | doc `:258-260`의 원장 close 계약이 깨진다. 감독 밖으로 뺀 대가이고, **`wg.Add(1)` 한 줄에 걸린다** | **R14** |

  **앞의 둘은 반대 방향이고 같은 한 줄에서 만난다.** 19판이 배달 루프를 a098로 분리할 때
  **앞의 것만 요구로 건너왔고**, 뒤의 것은 a092의 이 함수 산출물에만 남아 있었다
  (a092 §7.4.3 · 19라운드 B-P6이 기전을 정정).

- **`Health`는 붙이지 않는다 — 결정 9-2가 그것도 정했다.** `Health`는
  `SupervisedLoop`의 필드이고 `superviseHealth`(`:290`)가 `Loops`만 훑는다.
  배달 실행자가 `Loops` 밖이면 **열화 임계 사다리에 올라갈 자리가 없다.**
  그 대가는 *"사이클이 연속 실패하지만 살아 있는 배달 실행자"*를 아무도 안 센다는 것이고,
  **숨기지 않는다** — design D8.4가 그것을 적고 후속으로 남긴다.

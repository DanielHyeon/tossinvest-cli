# a087 §0.4 — CodeGraph hard evidence (네 심볼)

- 측정 순간: 2026-10-10, HEAD `6e844e116686b372dcf49c6324a910706e9c48a6`(브랜치 `feat/a112-four-family-runtime`)
- 도구: **codegraph 1.6.0** (index: 2,623 files · 46,504 nodes · 172,742 edges)
- `make sdd-sync` rc=2 (2회 실행, 동일) — CodeGraph `sync` 성공("Already up to date"), `.sdd/index-state.json`
  의 codegraph fingerprint `1f4578e1…2621` · head `6e844e11` 로 기록됨. **CodeGraphContext `update` 가 두 번 다
  300초 타임아웃**(advisory, 미갱신 — 마지막 성공 2026-10-06 head `5ea91e15`). GBrain 은 `gbrain serve`(pid 23201)
  점유로 busy → 이전 freshness 유지. 아래 수치는 CodeGraph(hard evidence)만 근거로 한다.
- 명령: `codegraph query|callers|impact <sym>`(impact 기본 depth 2, 아래 표는 `-d 5`),
  `codegraph affected <file> --filter '*_test.go' -q`

## 심볼별 정의·호출자

| 심볼 | 정의 (HEAD 대조 확인) | non-test 호출자 | test 호출자(직접) |
| --- | --- | --- | --- |
| `ExitObserver.sellIntent` | `internal/app/engine/exitloop.go:1697` `(m managed, quantity, observed string) (orderintent.PlaceIntent, error)` | **1** — `ExitObserver.submit` (`exitloop.go:1402`) | 0 |
| `checkOrderShape` | `internal/execgw/failclosed.go:53` `(intent orderintent.PlaceIntent) *RejectedError` | **1** — `CheckPlace` (`failclosed.go:40`) | 0 |
| `isProtective` | `internal/app/engine/exitloop.go:1382` `(p exitpolicy.Proposal) bool` | **2** — `ExitObserver.record` (`exitloop.go:1232`), `ExitObserver.submit` (`exitloop.go:1402`) | 1 — `IsProtectiveForTest` (`a091_export_test.go:8`, export shim) |
| `buildOrderCreate` | `internal/official/orders_write.go:100` `(intent orderintent.PlaceIntent) (any, error)` | **1** — `PlaceOrder` (`orders_write.go:182`) | 6 — `TestBuildOrderCreate{FractionalBuy,FractionalSell,Limit,Market,CarriesClientOrderID,Errors}` (`orders_write_test.go`) |

## impact (depth 5, 함수 단위 역호출)

| 심볼 | 영향 심볼 수 | 비시험 상류 경로 요지 |
| --- | --- | --- |
| `sellIntent` | 11 | `submit`(1402) ← `record`(1232) ← `judgeRatchet`(969)/`judgeLadder`(1023) ← `judge`(905) ← `ObserveOnce`(444) — 방향은 `codegraph callers` 로 단계별 확인 |
| `checkOrderShape` | 19 | `CheckPlace` ← `execgw.place`/`Place` ← engine `Place`(exitloop.go:138)·`submit`, `PlaceClaimedStrategy`(execgw·engine), `strategydispatch.PlaceStrategyEntry`, **`flatten.sell`/`Liquidate`** |
| `isProtective` | 24 | `record`(1232)·`submit`(1402, `record` 가 부름) ← `judgeRatchet`/`judgeLadder` ← `judge` ← `ObserveOnce` ← `Run`(382), `tracer.Run`(tracer.go:273) |
| `buildOrderCreate` | 36 | `official.PlaceOrder` ← `ops.PlacePendingOrder`/`placeHandler`, `trading.Place`, `execgw.place`, engine `Place`, `hybrid.PlaceOrder`, **`verifylive.PlaceOrder`** ← `verifylive` steps(idempotency·cancel·amend·sellBoundary) |

관찰(§4 회귀 방지 범위와 직결):

- `checkOrderShape` 의 상류에 `flatten.Liquidate` 가 있다 — §1 게이트 변경은 flatten 경로도 통과한다(tasks 4.1).
- `buildOrderCreate` 의 상류에 `verifylive` 가 있다 — tasks 4.2 범위.
- `checkOrderShape` 의 상류에 진입 경로(`strategydispatch.PlaceStrategyEntry`)가 있다 — tasks 1.3(진입은 여전히 시장가 거부) 범위.

### 이름 충돌로 보이는 거짓 엣지 (HEAD 직접 대조)

impact 에 잡혔으나 대상 심볼을 부르지 않는 시험 — CodeGraph 가 메서드 이름을 로컬 클로저·필드로 해소한 것:

- `internal/candidate/metrics_test.go:175` `TestAOneMinuteGapAndATenMinuteGapAreNotTheSameRate` — 로컬 `record := func(...)`
- `internal/journal/durability_test.go:539` `TestPrepareFailureBlocksSubmission` — 로컬 `submit := func(...)`
- `internal/app/engine/a091_replay_test.go:398` `a091LaterStopUnder` — 필드 `r.submit.record` (이건 engine 시험 하네스의 필드라 실제로 exit 경로를 몰 수는 있으나, 엣지 근거는 이름이다)

## 영향 시험 수 (파일 단위, `affected` depth 5)

| 정의 파일 | `--filter '*_test.go'` | 기본 판별식 |
| --- | --- | --- |
| `internal/app/engine/exitloop.go` | **1081** | 1 (`auth-helper/tests/test_cli.py`) |
| `internal/execgw/failclosed.go` | **1035** | 1 (`auth-helper/tests/test_cli.py`) |
| `internal/official/orders_write.go` | **1062** | 1 (`auth-helper/tests/test_cli.py`) |

기본 판별식은 Go `_test.go` 를 못 잡는다(`.claude/CLAUDE.md` 하네스 주석) — 이번에도 재현(1 대 1000+).
파일 단위 수치는 의존 폐포가 넓어 사실상 저장소 Go 시험 대부분이며, **변경 범위를 좁히는 근거가 아니다**.
범위 판단은 위의 함수 단위 impact 와 Function Logic Map(§2.1)으로 한다.

## 문서 인용과의 표류 (HEAD 기준)

- proposal 의 `exitloop.go:1452`("Automated orders are LIMIT only" 주석) → 현재 **`exitloop.go:1684`** (sellIntent 는 1697).
- tasks 1.2 의 `failclosed.go:84` `orderType != "limit"` — 현재도 **84** (일치).
- 줄 번호 표류는 의미 변경이 아니므로 이 게이트에서는 기록만 하고 proposal 은 고치지 않는다.

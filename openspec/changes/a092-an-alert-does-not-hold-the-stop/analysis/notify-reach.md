# `obs.Notifier.Notify`에 도달하는 프로덕션 경로 전수 (11판)

- 기준: `ec29dc72`
- 방법: `grep -rn "\.Notify(" --include=*.go internal/ cmd/ | grep -v _test.go` 전수 +
  `AnnounceOperatingMode` / `Announcer` 배선 추적
- **열거의 단위는 파일도 호출자 메서드도 아니라 `Notify`에 도달하는 경로다.**
  a091은 단위를 파일로 잡아 틀렸고, a092 초판은 `o.alert`으로 잡아 또 틀렸다.
- **3판 추가: 각 경로가 속한 루프의 주기와 래치를 함께 적는다.** 2판은 경로를 다 열거해
  놓고 **주기를 묻지 않았고**, 그래서 델타가 "가장 짧은 주기"를 exit 관측이라고 잘못 썼다.

## 직접 `.Notify(`를 부르는 프로덕션 자리 — 7곳

| # | 자리 | 루프/프로세스 | **주기** | 기한 | 래치 |
|---|---|---|---|---|---|
| P1 | `internal/obs/mode.go:57` | `AnnounceOperatingMode` — P1a~P1d의 종착지 | 호출자에 따름 | 없음 | `changed==true`일 때만 |
| P2 | `internal/app/engine/reconcileloop.go:556` | `ReconcileDriver.alert` — 대사 루프 | 60s (`DefaultReconcilePeriod`, `reconcileloop.go:95`) | 없음 | 호출자별 |
| P3 | `internal/app/engine/runtime.go:453` | `Runtime.alert` — 감독자 | **1s** (`DefaultHealthInterval`, `runtime.go:84`) | **30s** (`alertDeliveryBound`) | `takeLatch` (`runtime.go:383`) |
| P4 | `internal/app/engine/exitloop.go:1604` | `ExitObserver.alert` — **exit 관측 루프** | **5s** (`DefaultExitObservationInterval`) | 없음 | 호출자별 — 아래 표 |
| P5 | `internal/app/engine/exitwiring.go:103` | `notifierAlerter.ExternalPositionFound` — 대사 루프 | 60s | 없음 | 호출자 |
| P6 | `internal/app/engine/exitwiring.go:141` | `notifierAlerter.ManagedPositionClosedExternally` — 대사 루프 | 60s | 없음 | 호출자 |
| P7 | `internal/flatten/flatten.go:694` | `flatten.Saga` — **프로덕션에서 `Notifier`가 nil**(`cmd/tossctl/flatten.go:247`) | — | — | — |

**가장 짧은 주기의 루프는 P3(감독자, 1초)이지 exit 관측(5초)이 아니다.**

감독자를 예산 기준에서 면제할 수 있는 근거는 **`takeLatch` 하나뿐이다.**
`CheckHealth` B5(`runtime.go:383`)가 루프 이름당 래치하고 해제는 복구 시(B4 `:375`,
`consecutive == 0`)뿐이므로, 1초 루프가 매 초 전송을 기다리지는 않는다 — 두절당 1회다
(`internal-app-engine--runtime.takelatch` FLM, `…--runtime.checkhealth` FLM).

**"감독자는 이미 기한을 갖고 있다"는 근거로 쓸 수 없다.** 절반만 참이다 —
`Runtime.escalate`는 `Notify`에 **두 번** 도달하는데 앞의 `r.alert`(`runtime.go:396`)만
`alertDeliveryBound` 30초를 갖고, 뒤의 `EscalateOperatingMode`(`runtime.go:415`,
아래 P1b)는 **평범한 감독자 `ctx`를 그대로 넘긴다 — 기한이 없다**
(`internal-app-engine--runtime.escalate` FLM). 아래 P1 표의 "기한 = 없음"이 그것이다.

## P1(Announcer)에 도달하는 자리 — 4곳

`journal/operating_mode.go:479` `req.Announcer.AnnounceOperatingMode(ctx, ...)`가 P1이고,
`Announcer`는 프로덕션에서 `ectx.Notifier`다(`cmd/tossctl/engine.go:351`·`:370`).

| # | 자리 | 루프 | 오늘 발동하는가 |
|---|---|---|---|
| P1a | `internal/app/engine/exitloop.go:796` | **exit 관측** | **아니오** — 아래 |
| P1b | `internal/app/engine/runtime.go:415` | 감독자 (1s) | 아니오 — 같은 이유 |
| P1c | `internal/execgw/retry.go:386` | Retrier (호출자 goroutine) | 아니오 — 같은 이유 |
| P1d | `internal/execgw/riskguardian.go` (`announcer` 필드) | Guardian | 아니오 — 같은 이유 |

### P1은 오늘 어느 자리에서도 발동하지 않는다

`TransitionOperatingMode`는 계정이 이미 목표 모드면 `direction == 0`에서 **announce 전에**
`changed=false`로 반환한다(`operating_mode.go:409-415`). 위 네 트리거의 목표는 전부
`ModeEntryBlocked`다(`TargetModeForTrigger`, `operating_mode.go:537-549`).

그리고 계정은 **2026-07-31T09:55:49부터 `ENTRY_BLOCKED`이고 완화된 적이 없다**
(`operating_modes` 1행). 로그도 같은 말을 한다 — `AnnounceOperatingMode`가 쓴 모양의 줄
(`from_state` 필드를 가진 `engine.operating_mode`)은 **전체 로그에 line 372 하나뿐**이다
(`analysis/delivery-latency.md` §0).

> **⇒ 오늘 두절 사이클은 `Notify`에 한 번 도달한다. 두 번이 아니다.**
> 2 × 34s = 68s는 **NORMAL 계정에서만** 성립하는 상한이다.

## P4의 호출자 7곳과 래치 — 사이클 배수의 근거

| 자리 | 이벤트 | 등급 | **래치** |
|---|---|---|---|
| `exitloop.go:780` `checkOutage` | `exit.observation_outage` | critical | `o.outageRaised` (`:775-778`) — **두절 에피소드당 1회** |
| `exitloop.go:1430` `applyFloor` | `exit.proposal_capped` | normal | 없음 |
| `exitloop.go:1500` `alertUnmanaged` | `exit.position_unmanaged` | normal | `o.unmanaged[p.ID]` — 포지션당 1회 |
| `exitloop.go:1526` `alertRefused` | `exit.judgement_refused` | critical | `o.refused[m.position.ID]` (`:1517-1520`) — 포지션당 1회 |
| `exitloop.go:1550` `alertProposalRefused` | `exit.proposal_refused` | critical | **없음** (`:1548-1565`) |
| `exitloop.go:1580` `noteDelay` | `exit.liquidation_delayed` | critical | 지연 시계 |
| `exit_quarantine_announce.go:71` | `exit.snapshot_quarantined` | critical | 격리 시점 |

**`alertProposalRefused`에 래치가 없다 — AST 열거로 확정했다.**
`ast.json`이 **`"branches": null`·`"returns": null`**를 내놓는다(`…--exitobserver.alertproposalrefused`
FLM). 조건문이 0개이므로 조기 반환도 중복 억제도 어떤 게이트도 **없다.**
"래치가 없다"는 부재 주장이고, 손으로 읽어서는 못 본 것인지 없는 것인지 구별되지 않는다.

호출자는 `:1264`·`:1309` 두 곳이고 둘 다 `submit` 안이다. 한 사이클의 `range states`
순회에서 **포지션마다, 레벨마다** 올라갈 수 있다. 프로덕션 로그가 그것을 보인다 —
005930이 여러 사이클에 걸쳐 반복(line 6896·6899·6910·6912).

> **⇒ 한 사이클의 최악 체류는 `예산 × (그 사이클이 올린 알림 수)`이고,
> 그 수는 포지션 수에 비례한다. a092는 알림 **하나**의 예산을 정하지 사이클의
> 총합을 정하지 않는다.**

## `Notify` 아래의 직렬화 — `n.mu`

`deliver`는 `n.mu`를 **재시도 루프 전체(대기 포함)** 보유한다(`notifier.go:241-242`).
그리고 `*obs.Notifier`는 **하나**다 — `gateway.go:280`에서 만들어져
`Context.Notifier`(`engine.go:571`)로 게시되고 exit 관측(`exitwiring.go:341-342`) ·
대사 드라이버(`reconcileloop.go:367-368`) · Retrier(`exitwiring.go:54`) ·
Guardian(`guardian_wiring.go:81`) · 감독자(`cmd/tossctl/engine.go:368`·`:370`)에게
**같은 포인터**로 간다. 그 루프들은 `runtime.go:276-283`이 별도 goroutine으로 띄운다.

⇒ exit 관측의 critical 알림 하나가 실제로 쓰는 시간은
**자기 예산 + 다른 루프의 `deliver`를 기다린 시간**이다.

**a092는 이것을 고치지 않는다.** 고치는 것은 각 `deliver`가 뮤텍스를 쥐고 있는 **시간**이다
(34s → 4.2s). 경합 배수는 그대로 남고 a093 대상이다.

그러므로 **spec은 "동기 체류의 합"을 약속하면 안 된다.** 경합이 없어도 사이클에 여러
알림이 올라가고(위 `alertProposalRefused`), 경합이 있으면 합은 유계가 아니다.
약속할 수 있는 것은 **알림 하나가 자기 전송에 쓰는 시간**이다.

(`publishBestEffort`는 `n.mu`를 잡지 않는다 — 직렬화되는 것은 critical 배달뿐이다.)

## `Notify`를 지나지 않는 발송 경로

`Notifier.Flush`(`notifier.go:307-336`)는 `Publisher.Publish`를 직접 부른다.
**프로덕션 호출자가 0곳**이므로 오늘은 발송이 일어나지 않는다. a093 대상.

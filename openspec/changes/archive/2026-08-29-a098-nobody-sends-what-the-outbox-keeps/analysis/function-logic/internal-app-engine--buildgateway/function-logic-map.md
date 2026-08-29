# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go` (**234-355**) — 편집 전에는 `199-312`
- AST evidence: `ast.json` — AST 기준 branches **5** / returns **7** / calls 20 / assignments 19
  (편집 전 4 / 6 / 18 / 18)
- Risk scan: `risk-pattern-report.md`
- source SHA-256: 편집 전 `de44aae5…` → **편집 후 `cf483384…`**

## ⛔⛔ 이 번들은 승인된 task 가 지목하지 **않은** 함수다 — 왜 생겼는지 먼저 적는다

task 4.6 은 기동 복원의 자리를 **`RuntimeOptions.Recover` 콜백**이라고 적는다
(design D6 · 사용자 결정 5-1). 구현하면서 **더 이른 자리**를 찾았고, 그 자리가
이 함수다. 자리를 바꾸면 편집 대상 함수가 바뀌므로 **먼저 이 산출물을 만든다** —
`.claude/CLAUDE.md`의 「분기를 근거로 삼는 문서는 AST 를 먼저」가 여기 걸린다.

**결정의 근거 넷.** 넷 다 실측이다.

| # | 사실 | 근거 |
|---|---|---|
| 1 | 이 함수는 **어떤 루프보다도 먼저** 끝난다 — 조립이고, 루프는 `Runtime.Run` 이 띄운다 | `engine.go:490` 호출 · `runtime.go:277-283` |
| 2 | **같은 결함을 이미 고친 선례가 이 함수 안에 있다** — `tracker.Restore`(B2, `:226`)가 RECONCILE 계열 래치를 원장에서 재건한다. 주석: *"without this call a restart silently clears every block a disagreement raised"* | `gateway.go:216-227` |
| 3 | 원장(`in.journal`)과 게이트(`entry`)가 **둘 다 이 함수 안에 있다.** `Recover` 클로저(`cmd/tossctl/engine.go:373`)에는 게이트가 없다 | `gateway.go:214`·`:199` |
| 4 | **복구 시퀀스가 내 래치를 안 지운다** — `recovery.Run` 의 `Clear` 는 `ReasonRecoveryIncomplete` **하나뿐**이다 | `internal/reconcile/recovery.go:289` |

> **design 이 준 이유를 어기지 않는다.** D6 가 `Recover` 를 고른 논거는
> *"첫 tick 은 **루프가 이미 도는 뒤**라 그 사이에 진입이 열린다"* 였다 —
> 즉 **3판의 「루프 첫 tick」을 기각하는 논거**이지 `Recover` 를 유일한 자리로
> 지목하는 논거가 아니다. 이 함수는 그 논거를 **더 강하게** 만족한다(사실 1).
>
> **4번이 없었다면 이 자리를 못 골랐다.** 복구가 게이트를 통째로 비웠다면
> 조립 시점의 래치는 `Recover` 가 도는 사이에 사라졌을 것이다. `recovery.go:289`
> 한 줄이 그것을 정한다 — **추측이 아니라 읽어서 확인했다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.journal` | non-nil · 훅 바인딩 완료 | `engine.go:484` `bindApplyHooks` | B1 `:204` — `checkProjectionWired` 가 거절하면 **조립 실패** |
| `in.official` | **non-nil 필수** | `engine.go:493` | nil 이면 `:265` `in.official.BaseURL()` 에서 **패닉** — 검증이 없다 |
| `in.trading` | **non-nil 필수** | `engine.go:492` | nil 이면 B4 `:274` — `execgw: a trading service is required` |
| `entry` | 새로 만든다 | `:214` `NewEntryGate(in.clock, nil)` | nil 임계 → **기본 staleness**. 한 번도 관측 안 된 질의가 이미 진입을 막는다 |
| outbox 의 PENDING critical | 0..N | `alert_outbox` | **오늘 아무도 안 읽는다** — a098 R8 이 그 자리다 |

> **⚠ 기동 직후 `CheckEntry()` 는 이미 non-nil 이다** (`:211-213` 주석).
> 그래서 *"진입이 막힌다"* 로 관측하면 **4.6 이 있든 없든 통과한다.**
> R8 이 `Blocks()` 를 지정하는 이유이고, 래치 맵은 `Block` 이 부른 것만 담는다.

## Branches and early returns

`ast.json` 의 열거를 그대로 옮긴다.

| Branch | 위치 | Condition | Return |
|---|---|---|---|
| B1 | `:204` | `checkProjectionWired` 오류 | `:205` |
| B2 | `:226` | `tracker.Restore` 오류 | `:227` |
| B3 | `:250` | `NewPairedReadinessAdapter` 오류 | `:251` |
| B4 | `:274` | `execgw.New` 오류 | `:275` |

이탈 여섯 = 위 넷 + `:230`(`SetAuthorityRefresh` 클로저 안의 `return`) + `:297`(정상).

> **`:230` 은 `buildGateway` 의 이탈이 아니다.** `entry.SetAuthorityRefresh(func() error {
> return tracker.Refresh(...) })` 안의 것이고, AST 추출기가 함수 리터럴 안까지 센다.
> **여섯 중 하나는 이 함수가 반환하는 자리가 아니다** — 그 사실을 안 적으면
> 다음 사람이 이탈 수로 종료 경로를 세다가 하나를 더 센다.

## Calls and live bindings

| Callee | Why | 오류 계약 |
|---|---|---|
| `checkProjectionWired` | 훅 바인딩 확인 | B1 — 조립 실패 |
| `execgw.NewEntryGate` | 게이트 신설 | 오류 없음 |
| `tracker.Restore` | **RECONCILE 래치를 원장에서 재건** | B2 — 조립 실패 |
| `entry.SetAuthorityRefresh` | 권위 갱신 훅 | — |
| `protection.NewPairedReadinessAdapter` | 보호 준비도 | B3 |
| `execgw.New` | 실행 게이트웨이 | B4 |
| `newNotifier` (`:280`) | 알림 경로 — **게이트를 받는다** | 오류 없음 |
| `newRetrier` (`:281`) | 질의 재시도 | 오류 없음 |

## 편집 예측 (task 4.6)

`tracker.Restore` **바로 뒤**에 한 갈래를 더한다 — 원장→게이트 복원 둘이 붙어 있게.

| 자리 | 편집 | 예측 |
|---|---|---|
| `buildGateway` | `if err := restoreAlertEntryLatch(...); err != nil { return }` | **분기 4 → 5 · 이탈 6 → 7 · 호출 18 → 19** |
| `restoreAlertEntryLatch` | **신설 leaf 함수** | 새 함수이므로 `not-applicable` |

**분기가 5를 넘으면 판정을 하나 더 만든 것이다.** 4.6 은 *"`Block` 만 건다.
`Clear` 는 안 한다"* 이므로 갈래는 하나여야 한다 — 「미전달이 0이면 푼다」를
넣으면 갈래가 둘이 되고, 그것이 승인된 정본
(`openspec/specs/engine-safety/spec.md:147-152`, 전달 복구 뒤 **수동 확인**)보다
덜 보수적이 되는 자리다.

## ✅ 편집 후 재측정 — 둘은 맞고 하나는 틀렸다 (2026-08-12)

| | 편집 전 | 예측 | **편집 후 실측** | |
|---|---:|---:|---:|---|
| 분기 | 4 | 5 | **5** | 맞음 |
| 이탈 | 6 | 7 | **7** | 맞음 |
| 호출 | 18 | 19 | **20** | **틀림 — 하나 더** |
| 대입 | 18 | — | 19 | |
| 범위 | `:199-312` | — | `:234-355` | leaf 를 위에 넣어 밀렸다 |

**호출 예측이 하나 모자랐던 이유**: 새 갈래는 호출을 **둘** 만든다 —
`restoreAlertEntryLatch(...)` 와, 실패를 감싸는 **`fmt.Errorf`**. 앞엣것만 셌다.
**옆의 네 갈래가 전부 같은 모양(`if err := f(); err != nil { return fmt.Errorf(...) }`)
이었으므로 셀 수 있었다** — 못 본 것이지 못 셀 것이 아니었다.

> **분기·이탈이 맞은 것이 이 예측의 값이다.** 그 둘이 이 편집의 안전 성질을 진다 —
> 분기가 6이면 판정을 하나 더 만든 것이고, 이탈이 8이면 종료 경로를 하나 더 만든 것이다.
> 호출 수는 그 성질을 안 진다. **그래도 틀린 것은 틀렸다고 적는다.**

편집 후 분기·이탈 위치: B1 `:239` · B2 `:261` · **B5 `:269`(신설)** · B3 `:293` · B4 `:317`,
이탈 `:240`·`:262`·`:270`·`:273`·`:294`·`:318`·`:340`.

> ⚠ AST 가 붙인 **id 순서와 소스 순서가 다르다** — 신설 갈래는 `:269` 인데 `B5` 다.
> id 는 발견 순서가 아니라 추출기의 순회 순서다. **B 번호로 소스 순서를 추론하면 안 된다.**

## State mutations and fallbacks

이 함수는 **조립**이다. 원장에 쓰지 않고, 프로세스 메모리 안의 상태 둘을
원장에서 **재건**한다.

| Mutation | 무엇을 바꾸는가 | Fallback |
|---|---|---|
| `tracker.Restore` (B2) | `entry` 의 **RECONCILE 계열 래치** + tracker 의 block set | 없음 — 실패하면 조립을 거절한다 |
| **`restoreAlertEntryLatch` (B5, a098)** | `entry` 의 **`ReasonAlertUndelivered` 래치** | 없음 — 실패하면 조립을 거절한다. **미전달 수를 못 읽는 것은 게이트를 열 이유가 아니다** |
| `entry.SetAuthorityRefresh` | 게이트의 권위 갱신 훅 | — |
| `execgw.New` (B4) | 새 gateway 를 만든다 | 없음 |

**둘 다 잠그는 쪽으로만 간다.** `restoreAlertEntryLatch` 는 `Block` 만 부르고
`Clear` 를 **한 번도 안 부른다** — 「미전달이 0이면 푼다」는 승인된 정본
(`openspec/specs/engine-safety/spec.md:147-152`, 전달 복구 뒤 **수동 확인**)보다
덜 보수적이다. 전달이 복구된 것과 사람이 그 알림을 본 것은 다른 사실이고,
뒤엣것을 대표하는 것은 `Notifier.Acknowledge` 뿐이다.

**래치의 detail 은 개수만 싣는다** (불변식 8). 원장의 행은 알림 제목·본문·payload·
계좌를 들고 오고(`alertSelect`, `outbox.go:510-513`), 게이트 detail 은 게이트 상태가
읽히는 모든 자리에서 함께 읽힌다.

## Safety conclusion

- Safe edit boundary: `tracker.Restore` 뒤 한 갈래. **`execgw.New` 호출과 그 옵션은
  안 건드린다** — 게이트웨이 구성이 바뀌면 이 change 의 범위를 넘는다.
- High-risk impact: **yes** — 진입 게이트. 방향은 **보수적**이다(잠그기만 한다).
- 되돌리기: 그 갈래를 지우면 정확히 오늘로 돌아간다 — **그리고 오늘이 결함 상태다.**

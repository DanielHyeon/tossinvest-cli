# Review — console-owns-the-operating-toggles

## Pre-Edit Gate

- base-commit: `b50f991` (`interlock-gates-entry-not-exit`)
- 관측한 근거: `internal/console`의 세 기존 seam과 그 정적 가드
  (`static_test.go`·`orders_static_test.go`·`engineproc_test.go`),
  `internal/config/limits_io.go`의 per-key splice와 그것이 적은 이유,
  `internal/trading/service.go`의 정책 검사, `internal/app/engine/interlock.go` 3절,
  `.claude/CLAUDE.md` §0.7의 실제 문장.
- 사용자 지시 2026-07-30: "자꾸 수동으로 뭔가를 설정하라고 하지? 메뉴를 만들어 주던가 해야지."

## RED 관측

| 테스트 | RED |
|---|---|
| `TestGateOnRefusals/placing_is_disabled` | `interlock_test.go:553: startup must be refused` |
| `TestGateOnRefusals/cancelling_is_disabled` | 동상 |
| `TestTradingPolicyClauseNamesWhatIsOff` | `a policy that can sell and act live must pass: trading.place and trading.cancel is off` — 확장 직후, 기대값이 옛 요구를 그대로 담고 있어 잡혔다 |
| 정적 가드 3건 | `/settings/gate names "gate"`, `Options declares "TradingPolicy"`, `Gate carries verb exemptions` — 새 라우트·seam·예외가 전부 argued-for를 요구하는 가드에 정확히 걸렸다 |

GREEN: `go test ./...` **3,912건 통과, 실패 0** (57 패키지, +23). `go vet` 무이슈.

## 적대적 리뷰

### A1 — 게이트를 콘솔에 넣은 것이 §0.7 위반인가

**지적.** `.claude/CLAUDE.md` §0.7은 "운영 토글 flip과 live 검증은 사람이 직접 승인한다"이고,
스펙에는 "게이트 ON은 §0.7 콘솔 밖 절차 유지"가 명시돼 있었다.

**판정: 스펙 문장은 이 change가 개정했고, §0.7 자체는 위반하지 않는다.**
§0.7은 **승인의 주체**를 정한다. 승인의 장소를 정하지 않는다. 루프백 콘솔에서
세션+CSRF를 통과한 사람이 버튼을 누르는 것은 사람이 직접 승인하는 것이고, 그 사람은
`config.json`을 손으로 고치던 그 사람과 동일인이다.

바뀐 것은 **누가 승인하는가**가 아니라 **승인에 무엇이 따라붙는가**다:

| | 손편집 | 콘솔 |
|---|---|---|
| 다섯 한도 설정 확인 | 없음 | 사전 판정 |
| JSON 유효성 | 없음 | 무효면 저장 거부 |
| 무엇이 시작되는지 고지 | 없음 | 화면 문장 |
| 감사 기록 | **없음** | `automation_gate.toggle` |

### A2 — 게이트를 켜기 쉬워진 것 자체가 위험 아닌가

**지적.** 마찰이 곧 안전인 경우가 있다. 손편집의 번거로움이 실수로 켜는 것을 막고 있었다.

**판정: 부분적으로 유효하고, 교환은 명시적이다.**
막고 있던 것은 실수가 아니라 **의도한 행위의 지연**이었다. 4일 동안 손절이 0개였던 것이
그 대가다. 그리고 "번거로움"은 안전 장치로 설계되지 않았다 — 그것은 기능의 부재였고,
부재는 검증도 감사도 함께 없앤다.

실수로 켜는 것에 대한 실제 방어는 셋이다: 세션+CSRF, 켜기 전에 읽는 문장, 그리고
게이트 ON이 **아무것도 즉시 하지 않는다**는 사실 — 다음 엔진 기동에서 반영된다.
`tossctl engine run`이 별도 행위로 남아 있다.

**남는 위험**: 콘솔 세션을 가진 사람은 게이트를 켤 수 있다. 그 경계는 이 change의 것이
아니므로 넓히지 않았다 — localhost bind, 터미널 점유가 신뢰 근원, 다른 mutating 화면과
동일한 처리.

### A3 — 사전 판정이 두 번째 인터록이 되지 않았는가

**지적.** 화면이 조항 1과 3을 판정한다. 그것은 인터록의 규칙을 두 번째로 구현한 것이고,
두 구현은 반드시 어긋난다.

**판정: 유효한 위험이고, 세 가지로 좁혔다.**
① 판정은 **config 값에서만** 나온다 — `Limits().Validate()`는 인터록이 **부르는 바로 그
함수**이고, 거래 정책은 `TradingPolicy.Missing()`이 인터록과 같은 순서·같은 문자열로
열거한다. 규칙을 다시 쓴 곳이 없다.
② 판정 불가한 절을 **화면에 렌더한다.** 침묵이 아니라 "판정 불가"라고 적는다.
③ "사전 판정을 통과해도 기동이 보장되지 않는다"를 화면이 직접 말한다.

그리고 **저장은 판정에 걸리지 않는다** — 미충족 상태로도 ON을 기록한다. 저장을 거부하면
화면이 인터록의 권위를 흉내 내기 시작한다. 기록은 콘솔의 것이고 판단은 엔진의 것이다.

### A4 — 세 저장 경로의 분리가 실제로 지켜지는가

**지적.** 이전에는 타입이 보장했다("`GuardianLimits`에는 `enabled` 필드가 없다"). 이제
같은 패키지가 세 경로를 다 갖는다.

**판정: 유효했던 보장을 다른 층으로 옮겼고, 양쪽에서 측정한다.**
- 타입 층: `GateSwitch.Save(bool)`, `TradingPolicySettings.Save(TradingPolicy)` —
  리플렉션 테스트가 필드 수와 이름을 읽는다.
- 바이트 층: `operating_io_test.go`가 각 저장을 "건드리면 안 되는 값이 들어 있는 파일"에
  겨누고 남은 바이트를 단언한다. `TestTheLimitSaveStillCannotMoveTheSwitch`가
  기존 보장이 살아 있음을 새 경로가 생긴 뒤에 다시 확인한다.

세 member 목록(`gateMembersOf`·`gateSwitchMembers`·`tradingMembersOf`)이 서로소이고,
리뷰어는 세 함수만 보면 된다.

### A5 — 정적 가드 셋을 고친 것이 가드를 무디게 했는가

**지적.** `"gate"` 금지, 예외 seam 수, 파일 면제 목록 — 셋 다 이 change가 손댔다.
가드를 통과시키려고 가드를 고치는 것은 가장 흔한 자기기만이다.

**판정: 유효한 의심. 각각 좁혔지 지우지 않았다.**
- `"gate"`는 목록에 그대로 있다. 정확 경로 하나(`/settings/gate`)에, 그 단어 하나에
  한해 예외다. `/settings/gate/force`도 `/settings/Gate`도 통과하지 못한다.
  그리고 스펙이 그 이름을 **요구**한다 — 숨긴 이름은 감사를 무력화한다.
- 예외 seam은 "정확히 하나"에서 **열거**로 바뀌었다. 진짜 보장("모든 예외에 논증이
  적혀 있다")은 그대로이고, 열거되지 않은 seam이 예외를 가지면 여전히 실패한다.
  stale 항목(열거됐는데 예외가 없는)도 새로 잡는다.
- 파일 면제는 `templates_settings.go` 하나 늘었고, `settings_operating.go`는
  **면제가 필요 없어서 넣지 않았다** — 그 파일은 코드에서 그 키를 쓰지 않는다.
- 금지된 두 단어(`Interlock`, `ProtectionReady`)는 모든 파일에서 그대로 금지다.

### A6 — 인터록 3절 확장이 범위를 넘었는가

**지적.** 이 change는 콘솔 change다. 엔진 인터록을 건드리는 것은 범위 밖이다.

**판정: 범위 안이고, 빼는 쪽이 나빴다.**
메뉴가 `sell`과 `allow_live_order_actions`만 요구하는 3절을 물려받으면, 운영자가 화면에서
두 개를 켜고 게이트를 켜고 엔진이 뜨고 **첫 손절에서 거부된다.** 메뉴가 없던 때보다 나쁘다.
방향도 fail-closed(요구를 늘림)이고 §0.6의 보수 방향이다. 3절 주석이 원래 주장하던 것을
검사가 따라잡은 것이다.

## §0 재확인

- §0.1 — 주문 side effect 없음. 전 테스트 fake seam.
- §0.3 — 게이트 OFF 동작 무수정. 저장은 다음 기동부터 반영(기존 문구 유지).
- §0.4 — 손절 즉시성 개선 방향. `place`·`cancel` 검사가 "손절을 낼 수 없는 구성"을 막는다.
- §0.6 — 정책 수치 무수정. 인터록 요구는 늘었다.
- §0.7 — 게이트도 거래 정책도 이 change가 켜지 않는다. 사람이 누르고 audit에 남는다.
- §0.9 — 완화 없음. 손편집 경로 → 검증·고지·감사가 있는 경로로의 이동이다.

# Design — interlock-gates-entry-not-exit

## §0 준수

**§0.1 (사람 승인 없는 LIVE 주문 side effect 금지)** — 이 change는 주문을 내지 않는다.
자동 테스트는 전부 httptest·fake broker다. 게이트 ON과 엔진 기동은 사람이 한다.

**§0.3 (토글 OFF는 upstream 동작과 동일)** — `automation_gate.enabled = false`의 동작은
한 바이트도 바뀌지 않는다. `runInterlock`의 첫 분기(`if !gate.Enabled`)는 무수정이다.

**§0.4 (손절·비상 청산의 즉시성을 약화·지연하지 않는다)** — 이 change는 반대 방향이다.
현재 손절 즉시성은 0이다(엔진이 뜨지 않으므로). 이 change 후 exit observer가 돈다.
flatten-all 경로는 무수정.

**§0.5 (High-risk 경로)** — 인증·주문·Guardian 경로를 건드린다. 그래서 집행 지점을
**줄이지 않고 옮긴다**: 기동 시점 1회 검사 → mutation 1건마다 검사. 커버리지가 넓어진다.

**§0.6 (손절·익절·사이징은 보수 방향만)** — 수치를 건드리지 않는다. 사용자가 별도로 정한
`default_stop_pct` 0.025 → 0.05는 밴드를 **넓히는** 보수 방향이며 이 change의 코드 델타가
아니다(사용자 config).

**§0.7 (운영 토글 flip·live 검증은 사람 승인)** — 게이트 ON, `trading.sell`,
`allow_live_order_actions`를 이 change가 켜지 않는다. 셋 다 사람이 켠다.

**§0.9 (불명확하면 움직이지 않는다)** — 이 change가 완화하는 것은 **하나**다:
"게이트 ON 기동을 거부한다" → "게이트 ON 기동을 허용하되 raising mutation을 거부한다".
그 외 여덟 조항, 한도, 정책 수치, attestation 요구는 전부 불변이다. D4의 테스트가
그 불변을 필드 단위로 단언한다.

## 결정

### D1 — 왜 조항 6을 삭제하지 않고 옮기는가

삭제하면 2c-A가 진입 경로를 만드는 날 아무것도 막지 않는다. 이 change의 근거는
"지금 진입 경로가 없다"는 **관측**이고, 관측은 코드가 바뀌면 만료된다. 그래서 조항을
없애는 대신 **관측이 만료되는 지점에서 발동하도록** 옮긴다: 매수 intent가 게이트웨이에
도달하는 순간.

### D2 — 왜 `Gateway.Place`인가 (`IssueEntry`가 아니라)

`RiskGuardian.IssueEntry`도 후보였다. 게이트웨이를 고른 이유는 판정 근거의 성질이다.
`Gateway.Place`는 `raisesExposure`를 **intent의 형태에서 계산한다**:

```go
// gateway.go:335-338 — 원문
// Computed here, from the mutation itself, and never taken from the
// decision: a class the caller declares is a label, and this is the fact
// it has to agree with
raisesExposure: strings.EqualFold(intent.Side, "buy"),
```

발급자에서 막으면 "발급자를 안 거친 결정"이라는 우회가 생긴다. 게이트웨이에서 막으면
mutation 자체의 형태가 근거이므로 우회할 형태가 없다. 그리고 **한 곳에서만** 막는다 —
tracer.go 주석의 판단을 그대로 따른다: *"Adding a second gate here would be a second
place to get the answer wrong."*

### D3 — 왜 표지를 `execgw`로 옮기는가

집행이 `execgw`에서 일어나는데 표지가 `engine`에 있으면 `execgw`가 그것을 읽을 수 없다
(import 방향이 반대). 남는 선택지는 게이트웨이 생성자에 bool을 넘기는 것인데, 그것이
정확히 원 커밋이 거부한 형태다:

> "config 키도 Options 필드도 없다. 그런 것은 전부 '짓지 않고 준비됐다고 주장하는 길'이고
> 이 조항은 바로 그 주장을 거부하려고 있다." (`43b3fd6`)

`const`를 집행 지점으로 옮기면 그 성질이 보존된다. 여전히 config도 Options도 없고,
여전히 T2.3이 식별자 하나를 뒤집는 것이 마지막 단계다. 바뀌는 것은 상수가 사는 패키지뿐이다.

### D4 — 완화의 범위를 기계로 고정한다

`size-us-guardian-tier`에서 쓴 방법을 재사용한다: 움직인 것과 움직이지 않은 것을 이름으로
단언하는 테스트. 여기서는 —

- 조항 1~8 각각이 여전히 기동을 거부한다(8건, 개별)
- 조항 6 미충족만으로는 기동이 거부되지 않는다
- 조항 6 미충족 상태에서 `Gateway.Place`가 buy를 거부한다
- 같은 상태에서 sell은 통과한다
- `AutomationStatus.Protection`이 여전히 `UNWIRED`로 보고된다

완화가 아무도 논증하지 않은 조항으로 새지 못한다.

### D5 — 구조적 단언: 도달 가능한 매수 경로 없음

`deps_test.go`의 `TestEngineDependencyGraphExcludesWTSMutators`가 선례다. 런타임 테스트는
"오늘 이 경로가 매수를 안 했다"만 보이고, AST 단언은 "매수가 철자되지 않는다"를 보인다.

이 change의 논거가 관측(진입 경로 없음)에 기대므로, 그 관측을 테스트로 굳힌다:
`cmd/tossctl`의 `engine run` 경로에서 도달 가능한 심볼 그래프에 진입 발급이 없음을 단언한다.
`tracer.go`는 도달 불가이므로 통과하고, 2c-A가 그것을 배선하는 순간 빨개진다 — 그때
2c-A는 이 테스트를 지우는 게 아니라 게이트웨이 거부(D2)를 마주해야 한다.

### D6 — 로그와 콘솔은 무엇을 말해야 하는가

기동이 성공하는데 보호가 미배선인 상태는 **새로운 상태**다. 조용히 뜨면 운영자는 브로커에
손절이 걸려 있다고 오해한다. `engine.operating_mode`는 이미 `protection: "UNWIRED"`를
찍고 있으므로 필드는 그대로 두고, **기동 출력에 한 줄**을 추가한다: 보호는 이 프로세스가
살아 있는 동안만 유효하고 프로세스가 죽으면 손절이 사라진다는 사실.

이것은 확인 문구가 아니다(사용자 지시: 콘솔 UI에 타이핑 확인·추가 승인 마찰 금지).
읽는 문장이지 누르는 것이 아니다.

## 위험

**로컬 판단 오작동 매도** — exit observer가 잘못 판정하면 실제로 팔린다. 오늘은 엔진이 안
떠서 이 위험이 0이고, 대신 보호도 0이다. 사용자가 교환을 명시적으로 선택했다(2026-07-30).
완화: 편입 t0가 원가가 아니라 편입가라 오래된 승자가 즉시 청산되지 않고(`exitpolicy/adoption.go`),
`MinStopPct` 0.02 하한이 노이즈 폭 밴드를 거부하며, 사용자가 손절폭을 5%로 넓혔다.

**프로세스 사망 시 무보호** — 조항 6이 원래 지목한 실패이며 이 change가 해소하지 않는다.
D6의 기동 출력이 그것을 매번 말한다. T2.3이 닫는다.

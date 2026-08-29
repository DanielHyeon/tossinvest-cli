# Change: interlock-gates-entry-not-exit

> 2026-07-30. 사용자 결정: 조항 6을 분리해 보호 루프를 기동시킨다. 손절폭 5%, 자동 편입 유지.

## Why

인터록 조항 6은 게이트 ON 기동을 통째로 거부한다. 그 조항이 스스로 밝힌 목적은 좁다:

> "보호주문 도입 change가 이 표지를 배선하기 전에는 **자동 진입이 켜질 수 없다**"
> (engine-safety "자동화 게이트 기동 인터록" 6항)

의도는 자동 진입 차단인데 기제는 런타임 전체를 거부한다. 그 결과가 4일째 관측되고 있다 —
`ProtectionUnwired`가 컴파일 타임 상수(`add-core-domain` 5.2, 커밋 `43b3fd6`)라 설정으로는
만족시킬 수 없고, 상수를 뒤집을 T2.3은 착수되지 않았으며, 그동안 **계좌의 보유 종목은 손절이
0개**다.

조항 6이 막으려는 실패는 "손절 없는 새 포지션"이다. 그것을 막느라 실제로 유지되고 있는 상태는
"손절 없는 기존 포지션"이다. 같은 결함을, 더 넓은 범위에.

### 지금 빌드에 자동 진입이 없다는 사실

이 change의 근거는 규범이 아니라 관측이다. 기동 중인 엔진에서 매수 주문이 나갈 수 있는 경로가
없다:

- 엔진 런타임의 루프 집합은 셋이다 — reconcile driver, exit observer, 체결 감지
  (engine-safety "엔진 런타임 수명주기"). 진입 루프는 스펙에도 코드에도 없다.
- `internal/app/engine` 전체에서 `Side: "buy"`를 만드는 곳은 `tracer.go:444` 하나뿐이다.
- `tracer`는 `cmd/tossctl`에서 참조 0건이고 `runtime*.go`·`engine.go`도 부르지 않는다.
  자기 테스트에서만 도달한다(설계상 — D8 "실전 실행은 verify 트랙").

즉 조항 6은 존재하지 않는 진입 경로를 막기 위해 존재하는 exit 경로를 막고 있다.

## What Changes

- **인터록**: 조항 6이 기동 거부 조건에서 빠지고 `AutomationStatus.EntryPermitted`가 된다.
  조항 1~8이 기동을 결정하고, 조항 6은 **무엇이 허가되는지**를 결정한다.
- **집행 지점 이동**: 거부는 사라지지 않고 mutation chokepoint로 내려간다.
  `execgw.Gateway.Place`가 `raisesExposure`인 mutation을 보호 미배선 상태에서 거부한다.
  판정 근거는 호출자가 선언한 클래스가 아니라 **intent의 형태에서 계산된 사실**
  (`raisesExposure: side == "buy"`, gateway.go:338)이므로 라벨을 잘못 붙여 우회할 수 없다.
- **표지의 이사**: `ProtectionReadiness`와 `ProfileProtection` 상수가 `internal/execgw`로
  옮겨간다. 집행하는 곳에 표지가 있어야 한다. `const`인 성질은 그대로이므로
  "짓지 않고 준비됐다고 주장하는 길"은 여전히 없다.
- **구조적 단언**: `deps_test.go`의 선례를 따라, 엔진이 도달할 수 있는 매수 경로가 없음을
  AST로 증명하는 테스트를 추가한다. 이 change가 기대는 사실이 나중에 조용히 깨지지 않게.

### 이 change가 하지 않는 것

- `profileProtection`을 `ProtectionWired`로 뒤집지 않는다. 브로커측 보호는 여전히 미배선이고,
  그 사실은 로그·콘솔·기동 출력에 계속 표시된다.
- 자동 진입을 허용하지 않는다. 2c-A가 진입 경로를 만들 때 그 경로는 게이트웨이에서 거부된다 —
  이 change 전보다 **좁은 구멍이 아니라 같은 구멍**이며, 기동 시점 1회가 아니라 주문마다 걸린다.
- 손절·익절 수치를 건드리지 않는다. 래칫·래더 값은 `[미검증 — StockOS KOSPI 튜닝값]`으로
  남는다.
- `trading.sell`·`allow_live_order_actions`를 대신 켜지 않는다. 인터록 3항은 그대로이고,
  그 토글은 사람이 켠다(§0.7).

## Capabilities

### Modified Capabilities

- `engine-safety`: 인터록 조항 6을 기동 거부 조건에서 진입 허가 조건으로. 집행 지점을
  게이트웨이 mutation chokepoint로 명시. 런타임 수명주기 시나리오 갱신.

## Impact

- Affected code: `internal/execgw`(표지 이사 + `Gateway.Place` 거부),
  `internal/app/engine/interlock.go`(조항 6 → `EntryPermitted`), 구조 테스트 신규.
- 사용자 영향: 게이트 ON + 인터록 1~8 충족 시 **대사·exit 관측·체결 감지 루프가 기동한다.**
  편입된 보유에 손절·래칫·부분익절이 적용되고 실제 매도 주문이 나간다.
- 위험: 로컬 판단에 의한 오작동 매도가 가능해진다. 이것이 이 change가 사는 대가이며,
  사용자가 명시적으로 선택했다(2026-07-30). 프로세스 사망 시 보호가 사라지는 성질은
  변하지 않으며 T2.3이 그것을 닫는다.
- 선행: 없음. TSLA 발동 실측(3.3)도 T2.3도 이 change의 선행이 아니다 —
  둘은 브로커에 손절을 **얹는** 작업이고 이 change는 손절이 **있게** 하는 작업이다.

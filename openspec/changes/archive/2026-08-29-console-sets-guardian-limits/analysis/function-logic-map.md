# Function Logic Map — console-sets-guardian-limits (서술)

기계 검증 대상별 증거는 `analysis/function-logic/<pkg>--<symbol>/`에 있다. 이 문서는
그것들을 잇는 서술이고, §0 확인 두 건을 여기 적는다.

## §0.1 — 기본값이 게이트 ON 기동을 열지 않는가

초안 설계(로드 시점 암묵 기본값 주입)에서는 열었다. 오늘 "게이트 ON + 한도 없음"은
인터록 1번 조항 위반으로 기동 거부이고, 기본값이 주입되면 그 조합이 통과한다.

D1에서 그 설계를 버렸으므로 이 change는 파싱을 건드리지 않는다. 확인 방법:

- `internal/config/mergeEngine`은 이 change에서 **무수정**이다. diff에 없다.
- `TestParsingStillInventsNothing`(신규)과 기존 `TestAutomationGateDefaultsOff`가
  "한도 없는 게이트는 로드 후에도 `LimitsSet() == false`"를 함께 고정한다.
- 따라서 인터록이 보는 것은 파일뿐이고, 파일은 사람이 클릭해야 바뀐다.

부수적으로 확인해 둔 것: 한도가 채워져도 게이트 ON 기동은 여전히 거부된다.
`engine-safety`의 기동 인터록은 여섯 조항이고 한도는 1번뿐이다. 2번 capability
attestation, 3번 매도·실주문 정책, 5번 ExecutionGateway 구성, 6번 ProtectionReady가
남으며, 6번은 보호주문 도입 change 이전에는 충족될 수 없다고 스펙이 못박았다.

## §0.2 — 한도가 손절·비상 청산의 즉시성에 닿는가

닿지 않는다. 두 갈래로 확인했다.

1. **판정 체인의 위치.** `risk-management`의 체인은 "모든 자동 **진입** 의도"에
   적용된다. 주문 크기 한도·총 개방 노출 한도·일일 손실 한도는 전부 그 체인의 rung다.
2. **class 분류.** `engine-safety`가 "RISK_REDUCING 결정은 한도 스냅샷을 싣지 않으며
   **수량·금액 한도의 적용을 받지 않는다**(SHALL)"고 명시한다. 청산은
   RISK_REDUCING이다.

그러므로 한도를 최소값까지 조여도 손절·익절·flatten은 느려지지 않는다. 반대 방향
— 한도를 **올리면** 더 큰 진입이 허용된다 — 이 §0.6의 대상이고, D5의 상한
백스톱이 그 방향의 천장이다.

## 수정 대상 열 개의 관계

```
runConsole                         cmd/tossctl/console.go
 ├─ consoleSettingsSeam            (무수정 — diff hunk에만 걸림, base 리비전 증거)
 └─ consoleLimitSettingsSeam       (신규) → consoleLimitSettings{svc} → config.Service
                                                 ├─ LoadRawEngineGate
                                                 └─ SaveEngineGateLimits ─ 6키 splice
Console.routes                     internal/console/console.go
 └─ /settings/limits, /settings/limits/preset  (session0 + mutating)
Console.handleSettings             internal/console/settings.go
 └─ LimitSettings.Load → page.Gate  (표시 전용, 폼으로 되돌아가지 않음)

정적 가드 다섯 (전부 테스트 함수):
  TestEveryRouteGoesThroughTheSessionGate            하한 20 → 22
  TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate  map += 2
  routeFindings                                      actVerbs += limit, preset
  TestTheConsoleDecidesNothingAboutTheGate           금지어를 둘로 쪼갬 + 면제 2파일
  TestTheGateEditingExemptionIsNotIdle               (신규) 면제의 유휴 방지
```

## 신규 코드의 위험 경계

기계 검증 대상은 "기존 함수 내부 수정"만 요구하므로 아래 신규 심볼은 대상이 아니다.
그러나 이 change의 위험은 대부분 여기 있으므로 경계를 적어 둔다.

| 심볼 | 경계 | 고정 테스트 |
|---|---|---|
| `config.SaveEngineGateLimits` | 여섯 키만 쓴다. `enabled` 바이트 무접촉 | `TestSavingLimitsNeverRewritesEnabled`, `TestTheSaveCarriesNoEnabledKeyOfItsOwn` |
| `config.GuardianLimits.Validate` | 인터록과 동일 판정 | `TestConfigRefusesExactlyWhatTheInterlockRefuses` (생성된 코퍼스) |
| `config.GuardianCeiling` | 등록 티어의 필드별 최대, 미등록 통화는 fail-closed | `TestTheCeilingIsTheMaxAcrossRegisteredTiers`, `TestAnUnregisteredCurrencyFailsClosed` |
| `console.LimitSettings` | Save 인자에 `enabled` 필드가 없다 | `TestTheLimitSeamCannotCarryTheSwitch` (reflection) |
| `console.handleSettingsLimitPreset` | 등록된 티어만, 확인 1회 | `TestApplyingAPresetWritesAllFiveAtOnce`, `TestAnUnknownTierIsRefused` |
| `console.handleSettingsLimits` | 인터록 규칙 + 상한 + 통화 등록 | `TestABlockTheInterlockWouldRefuseIsNotWritten` 외 3건 |

## 읽기-수정-쓰기 유실이 이 경로에는 없다

편입 저장은 handler가 flock **밖**에서 Load한 블록을 통째로 다시 쓰므로 lost update
성질을 갖는다(그 change의 리뷰에 기록돼 있다). 한도 경로에는 그 성질이 없고, 이유는
D6의 부산물이다.

- handler는 read-modify-write를 하지 않는다. 프리셋은 레지스트리에서, 개별 기입은
  폼에서 여섯 값을 **전부** 만들어 넘긴다.
- writer는 flock을 잡은 **뒤에** 파일을 읽고, 그 바이트에 여섯 키만 splice한다.
  자기가 쓰지 않는 키(`enabled`·`attestation_file`·`adoption`·미지 키)는 읽은 그대로
  남는다.

`TestTheAdoptionBlockIsUntouched`와 `TestUnknownKeysSurviveToTheByte`가 그 성질을
바이트 수준에서 고정한다.

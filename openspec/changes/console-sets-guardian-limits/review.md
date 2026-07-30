# Review: console-sets-guardian-limits

날짜: 2026-07-30 · 위험 등급: **High-risk**

이 change는 승인된 스펙 문장 하나를 뒤집는다 — "Guardian 한도는 콘솔에서 편집할 수
없다(SHALL NOT)". 뒤집는 근거와 대신 세운 경계가 이 문서의 본론이다.

## Pre-Edit Gate

```text
change id / task id:  console-sets-guardian-limits / 1.1-9.4
대상 심볼 (기존 함수 내부 수정, 기계 검증 10건):
  main.runConsole                       (cmd/tossctl/console.go:162)  — Options 1줄
  main.consoleSettingsSeam              (cmd/tossctl/console.go:291)  — 무수정(hunk 인접)
  main.consoleLimitSettingsSeam         (cmd/tossctl/console.go:300)  — 신규
  console.(*Console).routes             (internal/console/console.go) — 라우트 2줄
  console.(*Console).handleSettings     (internal/console/settings.go)— seam 적재 1블록
  console 정적 가드 5종                  (static_test.go·engineproc_test.go)
가산 (신규):
  config.GuardianLimits/GuardianTier/GuardianCeiling/CeilingViolations/
  MatchingGuardianTier/GuardianCurrencies, config.Service.LoadRawEngineGate/
  SaveEngineGateLimits, console.LimitSettings + 두 핸들러 + 한도 섹션,
  main.consoleLimitSettings + recordLimitSave
기존 동작 파악 근거:
  analysis/function-logic-map.md + 기계 검증 10건
  읽은 파일: internal/config/{engine,service,adoption_io}.go,
             internal/execgw/guardian.go, internal/risk/chain.go,
             internal/app/engine/runtime_wiring.go,
             internal/console/{settings,console,overview,templates_settings}.go,
             cmd/tossctl/{console,adoptionsettings}.go,
             static_test.go·settings_static_test.go·engineproc_test.go·engine_test.go,
             openspec/specs/{operator-console,risk-management,engine-safety}/spec.md
  StockOS 원본: packages/trading/stockos_trading/{risk_profiles,profile_safety_gate}.py,
             apps/api/stockos_api/routes/strategy_profile.py
upstream 상속 테스트 영향: no — internal/console·engine 블록은 TossOS 전용이다
실패 테스트 선행 작성: yes (2.x-8.x 전부 RED 선행)
안전 불변식 §0 위반 여부 검토: 통과 — 조항 3·4·6·7을 아래 표에서 명시적으로 다룬다
```

## §0 대조

| 조항 | 이 change |
|---|---|
| 1 승인 없는 LIVE side effect 금지 | 무관 — 주문 경로 없음. 쓰는 것은 `engine.automation_gate`의 여섯 값뿐이고 계좌·브로커 무접촉 |
| 2 mutating 자동 실행 금지 | 콘솔 클릭은 사람의 행위다. 에이전트가 실행하지 않는다 |
| 3 토글 OFF = upstream 동작 | 파싱 무수정. 한도가 비면 오늘과 똑같이 `LimitsSet()==false`이고, 게이트 OFF 경로는 한 갈래도 바뀌지 않는다. `TestParsingStillInventsNothing`이 고정 |
| 4 손절·비상 청산 즉시성 | **확인 대상이었고 통과한다.** 한도는 진입 체인의 rung이고, `engine-safety`가 "RISK_REDUCING 결정은 수량·금액 한도의 적용을 받지 않는다(SHALL)"고 명시한다. 한도를 최소값까지 조여도 청산은 느려지지 않는다 (FLM §0.2) |
| 5 High-risk 경로 | 해당(Guardian) → full TDD + FLM 10건 + 적대적 리뷰 + gate |
| 6 손절·사이징은 보수 방향만 | 낮추기는 양수인 한 자유, 올리기는 **등록 티어 상한에서 멈춘다**(D5). 상한 위로 올리려면 콘솔 밖에서 파일을 열어야 한다 — 방향 비대칭이 구조로 들어가 있다 |
| 7 운영 토글 flip은 사람 | **이 change가 가장 조심한 지점이다.** 게이트 ON/OFF와 kill switch는 여전히 콘솔 밖이고, 그것을 규율이 아니라 문법으로 만들었다(D6·D7): seam의 Save 인자 타입에 `enabled` 필드가 없고, writer는 그 키의 바이트를 쓰지 않는다 |
| 8 시크릿·개인정보 미저장 | 기록은 숫자 다섯과 통화뿐 |
| 9 주문은 공식 Open API만 | 무관 |
| 10 실계좌 자동 테스트 금지 | 전 테스트가 fake seam + tmpdir config |

## 스펙 문장을 뒤집는 근거

개정 전 문장은 셋을 한 덩어리로 금지했다: automation gate·Guardian 한도·kill switch.
괄호에 적힌 이유는 하나였다 — "게이트 ON은 §0.7 콘솔 밖 절차 유지". 지키려던 것은
**스위치**였고, 한도는 같은 문장에 실려 갔다.

셋을 분리하면서 넣은 대체 경계는 넷이다.

1. **seam의 모양** — Save는 `config.GuardianLimits`를 받고 그 타입에 `enabled`가 없다.
   `TestTheLimitSeamCannotCarryTheSwitch`가 reflection으로 읽는다.
2. **writer의 모양** — 여섯 키를 개별 splice하고 `enabled` 바이트를 쓰지 않는다.
   `TestSavingLimitsNeverRewritesEnabled`(양방향)와
   `TestTheSaveCarriesNoEnabledKeyOfItsOwn`(바이트)이 고정한다.
3. **상한** — 등록 티어의 필드별 최대. 콘솔은 그 위로 못 올린다.
4. **어휘** — `routeOnlyAccountVerbs`의 `"gate"`를 **유지**했다. 게이트를 여는 라우트는
   이 change 이후에도 없어야 하므로, 이름에 `gate`를 담은 경로는 여전히 실패한다.

## 사용자 결정과의 대조

| 결정 | 이 change |
|---|---|
| 타이핑 확인·추가 승인 마찰 금지 (2026-07-27) | 지킨다. 프리셋은 `confirm()` 1회뿐이고 폼에 입력 요소가 없다. `TestThePresetControlsAskForNoTyping`이 못박는다 |
| 클릭 한 번으로 설정 (2026-07-30) | 프리셋이 1차 경로. 개별 기입은 고급 접힘 |
| 콘솔 웹에서 설정 (2026-07-30) | 이 change 자체. CLI 대안을 권했으나 사용자가 콘솔을 택했고, 스펙 개정을 선행했다 |

## 적대적 Eng 리뷰

날짜 2026-07-30 · 시점 구현 후 · 방식: "이 화면이 운영자에게 무엇을 믿게 만드는가,
그리고 그 믿음이 어디서 깨지는가"를 경로별로 추적.

### A1. 고급 폼이 미설정 칸에 `0`을 그렸다 — **P1**

같은 화면의 표는 `미설정`이라고 말하는데, 바로 아래 폼의 입력칸은 `0`이었다. 이
저장소는 "0은 무제한이 아니라 아무도 정하지 않았다"를 도처에서 되풀이하는데, 화면이
한 화면 안에서 그 둘을 다르게 그리고 있었다. 운영자가 `0`을 "한도가 0"으로 읽으면
잘못된 안심이고, 그대로 제출하면 이미 화면이 부정한 이유로 거부된다.

→ 미설정은 **빈 칸**으로 렌더한다. 테스트 `TestTheAdvancedFormShowsNoZeroForAnUnsetLimit`.

### A2. 지수 표기 — **P2**

Go 템플릿의 float64 기본 렌더는 `10000000`을 `1e+07`로 쓴다. `ParseFloat`은 되받으므로
기능은 멀쩡하지만, 텍스트 박스의 `1e+07`을 보고 "고쳐" 적는 운영자는 자릿수 하나
틀린 노출 상한에서 한 타 거리다.

→ 폼 값은 그룹 없는 평문 십진수로 포맷한다.
테스트 `TestTheAdvancedFormRendersNumbersTheOperatorCanRead`.

### A3. 미등록 통화를 "상한 초과"로 보고했다 — **P2**

`CeilingViolations`는 목록 하나를 돌려주므로 "JPY에는 등록된 티어가 없다"를 상한 위반
목록에 섞어 넣는다. 핸들러가 그것을 "등록된 티어 상한을 넘는다" 접두와 함께 보여주고
있었다. 운영자는 애초에 문제가 아니었던 숫자를 낮추러 간다.

→ 통화 등록 여부를 먼저 따로 검사하고, 등록된 통화 목록을 레지스트리에서 뽑아 함께
말한다. 테스트 `TestAnUnregisteredCurrencyIsRefusedForTheRightReason`.

### A4. 초안의 암묵 기본값 — **설계 단계에서 폐기 (P0급)**

가장 큰 것은 구현 전에 잡혔다. 로드 시점 주입은 화면이 1,000,000을 그리는 동안
엔진의 인터록은 파일의 빈 값을 보는 상태를 만든다. 화면과 엔진이 갈라지고, 갈라진
쪽이 "게이트가 구성돼 있다"고 말하는 쪽이다. 이 저장소는 손절폭 기본값에서 이미 같은
갈림길을 만나 "the engine still never runs on an implicit number"로 결론냈다.

→ D1. 파싱 무수정, 기본값은 권장 프리셋으로만 존재. 부수 효과로 엔진·인터록·상속
파싱 테스트가 전부 사정거리 밖으로 나갔다.

### 닫지 않고 남긴 것

- **US 티어로는 TSLA 한 주를 자동 진입할 수 없다.** 이식한 `us-small-live`의 주문
  notional 상한은 $300이고 TSLA는 그보다 비싸다. 상한은 등록 티어의 최대이므로
  콘솔에서 올릴 수도 없다. 지금 당장은 무해하다 — 대기 중인 실측(3.3)은 Guardian
  체인을 타지 않는 수동 경로다. 그러나 **2c-A가 US에서 자동 진입을 켜는 시점에는
  걸린다.** 그때 필요한 것은 이 상한을 임의로 올리는 것이 아니라, 실측 근거를 갖춘
  US 티어를 레지스트리에 추가하는 별도 change다.
- **`risk-management`의 암묵 기본값 요구는 미구현으로 남는다.** proposal의 "범위 밖"에
  적었다. 그 후반부는 엔진 배선을 포함하고 게이트 ON 기동 조건을 바꾸므로, §0.1의
  방향 논증을 지는 별도 change의 몫이다.
- **여섯 번 재스캔.** writer는 멤버마다 블록 span을 다시 찾는다. 빠른 방법이 아니라
  stale offset에 쓸 수 없는 방법을 골랐다. config.json 크기에서 문제되지 않는다.

## 검증 결과

| 항목 | 결과 |
|---|---|
| `go test ./...` | **3873 passed**, 0 failed (57 packages) — 직전 3821 대비 신규 52 |
| `make vet` | 통과 |
| `make validate` | 32 passed, 0 failed |
| `openspec validate --strict` | valid |
| 상속 테스트 회귀 | 0 — `internal/config` 파싱 테스트 무수정 통과, `internal/console` 전건 통과 |
| zero value 안전 | 한도가 비면 `LimitsSet()==false`이고 파싱·엔진 경로가 변경 전과 동일 |

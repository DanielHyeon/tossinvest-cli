# Function Logic Map — size-us-guardian-tier (서술)

기계 검증 대상별 증거는 `analysis/function-logic/<pkg>--<symbol>/`에 있다. 이 문서는
그것들을 잇는 서술이고, §0 확인 두 건을 여기 적는다.

## 편집 표면이 왜 이렇게 좁은가

이 change의 production diff는 `internal/config/limits.go` 한 파일이고, 그 안에서도
**함수 본문은 한 줄도 바뀌지 않았다**. 바뀐 것은 패키지 수준 `var guardianTiers`
슬라이스 리터럴에 원소 하나가 붙은 것과 파일 헤더 주석이다.

그래서 기계 검증이 요구한 일곱 심볼은 전부 `internal/config/limits_test.go`의
테스트 함수다 — 둘은 수정된 기존 함수(전사 검사·상한 검사), 다섯은 신규다.
gate가 테스트 함수도 diff 대상으로 계산하기 때문이고, 그것이 맞다: 이 change에서
"판정하는 코드"는 사실상 테스트뿐이다.

레지스트리를 읽는 다섯 함수(`GuardianTiers`·`GuardianTierByID`·`GuardianCeiling`·
`GuardianCurrencies`·`MatchingGuardianTier`)는 무수정이다. `GuardianCeiling`이
돌려주는 **값**은 바뀌지만 그것은 데이터가 바뀐 결과이고, 분기·early return·
mutation·fallback은 동일하다.

## §0.1 — 티어 추가가 기동 인터록을 바꾸는가

바꾸지 않는다. 호출 방향으로 확인했다.

```text
guardianTiers (var)
 ├─ GuardianTiers ─────────── console: 프리셋 카드 렌더
 ├─ GuardianTierByID ──────── console: 프리셋 적용 핸들러
 ├─ GuardianCeiling ───────── config.SaveEngineGateLimits (쓰기 경로) + console 핸들러
 ├─ GuardianCurrencies ────── console: 미등록 통화 거부 문구
 └─ MatchingGuardianTier ──── console: "이 티어와 일치" 표시

기동 인터록 ── execgw.Limits.Validate ── 상한 개념 없음, 이 파일 무접촉
```

인터록이 수용·거부하는 블록의 집합은 이 change 전후로 **동일하다**.
`TestTheCeilingIsNotTheInterlock`이 그 분리를 한 블록 위에서 직접 보인다: 상한
10배 블록은 콘솔이 거부하고 인터록은 통과시킨다.

부수적으로 확인해 둔 것은 이전 change와 같다. 한도가 다 채워져도 게이트 ON 기동은
여전히 거부된다 — 인터록 여섯 조항 중 한도는 1번뿐이고, 6번 ProtectionReady는
보호주문 change 이전에 충족될 수 없다고 스펙이 못박았다.

## §0.2 — 한도 상한 이동이 손절·비상 청산에 닿는가

닿지 않는다. `engine-safety`가 "RISK_REDUCING 결정은 한도 스냅샷을 싣지 않으며
수량·금액 한도의 적용을 받지 않는다(SHALL)"고 명시하고, 한도는 자동 **진입**
체인의 rung이다. 이 change는 그 rung의 천장을 **올리는** 방향이므로 청산 경로에는
더더욱 닿지 않는다.

## 일곱 증거의 역할 분담

| 심볼 | 무엇을 막는가 |
|---|---|
| `TestTheTiersTranscribeStockOS` | 두 코퍼스 어디에도 근거 없는 티어의 추가 |
| `TestTheCeilingIsTheMaxAcrossRegisteredTiers` | 상한이 조용히 움직이는 것 |
| `TestRegisteringTheUSTierMovedExactlyTwoCeilings` | 완화가 승인 범위 밖 필드나 KRW로 새는 것 |
| `TestTheUSTierMatchesItsRecordedDerivation` | 유도가 사후 서술로 바뀌는 것 |
| `TestTheMeasuredInstrumentFitsWithHeadroom` | 상한이 실측 계측기 쪽으로 되돌아가는 것 |
| `TestEveryTierWouldStart` | 누르면 엔진이 안 뜨는 프리셋 |
| `TestTheCeilingIsNotTheInterlock` | 상한을 인터록으로 착각하는 것 |

앞의 셋은 **완화의 크기**를, 가운데 둘은 **완화의 근거**를, 뒤의 둘은 **완화의
의미**를 지킨다.

## 상속 테스트 회귀

없다. `internal/console`의 한도 화면 테스트는 티어 수를 상수로 갖지 않고
`config.GuardianTiers()`를 순회하므로 새 티어를 자동으로 덮는다
(`TestEveryRegisteredTierIsOfferedWithItsNumbers`). `internal/config`의 writer
테스트는 티어를 ID로 지목하고 그 ID들은 그대로 남아 있다.
`internal/app/engine/limits_equivalence_test.go`는 코퍼스를 필드 집합에서
생성하므로 레지스트리와 무관하다.

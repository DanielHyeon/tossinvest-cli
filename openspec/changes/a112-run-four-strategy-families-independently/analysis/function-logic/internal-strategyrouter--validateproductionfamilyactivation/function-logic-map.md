# Function Logic Map: `validateProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go` (471-532)
- Function: `validateProductionFamilyActivation` in package `strategyrouter`
- File SHA-256: `411d612f7444c31744585b53c6128ccb71d0406a3e708efe9b7aee6232ca67d7`
- Pinned revision: `HEAD` (c1d1e295) — 편집 **전** 커밋된 파일의 AST 다. 같은 로트의 다른 함수 편집이 워크트리에 먼저 올라 좌표가 밀렸으므로, 이 함수의 편집 전 증거는 HEAD blob 에서 떴다. GREEN 뒤 `current` 로 다시 만든다.
- AST evidence: `ast.json` — AST branches 8.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 전 기록 (태스크 8.7.2, 2026-09-27).** 입력: 복호한 몸통, 적재 config(시장·다섯 결속·관측 시각). 결속·수명·서술자
집합을 본다. 만료 판정은 B3(492:2) `!now.Before(expires)` → `ErrProductionFamilyActivationExpired` 한 줄이다.

## Branches and early returns

- Measurement regime: 몸통 진입 count (하네스 `analysis/harness/a872_attribute.py`), strategyrouter 태그 스위트 전체 시험 67 개 per-test. 시험별 합 == 스위트(MISMATCH 0). 엔진 스위트에서는 여덟 분기 모두 진입 0 이다(엔진 시험은 매니페스트를 쓰지 않았다).

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 474:2 | arm entered 12x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`, `TestTheOtherMarketsWholeManifestInThisMarketsFilePromotesNothing` |
| B2 | if | 488:2 | arm entered 4x (strategyrouter tagged suite, pre-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` |
| B3 | if | 492:2 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestAnActivationOutsideItsApprovedLifetimePromotesNothing` |
| B4 | range | 510:2 | arm entered 41x (strategyrouter tagged suite, pre-edit); `TestAVerifiedActivationCarriesTheProtectionFloorTheOrderPathMustBind`, `TestAVerifiedFourFamilyActivationPromotesExactlyTheLanesItNames`, `TestAnActivationPromotesOnlyTheFamiliesItTurnsOn`, `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing`, `TestTheCommittedGoldenManifestMatchesItsPinAndPromotesTheFourLanes` |
| B5 | if | 512:3 | arm entered 5x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B6 | if | 519:3 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B7 | if | 523:3 | arm entered 2x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |
| B8 | if | 528:2 | arm entered 1x (strategyrouter tagged suite, pre-edit); `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validMarket` | 475:60 |
| `productionRouteIdentity` | 481:4 |
| `productionRouteTime` | 484:26 |
| `productionRouteTime` | 485:22 |
| `productionRouteTime` | 486:24 |
| `config.ObservedAt.UTC` | 487:9 |
| `issued.Before` | 488:47 |
| `issued.After` | 488:74 |
| `issued.Before` | 489:4 |
| `expires.Sub` | 489:30 |
| `now.Before` | 492:6 |
| `productionRouteDescriptors` | 501:10 |
| `make` | 509:11 |
| `len` | 509:72 |
| `validDesiredState` | 514:5 |
| `validDesiredState` | 514:47 |
| `len` | 528:5 |
| `len` | 528:19 |

## State mutations and fallbacks

없음. 서술자 map 을 만들어 돌려준다.

## Safety conclusion

- **편집 계획:** B3 의 조건을 `familyActivationRemaining(expires, now)` 호출로 바꾼다. 같은 판정을 `FamilyActivation.LeaseCeiling`
  이 쓰므로, 만료 규칙이 패키지 안에 **한 곳**만 남는다(Manager 조건 ①). 경계 의미(만료 시각에 닿은 순간부터 만료)는 그대로다 —
  `TestLoadingAndLeasingJudgeExpiryAtTheSameInstant` 가 경계 세 점(−1ns · 0 · +1ns)에서 두 판정이 같은 답을 내는지 잰다.
- High-risk impact: yes (활성화 발급 경로). 승격을 넓히는 편집이 아니다 — 거절 조건의 표현만 바뀐다.

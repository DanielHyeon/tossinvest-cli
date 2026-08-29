# Review — interlock-gates-entry-not-exit

## Pre-Edit Gate

- base-commit: `238e51ffa700b4359fc5e33afd47966117c1b27a`
- 관측한 근거: `internal/app/engine/interlock.go` 조항 6, `internal/execgw/gateway.go`
  `mutationPlan.raisesExposure`·`checkEntry`, `engine-safety` 스펙 "자동화 게이트 기동 인터록"과
  "엔진 런타임 수명주기", 커밋 `43b3fd6`의 메시지, `internal/app/engine/deps_test.go`의 구조 단언 선례.
- 사용자 결정 2026-07-30: 조항 6 분리, `default_stop_pct` 5%, 자동 편입 유지.

## RED 관측

| 테스트 | RED 관측 |
|---|---|
| `TestARaisingMutationIsRefusedWhileProtectionIsUnwired` | `protection_test.go:69: a buy was confirmed on a build with no broker-resident protective execution` |
| `TestAReducingMutationIsAdmittedWhileProtectionIsUnwired` | 처음부터 통과 — 이 change가 **보존**하는 불변식이므로 RED가 없는 것이 정답 |
| `TestAnUnwiredProfileReportsEntryNotPermitted` 외 2건 | 컴파일 실패: `eng.Automation.EntryPermitted undefined` |
| `TestTheStartupSaysProtectionDiesWithTheProcess` | `the gate decision record does not mention …` |
| `TestNoShippedFileClaimsProtection` | 중간 단계에서 `gateway.go:168 assigns ProtectionOverrideForTest`를 잡아냈고, 판정 규칙을 "전달"과 "값을 만들어냄"으로 좁혀 해소 |

GREEN: `go test ./...` **3,889건 통과, 실패 0** (57 패키지). `go vet ./...` 무이슈.

## 적대적 리뷰

### A1 — 내보내진 override는 5.2의 컴파일 타임 보장을 약화한다

**지적.** 5.2는 "config 키도 Options 필드도 없다"를 근거로 조항을 신뢰할 수 있게 만들었다.
이 change는 `Options.ProtectionOverrideForTest`를 내보낸다. 컴파일러가 막던 것을 테스트가 막는다.

**판정: 유효한 약화다. 축소했고, 숨기지 않는다.**
불가피한 이유는 하나다 — 매수를 내야 하는 스위트가 다른 패키지에 있고(`internal/reconcile`,
`internal/app/engine`), `export_test.go` 선언은 자기 패키지에서만 보인다.
복구 수단은 `TestNoShippedFileClaimsProtection`이며, 이 저장소가 WTS 주문 mutator에
이미 쓰고 있는 것과 같은 등급의 단언이다(`deps_test.go`).
**남는 차이**: 컴파일 오류는 지울 수 없고 AST 테스트는 지울 수 있다. 그 차이는 실재하며
이 문서가 그것을 기록하는 것이 전부다.

### A2 — execgw 테스트 바이너리의 기본값이 WIRED다

**지적.** `export_test.go`의 `init`이 `defaultProtection`을 뒤집어서, 이 패키지의 291개
테스트 중 출하 기본값을 실제로 겪는 것은 두 개뿐이다. 앞으로 추가되는 게이트웨이 테스트는
자기도 모르게 보호 배선된 세계에서 돈다.

**판정: 유효. 의도된 교환이고, 대안이 더 나빴다.**
대안은 19개 생성 지점 각각에 `ProtectionOverrideForTest`를 다는 것이었고, 실제로 한 번
그렇게 했다가 되돌렸다 — Function Logic Map 요구 대상이 47개로 늘었고, 그 21개는
"필드 한 줄 추가"를 문서화하는 일이었다. 되돌린 뒤 28개.
`evaluateChain`이 같은 파일에서 같은 방식으로 이미 산다.
**남는 위험**: 매수 거부를 관측해야 하는 새 테스트가 `newUnprotectedGateway`를 쓰는 것을
잊으면 조용히 통과한다. `protection_test.go`의 파일 주석이 그 두 헬퍼의 차이를 말한다.

### A3 — "도달 가능한 매수 경로 없음"은 문자열 검사에 기대고 있다

**지적.** `TestTheEngineSpellsExactlyOneBuy`는 `Side: "buy"` 리터럴을 센다.
`side := computeSide()` 로 만든 매수는 이 검사를 통과한다.

**판정: 유효한 한계. 그러나 이 change의 안전은 여기 기대지 않는다.**
그 테스트는 proposal의 *논거*가 만료되는 시점을 알리는 장치이고, 실제 차단은
게이트웨이의 `raisesExposure`다. 그 값은 `strings.EqualFold(intent.Side, "buy")` —
문자열 리터럴이 아니라 **런타임 값**에서 계산된다. 변수로 만든 매수도 거부된다.
`TestTheRuntimeCannotReachEntryIssuance`는 심볼 참조를 AST로 보므로 문자열 한계가 없다.

### A4 — 열거에서 조항 6을 뺀 것이 운영자에게 불친절한가

**지적.** "왜 자동 진입이 안 되는가"를 묻는 운영자가 이제 기동 거부 목록에서 답을 못 찾는다.

**판정: 무효 — 반대다.** 기동이 성공하므로 거부 목록 자체가 나오지 않는다.
답이 있는 곳은 세 군데다: `engine.operating_mode`의 `entry_permitted: false`,
그 옆의 설명 문장(D6), 그리고 실제로 매수를 시도했을 때의 거부 사유
(`broker_protection_not_wired`). 고칠 수 없는 것을 고치라고 목록에 올리는 쪽이
불친절하다.

### A5 — 이 change가 **편입을 살아나게 한다**

**지적.** 논의는 손절에 집중했지만, 기동하는 루프는 셋이다. 대사 루프의 마지막 단계가
편입이고, 편입은 되돌릴 수 없다. 엔진이 뜨는 순간 `adoption.enabled = true`인 계좌의
무기록 보유가 전부 편입 대상이 된다.

**판정: 유효하고, 이 리뷰에서 가장 중요한 항목이다.**
이 change의 proposal은 "손절이 있게 한다"고 썼는데, 같은 스위치가 편입도 켠다.
사용자는 자동 편입 유지를 명시적으로 선택했고(2026-07-30) TSLA는 제외해 두었지만,
**그 결정은 편입이 즉시 일어난다는 것을 알고 내린 것이어야 한다.**
인계 항목 5.4에 관측 대상으로 적는다. 코드 변경 없음 — 편입 동작은 이 change가
건드리지 않았고, 바꾸는 것은 별개 결정이다.

### A6 — 기동 조건이 실제로 충족되는가

**지적.** 조항 3이 `trading.sell`과 `allow_live_order_actions`를 요구하는데 사용자 config는
둘 다 false다. 이 change를 landed해도 엔진은 여전히 안 뜬다.

**판정: 유효. 의도된 것이고, 인계에 적혀 있다.**
그 두 토글은 "엔진이 내 주식을 팔아도 된다"는 승인이며 §0.7에 따라 사람이 켠다.
이 change가 대신 켜지 않는다. 5.2에 기록.

## §0 재확인

- §0.1 — 주문 side effect 없음. 전 테스트가 httptest·fake broker.
- §0.3 — 게이트 OFF 경로 무수정. `TestTheGateOffPathIsUntouched`가 단언.
- §0.4 — 손절 즉시성은 0에서 유효로 **개선**된다. flatten-all 무수정.
- §0.6 — 정책 수치 무수정.
- §0.7 — 게이트 flip, `trading.sell`, `allow_live_order_actions` 전부 사람 몫.
- §0.9 — 완화는 하나("게이트 ON 기동 거부" → "기동 허용 + raising mutation 거부").
  나머지 여덟 조항·한도·attestation은 불변이며 `TestNothingElseRefusesTheOperatorConfiguration`이
  이름으로 단언한다.

# Branch Test Map: `registeredRoutes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 패키지 전 파일을 읽는다 | `TestEveryRouteGoesThroughTheSessionGate`의 하한 19(다른 파일 등록 3건 포함) | yes — 확장 전에는 `console.go`만 읽었고 `/dashboard`·`/orders`·`/signals`가 전부 보이지 않았다 | yes |
| B2 | 정렬된 순서로 파싱 | 같은 위 | — | yes |
| B3 | 등록자를 값으로 받으면 실패 | 리뷰 라운드 2의 변이 시연(`register := mux.HandleFunc`) | yes — 그 변이에서 라우트 가드 다섯이 전부 통과했다 | yes |
| B4 | CallExpr가 아닌 노드는 건너뛴다 | 구조 분기 — 전 파일 파싱이 커버 | — | yes |
| B5 | registrar가 아닌 호출은 건너뛴다 | 같은 위 | — | yes |
| B6 | 경로 인자 없는 등록은 실패 | 직접 변이 대상 — 현재 표에는 해당 등록이 없다 | — | n/a |
| B7 | 리터럴 아닌 경로는 실패 | 같은 위 | — | n/a |
| B8 | 서브트리 패턴은 실패 | `TestTheOrdersExceptionDoesNotReachAnyPathBeneathOrBesideIt`의 `/orders/` 케이스가 판정부 쪽에서 같은 사실을 잰다 | yes | yes |
| B9 | 핸들러 인자를 순회한다 | 전 등록 | — | yes |
| B10 | 불투명 핸들러는 실패 | `opaqueHandler`의 표 테스트(아래 별도 target) | — | yes |
| B11 | 게이트 체인의 바깥 호출 인식 | `TestEveryRouteGoesThroughTheSessionGate` | — | yes |
| B12 | `session0` 인식 | 같은 위 | yes — 래퍼를 떼면 실패한다 | yes |
| B13 | 중첩 래퍼 순회 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` | — | yes |
| B14 | 중첩 호출의 selector 인식 | 같은 위 | — | yes |
| B15 | 래퍼 이름 분기 | 같은 위 + `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs` | — | yes |
| B16 | `mutating` 인식 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` | yes | yes |
| B17 | `readOnly` 인식 | `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs` | yes — 개명 후 래퍼 제거 변이로 확인(issues.md I-2) | yes |
| B18 | 라우트가 하나도 안 읽히면 Fatal | positive control | — | yes |

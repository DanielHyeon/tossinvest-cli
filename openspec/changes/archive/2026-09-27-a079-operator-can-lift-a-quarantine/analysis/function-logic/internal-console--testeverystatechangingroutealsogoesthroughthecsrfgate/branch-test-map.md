# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

이 함수는 테스트 자신이므로 "Test" 열은 자기 자신을 가리킨다. a063은 분기를 바꾸지
않고 allowlist 리터럴에 두 항목을 추가한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록된 모든 라우트를 순회한다 | 자기 자신 | no | pass |
| B2 | 각 라우트를 allowlist 대비로 분류한다 | 자기 자신 | no | pass |
| B3 | 목록에 있는데 CSRF 게이트가 없으면 실패한다 | 자기 자신 | no | pass |
| B4 | 목록에 없는데 CSRF 게이트가 있으면 실패한다 | 자기 자신 | **yes** | pass |
| B5 | allowlist 항목을 순회한다 | 자기 자신 | no | pass |
| B6 | 목록에 있는데 등록되지 않았으면 실패한다 | 자기 자신 | no | pass |

**B4의 RED는 실제 관측이다** (2026-08-04): 격리 preview/apply 두 라우트를 등록한 뒤
allowlist에 넣기 전 상태에서 이 테스트가 정확히 B4로 실패했다.

```
/position-management/quarantine/preview is a read route behind the CSRF gate; it would be unopenable
/position-management/quarantine/apply is a read route behind the CSRF gate; it would be unopenable
```

즉 이 편집은 검사를 통과시키려고 목록을 늘린 것이 아니라, 게이트가 요구한 선언을
한 것이다.

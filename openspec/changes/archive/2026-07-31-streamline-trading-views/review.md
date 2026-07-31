# Review: streamline-trading-views

## Proposal-freeze — 2026-07-31

### Round 1 — REVISE

보이스: Product UX, 접근성, 적대적 Engineering, CSP security, testability.

독립 UI/UX 리뷰는 방향을 수용하되 다음 여섯 항목을 구현 전 보강하라고 판정했다.

1. 표가 아니라 header/nav/summary/긴 ID까지 포함한 375px 전체 문서 overflow 측정.
2. caption, scope, 접근성 트리에 남는 thead, focus-visible, 44px touch target, aria-current.
3. 페이지 잘림·조건주문 주체 불명·unknown 상태·원본 시각·캐시 불확실성은 접지 않음.
4. CSP 회귀를 특정 세 handler가 아니라 모든 `on[a-z]+`, script, javascript URL로 검사.
5. 관리 상태 라벨과 변경 버튼을 분리하고 심볼을 포함한 접근 가능한 버튼 이름 제공.
6. journal 실패에 없는 시각을 만들어내지 않도록 요구 범위를 축소.

처분: 전부 수용해 design/spec/tasks에 반영했다. requirement 수정이므로 strict validation과 독립
재리뷰 후에만 freeze한다.

Round 2는 위 여섯 항목을 해결로 판정했으나, 새 편입 버튼과 기존 소개문 “편입 버튼은 없다”의 모순
1건을 추가로 발견했다. design/task에 소개문 교체와 회귀 테스트를 명시해 수용했다.

### Round 2/3 — ACCEPT/FREEZE

독립 재리뷰는 첫 여섯 항목이 모두 해결됐다고 판정했고, Round 2에서 추가 발견한 소개문 모순도
design/task에 반영한 뒤 Round 3에서 해소를 확인했다. 추가 동결 차단 사항 없음.

결론: **ACCEPT/FREEZE**. strict validation과 PM Story↔change 1:1 검사를 통과했으며 구현 착수 가능.

Function Logic Map: not-applicable — 제안 단계이며 아직 production Go 함수는 변경하지 않았다.

## Post-implementation review — 2026-07-31

보이스: 독립 UI/UX, 접근성, 적대적 Engineering, CSP security, testability.

독립 리뷰는 다음 구현 증거를 확인했다.

1. positions 렌더 결과에 인라인 이벤트 handler, inline script, `javascript:` URL이 없고,
   편입·제외 동작은 심볼을 포함한 접근 가능한 이름과 CSRF를 가진 명시적 POST 버튼이다.
2. 편입·제외는 설정 endpoint만 호출하며 계좌나 주문을 변경하지 않는다. 상태 표시는 동작 버튼과
   분리되어 있다.
3. positions 기본 행은 기존 원장 값인 `Exit.Baseline`, `InitialStop`, `Taken/Rung`을 그대로
   표시하고 재계산하지 않는다.
4. orders는 주문 동작 form/button/input 없이 7개 기본 열로 축약됐고, 주문 ID와 평균 체결가는
   native details에 남겼다.
5. 원장·브로커 누락, stale, truncated, unknown, 원본 시각 안내는 펼쳐진 상태로 유지됐다.
6. deterministic fake fixture의 375px 검증에서 positions와 orders 모두
   `document.documentElement.scrollWidth == 375`를 만족했다.
7. caption, row/column scope, focus-visible, 현재 메뉴 표시와 모바일 touch target을 확인했다.

결론: **APPROVE**. 차단 finding 없음. Chrome 확장 프로그램이 요청하는
`/.well-known/appspecific/com.chrome.devtools.json`의 CSP 경고는 TossOS 문서의 inline handler와
무관한 확장 프로그램 측 probe이며, 이를 허용하기 위해 앱 CSP를 완화하지 않는다.

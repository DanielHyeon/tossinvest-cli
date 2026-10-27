## ADDED Requirements

### Requirement: 콘솔 테마는 토큰 계약과 모션 계약을 따른다

콘솔 스타일시트(`{{define "style"}}`)의 색·여백·크기는 `:root` 토큰에서만 읽어야 하고(SHALL), 다크 테마는 토큰 재정의로만 만든다(SHALL). 본문 글자 토큰은 세 면 위에서 4.5:1, UI 경계 토큰은 3:1 이상이어야 한다(SHALL). 전환은 compositor 속성(`translate`·`scale`·`opacity`·짧은 색 전환)만 쓰고(SHALL), POST 폼과 그 버튼·승인 영역에는 전환을 두지 않는다(SHALL NOT — 주문·토글·승인·청산의 즉시성). 측면 띠(side-tab) 강조와 `backdrop-filter`·blur 는 쓰지 않는다(SHALL NOT). `prefers-reduced-motion` 에서는 모션 토큰이 0 으로 리셋된다(SHALL). 클라이언트 스크립트와 CSP 완화는 도입하지 않는다(SHALL NOT).

#### Scenario: 두 테마의 대비
- **WHEN** 라이트·다크 토큰을 세 면 위에서 재면
- **THEN** 글자 토큰은 4.5:1 이상, 경계 토큰은 3:1 이상이다

#### Scenario: POST 폼의 전환 없음
- **WHEN** 스타일시트의 전환 규칙을 검사하면
- **THEN** `form:not([method=get])` 과 그 버튼·`.sticky-save`·`.approval-actions`·`.confirm-check`·`.console-approval` 은 `transition: none` 이다

#### Scenario: 앰비언트 톤 배경의 가독성
- **WHEN** 상태 톤(ok·warn·bad)별 배경 틴트를 라이트·다크에서 재면
- **THEN** 톤 쌍의 ΔE76 은 12 이상이고 그 위 글자 대비는 4.5:1 이상이며, 자동 재로드 화면과 reduced-motion 에서는 움직임이 멈춘다

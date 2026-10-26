# a174 디자인·접근성 검토

점검일: 2026-10-26
대상: `readiness` 절 여섯 행, Guardian 프리셋 비교표, 거래 정책 네 행, 외부 편입 링크.

## 적용성 및 방법

- UI 영향 있음. `impeccable`의 변경 대상 검사와 `fixing-accessibility`·`fixing-motion-performance` 지침을 적용함.
- 전용 `design-critic` 도구는 설치되어 있지 않았음. 대체로 템플릿 원문, a173 측정 하네스의 1280/390 light/dark 렌더, 콘솔 디자인 규칙을 수동 검토함.
- `impeccable detect --json internal/console/templates.go internal/console/templates_settings.go internal/console/templates_position_policy.go` 결과는 `[]`임.

## 확인 결과

| 항목 | 결과 |
|---|---|
| 준비도 | 절 여섯 행을 1280에서 3열, 좁은 화면에서 1열로 표시함. 상태와 사유를 분리했고 합산 기동 판정을 만들지 않음. |
| 프리셋 표 | 표 머리와 행 머리에 scope를 두고, 좁은 화면은 구조를 바꾸지 않고 `table-scroll`로 가로 스크롤함. 프리셋 기록 버튼의 접근 가능한 이름은 적용 후 비교 대상을 설명함. |
| 거래 정책 | 각 체크박스가 label에 포함되고, 상태 배지·이유가 같은 행에 있음. POST form·hidden CSRF·필드명은 유지함. |
| 링크 | 준비도 소유 탭 링크와 외부 편입의 상시 설정 링크는 목적지를 문장 안에 명시함. |
| 모션·즉시성 | 추가 CSS에 transition·animation을 넣지 않았음. 저장·주문·엔진 제어의 클릭 동작과 클릭 수를 바꾸지 않았음. |
| 렌더 | 격리 콘솔 GET-only 96 화면에서 `/settings/daily` desktop 2445px·notice 3, mobile 3365px·notice 3을 확인함. |

## 남은 한계

실제 운영 콘솔이나 운영 설정을 열지 않았고, Manager 4.5의 직접 브라우저 검증은 수행하지 않았음. 전용 `design-critic` 점수는 도구 부재로 제공하지 않음.

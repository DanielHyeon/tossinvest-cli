# a175 디자인·접근성 검토

## design-critic

- 시각 근거: `after-review-20261026T1130Z`의 light 390px 캡처에서 후보 필터, preset 비교표, strategy runtime을 직접 확인함. 후보 필터는 Help 설명만 기본 닫힘이고 preview 계약·CAS·사유 코드가 본문에 남음.
- `impeccable detect --json internal/console/templates.go internal/console/templates_optimization.go internal/console/templates_signals.go` 결과: `[]`, exit=0 (`impeccable-detect.json`).
- 판정: Operate 표면으로서 status·사유·읽기 경로를 색이나 hover에만 의존하지 않음. 390px preset 표의 가로 스크롤은 의도된 `.table-scroll` 안에 있음.
- 남은 관찰: 후보 필터는 Help 설명만 접어도 높이 목표 1200을 초과함. preview 계약·값·미승인·미측정·누락 사유·CAS를 접거나 행 간격을 누르지 않았음.

## fixing-accessibility

| 점검 | 근거 | 결과 |
|---|---|---|
| 표 머리 | preset·veto 표의 `scope=col`, 행의 `scope=row` | 통과 |
| 키보드 표 영역 | 새 표가 `role=region`·고유 `aria-label`·`tabindex=0` 보유 | 통과 |
| 링크·폼 의미 | 기존 native `<a>`·`<form>`·`<button>` 유지, 새 custom control 없음 | 통과 |
| 접기 범위 | `shadow-band`, `filter-description`은 설명만 포함하고 `.notice`·`.danger` 없음 | 통과 |
| 모바일 summary | a175 CSS에 `details.shadow-band summary, details.filter-description summary { display: list-item; }` 복원 | 통과 |
| motion | `TestAmbientMotionContract`가 POST form을 품는 `.preset-compare`를 glow 대상에서 배제 | 통과 |

정적 추적은 `accessibility-static-check.txt`, 동작 계약은 `green-console-output.txt`의 대상·전체 시험으로 보강함.

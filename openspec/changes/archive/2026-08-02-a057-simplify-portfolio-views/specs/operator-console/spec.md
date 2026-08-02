## MODIFIED Requirements

### Requirement: 거래 화면 핵심 정보 계층과 반응형 표시
콘솔의 대시보드·포지션·주문 화면은 운영자가 첫 화면에서 판단해야 하는 사실을 1차 정보로 표시해야 한다(SHALL). 대시보드와 포지션 화면은 동일한 read-only 보유 projection을 사용해야 하며(SHALL), 둘 중 하나가 lifecycle·desired/effective policy 또는 exit evidence를 다르게 해석해서는 안 된다(SHALL NOT).

각 보유의 1차 행은 종목명/심볼, 수량, 평균가, 현재가, 익절, 손절, 추적 회수, 기준선, 고점(high-water), 총금액, 미실현 PnL을 항상 표시해야 한다(SHALL). `익절`은 canonical exit projection의 다음 익절가, `손절`은 최초 손절, `추적 회수`는 다음 보호선, `기준`은 현재 보호선, `고점`은 high-water를 의미해야 한다(SHALL). canonical snapshot이 없거나 stale·lifecycle generation 불일치·runtime unknown이면 actionable 가격을 계산하거나 raw 원장값으로 대체하지 않고 `—`와 상태/사유를 표시해야 한다(SHALL NOT). 저장된 비실효 기준은 상세 증거로만 표시할 수 있다.

관리 상태와 차단·stale·미측정 verdict는 접지 않고 표시해야 한다(SHALL). 낮은 우선순위의 자격 provenance, 원장 수량·평단, position/policy generation, decision/snapshot/observation ID, 평가 시각, 원본 저장 기준은 키보드로 여는 native HTML 상세 영역에 둘 수 있다. 상세 영역을 열지 않아도 앞 문단의 1차 필드와 안전 verdict를 모두 확인할 수 있어야 한다(SHALL).

주문은 미체결 일반/조건/합계, 필터 결과, 각 주문의 시각·심볼/시장·방향·상태·수량·가격·발주 주체를 1차 정보로 표시해야 한다(SHALL). 낮은 우선순위의 원장 식별자·래칫 진단·주문번호·평균체결가는 native HTML 상세 영역에 둘 수 있다.

세 화면은 375 CSS pixel viewport에서 header, navigation, summary, table, 긴 식별자를 포함한 문서 전체의 수평 overflow 없이 핵심 정보를 읽고 조작할 수 있어야 한다(SHALL). 표는 접근 가능한 이름, 열/행 header 관계를 유지하고(SHALL), 현재 navigation은 `aria-current`를 제공해야 한다(SHALL). 버튼·summary·필터 링크는 keyboard focus가 보이고 모바일 조작 대상은 최소 44 CSS pixel 높이를 가져야 한다(SHALL). 이 반응형 표시는 JavaScript나 외부 asset을 요구하지 않아야 한다(SHALL NOT). 브로커·journal 미측정, 페이지 잘림/하한 건수, 조건주문 발주 주체 불명, 해석 불가능 상태, 파싱되지 않은 원본 시각, 캐시 실패·보류·stale 안내는 접지 않고 해당 수치와 함께 노출해야 한다(SHALL). 시각이 기록된 캐시·측정에는 그 시각을 표시하되, 시각이 없는 journal 실패에 화면이 시각을 만들어내지 않아야 한다(SHALL NOT).

#### Scenario: 대시보드와 포지션의 같은 보유
- **WHEN** 같은 캐시와 journal 상태에서 `/dashboard`와 `/positions`를 렌더하면
- **THEN** 두 화면의 해당 종목은 같은 관리 verdict와 같은 익절·손절·추적 회수·기준·고점 값을 표시한다

#### Scenario: 관리 포지션의 핵심 정보 스캔
- **WHEN** fresh canonical exit snapshot이 있는 관리 포지션을 렌더하면
- **THEN** 상세 영역을 열지 않은 같은 1차 행에서 종목명/심볼, 수량, 평균가, 현재가, 익절, 손절, 추적 회수, 기준, 고점, 총금액, 미실현 PnL을 확인할 수 있다

#### Scenario: canonical 근거가 없는 저장 기준
- **WHEN** seed-only raw exit state 또는 stale·generation mismatch 상태를 렌더하면
- **THEN** 1차 actionable 라인은 `—`와 상태/사유를 표시하고 저장된 raw 가격은 상세 증거로만 표시한다

#### Scenario: 보조 진단 접기
- **WHEN** 관리 포지션 행을 처음 렌더하면
- **THEN** 자격 provenance와 journal/snapshot 식별자는 native 상세 영역 안에 있고 관리·stale·미측정 verdict는 상세 영역 밖에 있다

#### Scenario: 주문 화면의 작은 viewport
- **WHEN** 주문이 있는 화면을 375 CSS pixel 폭으로 렌더하면
- **THEN** `documentElement.scrollWidth`가 viewport 너비를 넘지 않고 각 주문의 핵심 필드 라벨과 값을 읽을 수 있다

#### Scenario: 보유 화면의 작은 viewport
- **WHEN** KR 또는 US 보유가 있는 대시보드나 포지션 화면을 375 CSS pixel 폭으로 렌더하면
- **THEN** 1차 보유 필드는 라벨과 함께 card flow로 읽히고 문서 전체의 수평 overflow가 생기지 않으며 상세 summary는 최소 44 CSS pixel 높이다

#### Scenario: 미측정 상태의 진실성
- **WHEN** 브로커 또는 journal 읽기가 실패하거나 캐시가 stale이면
- **THEN** 해당 사유와 기록된 시각이 접히지 않은 안내로 표시되고, 기록되지 않은 시각을 만들거나 빈 목록/현재 측정값으로 위장하지 않는다


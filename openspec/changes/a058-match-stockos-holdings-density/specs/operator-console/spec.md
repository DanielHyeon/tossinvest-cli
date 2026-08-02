## MODIFIED Requirements

### Requirement: 거래 화면 핵심 정보 계층과 반응형 표시
콘솔의 대시보드·포지션·주문 화면은 운영자가 첫 화면에서 판단해야 하는 사실을 1차 정보로 표시해야 한다(SHALL). 대시보드와 포지션 화면은 동일한 read-only 보유 projection과 동일한 shared holdings template을 사용해야 하며(SHALL), 둘 중 하나가 lifecycle·desired/effective policy 또는 exit evidence를 다르게 해석하거나 다른 열 구조로 표시해서는 안 된다(SHALL NOT).

각 보유의 desktop 1차 표는 정확히 `종목, 수량, 평균가, 현재가, 라인, 총금액, 미실현 PnL` 일곱 열을 이 순서로 표시해야 한다(SHALL). header와 body는 같은 명시적 column grid를 사용하고(SHALL), 수량·가격·금액·손익 header와 값은 같은 오른쪽 정렬 축과 tabular numerals를 사용해야 한다(SHALL). holdings table의 desktop header/body typography와 cell spacing은 StockOS reference의 compact scan density(약 10/12 CSS px, 15/18 CSS px line-height, 8×12 CSS px padding)에 맞춰야 한다(SHALL).

`라인` 열은 익절, 손절, 추적 회수, 기준, 고점을 이 순서의 compact stack으로 표시해야 한다(SHALL). `익절`은 canonical exit projection의 다음 익절가, `손절`은 최초 손절, `추적 회수`는 다음 보호선, `기준`은 현재 보호선, `고점`은 high-water를 의미해야 한다(SHALL). canonical snapshot이 없거나 stale·lifecycle generation 불일치·runtime unknown이면 actionable 가격을 계산하거나 raw 원장값으로 대체하지 않고 `—`를 표시해야 한다(SHALL NOT).

관리·차단·pending·excluded 및 exit evidence의 concise verdict는 접지 않고 표시해야 한다(SHALL). 반복되거나 긴 reconciliation detail, management explanation, exit reason, 자격 provenance, 원장 수량·평단, generation과 decision/snapshot/observation ID, 평가 시각, 원본 저장 기준은 키보드로 여는 native HTML 상세 영역에 두어야 한다(SHALL). 상세 영역을 열지 않아도 일곱 열의 핵심 값과 concise safety verdict를 확인할 수 있어야 한다(SHALL).

375 CSS pixel viewport에서는 같은 일곱 필드를 label-value card flow로 표시하고 문서 전체에 수평 overflow가 생기지 않아야 한다(SHALL). 상세 summary는 최소 44 CSS pixel 높이와 visible keyboard focus를 유지해야 한다(SHALL). 이 표시는 JavaScript, 입력 control, 외부 asset을 요구하지 않아야 한다(SHALL NOT).

#### Scenario: desktop 일곱 열 정렬
- **WHEN** 대시보드 또는 포지션 화면을 desktop viewport에서 렌더하면
- **THEN** 일곱 header와 모든 보유 값은 같은 explicit column grid에 놓이고 수량·평균가·현재가·총금액·미실현 PnL은 오른쪽 축으로 정렬된다

#### Scenario: StockOS형 compact line stack
- **WHEN** 보호 근거가 있는 KR 또는 US 보유를 렌더하면
- **THEN** `라인` 열은 concise verdict 다음에 익절·손절·추적 회수·기준·고점을 12 CSS px 수준의 compact stack으로 표시하고 verbose reason은 primary row에 반복하지 않는다

#### Scenario: 근거가 없는 보유의 진실성
- **WHEN** canonical exit evidence가 없거나 stale 또는 generation mismatch이면
- **THEN** visible concise verdict와 다섯 개 `—` 값이 상태를 알리고 긴 설명과 raw evidence는 상세 영역에만 표시된다

#### Scenario: 작은 viewport card flow
- **WHEN** 같은 표를 375 CSS pixel 폭에서 렌더하면
- **THEN** 일곱 필드가 label-value card flow로 읽히고 문서의 수평 overflow가 없으며 상세 summary는 최소 44 CSS pixel 높이다

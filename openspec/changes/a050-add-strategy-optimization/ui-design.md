# `/optimization` 설정 화면 설계

## 1. Metadata

- Change: `a050-add-strategy-optimization`
- Story: `STORY-TOS-a050`
- Status: Draft, proposal-freeze review required
- Date: 2026-07-31
- Reviewers: Manager, UI/UX, adversarial Eng, security
- References: StockOS `OptimizationPage`, `CommonExitPolicyCard`, lane console, settings history, stop-protection confirmation; TossOS a041~a051 contracts
- Operator input rule: 자유 text/textarea/number/contenteditable/symbol 입력과 typed confirmation 금지. server-defined option만 선택.

## 2. Context

현재 TossOS `/optimization`은 `BALANCED`, `RUNNER`, `HYBRID 50` 세 정책의 ladder를 보여주고 정책 ID 하나만 저장한다. 향후 후보 필터, 전략 lane, 시장 스케줄, 성과, rollback까지 같은 메뉴에 들어오면 설정을 단순 나열하는 방식은 사용자가 적용 범위와 위험을 혼동하게 만든다.

이 화면의 목적은 “값을 많이 보여주는 것”이 아니라 사용자가 다음 네 가지를 저장 전에 답할 수 있게 하는 것이다: 무엇을 바꾸는가, 기본값과 현재 적용값은 무엇인가, 어느 포지션·시장·기동부터 적용되는가, LIVE 권한이나 기존 보호가 같이 바뀌는가.

## 3. Information architecture

`/optimization`은 아래 여섯 카테고리를 고정 순서로 제공해야 한다. 데스크톱은 좌측 보조 탐색, 폭 768px 미만은 상단 `<select>` 또는 동일 의미의 disclosure 탐색을 사용한다. 카테고리 이름과 순서는 API와 모바일 소비자도 공유한다.

| 순서 | category ID | 사용자 메뉴 | 소유 change | 기본 화면 상태 |
|---:|---|---|---|---|
| 1 | `overview` | 개요 | a050 | 펼침, 읽기 전용 |
| 2 | `exit-protection` | 익절·보호 | a041, a042, a045, a050 | 공통 정책 기본 탭 |
| 3 | `position-management` | 종목별 관리 | a044 | 공통 정책 상속 목록 |
| 4 | `candidate-filters` | 후보 필터 | a046 | 승인 상태 요약 |
| 5 | `strategy-runtime` | 전략·실행 | a047, a048 | 모든 진입 토글 OFF |
| 6 | `performance-history` | 성과·이력 | a049, a050 | 최근 30일, 전체 시장, 읽기 전용 |

카테고리 deep link는 `/optimization?category=<category-id>`를 사용한다. 알 수 없는 ID는 `overview`로 이동시키고 경고를 표시한다. 설정 저장·LIVE 승인·rollback은 URL 이동만으로 실행되어서는 안 된다.

## 4. Layout contract

```text
┌ 최적화 ─ desired v12 / effective v11 ─ LIVE 상태: OFF ─ 재시작 필요 ┐
│ [개요]                │ 카테고리 제목 · 한 문장 목적                 │
│ [익절·보호]           │ 현재값 / 기본값 / 적용 범위 요약             │
│ [종목별 관리]         │                                                │
│ [후보 필터]           │ 설정 그룹                                     │
│ [전략·실행]           │  label · 설명 · 단위 · 입력 · 허용범위        │
│ [성과·이력]           │  현재 / 기본 / 적용시점 / 안전방향            │
│                       │                                                │
├───────────────────────┴ 변경 3개 · [초기화] [미리보기] [승인·저장] ┤
```

- 첫 시각 계층은 `desired/effective version`, LIVE 상태, 재시작 필요 여부다.
- 두 번째는 현재 카테고리의 목적과 적용 범위다.
- 세 번째가 설정 입력이다. 카드 중첩 대신 의미가 있는 `fieldset`과 표/행을 사용한다.
- 변경이 없으면 하단 저장 바는 숨기거나 비활성화한다. 변경이 생기면 항상 변경 개수와 적용 대상을 표시한다.
- 한 저장은 현재 카테고리가 소유한 변경 key만 전송한다. 다른 카테고리의 미저장 draft를 암묵적으로 함께 저장하지 않는다.

## 5. Field presentation contract

모든 설정 행은 다음 정보를 생략 없이 표시해야 한다.

| 필드 | 필수 표시 |
|---|---|
| 이름 | 운영자가 이해하는 한국어 label과 안정된 parameter key |
| 설명 | 값이 커지거나 작아질 때 실제 주문·보호·선별에 생기는 변화 |
| 값 | draft, 현재 desired, 현재 effective, registry 기본값 |
| 제약 | 단위, min/max, step, enum, 시장 범위 |
| 적용 | 즉시/다음 평가/다음 엔진 기동/신규 포지션만 |
| 안전 | 더 보수적/위험 확대/중립, LIVE 확인 필요 여부 |
| 출처 | 소유 change, policy/version, 성과 evidence 상태 |

기본값은 control 내부 placeholder가 아니라 `기본값 5.0%`처럼 별도 텍스트로 표시한다.
`기본값으로 되돌리기`는 server-defined option을 고르는 실제 draft 변경이다. `0`, 빈 문자열,
미승인, 측정 불가를 서로 대신 사용하지 않는다. 모든 변경 control은 preset radio tile, select/chip,
toggle, discrete step 또는 현재 row action이며 운영자가 임의 문자열이나 숫자를 입력하지 않는다.

## 6. Category fields and defaults

### 6.1 개요

- 저장 기능이 없다.
- 카테고리별 변경 가능 여부, desired/effective version, restart 필요, 마지막 적용 actor/time/reason, stale/unavailable 상태를 보여준다.
- LIVE·lane·automation·kill switch는 설정값과 분리된 상태 요약으로만 표시한다.

### 6.2 익절·보호

기본 모드는 세 공통 정책 중 하나를 선택하는 preset 화면이다. 아직 운영자가 승인하지 않았다면 effective 기본은 `미승인 · 기존 RATCHET 유지`이고 추천 선택만 `COMMON_LADDER_HYBRID_50`이다. 추천값을 자동 저장해서는 안 된다.

| preset | 표시명 | T1 | T2 | T3 | T4 | runner |
|---|---|---|---|---|---|---|
| `COMMON_LADDER_BALANCED` | 균형형 | `+1.5 / stop 0 / partial 0` | `+2.5 / +1.0 / 25%` | `+4.0 / +2.0 / 25%` | `+6.0 / +3.5 / 전량` | 없음 |
| `COMMON_LADDER_RUNNER` | 러너형 | `+2.5 / stop 0 / partial 0` | `+4.5 / +2.0 / 15%` | `+7.0 / +3.5 / 0` | 고정 목표 없음 / `+5.0` 보호 | 없음 |
| `COMMON_LADDER_HYBRID_50` | 하이브리드 50, 추천 | `+1.8 / stop 0 / partial 0` | `+3.0 / +1.2 / 25%` | `+4.8 / +2.5 / 잔량 1/3` | `+6.5 / +3.8 / 0` | high-water `-6.5%` |

표의 각 셀은 `목표 수익률 / 진입가 대비 보호 / 잔량 기준 익절`임을 header와 도움말에서 설명한다. `999%` sentinel은 입력값으로 노출하지 않고 `고정 목표 없음`으로 표시한다.

`고급 설정` disclosure를 열면 선택 preset을 base로 registry가 제공하는 discrete option을 선택하는
versioned candidate를 만든다. 자유 숫자 입력이나 임의 step은 없다.

- rung별 `target_pct`, `stop_pct`, `partial_ratio`
- `runner_trail_pct` 또는 `없음`
- 최종 전량익절 여부

검증은 target 엄격 증가, stop 비감소, stop < target, partial 0~1, 전량 후 후속 rung 금지, runner와 최종 전량의 모순 금지를 포함한다. 1주 preview는 중간 rung마다 `매도 0주 · 보호선만 승격`, 최종/보호 breach는 `1주 전량`으로 표시한다.

a045 브로커 보호는 `미검증/사용 불가`, `준비됨·기본 OFF`, `ACTIVE`, `복구 필요` 상태를 표시한다. attestation 전에는 임의 주문유형 기본값을 만들지 않는다.

### 6.3 종목별 관리

| parameter | 기본값 | UI 형식 | 설명 |
|---|---:|---|---|
| `adoption.enabled` | `OFF` | switch | 외부 매수 보유를 자동 관리 대상으로 편입 |
| `adoption.default_stop_pct` | `5.0%` | registry discrete step, `2.0~20.0%`, step `0.5%` | 편입 관측가 아래의 합성 최초 손절폭 |
| `adoption.include_symbols` | 빈 목록 | 현재 position/candidate 행의 `명시 편입` action | 전역 OFF여도 명시 편입할 종목 |
| `adoption.exclude_symbols` | 빈 목록 | 현재 position/candidate 행의 `제외` action | 편입하지 않을 종목, include보다 우선 |
| per-symbol exit policy | `공통 정책 상속` | preset selector | 해당 generation에만 적용할 정책 override |

종목 행은 계좌·시장·symbol·관리 상태·generation·현재 정책·상속/override·현재 보호선을 먼저 보여준다. `release`와 `re-adopt`는 일반 저장과 분리된 위험 작업이며 미체결 보호/청산이 있으면 비활성화 사유를 같은 위치에 표시한다.

### 6.4 후보 필터

a046 registry가 선언한 threshold만 렌더링한다. 각 항목은 시장, feature key, 비교 방향, 단위, 승인값, evidence digest와 표본 시각을 표시한다. 최초 기본은 `미승인 · 구조적 0 · verdict 비활성`이며 숫자 0을 임계 기본값처럼 표시하지 않는다.

분석 결과가 없거나 stale이면 입력과 저장을 비활성화하고 누락된 evidence를 설명한다. 시장 간 값을 복사하는 control은 제공하지 않는다.

### 6.5 전략·실행

| parameter | 기본값 | 설명 |
|---|---|---|
| lane desired state | `OFF` | 신규 진입만 제어, exit/reconcile은 계속 실행 |
| engine autostart | `OFF` | 콘솔 부팅 시 엔진 시작 시도, 주문 권한을 만들지 않음 |
| market scope | 선택 없음 | 승인된 KR/US 시장만 선택 |
| session scope | 정규장만 | 장전·장후 신규 진입은 금지 |
| scheduler desired state | `OFF` | 휴장·DST·API budget을 반영한 진입 스케줄 |

전략 parameter는 a047 source policy에서 provenance와 범위가 고정된 key만 노출한다. lane ON, autostart, automation gate, LIVE trading은 서로 다른 행과 승인 동작으로 표시하며 하나의 `모두 켜기` 버튼을 두지 않는다.

### 6.6 성과·이력

- 기본 필터는 `최근 30일`, `전체 시장`, `전체 lane`, `complete lineage만`이다.
- 비용 후 P&L, realized R, PF, MDD, slippage, 5/15/30분 markout, MFE/MAE와 표본 수를 표시한다.
- `link_missing`, `not_measured`, `insufficient_sample`은 0과 다른 상태로 설명한다.
- history와 snapshot은 변경 key, before/after, actor, reason, 적용 시각, settings/evidence digest를 표시한다.
- rollback은 과거 row를 수정하지 않고 새 candidate를 만들어 preview 후 적용한다.

## 7. Interaction states

1. **Loading:** 입력 skeleton을 실제 기본값처럼 보이지 않게 하고 저장 control을 비활성화한다.
2. **Loaded/clean:** current·effective·default를 표시하고 저장 bar는 비활성화한다.
3. **Dirty:** 변경 필드와 개수를 표시하고 `초기화`, `미리보기`를 활성화한다.
4. **Preview:** before/after, 영향 대상, 적용 시점, restart, safety direction, 1주 예시와 evidence를 표시한다.
5. **Risk confirmation:** 손절폭 확대나 보호 약화 설정은 별도 dialog에서 3초 대기와 명시 체크를 요구한다. typed phrase와 자유 reason 입력은 없고 server-defined reason chip을 선택한다. lane/LIVE/automation gate/activation-manifest authority는 optimization mutation이 아니며 이 화면에서는 상태와 별도 승인 경로만 표시한다. 설정 저장과 LIVE 승인을 결합하지 않는다.
6. **Applying:** 중복 제출을 막고 idempotency key를 유지한다.
7. **Success:** 새 desired/effective version, 즉시/재기동 적용 여부와 audit ID를 표시한다.
8. **CAS conflict:** 412 후 서버 최신값을 다시 읽고 사용자의 draft를 자동 재적용하지 않는다. field diff를 보여준다.
9. **Stale/unavailable:** 마지막 값은 stale badge와 시각을 붙여 읽기만 허용한다.
10. **Insufficient evidence:** 추천·apply를 비활성화하고 부족한 표본·lineage·metric을 열거한다.

## 8. API contract

```ts
type OptimizationCategory =
  | "overview"
  | "exit-protection"
  | "position-management"
  | "candidate-filters"
  | "strategy-runtime"
  | "performance-history";

interface OptimizationFieldDescriptor {
  key: string;
  category: OptimizationCategory;
  label: string;
  description: string;
  type: "boolean" | "decimal" | "integer" | "enum" | "symbol-list";
  unit?: "%" | "ratio" | "minutes" | "count";
  defaultState: "value" | "unapproved" | "not-applicable";
  defaultValue?: boolean | string | number | string[];
  min?: number;
  max?: number;
  step?: number;
  choices?: { value: string; label: string; description: string }[];
  control: "radio-tile" | "select" | "chip" | "toggle" | "discrete-step" | "row-action" | "read-only";
  ownerChange: string;
  applyTiming: "immediate" | "next-evaluation" | "next-engine-start" | "new-position-only";
  safetyDirection: "safer-when-higher" | "safer-when-lower" | "neutral" | "contextual";
}

interface OptimizationPreview {
  baseVersion: string;
  changes: Array<{ key: string; before: unknown; after: unknown }>;
  affectedMarkets: string[];
  affectedPositions: string[];
  existingPositionsUnchanged: boolean;
  restartRequired: boolean;
  liveStateUnchanged: boolean;
  riskConfirmationRequired: boolean;
  validationErrors: Array<{ key: string; code: string; message: string }>;
  evidenceDigest?: string;
}
```

## 9. Non-functional requirements

- NFR-1: 360px 폭에서 수평 페이지 overflow가 없어야 한다. 넓은 비교표만 자체 scroll 영역을 가질 수 있다.
- NFR-2: 모든 touch target은 최소 44×44px이어야 한다.
- NFR-3: keyboard만으로 category 이동, field 편집, preview, 취소가 가능해야 한다.
- NFR-4: 상태와 위험은 색만으로 전달하지 않고 text/icon/ARIA status를 함께 사용해야 한다.
- NFR-5: 입력 오류는 해당 field와 summary에 연결되고 사용자가 입력한 draft를 보존해야 한다.
- NFR-6: 기본값·현재값·effective 값은 3초 scan에서 구분되어야 한다.
- NFR-7: inline event handler를 사용하지 않고 현재 CSP를 유지해야 한다.
- NFR-8: optimization 화면은 `input[type=text]`, `input[type=number]`, `textarea`,
  `contenteditable`, 임의 symbol 입력과 typed confirmation phrase를 제공하지 않는다.
- NFR-9: 모든 변경 payload는 registry가 제공한 stable option ID 또는 현재 row의 stable identity로만 구성한다.

## 10. Acceptance criteria

- AC-1: Given 사용자가 어느 카테고리를 열든, When 설정 행을 본다, Then label·설명·단위·기본값·현재값·effective 값·적용 시점이 함께 표시된다.
- AC-2: Given 1주 포지션 preview, When partial rung 값을 변경한다, Then 중간 매도 0주와 보호선 승격, 최종 1주 전량이 표시된다.
- AC-3: Given LIVE 상태에서 보호를 약화하는 draft, When 저장을 누른다, Then before/after와 3초 확인 없이 apply되지 않는다.
- AC-4: Given 모바일 360px viewport, When 여섯 카테고리와 저장 bar를 사용한다, Then 페이지 수평 overflow 없이 모든 control이 44px 이상이다.
- AC-5: Given CAS conflict, When 서버 version이 바뀌었다, Then 최신값과 draft 차이를 보여주고 자동 재시도하지 않는다.
- AC-6: Given 미승인 후보 threshold, When 후보 필터를 연다, Then `미승인`을 표시하고 0을 기본 임계값으로 표시하지 않는다.
- AC-7: Given 어떤 설정을 저장한다, When apply가 성공한다, Then lane/LIVE/automation/kill-switch 상태는 별도 승인 없이는 동일하다.
- AC-8: Given 운영자가 설정을 변경한다, When control을 사용한다, Then 직접 값을 입력하지 않고 preset/선택/토글/discrete step만으로 완료한다.

## 11. Edge cases

- EC-1: registry에는 있으나 현재 시장에 적용되지 않는 field는 `해당 없음`으로 읽기 전용 표시한다.
- EC-2: unknown policy ID는 선택기에 추가하지 않고 현재 설정 거부 상태와 원문 ID를 표시한다.
- EC-3: 기존 position snapshot과 새 공통 정책이 다르면 둘을 동시에 표시하고 자동 rebind하지 않는다.
- EC-4: history 대상 key가 최신 registry에서 제거됐으면 raw key/value를 보존하되 apply는 금지한다.
- EC-5: category 이동 중 dirty draft가 있으면 유지하고, 페이지 이탈 전 미저장 변경을 경고한다.
- EC-6: include/exclude 변경은 symbol 문자열 입력이 아니라 positions/candidates의 현재 행 action에서 수행하고 canonical symbol ID를 그대로 사용한다. exclusion이 inclusion보다 우선함을 표시한다.

## 12. Out of scope

- 자동 추천 알고리즘과 자동 apply. 근거가 충분해도 첫 버전은 사람 candidate만 적용한다.
- 설정 저장과 LIVE/lane 승인 결합.
- 공용 인터넷용 계정·권한 UI.
- 모바일 네이티브 앱 구현. a051은 동일 schema를 제공하는 데까지만 담당한다.

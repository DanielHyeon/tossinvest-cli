# a175 표시 요구 목록

## 조사 범위

- 기준 commit: `526650e50ee99fbd6c9bd6ff4121101458416612`
- 원문: `openspec/specs/strategy-optimization/spec.md`, `candidate-veto-thresholds/spec.md`, `strategy-runtime/spec.md`
- 구현 원칙: 값·레지스트리·핸들러를 바꾸지 않고 읽기 전용 표시만 재배치함. 값 미확정·미측정·배선 부재를 0 또는 정상 상태로 바꾸지 않음.

## strategy-optimization

| 표시 대상 | 화면 요구 |
|---|---|
| 최적화 카테고리 | 한 내비게이션에서 6개 카테고리를 찾을 수 있어야 함. 개요 KPI는 개요에만 표시함. |
| 공통 exit 프리셋 | 등록 레지스트리의 정책·사다리 단계·정책 identity를 한 비교표에서 읽을 수 있어야 함. 각 정책의 기존 preview→apply form 계약은 변경하지 않음. |
| runner 마지막 단계 | 레지스트리 값 `999.0`은 화면에 노출하지 않고 `고정 목표 없음`으로 표시함. 원본 값·계산·정책은 바꾸지 않음. |
| position·strategy·performance relay | 중계 설명은 한 문단으로 제한하고 기존 링크 수 1·2·1과 성과·이력 감사 표를 보존함. |
| hypothetical·exit-protection | action, hidden field, button 문구, disabled 분기를 바꾸지 않음. |

## candidate-veto-thresholds

| 표시 대상 | 화면 요구 |
|---|---|
| 후보 필터 key×시장 | `seen_late`, `extended`, `near_high`의 KR·US 값을 각각 읽을 수 있어야 함. 시장별 수치·세션·사유를 섞지 않음. |
| 공통 값 | KR·US의 표시 값이 모두 같을 때만 공통 행으로 한 번 표시함. 하나라도 다르면 두 시장 셀에 각각 표시함. |
| 미승인·미측정 | `unapproved`, `not_measured`, missing evidence 코드와 preview 계약을 숨기거나 0·통과로 바꾸지 않음. |
| 증거 | sample count, missing rate, evidence digest, human activation record 부재를 사유 코드로 모두 노출함. |
| 쓰기 경계 | dormant registry의 읽기 전용·CAS·preview 요구를 표시해도 정책 활성화·값 변경 기능을 추가하지 않음. |

## strategy-runtime

| 표시 대상 | 화면 요구 |
|---|---|
| 시장 projection | KR·US는 독립 열로 표시함. 한 시장 load 실패·미배선이 다른 시장의 값·상태를 가리지 않음. |
| 안전 상태 | `UNWIRED`, `NOT_CONFIGURED`, `UNKNOWN`, `OFF` 등 상태와 사유는 접지 않고 화면 본문에 남김. |
| schema·golden | projection schema와 golden registry의 설명은 `details.tech`에 접을 수 있으나, 수치·사유·안전 상태는 넣지 않음. |
| 보호 상태 | `ProtectionReady`와 exit/reconcile/protection의 상태를 entry OFF와 혼동하지 않게 표시함. |
| 미확정 값 | unknown·미설정 값은 기본값·0·정상값으로 대체하지 않고 원래 코드와 사유를 표시함. |

## 공통 접근성·반응형 요구

- 모든 비교표 머리에는 `scope`를 둠.
- `<details>`는 설명만 포함하며 `.notice`·`.danger`를 포함하지 않음.
- 390px에서 후보 필터 행렬은 잘리지 않게 표시하고, 프리셋 표는 `table-scroll`로 수평 스크롤할 수 있게 함.
- reduced-motion 환경에서 새 전환은 재생하지 않으며, keyboard focus와 링크·form의 기존 조작 계약을 보존함.

## research-first 기록

- 저장소의 `html/template`, `table-scroll`, `details.tech`, 후보·정책·runtime 레지스트리와 DOM 시험 helper를 확인함.
- 외부 UI 표 라이브러리·새 의존성은 도입하지 않음. 현재 Go 표준 라이브러리와 기존 템플릿 구조로 요구를 충족하며, 의존성·라이선스·호환성 위험을 늘리지 않음.

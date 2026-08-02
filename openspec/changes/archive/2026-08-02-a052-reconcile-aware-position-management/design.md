## Context

실행 중 엔진의 `PositionPolicyCommandService`는 startup config와 `reconcile.Tracker`를 소유하지만 RPC는 position lifecycle의 List/Preview/Apply만 제공한다. 웹 콘솔과 a051 API는 `positionpolicy.Descriptor()` 기본값을 렌더하므로 config desired, running-engine effective, active reconcile block이 서로 분리되어 보인다.

현재 `Tracker.Resolve`는 active block들을 row별 transaction으로 해제한 뒤 memory/gate를 clear한다. 또한 exact active journal cause 전체나 authoritative stable re-query 증거를 capability에 묶지 않는다. 따라서 a052에서 이 command를 UI에 노출하지 않는다.

## Goals / Non-Goals

**Goals:**

- 저장 config와 running engine snapshot을 구분한 desired/effective adoption read model을 만든다.
- position 행이 provenance, candidate designation과 reconcile scope를 합쳐 같은 stable enum/문구를 사용하게 한다.
- raw exit-state evidence와 canonical effective snapshot을 분리해, 기존 원장 기준선 정보는 보이되 실효값을 추측하지 않는다.
- US include-only 보유분의 정상 fold→adopt→exit t0 경로와 reconcile-blocked 경로를 회귀 테스트한다.

**Non-Goals:**

- 영구 대사 차단 자동/수동 해제 surface, journal mutation 또는 capability 발급.
- baseline/snapshot 재계산, legacy snapshot 백필, journal schema 변경.
- LIVE 주문, 운영 toggle, lane/gate/kill-switch mutation.
- official price transport 자체에 새 market 필드를 추가하거나 broker API contract를 변경하는 일. a052는 existing quote currency로 market identity를 fail-closed 검증한다.

## Decisions

### 1. engine-owned `ManagementRuntime`가 effective와 adoption-blocking tracker projection의 권위다

`positionpolicy`에 transport-neutral `AdoptionSettings`, `ManagementRuntime`, `ReconcileBlock` DTO를 둔다. engine command service가 startup 때 로드한 `config.Adoption`과 adoption driver의 `blocked()`가 사용하는 동일 `reconcile.Tracker.Blocks()` projection을 투영한다. console/API의 config seam은 file desired만 제공한다. runtime GET이 unavailable이면 desired는 유지하고 effective-known=false로 표시한다.

기존 lifecycle command endpoint는 engine process의 loopback에 그대로 남긴다. Compose의 console과 HTTP API는 서로 다른 network namespace이므로 API가 loopback descriptor를 재사용해서는 안 된다. engine은 공유 engine directory 아래에 별도 authenticated Unix-domain runtime read endpoint를 게시하며, 이 endpoint는 `GET runtime`만 제공하고 Preview/Apply 또는 reconcile route를 등록하지 않는다. HTTP API는 mutation method가 없는 narrow reader로 이 endpoint만 연다.

tracker projector는 account/market/symbol scope, reason/detail을 sanitized DTO로 변환한다. 이는 모든 journal reconcile cause의 목록이 아니라 실제 adoption을 보류하는 runtime projection임을 source field로 명시한다. 이 read는 journal을 쓰거나 release authority를 전달하지 않는다.

### 2. candidate를 먼저 계산하는 순수 우선순위 함수 하나를 사용한다

candidate/status는 running effective settings로 계산한다. runtime unavailable이면 이미 journal evidence로 managed인 행은 `MANAGED`를 유지하지만 나머지 행은 desired 설정을 effective로 위장하지 않고 `UNKNOWN`/`RUNTIME_UNAVAILABLE`로 둔다. runtime known일 때 `candidate=(globalEnabled || included) && !excluded`로 먼저 계산한 뒤 상태는 다음 순서로 결정한다.

| 우선순위 | Stable enum | 기본 표시 |
|---|---|---|
| journal or runtime evidence unavailable | `UNKNOWN` | 관리 여부 불명 |
| already managed lifecycle | `MANAGED` | 엔진 관리 |
| operator-released lifecycle | `UNMANAGED` | 관리 외(운영자 해제) |
| excluded | `EXCLUDED` | 관리 제외 |
| candidate + covering block | `RECONCILE_BLOCKED` | 관리 편입 · 대사 차단으로 대기 |
| candidate | `ADOPTION_PENDING` | 관리 편입 · 편입 예약됨 |
| otherwise | `UNMANAGED` | 관리 외(미편입) |

이미 managed인 행은 장래 candidate 설정인 exclude/include로 강등하지 않는다. operator가 release한 lifecycle은 stored adoption ID가 남아 있어도 현재 exit working set에서 제외되므로 candidate/block보다 먼저 `UNMANAGED`로 판정한다. global OFF·미지정 행은 account block이 있어도 blocked로 오표시하지 않는다. block reason은 typed scope/market/symbol/detail/started-at 필드에서 sanitized display text를 만든다.

### 3. raw exit evidence와 actionable effective line을 분리한다

`operatorview.ExitLineView`는 canonical persisted effective snapshot이 있을 때만 actionable 보호선/다음 익절을 표시한다. `journal.ExitState`의 t0 entry, initial stop, baseline, high-water가 존재하지만 canonical snapshot이 없으면 별도 `StoredExitEvidence`로 표시하며 `원장 기록 · 실효 미확인`이라고 이름 붙인다. raw baseline을 current effective protection으로 복사하지 않는다.

### 4. web/API는 같은 projector를 사용한다

console과 a051 adapter는 동일한 transport-neutral projector와 runtime GET을 사용한다. 현재 lifecycle은 console의 engine-owned List와 API의 SELECT-only journal projection에서 읽어 `RELEASED`를 raw entry/adoption eligibility와 구분한다. `/api/v1/positions`는 stable status, `statusKnown`, typed reason, included/excluded/candidate와 각 boolean의 `designationKnown`, optional covering block을 반환한다. runtime unavailable이면 booleans는 false zero-value라도 `designationKnown=false`라서 실제 false로 해석되지 않는다. `/api/v1/optimization`은 desired/effective/known과 include/exclude를 반환한다. API mutation allowlist는 변경하지 않는다.

### 5. 미국 시장 회귀와 quote identity는 실제 engine boundary에서 검증한다

US holding + include-only + fresh USD US quote의 RunOnce가 adoption, exit_state t0와 external-adoption provenance를 생성하는 테스트를 추가한다. adoption quote mapping은 candidate market의 expected currency(KR→KRW, US→USD)를 만족하는 quote만 사용한다. currency가 비었거나 다르면 fail closed로 연기하며, 같은 symbol이 서로 다른 market candidate에 동시에 나타나면 market identity를 증명할 수 없으므로 양쪽 모두 연기한다. 별도 테스트는 account-wide permanent quantity mismatch에서 price read/adoption transaction이 생성되지 않고 projector가 `RECONCILE_BLOCKED`로 설명하는지 검증한다.

## Risks / Trade-offs

- [영구 차단이 배포 후에도 남음] → 화면은 정확한 원인을 말한다. 증거 없는 자동/수동 해제보다 안전하며 atomic resolution은 별도 change로 격리한다.
- [config desired와 engine effective가 재시작 전 다름] → 두 값을 분리하고 effective unknown을 기본값으로 대체하지 않는다.
- [raw baseline 오해] → 별도 evidence 영역과 `실효 미확인` 문구를 강제하고 actionable field에는 계속 em dash를 둔다.
- [sidecar가 loopback engine endpoint에 닿지 못함] → command loopback은 유지하고 공유 private directory의 별도 authenticated Unix read endpoint만 API에 배선한다. 이 endpoint에는 mutation route가 없다.
- [released lifecycle을 adoption ID만으로 managed로 오판] → authoritative lifecycle state를 별도 읽고 `OPERATOR_RELEASED`를 managed/candidate보다 먼저 판정한다.
- [official quote의 market 필드가 비어 있음] → existing currency provenance를 최소 market identity로 사용하고 empty/mismatch/duplicate-market symbol은 fail closed로 연기한다.

## Migration Plan

1. additive DTO/RPC/read projection과 테스트를 배포한다. DB migration은 없다.
2. console/httpapi/engine 이미지를 함께 재시작해 같은 binary contract를 사용한다.
3. 배포 canary에서 desired/effective, 미국 행 reconcile-blocked, raw evidence/effective unknown을 확인한다.
4. 운영 대사 상태는 변경하지 않는다.
5. rollback은 이전 image로 되돌리며 journal/config에는 쓰기가 없으므로 data rollback은 없다.

## Open Questions

없음. 대사 차단 해제는 a052 범위 밖이며 authoritative 재조회 증거와 atomic release가 설계되기 전에는 제공하지 않는다.

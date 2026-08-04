# Function Logic Map: `ExitObserver.workingSet`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a074-critical-events-reach-the-operator/base-commit.txt`
- 위험 등급: **High-risk** — 이 함수가 반환하는 집합이 "이번 사이클에 손절 평가를 받는
  포지션"의 전부다. 빠지면 그 포지션은 보호되지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Journal.Positions(accountRef)` | 계좌의 포지션 전체 | 원장 | error → 반환, 사이클 실패 |
| `Journal.OpenExitStateResults` | 열린 exit state | 원장 | error → 반환, 사이클 실패 |
| `p.State`/`p.Quantity` | CLOSED·0수량은 제외 | 원장 | — |
| `p.ExitEligible()` | false면 unmanaged 알림 후 skip | 원장 | — |
| `result.Corruption` | non-nil이면 `stored_snapshot_corrupt` 격리 | 원장 read | 격리 실패는 `cycle.Err` |
| `ActiveExitSnapshotQuarantine` | 활성 격리면 판정 거부 | 원장 | error → `cycle.Err` |
| `managedPolicyIdentity` | error면 `legacy_policy_identity_unknown` 격리 | 상태 + adoption | 격리 실패는 `cycle.Err` |

**불변식 1 (유지)**: 격리된 세대는 `refused`로 분류되어 판정을 받지 않는다. 이것은
fail-closed이며 a074가 **완화하지 않는다.**

**불변식 2 (유지)**: `return append(out, refused...)` — 격리 행이 마지막에 온다. 함수
주석이 이유를 적고 있다: "alertRefused may synchronously retry a publisher, so every
valid position—including an emergency breach—is recorded/armed/submitted before alert
delivery can wait." **a074가 publisher를 실제로 배선하므로 이 순서가 처음으로 의미를
갖게 된다.** 순서를 바꾸지 않는 것이 이 change의 §0.3 조건이다.

**a074가 바꾸는 것**: 격리를 만드는 두 경로(B11 `stored_snapshot_corrupt`, B18
`legacy_policy_identity_unknown`)에서 격리 생성 이벤트를 발행한다. 발행은 격리 행이
`refused`에 append된 **뒤**, 즉 분류가 끝난 뒤에 한다.

**a074가 바꾸지 않는 것**: 격리 조건, 분류 규칙, 반환 순서, `cycle.Err` 설정 위치.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (448) | Positions 실패 | 없음 | `nil, err` | 기존 |
| B2 (452) | OpenExitStateResults 실패 | 없음 | `nil, err` | 기존 |
| B3 (456) | state 결과 인덱싱 | `byPosition` | — | 기존 |
| B4 (461) | 포지션 순회 | — | — | 기존 |
| B5 (462) | CLOSED 또는 0수량 | 없음 | skip | 기존 |
| B6 (465) | `!ExitEligible()` | `cycle.Unmanaged++`, unmanaged 알림 | skip | 기존 |
| B7 (476) | 열린 state 없음 | `openState` | — | 기존 |
| B8 (478) | openState 실패 | `cycle.Err` | skip | 기존 |
| B9 (479) | `cycle.Err == nil` | `cycle.Err = err` | — | 기존 |
| B10 (484) | 완결된 state | 없음 | skip | 기존 |
| B11 (493) | `result.Corruption != nil` | **격리 생성** + `refused` append | skip | **3.3** |
| B12 (496) | 격리 write 실패 | `cycle.Err` | skip | 기존 |
| B13 (497) | `cycle.Err == nil` | `cycle.Err = qerr` | — | 기존 |
| B14 (506) | 활성 격리 read 실패 | `cycle.Err` | skip | 기존 |
| B15 (511) | 활성 격리 있음 | `refused` append | skip | **3.7** (재발행 금지) |
| B16 (507) | `cycle.Err == nil` | `cycle.Err = qerr` | — | 기존 |
| B17 (511) | else-if 본문 | `refused` append | skip | 기존 |
| B18 (517) | identity 해석 실패 | **격리 생성** + `refused` append | skip | **3.4** |
| B19 (520) | 격리 write 실패 | `cycle.Err` | skip | 기존 |
| B20 (521) | `cycle.Err == nil` | `cycle.Err = qerr` | — | 기존 |

B11과 B18의 성공 경로에만 이벤트 발행이 더해진다. **B15(이미 활성인 격리)에는 더하지
않는다** — 그것은 생성이 아니라 관측이며, 매 사이클 반복된다. B15에 발행을 넣으면
in-process latch가 있더라도 재기동마다 이미 알린 격리를 다시 알리게 된다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.Positions` | 보유 목록 | error → 사이클 중단 | AST |
| `Journal.OpenExitStateResults` | 열린 state | error → 사이클 중단 | AST |
| `Journal.QuarantineExitSnapshot` | 격리 생성 | 이미 활성이면 **기존 행을 반환**한다 (`exit_snapshot.go:712-716`) | AST + 소스 |
| `Journal.ActiveExitSnapshotQuarantine` | 활성 격리 확인 | error → 사이클 중단 | AST |
| `managedPolicyIdentity` | 정책 정체성 | error → 격리 | AST |
| `o.alertUnmanaged` | 미관리 보유 알림 | 포지션당 latch | AST |
| `o.alertQuarantined` (신규) | 격리 생성 알림 | `positionID\|gen\|version` latch | 신규 |

**`QuarantineExitSnapshot`이 기존 행을 반환한다**는 사실이 latch가 필요한 이유다.
B11은 활성 격리 확인(B14) **앞**에 있으므로, 이미 격리된 corrupt 포지션은 매 사이클
같은 행을 돌려받는다. version 포함 키의 latch가 그 반복을 흡수한다.

## State mutations and fallbacks

- `cycle.Unmanaged`, `cycle.Opened`, `cycle.Err`: 그대로.
- `o.unmanaged` map: 그대로.
- 신규 `o.quarantineAlerted` map: 알림 중복만 막는다. **분류에 영향을 주지 않는다** —
  latch가 걸려 있어도 포지션은 여전히 `refused`로 분류된다.
- fallback 없음. 알림 실패는 `o.alert`가 로그로 흡수한다(`exitloop.go:1444-1451`).

## Safety conclusion

- Safe edit boundary: B11·B18의 성공 경로에서 `refused` append 직전/직후의 알림 호출
  한 줄씩. 조건식·append 대상·`continue` 위치는 손대지 않는다.
- High-risk impact: **yes** — 다만 편집은 관측 추가뿐이고 **어떤 포지션도 집합에서
  빼거나 넣지 않는다.** 반환값의 내용은 편집 전후로 동일하다.
- §0.3: 새 호출은 `o.alert` → `Notifier.Notify`이고, 격리된 포지션에 대해서만 실행된다.
  격리된 포지션은 정의상 이번 사이클에 청산 발의를 하지 않는다. 그리고 이 함수는
  판정 이전 단계이므로, 유효 포지션의 발의는 아직 시작도 하지 않았다 — 함수 주석의
  append 순서 결정이 지키려는 것이 정확히 이것이고 a074는 그 순서를 유지한다.
- §0.7: 격리를 자동 해제하지 않는다. 알림만 한다.

# Function Logic Map: `scanExitStateResult`

- Source: `internal/journal/exit_snapshot.go` (L574-702)
- AST evidence: `ast.json` — 분기 22, return 12
- Risk scan: `risk-pattern-report.md`

## 이 함수가 대상인 이유

a100은 exit 상태 행에 보호 컬럼을 추가한다(design D4). 이 함수가 그 행을 읽는 **단일 지점**이고
(`ExitState`·`OpenExitStates` 둘 다 `scanExitState` → 이 함수), 컬럼이 늘면 `row.Scan`의 인자
목록이 바뀐다. **기존 함수 내부를 편집**하므로 편집 전에 이 산출물이 필요하다.

## 이 함수는 스캐너가 아니라 **부패 판정기**다

22개 분기 중 스캔 자체는 2개(B1·B2)뿐이고, 나머지 20개는 **행이 온전한지 판정**한다.
잘못 건드리면 멀쩡한 행이 부패로 읽힌다.

판정은 세 층이다.

| 층 | 코드 | 무엇을 결정하나 |
|---|---|---|
| L1 legacy 판별 | `v10Evidence` 17개 + `anyV10` (L620-628) | 이 행이 v10 이전인가 |
| L2 튜플 완결성 | B14·B15·B16·B17·B18·B20 (L645-680) | v10 튜플이 부분적으로 쓰였는가 |
| L3 평탄화 일치 | B22 (L689-695) | 평탄화 컬럼이 저장된 스냅샷 JSON과 같은가 |

## a100이 반드시 지켜야 할 것 — 보호 컬럼은 세 판정 중 **어디에도** 들어가면 안 된다

**(1) `v10Evidence`에 넣으면 안 된다.** 그 리스트(L620-623)는 "이 행이 v10 스냅샷 튜플을
가졌는가"의 증거다. 보호 컬럼이 채워진 것은 v10 스냅샷의 증거가 아니다. 넣으면 v10 이전
행에 보호를 설치하는 순간 `anyV10`이 참이 되고, 그 행은 legacy 경로(B9→L643)를 벗어나
L2 완결성 검사로 떨어져 **`partial_snapshot_tuple`로 부패 판정**된다.

**(2) `full` 판정(L674-676)에 넣으면 안 된다.** 보호가 없는 정상 포지션이 즉시
`partial_evaluated_tuple`이 된다.

**(3) 평탄화 비교(L689-695)에 넣으면 안 된다.** 그 비교는 저장된 스냅샷 JSON과 컬럼을
대조한다. 보호 컬럼은 그 JSON에 없으므로, 넣는 순간 **모든 기존 행이
`flattened_snapshot_mismatch`가 된다.**

세 경우 모두 결과가 같다 — **exit 스냅샷이 부패로 판정되면 그 포지션의 exit 정책이 멈춘다.**
보호를 설치하려다 손절을 끄는 것이므로 안전 불변식 §0-4에 직접 걸린다.

⇒ **보호 컬럼은 `row.Scan` 인자에 추가되고 `ExitState` 필드에 담기되, 세 판정 리스트에는
추가되지 않는다.** 이것이 spec의 "additive이고 nullable이며 기존 컬럼의 의미를 바꾸지
않는다"가 이 함수에서 갖는 구체적 의미다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `row` | `exitStateSelect`(apply_hook.go:580-587) 순서와 **정확히** 일치 | SELECT 상수 | 개수·순서 불일치 = B2 스캔 에러 |
| `rung` | NULL 또는 rung index | `active_rung` | NULL이면 `NoRung`(B3) |
| `s.LifecycleGeneration` | ≥1 | `coalesce(lifecycle_generation,1)` | <1이면 부패(B4) |
| `effectiveJSON` | 저장된 스냅샷 JSON | `effective_snapshot_json` | 디코드 실패 = 부패(B21) |

**불변식 1 — SELECT 상수와 Scan 인자는 한 쌍이다.** 컬럼을 추가하려면 두 곳을 같은 순서로
바꿔야 한다. 순서가 어긋나면 컴파일은 되고 **값이 서로 다른 필드로 들어간다.**

**불변식 2 — 부패는 에러가 아니라 값이다.** L2·L3 판정은 `error`를 반환하지 않고
`result.Corruption`에 담아 `nil` 에러로 반환한다(B14·B15·B16·B18·B20·B21·B22 전부
`return result, nil`). 호출자가 `Corruption`을 보지 않으면 **부패한 행을 정상으로 쓴다.**

## Branches and early returns

| Branch | 조건 (L) | 결과 | Required test |
|---|---|---|---|
| B1 | `sql.ErrNoRows` (590) | `ErrExitStateNotFound` | 없는 포지션 |
| B2 | 스캔 에러 (593) | wrap된 err | 컬럼 불일치 |
| B3 | `rung.Valid` (597) | `ActiveRung` 설정 | LADDER 행 |
| B4 | `LifecycleGeneration < 1` (601) | `ErrExitSnapshotCorrupt` | 손상된 generation |
| B5 | `policyID.Valid` (606) | `PolicyID` 설정 | 정책 있음 |
| B6 | `generation.Valid` (609) | `PositionGeneration` 설정 | v10 행 |
| B7 | `status.Valid` (612) | `SnapshotStatus` 설정 | v10 행 |
| B8 | `range v10Evidence` (625) | `anyV10` 누적 | 항상 |
| B9 | `!anyV10` (628) | legacy 경로 | **v10 이전 행** |
| B10 | RATCHET 또는 비-RUNNER ladder (633) | legacy 정체성 해석 시도 | legacy RATCHET |
| B11 | else (640) | `legacy_adoption_context_required` | legacy RUNNER ladder |
| B12 | `identityErr == nil` (635) | 정체성 채택 | 해석 성공 |
| B13 | else (637) | `legacy_policy_identity_unknown` | 해석 실패 |
| B14 | `!status.Valid` (645) | `partial_snapshot_tuple` 부패 | 부분 기록 |
| B15 | policy 튜플 부분 (651) | `partial_policy_tuple` 부패 | 부분 기록 |
| B16 | `identity.Validate()` 실패 (657) | `invalid_policy_identity` 부패 | 잘못된 정체성 |
| B17 | `status == Seed` (663) | seed 경로 | 평가 전 행 |
| B18 | seed인데 평가 증거 있음 (664) | `partial_seed_tuple` 부패 | 모순 행 |
| B19 | else (669) | `not_evaluated_yet` | 정상 seed |
| B20 | `!full` (677) | `partial_evaluated_tuple` 부패 | 부분 평가 |
| B21 | 스냅샷 디코드 실패 (683) | `invalid_effective_snapshot` 부패 | 깨진 JSON |
| B22 | 평탄화 불일치 (689) | `flattened_snapshot_mismatch` 부패 | 컬럼/JSON 어긋남 |

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `row.Scan` | 컬럼 → 필드 | 에러 = B1(no rows) 또는 B2(wrap) | AST L584 |
| `legacyPolicyIdentity` | v10 이전 행의 정책 정체성 해석 | 실패는 에러가 아니라 `UnknownReason`(B13) | AST L634 |
| `identity.Validate` | 정책 정체성 검증 | 실패 = `Corruption` 값(B16) | AST L657 |
| `decodeStoredSnapshot` | 저장된 스냅샷 JSON 디코드 | 실패 = `Corruption` 값(B21) | AST L682 |

호출부는 `scanExitState` 하나뿐이고, 그것을 `Journal.ExitState`와 `Journal.OpenExitStates`가
쓴다(apply_hook.go:601-644). **exit 상태를 읽는 모든 경로가 이 함수를 지난다.**

## State mutations and fallbacks

- 이 함수는 **DB를 쓰지 않는다.** 그러나 `result.State.Snapshot.UnknownReason`과
  `result.Corruption`을 채워 **호출자의 판단을 결정한다.**
- **fallback의 방향이 두 가지다.** legacy 경로(B9~B13)는 디스크를 건드리지 않고 메모리에서만
  정체성을 해석한다(주석 L630-632). 부패 경로(B14~B22)는 값을 채우고 `nil` 에러로 반환한다.
  ⇒ **부패는 조용하다.** 호출자가 `Corruption`을 확인하지 않으면 아무 일도 일어나지 않는다.
- 보호 컬럼이 추가되면 그 값도 `s`에 담기고 `result.State`로 나간다. **`UnknownReason`이나
  `Corruption`에는 영향을 주지 않아야 한다** — 위의 세 판정 리스트를 건드리지 않는다는 규칙이
  곧 그 뜻이다.

## Safety conclusion

- Safe edit boundary: `row.Scan` 인자 추가 + `ExitState` 필드 추가 + `exitStateSelect` 상수
  갱신. **세 판정 리스트(`v10Evidence`, `full`, 평탄화 비교)는 건드리지 않는다.**
- High-risk impact: **yes.** 이 함수의 오판은 exit 정책 정지로 이어지고, 그것은 손절 정지다.
- **RED 테스트 의무:** 보호 컬럼이 채워진 v10 이전 행이 여전히 `legacy_snapshot_absent`
  경로(B9)로 가는지, 보호 컬럼이 NULL인 정상 v10 행이 여전히 부패 없이 읽히는지 —
  두 방향 모두 편집 전에 실패하는 테스트로 고정한다.

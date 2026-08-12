# Branch Test Map: `scanExitStateResult`

> **측정 방법**: `go test -covermode=set -coverprofile ./internal/journal`(패키지 75.0%).
> 분기의 *조건*이 아니라 **분기 본문 블록의 실행 여부**를 측정했다. 조건이 여러 줄인 두 분기
> (L664·L689)는 본문 블록이 조건 마지막 줄에서 시작하므로 그 블록으로 판정했다
> (L666.82 count=1, L695.127 count=0).
> 측정일 2026-08-11, source SHA-256 `0e376d1ca4f6e29b…`(`ast.json`).

| Branch | Scenario | Test | 본문 실행됨 | 비고 |
|---|---|---|---|---|
| B1 | `sql.ErrNoRows` → `ErrExitStateNotFound` | — | **no** (L590) | 없는 포지션 |
| B2 | 스캔 에러 → wrap | — | **no** (L593) | **컬럼 불일치가 나오는 곳** |
| B3 | `rung.Valid` → `ActiveRung` | `internal/journal` 패키지 | **yes** (L597) | LADDER |
| B4 | `LifecycleGeneration < 1` → 부패 | — | **no** (L601) | 손상된 generation |
| B5 | `policyID.Valid` | 동상 | **yes** (L606) | |
| B6 | `generation.Valid` | 동상 | **yes** (L609) | |
| B7 | `status.Valid` | 동상 | **yes** (L612) | |
| B8 | `v10Evidence` 순회 | 동상 | **yes** (L625) | **a100이 건드리면 안 되는 판정 1** |
| B9 | `!anyV10` → legacy 경로 | 동상 | **yes** (L628) | v10 이전 행 |
| B10 | legacy RATCHET·비-RUNNER ladder | 동상 | **yes** (L633) | |
| B11 | legacy RUNNER ladder → 편입 문맥 필요 | — | **no** (L640) | |
| B12 | legacy 정체성 해석 성공 | 동상 | **yes** (L635) | |
| B13 | legacy 정체성 해석 실패 | — | **no** (L637) | |
| B14 | `!status.Valid` → `partial_snapshot_tuple` | 동상 | **yes** (L645) | **legacy 오판이 떨어지는 곳** |
| B15 | policy 튜플 부분 → 부패 | 동상 | **yes** (L651) | |
| B16 | 정체성 검증 실패 → 부패 | — | **no** (L657) | |
| B17 | `status == Seed` | 동상 | **yes** (L663) | |
| B18 | seed인데 평가 증거 존재 → 부패 | 동상 | **yes** (L666 본문) | |
| B19 | 정상 seed → `not_evaluated_yet` | 동상 | **yes** (L669) | |
| B20 | `!full` → `partial_evaluated_tuple` | 동상 | **yes** (L677) | **a100이 건드리면 안 되는 판정 2** |
| B21 | 스냅샷 디코드 실패 → 부패 | 동상 | **yes** (L683) | |
| B22 | 평탄화 불일치 → 부패 | — | **no** (L695 본문) | **a100이 건드리면 안 되는 판정 3** |

**측정 결과: 22개 중 7개 미실행** (B1, B2, B4, B11, B13, B16, B22).

## 세 판정의 커버리지가 서로 다르다 — 그리고 그 차이가 위험의 크기다

| 판정 | 잘못 건드렸을 때 | 기존 테스트가 잡아 주는가 |
|---|---|---|
| `v10Evidence`(B8) → `!anyV10`(B9) | v10 이전 행이 legacy 경로를 벗어나 **B14로 떨어져 부패 판정** | **잡는다.** B9·B14 둘 다 실행되므로 legacy 행이 부패로 바뀌면 기존 테스트가 깨진다 |
| `full`(B20) | 보호 없는 정상 행이 `partial_evaluated_tuple` | **잡는다.** B20이 실행되는 경로가 있다 |
| 평탄화 비교(B22) | **모든 기존 행이 `flattened_snapshot_mismatch`** | **부분적으로만.** B22의 true 본문은 한 번도 실행된 적이 없다 — 즉 **불일치 검출 자체를 검증한 테스트가 없다.** 다만 조건 평가 블록(L688.2-695.127)은 실행되므로, 정상 행이 불일치로 바뀌면 그 테스트들이 `Corruption != nil`로 깨진다 |

⇒ **세 판정 모두 「잘못 건드리면 기존 테스트가 깨진다」가 성립한다.** 이것이 이 편집의 안전
근거다. 다만 B22는 **깨지는 방식으로만** 보호되고 **의도한 검출을 확인하는 테스트는 없다.**

## a100의 RED 대상

- **스캔 에러(L593) — RED 필수.** 컬럼 추가가 만들 수 있는 유일한 새 실패다. 개수·순서
  불일치가 부분 읽기가 아니라 에러로 나오는지 고정한다.
- **legacy 경로(L628) — RED 필수.** **보호 컬럼이 채워진 v10 이전 행**이 여전히
  `legacy_snapshot_absent`로 가는지. 이 테스트가 「`v10Evidence`에 넣지 않았다」의 기계적 증거다.
- **`full` 판정(L677) — RED 필수.** 보호 컬럼이 NULL인 정상 v10 행이 부패 없이 읽히는지.
- **평탄화 비교(L695) — RED 필수.** 보호 컬럼이 채워진 정상 행이 `Corruption == nil`인지.
  **미실행 분기이므로 a100이 처음으로 이 판정에 테스트를 붙인다.**
- **B1·B4·B11·B13·B16 — a100 범위 밖.** 도달 조건이 달라지지 않는다.

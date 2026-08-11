# Branch Test Map: `runInterlock`

> **측정 방법**: `go test -covermode=set -coverprofile` (패키지 전체). 분기의 *조건*이 아니라
> **true 결과 본문의 실행 여부**를 측정했다. 조건 statement가 covered인 것은 조건이 평가됐다는
> 뜻일 뿐 그 분기를 탔다는 뜻이 아니므로, 본문 행을 따로 측정했다.

| Branch | Scenario | Test | true 결과 실행됨 | 비고 |
|---|---|---|---|---|
| B1 | 게이트 OFF면 검증도 허용도 없다 | `internal/app/engine` (435 tests) | **yes** (L383) | |
| B2 | `verifyGate` 실패 시 기동 거부 | 동상 | **yes** (L389) | |
| B3 | **수락 audit 기록 실패 시 기동 실패** | — | **NO (미실행)** | **공백** |
| fall-through | 검증 통과 → `EntryPermitted = (Protection==Wired)` | `guardian_test.go:131`, `interlock_entry_test.go:70` | **yes** (L406) | 두 테스트 모두 `EntryPermitted == false`를 단언한다 |

## 발견된 공백

**B3(audit 기록 실패 경로)이 한 번도 실행되지 않는다.** 운영 설정 변경이 audit로 추적 가능해야
한다는 안전 불변식(§0-5)의 실패 방향이 미검증이라는 뜻이다. a100이 `Protection`을 `WIRED`로
만들 수 있게 되면 이 경로의 의미가 커진다 — 진입이 허용된 사실이 기록되지 못했는데 기동이
계속되면 추적 불가능한 자동매매가 된다. **a100 tasks에 RED 테스트로 포함한다.**

`guardian_test.go:132`와 `interlock_entry_test.go:71`은 현재 `EntryPermitted == true`를 **실패로**
단언한다. a100이 `WIRED`를 생산 가능하게 만들면 이 두 단언의 전제가 바뀌므로 같은 change에서
함께 갱신해야 한다.

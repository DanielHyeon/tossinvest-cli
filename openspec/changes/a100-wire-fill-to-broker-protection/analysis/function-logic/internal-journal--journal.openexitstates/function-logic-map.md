# Function Logic Map: `Journal.OpenExitStates`

- Source: `internal/journal/apply_hook.go` (L621-644)
- AST evidence: `ast.json` — 분기 4, return 4
- Risk scan: `risk-pattern-report.md`

## 이 함수가 대상인 이유

수렴 워커의 입력이 여기서 나온다(design D2 — "exit 관리 상태가 열려 있고 유효한 stop을 가진
모든 포지션"). a100은 이 함수가 반환하는 행에 보호 상태를 실어야 하므로 **편집 대상**이다.

## 이 함수는 두 가지를 겸한다 — 그리고 그것이 의도다

주석(L614-620)이 명시한다.

> It is two things at once, and **deliberately not two functions.** It is the observation loop's
> working set (task 7.4), and it is the crash restore … A separate "restore" query would be a
> second definition of what is outstanding, and the two could disagree.

⇒ **수렴 워커를 위해 별도의 조회 함수를 새로 만들면 그 경고에 정면으로 걸린다.** "미보호
포지션 목록"이 이 함수와 다른 정의를 갖게 되고, 두 정의는 언젠가 어긋난다.
워커는 **이 함수를 쓰거나, 이 함수를 확장한다.**

## 대상 집합을 정하는 것은 Go 코드가 아니라 SQL이다

분기 4개는 전부 에러·순회이고 **필터가 없다.** 대상 집합은 L623-626의 SQL이 정한다.

| 조건 | 출처 | 의미 |
|---|---|---|
| `p.account_ref = ?` | 인자 | 계좌 한정 |
| `e.completed = 0` | `exit_states` | 아직 끝나지 않은 exit |
| `currentManagedExitLifecycle` | apply_hook.go:589-599 | lifecycle이 없거나, **현재 세대가 MANAGED이고 더 새 세대가 없을 때만** |

세 번째가 중요하다. 편입 세대가 갱신되면 이전 세대의 exit 상태는 이 목록에서 빠진다.
⇒ **수렴 워커가 이 목록만 보면, 재편입 중인 포지션은 잠시 대상에서 사라진다.** 그 순간
브로커에는 이전 세대가 등록한 상주 주문이 남아 있을 수 있다. 목록에서 사라진 것과
「보호가 필요 없다」는 같지 않다.

⇒ 워커의 취소·정리 판정은 **이 목록의 부재만으로 내리면 안 된다.** 브로커측 관측을 함께
본다(design D2의 「flat이 되면 남은 상주 주문을 취소한다」가 이 이유로 필요하다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` | 공백 제거된 계좌 참조 | 호출자 | 빈 문자열이면 0행 |
| `exitStateSelect` | 컬럼 목록 상수 | apply_hook.go:580 | `scanExitState`와 한 쌍 |
| `positions` join | `p.id = e.position_id` | 스키마 | 고아 exit 상태는 나오지 않는다 |

**불변식 — 정렬은 `market, symbol, instance_seq`다**(L626). 워커가 이 순서에 의존하면
같은 순서로 브로커를 호출하게 되므로 시장별 rate limit 소모가 편중될 수 있다.

## Branches and early returns

| Branch | 조건 (L) | 결과 | Required test |
|---|---|---|---|
| B1 | 쿼리 실패 (627) | wrap된 err | DB 오류 |
| B2 | `rows.Next()` (633) | 행 순회 | 정상 |
| B3 | `scanExitState` 실패 (635) | 첫 실패에서 **전체 중단** | 부패 행 |
| B4 | `rows.Err()` (640) | wrap된 err | 순회 중 오류 |
| — | 정상 (643) | `[]ExitState, nil` | 정상 |

**B3이 위험 지점이다.** 한 행의 스캔 실패가 **나머지 포지션 전부를 반환하지 못하게 한다.**
`scanExitStateResult`의 부패 판정은 에러가 아니라 값이므로 여기까지 오지 않지만(별도 산출물
참조), **스캔 자체의 실패**(컬럼 개수 불일치 등)는 온다. 보호 컬럼 추가가 `Scan` 인자와
어긋나면 **모든 포지션의 exit 관측이 한꺼번에 멈춘다.**

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `j.db.QueryContext` | 열린 exit 상태 조회 | 에러 = wrap 후 반환(B1) | AST L623 |
| `scanExitState` | 행 → `ExitState` | 첫 실패에서 **전체 중단**(B3) | AST L634, 별도 산출물 |
| `rows.Err()` | 순회 후 오류 확인 | 에러 = wrap 후 반환(B4) | AST L640 |
| `rows.Close()` | defer 해제 | — | AST L630 |

호출부: exit 관측 루프의 working set이자 크래시 복원 경로(주석 L614-620). **a100의 수렴
워커가 세 번째 호출부가 된다.**

## State mutations and fallbacks

- 이 함수는 **읽기 전용이다.** 상태를 바꾸지 않는다.
- fallback 없음 — 실패는 부분 목록이 아니라 에러다. **부분 결과를 반환하지 않는 것이 이
  함수의 계약이다**(B3에서 `out`을 버리고 에러만 반환한다).
- 정렬은 `market, symbol, instance_seq` 고정. 호출자가 순서를 바꾸려면 자기 쪽에서 한다.

## Safety conclusion

- Safe edit boundary: SELECT 상수와 `scanExitState`를 같은 순서로 갱신. WHERE 절은 바꾸지
  않는다 — 바꾸면 exit 관측 루프의 working set이 함께 바뀐다.
- High-risk impact: **yes.** 이 함수의 실패는 exit 관측 루프의 실패이고, 그것은 손절 정지다.
- **워커의 대상 집합은 이 함수 + 브로커 관측이다.** 이 함수만으로는 「보호를 지워야 할
  포지션」을 알 수 없다.

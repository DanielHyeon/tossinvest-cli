# Function Logic Map: `TestProjectionLivenessClausesEachDecideOnTheirOwn`

- Source: `internal/strategyprojectionrpc/a108_publication_is_total_test.go`
- AST evidence: `ast.json` — 구현 후(:460–519, 분기 6). base `54004f44` 에서는 행 5개(분기 5)였다.
- Risk scan: `risk-pattern-report.md`

a108 의 순수 probe 판정 표다. a113 이 **계약이 바뀐 행 하나를 교체하고 하나를 더한다**: 편집 전
"owner 쓰기 비트가 없다"(죽은 0500 → false)는 삭제된 추정 절을 고정하던 행이다. 그 사실(죽은
0500 잔재는 회수된다)은 이제 회수 전용 probe 의 표(`TestTheStaleProbeAsksInsteadOfGuessing`)와
`TestStartRecoversFromUnwritableSocketLeftover` 가 지킨다. 순수 probe 에서는 같은 디스크 상태가
EACCES = 생존이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 표 행 | 6행(부재·거부·수락·산 0400·죽은 0500·ENOTDIR) | 이 함수 | 행마다 `t.Errorf` |
| euid | 비root 에서만 EACCES 행 성립 | `os.Geteuid` | root 면 두 행 `t.Skip` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 표 순회 | 행마다 하위 테스트 | — | 자기 자신 |
| B2 | 산 0400 행: root | 없음 | `t.Skip` | 자기 자신(비root 에서 실행됨) |
| B3 | 산 0400 행: chmod 실패 | 없음 | `t.Fatal` | 자기 자신 |
| B4 | 죽은 0500 행: root | 없음 | `t.Skip` | 자기 자신 |
| B5 | ENOTDIR 행: fixture 쓰기 실패 | 없음 | `t.Fatal` | 자기 자신 |
| B6 | 판정 ≠ 기대 | 없음 | `t.Errorf` | 자기 자신 — 뮤테이션 N1·N4 가 이 분기로 사망 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `projectionSocketAccepts` | 판정 대상 | — | AST call |
| `a108LiveSocket`·`a108DeadSocketWithMode`·`a108MakeControlDir` | fixture | `t.Fatal` | AST calls |

## State mutations and fallbacks

- 임시 디렉터리(`shortRuntimeDir`) 안에서만 파일을 만든다. umask 를 건드리지 않는다(a108 관례).

## Safety conclusion

- Safe edit boundary: 표의 행 교체 1·추가 1과 머리 주석. 루프·판정 호출 무변경.
- High-risk impact: no — 테스트.

# Function Logic Map: `wtsPopularityReader`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `base`, L495–500, 분기 1개)
- Risk scan: `risk-pattern-report.md`

**이 파일에서는 삭제되었다.** `buildCandidatePanel`과 함께 `cmd/tossctl/candidatepanel.go`로
옮겨졌고, 본문은 **byte 동일**하다. `ast.json`은 base commit에서 뜬 것이다.

nil `*tossclient.Client`를 nil 인터페이스로 바꾼다. non-nil 인터페이스 안의 typed nil은
"WTS 세션 없음"이 세 층 아래에서 nil dereference가 되는 고전적인 경로이고, `Panel`의 계약은
nil reader가 곧 소스 부재라는 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | `*tossclient.Client` 또는 nil | `buildCandidatePanel` | nil이면 nil 인터페이스 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c == nil` | 없음 | `nil` (typed nil이 아니다) / 아니면 `c` | **커버 없음** — 계약의 반대편만 `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly`가 잡는다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음 — 순수 변환, base와 byte 동일.
- fallback 없음. 이 함수가 하는 일이 정확히 fallback을 **막는 것**이다.

## Safety conclusion

- Safe edit boundary: **파일 이동**만. 본문 byte 동일.
- High-risk impact: no (nil 처리). 이 변환이 없으면 typed nil이 `Panel`을 통과해 매 스캔 nil dereference가 된다.

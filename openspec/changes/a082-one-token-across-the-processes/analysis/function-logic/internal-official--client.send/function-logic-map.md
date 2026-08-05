# Function Logic Map: `Client.send`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**High-risk.** 인증된 요청이 전부 이 함수를 지난다. 이 change의 편집은 401 갈래
하나이고, 나머지 여덟 갈래는 손대지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `makeReq` | 토큰을 받아 요청을 만드는 함수. **두 번 이상 호출 가능해야 한다** | 호출하는 verb | 재사용 불가한 body reader를 쓰면 재시도가 깨진다 |
| `tok` | 현재 토큰 | `c.tm.token(ctx)` | 획득 실패는 그대로 올라간다 |
| 불변식 1 | `refresh`에 넘기는 `refused`는 **방금 요청에 실었던 토큰**이다 | 루프가 `tok`을 갱신하며 돈다 | 어기면 `refresh`가 죽은 토큰을 되돌려준다 |
| 불변식 2 | 갱신 시도는 **상한이 있다** (2회) | 루프 조건 | 어기면 401 무한 루프 |
| 불변식 3 | 채택하지 않은(=발급한) 토큰이 거부당하면 더 돌지 않는다 | `if !adopted { break }` | 어기면 죽은 자격증명에 재시도를 낭비한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.tm.token` 실패 | 없음 | 그 오류 | 기존 |
| B2 | `makeReq` 실패 (최초) | 없음 | 그 오류 | 기존 |
| B3 | `doRequest` 실패 (최초) | 없음 | 그 오류 | 기존 |
| B4 | **401 갱신 루프** — 최대 2회 | `c.tm.refresh` (채택 또는 발급) | — | `TestGetUnwrapsEnvelopeAndRetriesOn401`(기존), `TestAnAdoptedTokenThatIsAlsoRefusedStillEndsOnAMintedOne` |
| B5 | 루프 안 `refresh` 실패 | 없음 | 그 오류 | 기존 |
| B6 | 루프 안 `makeReq` 실패 | 없음 | 그 오류 | 기존 |
| B7 | 루프 안 `doRequest` 실패 | 없음 | 그 오류 | 기존 |
| B8 | **채택하지 않았으면 루프 종료** | 없음 | — | `TestGetUnwrapsEnvelopeAndRetriesOn401`(기존, 갱신 1회로 끝난다) |
| B9 | 최종 코드가 2xx 밖 | 없음 | `classifyStatus(code, body)` | `TestRawReadsClassifyErrorsLikeEveryOtherRead` |

**B4와 B8이 편집이다.** base는 `if code == 401 { 한 번 갱신하고 재시도 }`였다.
채택한 토큰은 만료 시각에서 **추론**한 것이지 검증된 것이 아니므로, 그것이 다시
거부당하면 유일한 재시도를 추측에 쓴 셈이 된다 (design D5). 그때만 한 번 더 돈다.

두 번째 회차가 채택으로 끝날 수 없는 이유: `refresh`는 자기가 거부당한 토큰을
절대 반환하지 않으므로(`refresh` 불변식 1), 첫 회차의 추측이 두 번째 회차의
채택 후보에서 배제되고 fallthrough로 간다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.tm.token` | 토큰 획득 | 캐시 우선. 네트워크는 만료 시에만 | AST calls |
| `c.tm.refresh` | 401 응답 | **신규 인자** `tok`, **신규 반환** `adopted` | AST calls |
| `makeReq` | 요청 (재)생성 | 매 시도 새로 만든다 — 소비된 body reader 재사용 금지 | AST calls |
| `c.doRequest` | 전송 | ctx 존중 | AST calls |
| `classifyStatus` | 비-2xx 분류 | 이 change가 인증 거부에 상태 코드를 싣는다 | AST calls |
| `unwrapAndDecode` | `result` 봉투 해제 | — | AST calls |

## State mutations and fallbacks

- `tok`·`req`·`code`·`body`가 루프 안에서 갱신된다. 프로세스 밖 side effect는
  `refresh`가 교환할 때의 파일 쓰기뿐이고 이 함수가 직접 하지 않는다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **401 갱신 구간뿐.** B1·B2·B3의 사전 처리, B9의 분류,
  봉투 해제는 글자 그대로 보존한다.
- 상한 2회는 명시적이어야 한다. 무한 루프는 인증 경로에서 브로커에 대한 부하이고
  `m.mu`를 반복해서 잡는다.
- High-risk impact: **yes.** 인증된 요청 전부가 여기를 지나고, 여기서 새어 나간
  `ErrAuth`는 엔트리 게이트를 잠근다.

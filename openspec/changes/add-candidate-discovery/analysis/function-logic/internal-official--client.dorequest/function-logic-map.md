# Function Logic Map: `Client.doRequest`

- Source: `internal/official/client.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 함수 — 이 change(`add-candidate-discovery`)가 본문에 **한 줄**을 넣었다:

```
c.rates.record(readRateBudget(req.URL.Path, resp.Header, time.Now()))
```

`git diff --stat`으로 `internal/official/client.go`는 `+12 -0`이며, 그 12줄은
구조체 필드 `rates rateBudgets` 4줄(주석 포함), doc 주석 7줄, 본문 1줄이다.
**삭제 0줄.** 반환값(`int, []byte, error`) 셋과 오류 사상은 손대지 않았다.

## 이 한 줄이 더한 것과 더하지 않은 것

- **더한 것**: 응답의 `X-RateLimit-Limit` / `-Remaining` / `-Reset`을 읽어 클라이언트의
  경로별 맵에 기록한다. 그 전까지 이 헤더는 **어디에서도 읽히지 않았다** —
  `doRequest`가 `(status, body)`만 돌려주고 상태는 sentinel로 사상되므로, 예산을 물어볼 수
  있는 시점에는 응답이 이미 사라진 뒤다.
- **더하지 않은 것**: 판단. `rates`는 이 패키지의 로직이 한 번도 읽지 않는다
  (`RateBudget`/`RateBudgets` 두 exported 접근자만 읽는다). 기록하지 결정하지 않는다.
- **위치가 요점**: `record`는 `resp.Body.Close()` defer **직후**, 본문 읽기와
  상태 분류보다 **앞**이고 status에 대한 분기가 **없다**. 그래서 429 응답 —
  이 측정에 가장 정보량이 많고, 기존 오류 사상이 정확히 버리던 응답 — 도 기록된다.
- **키**: `budgetKey`가 식별자 세그먼트를 `{id}`로 접는다. 그래서 취소·정정처럼 경로에 주문
  번호가 박히는 요청이 장기 실행 프로세스에서 맵을 무한히 키우지 않는다
  (`TestTheBudgetMapDoesNotGrowWithOrderIDs`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req` | `send`가 `makeReq(tok)`로 만든 요청(401 재시도 시 재생성) | `client.go` `send` | — |
| `c.hc` | `http.Client{Timeout: 15s}` 또는 주입값 | `New`/`WithHTTPClient` | 전송 실패 → `ErrTransport` |
| `resp.Header` | 헤더 없음도 정상 | 브로커 | 헤더 0건이면 `Reported=false` — **0 remaining이 아니다** |
| `c.rates` | 경로별 맵, `sync.RWMutex` | `ratebudget.go` | 첫 기록 때 lazily 생성 |

불변식: 반환 계약 무변경. `(0, nil, ErrTransport)` / `(status, nil, ErrTransport)` /
`(status, body, nil)` 세 형태가 base와 같다. 예산 기록은 어떤 반환값에도 영향을 주지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L95) | `c.hc.Do` 실패 | **없음** — 기록 지점 이전이다(응답이 없으니 헤더도 없다) | `0, nil, %w: ErrTransport` | `TestGetNon2xxReturnsClassifiedError` 및 transport 케이스 |
| B2 (if, L101) | `io.ReadAll` 실패 | 예산은 **이미 기록됨**(record가 앞선다) | `resp.StatusCode, nil, ErrTransport` | 동상 |

정상 경로(무분기 꼬리): `record` → 본문 읽기 성공 → `(status, body, nil)`.
`record` 자체에는 분기가 없다 — status와 무관하게 항상 호출된다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.hc.Do` | 실제 전송 | `http.Client.Timeout`; retry는 상위 `send`의 401 1회뿐 | ast.json calls |
| `resp.Body.Close` (defer) | 커넥션 반환 | — | ast.json defers |
| `readRateBudget` (신규 호출) | 헤더 → `RateBudget` | 오류를 만들지 않는다. 헤더 부재는 `Reported=false` | ast.json calls |
| `c.rates.record` (신규 호출) | 경로별 저장 | `sync.Mutex`, 오류 없음 | ast.json calls |
| `io.ReadAll` | 본문 | 실패는 `ErrTransport` | ast.json calls |

`readRateBudget`은 `time.Now()`를 받는다 — reset 헤더를 delta로 읽을 때의 기준이며,
파싱 불가·비현실적 값은 `ResetUnparsed`로 두고 원문을 보존한다
(`TestAnImplausibleResetIsNotPresentedAsADetermination`).

## State mutations and fallbacks

- 새 상태 하나: `c.rates`(경로 → 마지막 `RateBudget`). 프로세스 내 메모리, 디스크·계좌 무관.
- fallback 경로 무변경. `ShouldFallback`의 입력은 여전히 `classifyStatus`의 결과다.
- 이 함수는 주문을 내지 않는다. 주문 경로(POST/DELETE)도 같은 `send`를 지나므로 예산이
  같은 방식으로 기록될 뿐, 그 요청이 만들어지는 조건은 손대지 않았다.

## Safety conclusion

- Safe edit boundary: 본문 1줄(기록) + 필드 1개 + doc 주석. 삭제·재배치 0.
- High-risk impact: **yes** — `doRequest`는 이 패키지의 **모든** HTTP 요청이 지나는 단일
  지점이고, 그 안에는 주문·취소·정정이 포함된다. 여기서 panic·블로킹·오류 사상 변경이
  일어나면 손절·비상 청산의 즉시성에 직접 닿는다.
  그럼에도 안전한 이유: (a) 추가된 호출은 헤더 파싱과 맵 쓰기뿐으로 오류를 만들 수 없고
  panic 경로가 없다(`headerInt`는 `strconv.Atoi` 실패를 `(0,false)`로 흡수), (b) 잠금은
  `rateBudgets`의 자체 mutex이며 이 함수가 다른 잠금을 들고 있지 않다 —
  `ensureAccountSeq`의 `c.mu`는 요청 생성 전에 이미 풀렸다, (c) 맵 크기는 `budgetKey`가
  경로 템플릿으로 접어 상한이 있다, (d) 반환값·오류가 base와 동일해 상위 재시도 행렬이
  보는 것이 전혀 바뀌지 않는다.

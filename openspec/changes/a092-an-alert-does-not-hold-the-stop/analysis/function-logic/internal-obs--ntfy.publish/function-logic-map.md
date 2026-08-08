# Function Logic Map: `Ntfy.Publish`

- Source: `internal/obs/ntfy.go` (85-139)
- AST evidence: `ast.json` — branches 9, returns 5, calls 29, assignments 14,
  **defers 2**(`cancel` `:100`, `resp.Body.Close` `:129`), **go_statements 0**
- Risk scan: `risk-pattern-report.md`

**10초 상한이 여기서 정해진다.** 34초 예산의 곱해지는 항.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Topic` | 비어 있으면 미설정 | 설정 `engine.notifications` | B1 `:87` → `ErrNtfyNotConfigured` |
| `n.BaseURL` | 비면 `DefaultNtfyBaseURL` | 설정 | B2 `:91` |
| **`n.Timeout`** | **0이면 10s** | 조립 — 프로덕션 `&obs.Ntfy{BaseURL, Topic, Token}` (`notifications.go:101`)는 **채우지 않는다** | B3 `:96` → `10 * time.Second` |
| **`n.HTTPClient`** | **nil이면 `&http.Client{Timeout: timeout}`** | 같은 조립 — **채우지 않는다** | B7 `:122` |
| `ctx` | 호출자의 컨텍스트 | `deliver:256` / `publishBestEffort:142` / `Flush:325` | `:99`가 `timeout`으로 **한 번 더** 감싼다 |

**프로덕션에서 두 필드가 모두 비어 있다는 것이 10초의 근거다.** 콘솔 쪽 조립
(`cmd/tossctl/notificationsettings.go:151`)도 같은 세 필드만 채운다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:87` | `topic == ""` | 없음 | `return ErrNtfyNotConfigured` `:88` |
| B2 `:91` | `base == ""` | `base = DefaultNtfyBaseURL` | — |
| B3 `:96` | `timeout <= 0` | **`timeout = 10 * time.Second`** | — |
| B4 `:104` | 요청 생성 오류 | 없음 | `return fmt.Errorf(…)` `:105` |
| B5 `:110` | 제목 있음 | `Title` 헤더 | — |
| B6 `:117` | 토큰 있음 | `Authorization` 헤더 | — |
| B7 `:122` | `client == nil` | **`client = &http.Client{Timeout: timeout}`** | — |
| B8 `:126` | `client.Do` 오류 | 없음 | `return fmt.Errorf("obs: publishing to ntfy: %w", err)` `:127` |
| B9 `:134` | 2xx 아님 | 없음 | `return fmt.Errorf("obs: ntfy refused the message: …")` `:135` |
| — | 성공 | — | `return nil` `:138` |

**이탈 5개. 시간을 쓰는 것은 B8로 가는 경로 하나뿐이다** — `client.Do`(`:125`).

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `context.WithTimeout` `:99` | 요청 기한 | **10s**(B3). defer `cancel` `:100` | AST calls + defers |
| `http.NewRequestWithContext` `:102` | 요청 조립 | 오류는 B4 | AST calls |
| **`client.Do` `:125`** | **원격 왕복** | **상한 10s — ctx와 `http.Client.Timeout` 양쪽**. **재시도 없음** | AST calls |
| `resp.Body.Close` `:129` (**defer**) | 연결 재사용 | — | AST defers |
| `io.ReadAll(io.LimitReader(resp.Body, 4096))` `:132` | 오류 본문 | 상한 4KB | AST calls |

**이 함수는 재시도하지 않는다**(AST branches에 루프 없음 — B1~B9 전부 `if`).
재시도는 호출자(`deliver` B2 `:251`)의 것이다.

## State mutations and fallbacks

- 원격 상태만 바꾼다. 로컬 상태 변경 없음.
- fallback 둘: B2(기본 URL), B3(기본 timeout), B7(기본 client). **셋 다 프로덕션에서
  실제로 발동한다.**
- **goroutine 없음**.

## Safety conclusion

- **Safe edit boundary**: 이 change는 이 함수를 **바꾸지 않는다.** 예산의 근거로만 쓴다.
- **High-risk impact**: no(직접) / **yes**(예산 제공자로서).
- **왜 10초가 실제로 소모될 수 있는가**: `client.Do`는 DNS·TLS·응답 대기를 포함한다.
  연결 거부는 즉시 실패하지만 **응답 없는 서버·패킷 유실·중간 프록시 정체**는 기한까지
  기다린다. `Timeout` 필드가 존재한다는 것 자체가 그 실패 모드를 전제한 설계다.
- **실측 대조 — 무엇이 왕복이고 무엇이 아닌가**:
  건강한 상태의 왕복으로 **유효한** 표본은 §3의 **6건뿐**이고 범위는 **0.1983 ~ 0.7540초**다
  (critical 경로, 짝지은 줄이 모두 인접함을 확인).
  `engine.log`의 연속 `exit.position_unmanaged` 줄 간격에서 얻은 **0.2~1.8초(표본 9)** 는
  **왕복이 아니다** — `analysis/delivery-latency.md` §5.3대로 그 타입은 조립 자리가 4곳이고
  셋이 대사 루프라 줄 간격이 어느 루프의 체류도 재지 못한다.
  **이전 판이 이 값을 "왕복 실측"이라고 부른 것은 틀렸다**(5라운드 H3).
- 10초는 상한이지 통상값이 아니다 — 이 change의 논거는 "항상 느리다"가 아니라
  **"느려질 수 있는데 손절 루프가 그것을 기다린다"**이다.
- **1.3초 상한과의 관계**: 유효 표본의 최댓값은 냉 연결 1건의 0.7540초이고
  1.3초는 그 위에 1.72배다. 무효 표본의 1.836초가 만약 왕복이었다면 1.3초는 작다 —
  그것을 **측정으로 배제한 것이 아니라 생산자 추적으로 배제했다**는 것이 이 유도의
  약한 고리이고, design D2의 재검토 조건(냉 publish가 1.3초를 넘으면 재유도)과
  a090의 계측이 그 고리를 대신 지킨다.

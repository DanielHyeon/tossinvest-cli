# Function Logic Map: `resolveNotificationPublisher`

- Source: `internal/app/engine/notifications.go` (67-106)
- AST evidence: `ast.json` — branches 5, returns 4, calls 8, assignments 10,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**a092가 편집하는 두 함수 중 둘.** 34초 예산의 셋째 항(publish 1회 **10초**)이
여기 `:101`의 구조체 리터럴에서 비어 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cfg config.Notifications` | 기본 `Enabled:false` | `config/notifications.go:39-58` | B3 `:83`이 꺼짐을 조용히 통과(오류 아님) |
| `getenv func(string) string` | 주입, **nil 허용** | `engine.go:444` `opts.Getenv` | nil이면 B1 `:69`이 `os.Getenv`로 채운다 — **반환하지 않는다** |
| `cfg.Rejected` | 공백이면 정상 | `mergeNotifications` | B2 `:77`이 사유를 옮겨 담고 반환 |
| `cfg.Topic` / `TOSSCTL_NTFY_TOPIC` | 하나는 있어야 함 | 파일 ⊕ 환경 | B5 `:94`가 사유를 담고 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | publisher |
|---|---|---|---|---|
| B1 `:69` | `getenv == nil` | `getenv = os.Getenv` `:70` | **없음** | — |
| B2 `:77` | `resolution.Refused != ""` | 없음(`:75`가 이미 담았다) | `:81` | **nil** |
| B3 `:83` | `!cfg.Enabled` | 없음 | `:87` | **nil** |
| B4 `:91` | 환경 topic이 비었음 | `topic = cfg.Topic` `:92` | **없음** | — |
| B5 `:94` | topic이 여전히 비었음 | `resolution.Refused = ...` `:95` | `:97` | **nil** |
| — `:101` | — | **`ntfy := &obs.Ntfy{BaseURL, Topic, Token}`** | `:105` | `*obs.Ntfy` |

**AST 반환은 넷이다 — `:81`·`:87`·`:97`·`:105`.** 다섯 분기 중 **B1과 B4는 반환하지
않는다**: B1은 `getenv`의 nil 기본값을 채우고, B4는 파일 topic으로 넘어간다.
`resolution.Refused`는 B2가 아니라 `:75`의 구조체 리터럴이 `cfg.Rejected`에서 채운다.

**세 이탈(B2·B3·B5)이 nil publisher를 돌려준다.** 그것이 오늘 원장의 상태다 —
`alert_outbox` PENDING 9행 전부 `attempts=0`이고 로그의 오류는
`no notification publisher is configured`다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` `:74`·`:75`·`:90`·`:92`·`:99` (5회) | 정규화 | 즉시 | AST calls |
| `getenv` `:90`·`:99` | 환경에서 topic·token | 즉시 | AST calls |
| `ntfy.UsesPublicService` `:104` | 공개 서비스 경고 | 즉시 | AST calls |

**네트워크 호출 없음.** 이 함수는 값만 만든다.

### `:101`이 비우는 필드

| `obs.Ntfy` 필드 | 이 함수 | 안 채우면 |
|---|---|---|
| `BaseURL` | 채움 | — |
| `Topic` | 채움 | — |
| `Token` | 채움 | — |
| `HTTPClient` | **안 채움** | `Publish` B7 `:122` → `&http.Client{Timeout: timeout}` |
| **`Timeout`** | **안 채움** | `Publish` B3 `:96` → **10s** |

## State mutations and fallbacks

- `resolution`(값 타입)만 채운다. AST assignments 10은 전부 지역 또는 `resolution` 필드.
- 프로세스 상태를 바꾸지 않는다.
- goroutine·defer 없음.

## Safety conclusion

- **Safe edit boundary**: `Timeout`을 `:101`의 리터럴에 추가하는 편집은 **B1~B5의
  조건·순서·반환값을 바꾸지 않는다.** 편집 후 AST branches가 5, returns가 4로
  유지되면 제어 흐름 무변화가 증명된다.
- **High-risk impact**: **yes** — §0.5. publisher가 nil이 되는 세 경로는 그대로 둔다.
- **CLI 시험 발송은 다른 자리다**: `cmd/tossctl/notificationsettings.go:151`이 별도
  `&obs.Ntfy{}`를 만든다. **거기는 엔진 루프가 아니므로 a092의 범위 밖**이고
  10초 기본값을 유지한다 — 운영자가 손으로 한 번 보내는 시험에 짧은 기한을 씌우면
  차가운 연결에서 헛되이 실패한다.

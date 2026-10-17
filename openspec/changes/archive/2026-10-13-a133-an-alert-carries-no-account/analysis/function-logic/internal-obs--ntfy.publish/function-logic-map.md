# Function Logic Map: `Ntfy.Publish`

- Source: `internal/obs/ntfy.go` (85-142, 편집 뒤 — base 85-139)
- Qualified: `Ntfy.Publish`
- AST evidence: `ast.json` (편집 뒤 재추출, `source_sha256` 0c451499c6b640fd… · 편집 전 f777114c1c240ef9…) — AST branches 9 · returns 5
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 코드의 유일한 `obs.Publisher` 구현 — 알림 하나를 ntfy 토픽으로 POST 한다. 발행 호출 자리 7(design D1 P1~P7)이 전부 여기로 모인다.
반환이 곧 전달 판정의 입력이다: nil = 나감, 오류 = 안 나감(outbox 정산 · a124 시도 수 · 진입 래치 · 모드 승격이 이것을 읽음).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Topic` | 비어 있지 않음(a075 기계 생성) | 설정 · 환경 변수 | 빈 값 → `ErrNtfyNotConfigured`(B1) |
| `n.BaseURL` | 비면 ntfy.sh | 설정 | B2 기본값 |
| `n.Timeout` | 0 이하 → 기본 | 호출자 | B3 |
| `n.Token` | 선택 | 환경 변수 | 비면 Authorization 없음(B6) |
| `n.HTTPClient` | nil 가능 | 시험 주입 | B7 기본 클라이언트 |
| `msg.Title` · `msg.Body` · `msg.Tags` · `msg.Type` | 임의 문자열 — **사건 내용(계좌 포함 가능)** | 호출자(알림기 · 전달기 · 콘솔) | base: 그대로 헤더 · 본문으로 나감 |
| `msg.Severity` · `msg.Priority` | | 호출자 | `priorityFor` |

## Branches and early returns

| Branch | AST kind | Source line | a133 |
|---|---|---|---|
| B1 | if at 90:2 | `if topic == "" {` | 무편집(앞에 치환 대입 한 문장이 섬) |
| B2 | if at 94:2 | `if base == "" {` | 무편집 |
| B3 | if at 99:2 | `if timeout <= 0 {` | 무편집 |
| B4 | if at 107:2 | `if err != nil {` (요청 구성 실패) | 무편집 — 본문은 치환된 값 |
| B5 | if at 113:2 | `if title := strings.TrimSpace(msg.Title); title != "" {` | 편집 영향 — 치환된 Title 이 헤더로 |
| B6 | if at 120:2 | `if token := strings.TrimSpace(n.Token); token != "" {` | 무편집 |
| B7 | if at 125:2 | `if client == nil {` | 무편집 |
| B8 | if at 129:2 | `if err != nil {` (전송 실패) | 무편집 |
| B9 | if at 137:2 | `if resp.StatusCode < 200 \|\| resp.StatusCode >= 300 {` | 무편집 |

Exact AST return positions(편집 뒤): 91:3, 108:3, 130:3, 138:3, 141:2 — 편집 전과 같은 다섯이 3줄씩 밀림(분기 · 반환 수 불변: 9 · 5)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewRequestWithContext` | POST base/topic, 본문 = `msg.Body` | 오류 → B4 | AST calls |
| `req.Header.Set` Title · Tags · Priority · Content-Type · Authorization | ntfy 메타데이터 | 없음 | AST calls |
| `client.Do` | 전송 | 오류 → B8, 시한 = `timeout` | AST calls |
| `io.ReadAll(LimitReader 4096)` | 응답 본문(오류 문구) | 무시 | AST calls |
| `priorityFor` | 우선순위 | 순수 | AST calls |
| `redactForTransport`(a133 신설) | Title · Body · Tags · Type 을 `redact.String` 으로 | 순수 — 오류 · 패닉 경로 없음 | design D2 |

## State mutations and fallbacks

- 반환 외 부작용은 네트워크 요청 하나. `msg` 는 값 매개변수 — a133 의 재대입은 호출자의 값을 바꾸지 않는다(Tags 는 새 슬라이스).

## Safety conclusion

- 착지(1.2): B1 앞 `msg = redactForTransport(msg)`(88줄) — 주석 두 줄 포함 +3줄.
- Safe edit boundary: B1 앞에 `msg = redactForTransport(msg)` 한 문장. 분기 조건 · 반환 · 오류 문구 불변 → 같은 서버 응답에 같은 반환.
- High-risk impact: 인접 — 반환이 전달 판정의 입력이나 반환은 불변. 주문 · 청산 경로는 이 함수를 기다리지 않는다(design D5).

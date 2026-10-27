# Function Logic Map: `Console.chromeFor`

- Source: `internal/console/chrome.go` (184-247) — 편집 뒤(a173 rebase 후 main).
- Qualified: `Console.chromeFor`
- AST evidence: `ast.json` (`source_sha256` 367463d91837a4ab…) — AST branches 16 · returns 1
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-25-a173-console-ux-common-cleanup` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/chrome.go` = `367463d91837…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: 완료 게이트 직전(2026-10-25, 종결 에이전트). tasks 3.1 의 비례 원칙 면제는 required 7 이라 gate ⑤ 가 받지 않아 사후에 만든 산출물임(FLM 이 구현 판단에 쓰이지 않음).

**역할.** 모든 화면의 머리 상태 표시줄(엔진 · 데이터 신선도 · 세션 · 승인 대기) 뷰 모델을 만든다. 읽기 전용 렌더.

a173 편집: dataUnavailable 갈래에서 긴 사유(`f.Reason`)를 칸 본문(`DataNote`)이 아니라 `DataTitle`(툴팁)로 옮기고, 열거 사유 코드 `f.Code` 를 `읽지 못함 · <코드>` 로 한 단어만 붙임(B11 추가). `LastAt` 은 `마지막 성공 …` 단독 문구(B12), 갱신 보류는 `갱신 보류` 한 단어 + 긴 문구는 `DataTitle`(B14). 칸을 한 줄로 만드는 표시 변경이고 판정 · 상태 · 쓰기 경로 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `f freshness` | Mode 셋(cache · unavailable · onRequest) · Code 는 열거 사유 또는 빈 값 | 각 화면의 `freshness()` | 미지 Mode 는 default(요청 시 읽음)로 |
| `c.opts.Binary` | nil 또는 seam | 콘솔 Options | nil 이면 설치 stamp 영값 — 엔진 Stale 판정에만 영향 |
| `c.currentRun()` | nil 또는 진행 중 검증 | verify 실행 상태 | nil 이면 승인 대기 표시 없음 |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a173 |
|---|---|---|---|---|
| B1 | `188:2` | if | Binary seam 배선 → 설치 stamp 읽기 | 무변경 |
| B2 | `195:2` | switch | 엔진 상태 switch | 무변경 |
| B3 | `196:2` | case | !Wired → 미배선 | 무변경 |
| B4 | `200:2` | case | Running → 실행 중 | 무변경 |
| B5 | `203:3` | if | PID>0 → PID 표기 | 무변경 |
| B6 | `206:3` | if | Stale → 다른 실행 파일 경고 | 무변경 |
| B7 | `209:2` | case | default → 정지 | 무변경 |
| B8 | `216:2` | switch | 데이터 Mode switch | 무변경 |
| B9 | `217:2` | case | dataCache → 시각(n초 전) | 무변경 |
| B10 | `219:2` | case | dataUnavailable → `읽지 못함`, `DataTitle`=Reason | a173: Reason 을 본문에서 title 로 |
| B11 | `221:3` | if | Code 있음 → ` · <코드>` | a173 추가 |
| B12 | `225:3` | if | LastAt 있음 → `마지막 성공 …` | a173: Reason 이어붙이기 대신 단독 문구 |
| B13 | `228:2` | case | default(요청 시 읽음) → 렌더 시각 | 무변경 |
| B14 | `234:2` | if | Hold 있음 → `갱신 보류`, `DataTitle`=Hold | a173: 긴 문구를 title 로 |
| B15 | `239:2` | if | 검증 실행 있음 | 무변경 |
| B16 | `240:3` | if | 승인 대기 → PendingApproval · 링크 | 무변경 |

Exact AST return positions: 246:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.now` | 렌더 시각 | 순수 | AST calls 185 |
| `c.opts.Binary` | 설치 stamp | 오류 무시(영값) — 무변경 | AST calls 189 |
| `c.readEngine` | 엔진 마커 읽기 | 무변경 | AST calls 191 |
| `verifylive.SessionAdvisoryFor` | 세션 advisory 문구 | 순수 | AST calls 193 |
| `strconv.Itoa` · `f.Age.Round` · `f.tone` | 표기 | 순수 | AST calls 204 · 215 · 218 |
| `renderInstant` | 요청 시 읽음 시각 | 순수 | AST calls 231 |
| `c.currentRun` · `run.snapshot` | 승인 대기 | 읽기만 | AST calls 239 · 240 |

## State mutations and fallbacks

- 지역 `s statusStrip` 만 채워 반환. a173 은 `DataTitle` 필드에 쓰는 대입 셋과 `DataText` 덧붙임 하나를 더함. 쓰기 · I/O 추가 0.

## Safety conclusion

- Safe edit boundary: 표시 문자열 배치만 바뀜. 분기 조건 · 순서 · 반환 무변경, 추가 분기 B11 은 빈 Code 면 이전과 같은 `읽지 못함`.
- High-risk impact: no — 콘솔 읽기 전용 렌더(주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉).

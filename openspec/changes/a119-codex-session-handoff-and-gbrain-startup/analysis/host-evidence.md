# a119 호스트 증거 (task 2.2) — 살균본

- 측정일: 2026-09-25 (KST)
- 측정자: a119 구현 Teammate (Opus), Manager 독립 검증 전
- 살균 규칙: 세션 본문·절대 홈 경로·시크릿·계정 식별자·계좌 정보는 옮기지 않음.
  세션 ID 는 8자 접두어만, 경로는 `~` 또는 저장소 상대경로로만 적음.
- 이 문서는 **정적 산출물·로그의 판독**이다. 이 세션은 Codex 호스트를 띄우지 않았다
  (3.3 관측은 pending — 아래 §4).

## 1. 조사한 호스트와 산출물

| 대상 | 값 | 비고 |
| --- | --- | --- |
| Codex Desktop 번들 CLI | `codex-cli 0.155.0-alpha.16.3` (앱 `26.917.62051`) | `/usr/lib/chatgpt/resources/codex` |
| 사용자 PATH CLI | `codex-cli 0.154.0` | `~/.local/bin/codex` |
| 세션 기록 | `~/.codex/sessions/2026/**/rollout-*.jsonl` | TossOS cwd 세션 2026-08-20 이후 64개 |
| 로그 DB | `~/.codex/logs_2.sqlite` (읽기 전용 열람) | 12,228행, 2026-09-17~09-25 만 보존 |
| 원 진단 기록 | `~/.codex/memories/rollout_summaries/2026-08-29T15-33-36-…gbrain_duplicate_mcp_diagnosis.md` | a119 proposal 의 출처 세션 |
| 저장기 산출물 | `.codex-context/session-summary.md`, `backups/` 5개 | 마지막 저장 2026-09-20 02:03Z |

## 2. PostToolUse 이벤트 이름

### 2.1 관측된 것

| 이름 | 증거 종류 | 내용 |
| --- | --- | --- |
| `Bash` | **전달 관측(추론 1단계)** + 바이너리 상수 | 저장기가 쓴 `.codex-context/session-summary.md` 의 Session 필드가 `codex exec` 0.154.0 세션 `01a0bc8a…`(2026-09-20 02:00~02:03Z)이고, 백업 5개가 02:02:42~02:03:33Z 에 생성됨. 그 세션의 도구 항목은 `CommandExecution` 23건·`FileChange` 0건이므로 현 matcher `^(Bash|apply_patch)$` 에 걸린 이름은 `Bash` 로 추론함(`apply_patch` 는 `FileChange` 항목을 남긴다는 전제). 바이너리 0.155 에는 `Bash` 상수가 `command`·`description`(Claude 호환 tool_input 키)과 인접해 여러 번 나타남 |
| `apply_patch` | 바이너리 상수만 | 0.155 바이너리의 한 함수 안에서 `Bash` 와 `apply_patch` 상수가 나란히 비교됨(오프셋 약 101,035,596). **전달은 관측 못 함** — 관측 창의 세션들에 `FileChange` 뒤 저장 흔적을 가를 방법이 없음(백업이 5개로 잘림) |

### 2.2 관측되지 않은 것 — 이름을 지어내지 않음

다음은 세션 기록에서 **모델이 보는 도구 호출 이름**으로 나타났지만, 그것이 PostToolUse
훅의 `tool_name` 으로 전달되는지는 어떤 산출물에서도 확인되지 않았다.

- code mode 최상위 호출 `exec` (TossOS 세션 `custom_tool_call` 1,160건)
- 다중 에이전트 도구 `send_message`·`wait`·`wait_agent`·`spawn_agent`·`followup_task`·`list_agents`·`sleep`·`request_user_input_async`
- MCP 도구(`McpToolCall` 항목: 서버 `codegraph`·`codex_app`·`codex_apps`). 바이너리 문자열에
  `mcp__server__tool` 명명 규칙이 있으나 그것은 **모델 쪽 도구 이름** 설명이며 훅 페이로드
  이름의 증거가 아님

세션 기록 전수(2026 전체)에서 훅 실행 자체를 남긴 레코드(`*hook*` 타입)는 **0건**이다.
로그 DB 에서도 `PostToolUse`/`post_tool_use` 문자열 0건. 따라서 추가 이름은 **미관측**이다.

### 2.3 결론 — matcher

살균 픽스처로 확립된 PostToolUse 이름은 `Bash`·`apply_patch` 둘이고 둘 다 현 matcher 에 이미
걸린다. **확립된 추가 이름은 0개**이므로 스펙(`codex-session-save` 델타)이 요구하는 추가
이름 적용 대상이 없다. proposal 의 "matcher 가 현 호스트 이벤트를 못 덮는다"는 원인 가설은
이 증거로 **지지되지 않는다** — 현 호스트(0.154 exec)에서 저장기는 실제로 돌았다.
2026-08-29 시점의 정체(08-17 이후 미갱신)는 당시 호스트(VS Code 확장 0.150/0.151 alpha)에서
code mode 중첩 명령이 `CommandExecution` 항목조차 남기지 않은 세션이 있었다는 점(예: 08-29 13:22Z
세션 `exec` 115회·`CommandExecution` 0건)과 부합하지만, 그 판본의 훅 전달 여부는 지금 재현할 수 없다.

**부수 위험(측정)**: Codex 는 훅 신뢰를 `~/.codex/config.toml` 의
`[hooks.state."<repo>/.codex/hooks.json:post_tool_use:<group>:<handler>"] trusted_hash` 로 기억한다.
matcher 문자열을 바꾸면 그룹 해시가 바뀌어 사용자가 `/hooks` 에서 다시 승인하기 전까지 저장기가
돌지 않을 수 있다(원 진단 세션의 "trust 미승인" 오진 기록과 같은 축). 근거 없는 matcher 확장은
갱신을 늘리기는커녕 끊을 수 있다.

## 3. GBrain MCP 등록과 "두 번 시작"

### 3.1 호스트가 실제로 로드하는 등록 수

호스트 자신의 설정 로더로 셌다(`codex mcp list --json`, 읽기 전용; 실행 전후
`~/.codex/config.toml` mtime·크기 불변 확인).

| 실행 위치 | 0.155.0-alpha.16.3 | 0.154.0 |
| --- | --- | --- |
| TossOS 저장소 루트 | `gbrain` **1개** (`python3 tools/sdd/gbrain_project.py serve`), `codegraph` 1개 | 동일 |
| 저장소 밖(scratchpad) | `gbrain` **0개**, `codegraph` 0개 | (미실행) |

- 사용자 전역 `~/.codex/config.toml` 에는 `gbrain` 등록 **0개**(섹션 헤더 전수 확인).
- 따라서 유일한 `gbrain` 은 신뢰된 프로젝트 층 `.codex/config.toml` 에서 온다.
- `.mcp.json`(Claude 소유) 의 `gbrain` 은 호스트의 유효 목록에 **나타나지 않는다**. 바이너리에는
  `.mcp.json` 을 플러그인 루트·executor capability discovery 에서 읽는 경로가 있으나, 그것이
  워크스페이스 루트에 적용되는지는 미관측이다(`mcp list` 는 세션 시점 capability discovery 를
  돌지 않음) — **잔여 불확실성**으로 남김.

즉 스펙 `gbrain-codex-mcp-startup` 의 "유효 설정에 정확히 하나" 는 **이미 성립**한다.
정적 텍스트만 보고 `.mcp.json` 쪽을 지우는 것은 design.md 가 금지한 추론이며 증거로도 불필요하다.

### 3.2 "두 번 시작"의 실제 원인 — 스레드마다 MCP 기동

- 원 진단(2026-08-29 15:39Z) 사용자 화면: `MCP client for gbrain failed to start: … connection closed:
  initialize response` / `MCP startup incomplete (failed: gbrain)` — 실패한 서버 이름은 `gbrain`
  **하나**뿐이었다. 당시 락 소유자는 "VS Code Codex app-server 의 자식"이었고 직접 probe 는
  `[gbrain-project] busy: … exit=75` 였다(원 진단 기록).
- 같은 시각 TossOS cwd 의 Codex 스레드는 **셋이 동시에 살아 있었다**: `01a04daf…`(13:22Z 시작),
  `01a04e24…`(15:30Z, 0.150.0-alpha.12.2), `01a04e27…`(15:33Z, 0.151.0-alpha.7.1, 오류 표시 스레드).
- 로그 DB(09-20·09-24, 비-TossOS 스레드)에서 **한 app-server 프로세스 안의 서로 다른 스레드
  resume 요청마다** `node_repl`·`cua_repl` MCP 기동이 따로 기록된다(요청 ID 별 그룹: 한 프로세스
  3953 에서 두 요청 각각 기동 시도). 즉 호스트는 MCP 서버를 **스레드(세션) 단위로 기동**한다.
- 결론: "두 번 시작"은 등록 중복이 아니라 **동시 스레드 N 개 × 스레드당 1회 기동**이고, 두 번째
  이후 기동은 정본 `sdd-workflow` 계약(단일 writer flock, exit 75 busy)이 의도한 대로 끝난다.
  이는 기억 `EP-TOS-20260730-004`("두 번째 에이전트에 동시 GBrain MCP 를 제공하지 않으며 그 요구는
  HTTP broker 또는 local PostgreSQL backend 를 다루는 별도 change 가 필요")와 일치한다.

### 3.3 현재 프로세스 상태(참고, 읽기 전용)

2026-09-25 측정 시 TossOS 홈의 `gbrain serve` 소유자는 Claude 세션의 자식이었고(락 JSON pid 일치,
heartbeat 신선), Codex 프로세스는 없었다. 어떤 프로세스도 종료·락 삭제하지 않았다.

## 4. 관측하지 못한 것 (3.3 으로 넘김)

- 현 Desktop 호스트(0.155 alpha)에서 일반 도구 호출 뒤 `.codex-context/session-summary.md` 갱신 — **pending**
- `apply_patch` 이벤트의 실제 전달 — **pending**
- Desktop 호스트에서 워크스페이스 한 스레드 기동 시 wrapper 기동 횟수(= 1 기대) — **pending**
- 워크스페이스 `.mcp.json` 이 executor capability discovery 로 로드되는지 — **pending**

이 세션은 Codex 호스트를 띄우지 않았다: 띄우면 사용자 계정으로 원격 모델 호출이 나가고,
새 훅 신뢰가 사용자 전역 `~/.codex/config.toml` 에 기록된다. 둘 다 이 change 가 승인받지 않은
부작용이다.

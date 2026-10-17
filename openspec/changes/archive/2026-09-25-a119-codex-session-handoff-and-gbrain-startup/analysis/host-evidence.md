# a119 호스트 증거 (task 2.2) — 살균본

- 측정일: 2026-09-25 (KST)
- 측정자: a119 구현 Teammate (Opus), Manager 독립 검증 전
- 살균 규칙: 세션 본문·절대 홈 경로·시크릿·계정 식별자·계좌 정보는 옮기지 않음.
  세션 ID 는 8자 접두어만, 경로는 `~` 또는 저장소 상대경로로만 적음.
- 이 문서는 **정적 산출물·로그의 판독**이다. 이 세션은 Codex 호스트를 띄우지 않았다
  (what stays unobserved and which named follow-up owns it: §4, task 3.3).

## 1. 조사한 호스트와 산출물

| 대상 | 값 | 비고 |
| --- | --- | --- |
| Codex Desktop 번들 CLI | `codex-cli 0.155.0-alpha.16.3` (앱 `26.917.62051`) | Desktop 앱 번들 경로(시스템 경로, 절대 경로는 살균 규칙에 따라 생략) |
| 사용자 PATH CLI | `codex-cli 0.154.0` | `~/.local/bin/codex` |
| 세션 기록 | `~/.codex/sessions/2026/**/rollout-*.jsonl` | TossOS cwd 세션 2026-08-20 이후 64개 |
| 로그 DB | `~/.codex/logs_2.sqlite` (읽기 전용 열람) | 12,228행, 2026-09-17~09-25 만 보존 |
| 원 진단 기록 | `~/.codex/memories/rollout_summaries/2026-08-29T15-33-36-…gbrain_duplicate_mcp_diagnosis.md` | a119 proposal 의 출처 세션 |
| 저장기 산출물 | `.codex-context/session-summary.md`, `backups/` 5개 | 마지막 저장 2026-09-20 02:03Z |

## 2. PostToolUse 이벤트 이름

### 2.1 관측된 것

| 이름 | 증거 종류 | 내용 |
| --- | --- | --- |
| `Bash` | **전달 추론**(`codex exec` 0.154 에서만) + 바이너리 상수 — 페이로드에서 이름을 직접 본 적은 없음 | 저장기가 쓴 `.codex-context/session-summary.md` 의 Session 필드가 `codex exec` 0.154.0 세션 `01a0bc8a…`(2026-09-20 02:00~02:03Z)이고, 백업 5개가 02:02:42~02:03:33Z 에 생성됨. 그 세션의 도구 항목은 `CommandExecution` 23건·`FileChange` 0건이므로 현 matcher `^(Bash|apply_patch)$` 에 걸린 이름은 `Bash` 로 추론함(`apply_patch` 는 `FileChange` 항목을 남긴다는 전제). 바이너리 0.155 에는 `Bash` 상수가 `command`·`description`(Claude 호환 tool_input 키)과 인접해 여러 번 나타남 |
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

살균 픽스처에 담을 수 있는 이름은 `Bash`(전달 추론, `codex exec` 0.154 한정)와 `apply_patch`(바이너리
상수만)뿐이고 둘 다 현 matcher 에 걸린다. **확립된 추가 이름은 0개**다.

proposal 의 "matcher 가 현 호스트 이벤트를 못 덮는다"는 원인 가설은 **확립되지도 반박되지도 않았다**.
증상이 난 호스트는 대화형 VS Code/Desktop 이고, 저장 전달은 **다른 호스트**(`codex exec`)에서만 추론됐다.
- 반증 자료(2026-09-25 적대 리뷰가 지적): 세션 `01a04daf`(vscode 0.150.0-alpha.8)에는 08-29 14:39~15:24Z 에
  `FileChange` 10건이 있는데 15:33Z 진단 시점에 핸드오프는 여전히 08-17 이후 미갱신이었다. 당시 훅 신뢰가
  승인 전이었는지(원 진단 기록은 "초기엔 trust hash 없음, 사용자 정정 뒤 존재")로 설명될 수도 있으나
  **확인되지 않았다 — 미해명**.
- 현 대화형 호스트는 **미측정**: vscode 0.153.0(09-05/06, 새 경로) 세션들은 code mode `exec` 1,394회·`FileChange`
  260건·`CommandExecution` **0건**이다. 여기서 PostToolUse 가 어떤 이름으로(혹은 아예) 나가는지 모른다.

**부수 위험(추론 — 측정 아님)**: Codex 는 훅 신뢰를 `~/.codex/config.toml` 의
`[hooks.state."<repo>/.codex/hooks.json:post_tool_use:<group>:<handler>"] trusted_hash` 로 기억한다.
해시가 그룹 matcher 까지 덮는지는 확인하지 않았으나, 덮는다면 matcher 문자열을 바꿀 때 해시가 바뀌어 사용자가 `/hooks` 에서 다시 승인하기 전까지 저장기가
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

즉 **CLI 설정 로더 기준으로는** 스펙 `gbrain-codex-mcp-startup` 의 "유효 설정에 정확히 하나" 가 성립한다
(app-server 세션 시작 경로의 capability discovery 는 미검증). 출처 층 `.codex/config.toml` 은 `mcp list` 출력에
없는 값이며 **소거법으로 추론**했다(사용자 층 0개, 저장소 밖 0개; managed/system 층은 열거하지 않음).
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
- 결론(**추론**; 로그 DB 에 `gbrain` 행 0건, 스레드별 기동은 비-TossOS `node_repl`/`cua_repl` 행을 0.154/0.155 에서 본 것을 0.150/0.151 vscode 에 적용한 것): "두 번 시작"은 등록 중복이 아니라 **동시 스레드 N 개 × 스레드당 1회 기동**이고, 두 번째
  이후 기동은 정본 `sdd-workflow` 계약(단일 writer flock, exit 75 busy)이 의도한 대로 끝난다.
  이는 기억 `EP-TOS-20260730-004`("두 번째 에이전트에 동시 GBrain MCP 를 제공하지 않으며 그 요구는
  HTTP broker 또는 local PostgreSQL backend 를 다루는 별도 change 가 필요")와 일치한다.

### 3.3 현재 프로세스 상태(참고, 읽기 전용)

2026-09-25 측정 시 TossOS 홈의 `gbrain serve` 소유자는 Claude 세션의 자식이었고(락 JSON pid 일치,
heartbeat 신선), Codex 프로세스는 없었다. 어떤 프로세스도 종료·락 삭제하지 않았다.

## 4. Unobserved — owned by named follow-ups, not by this change (task 3.3, 2026-09-26)

Scope decision (a) (issues I-1, 2026-09-25) removed runtime observation from a119. Every item below stays
**unobserved**. This change makes no runtime claim about any of them. The regression pins landed in 3.1
(`c202b804`) are static: they check the repository configuration against sanitized fixtures. They do not observe
host delivery or host startup (spec scenarios "Host coverage has not been observed" and "Concurrent threads
start the workspace").

The follow-ups are the ones named in `proposal.md` "Follow-ups". They get numbers when they are opened.

| # | Unobserved item | Owning follow-up |
| --- | --- | --- |
| U1 | On the interactive VS Code/Desktop host (0.155 alpha), whether an ordinary tool call refreshes `.codex-context/session-summary.md` | 1 — interactive-host handoff refresh |
| U2 | Whether an `apply_patch` event is actually delivered to the PostToolUse hook (binary constant only, §2.1) | 1 — interactive-host handoff refresh |
| U3 | Which `tool_name` (if any) the interactive host sends for code-mode `exec` / `FileChange` calls (§2.3) | 1 — interactive-host handoff refresh; the same observation is the prerequisite of follow-up 3 |
| U4 | The new-path trust entry `post_tool_use:1:0` in `~/.codex/config.toml` has `trusted_hash` but no `enabled = true`. The entry for the pre-rename repository path has both. Saver delivery on `exec` 0.154 was inferred with this entry (§2.1, an inference); its effect on the interactive host is unknown | 1 — interactive-host handoff refresh |
| U5 | How many times the wrapper starts when one workspace thread starts on the interactive host (expected 1 per thread, §3.2 — an inference) | 2 — per-thread MCP startup warning |
| U6 | Whether the workspace `.mcp.json` is loaded through the app-server's executor capability discovery (§3.1 residual uncertainty) | 2 — per-thread MCP startup warning |
| U7 | Whether Codex emits any name matched by the SDD agent-save matcher `Write\|Edit\|MultiEdit\|NotebookEdit` (issues I-2) | 3 — SDD agent-save handler on Codex |

Observation route for follow-up 1 (needs human approval; not run here): a temporary probe hook with matcher `.*`
that writes only `tool_name` under `.codex-context/`. A human approves its trust, runs one session, and then removes it.

This session did not start a Codex host. Starting one sends remote model calls on the user's account and writes a
new hook trust into the user-global `~/.codex/config.toml`. This change is not approved for either side effect.

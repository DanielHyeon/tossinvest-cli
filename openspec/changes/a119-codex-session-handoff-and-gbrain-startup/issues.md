# a119 issues

## I-1 (2026-09-25, 분류: ① **blocking** — proposal-freeze 거절, Manager·사람 범위 결정 필요)

**proposal 의 두 원인 가설은 호스트 증거로 확립되지 않았다(반박된 것도 아니다).** 근거: `analysis/host-evidence.md`.

1. **matcher 가설** — "PostToolUse matcher 가 현 호스트 이벤트를 못 덮는다". 현 호스트(codex exec 0.154.0)
   에서 저장기는 `^(Bash|apply_patch)$` 로 돌았다고 추론되지만(2026-09-20), 증상 호스트(대화형 VS Code/Desktop)는
   미관측이고 08-29 의 `FileChange` 10건 뒤 미갱신은 미해명이다. 살균 픽스처로 확립된 추가 이름은 0개다.
   스펙 델타는 "확립된 추가 이름"만 요구하므로 **matcher 를 바꾸지 않는 것이 스펙에 맞는 최소 수정**이다.
   근거 없이 넓히면 `~/.codex/config.toml` 의 훅 `trusted_hash` 가 어긋나 오히려 저장이 끊길 수 있다.
2. **중복 등록 가설** — "Codex 가 두 등록(`.codex/config.toml`·`.mcp.json`)을 로드해 두 번 띄운다".
   CLI 설정 로더(`codex mcp list --json`, 0.155 alpha·0.154)가 TossOS 에서 로드하는 `gbrain` 은 **1개**다(app-server
   capability discovery 는 미검증). 관측된 "두 번 시작"은 **동시 Codex 스레드마다 MCP 를 따로 기동**한 결과이고(08-29 에 스레드 3개 동시),
   두 번째 이후는 정본 단일-writer 계약대로 exit 75 로 끝난다는 것이 **추론**이다.

**이 change 가 하는 것**: 두 사실을 회귀 시험으로 못 박는다(픽스처 이름 ⊆ matcher, Codex 유효 등록 = wrapper
1개·raw `gbrain serve` 금지). 설정 파일은 바꾸지 않는다.

**이 change 가 못 하는 것(사람 결정 필요)**: 여러 Codex 스레드(또는 Codex+Claude)가 동시에 열릴 때 뒤의
스레드에 뜨는 `MCP startup incomplete (failed: gbrain)` 경고는 이 스펙의 범위 안에서 없앨 수 없다.
없애는 길은 모두 이 스펙이 금지하거나 새 결정을 요구한다 —
(a) wrapper 의 busy 동작 변경(스펙 `Preserve project ownership` 이 보존을 요구), (b) Codex 등록 비활성화
(스펙 "정확히 하나의 enabled 등록"과 충돌), (c) HTTP broker/공유 백엔드(기억 `EP-TOS-20260730-004` 가 별도
change 로 지정). proposal 의 "exactly once" 문구를 "스레드당 최대 1회"로 읽을지, 후속 change 를 열지는
Manager·사람이 정한다.

**2026-09-25 proposal-freeze 결과: 거절(REJECT).** 계획(설정 불변 + 시험만)이 proposal 의 Why·What Changes 와
어긋난 채 동결하면 목적 미충족으로 보관될 수 있다. 동결 전에 다음 중 하나를 사람이 정해야 한다 —
(a) proposal 을 "증거와 회귀 고정만"으로 다시 쓰고, 정체된 핸드오프·스레드별 경고를 이름 붙인 후속 change 로
넘김, 또는 (b) 범위를 유지하고 두 증상이 해소될 때까지 보관을 막음(대화형 호스트 관측·사람 승인 probe 필요).
Teammate 는 이 결정 전에는 3.x 구현 커밋을 하지 않는다. 회귀 시험 초안은 `wip/a119-3.1` 브랜치에 보존했다.

**2026-09-25 사용자 결정: (a).** proposal 을 「증거 + 회귀 고정」으로 다시 썼고, 대화형 호스트 핸드오프 정체 ·
스레드별 경고 · I-2 는 이름 붙인 후속으로 넘겼다(proposal 「Follow-ups」). `gbrain-codex-mcp-startup` 델타에서
「Workspace startup is observed」 시나리오를 뺐고 스레드별 busy 종료를 이 change 가 없애지 않는다고 적었다.
동결 재리뷰(적대 보이스 1)는 Opus 리셋 뒤. 그 전에는 3.x 커밋 없음.

## I-2 (2026-09-25, 분류: 기록만 — 이 change 범위 밖)

`.codex/hooks.json:5` 의 SDD agent-save 핸들러 matcher `Write|Edit|MultiEdit|NotebookEdit` 는 Codex 가 낸다고
알려진 이름(`Bash`·`apply_patch`) 어느 것에도 걸리지 않는다. Codex 에서는 SDD agent-save 이벤트가 아마 생기지
않는다(`.sdd/history/events/agent-saves.jsonl` 의 actor `codex` 는 6,003건 중 3건). 기존 시험은 핸들러 **존재**만
본다. 스펙의 "coexist" 의도와 어긋날 수 있으나 이 change 에서 고치지 않는다 — 후속 change 후보.

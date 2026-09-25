# a119 issues

## I-1 (2026-09-25, 분류: ② safe local — 구현 비차단, Manager 판단 필요)

**proposal 의 두 원인 가설이 호스트 증거로 지지되지 않는다.** 근거: `analysis/host-evidence.md`.

1. **matcher 가설** — "PostToolUse matcher 가 현 호스트 이벤트를 못 덮는다". 현 호스트(codex exec 0.154.0)
   에서 저장기는 `^(Bash|apply_patch)$` 로 실제 돌았고(2026-09-20), 살균 픽스처로 확립된 추가 이름은 0개다.
   스펙 델타는 "확립된 추가 이름"만 요구하므로 **matcher 를 바꾸지 않는 것이 스펙에 맞는 최소 수정**이다.
   근거 없이 넓히면 `~/.codex/config.toml` 의 훅 `trusted_hash` 가 어긋나 오히려 저장이 끊길 수 있다.
2. **중복 등록 가설** — "Codex 가 두 등록(`.codex/config.toml`·`.mcp.json`)을 로드해 두 번 띄운다".
   호스트 로더(`codex mcp list --json`, 0.155 alpha·0.154)가 TossOS 에서 로드하는 `gbrain` 은 **1개**다.
   관측된 "두 번 시작"은 **동시 Codex 스레드마다 MCP 를 따로 기동**한 결과이고(08-29 에 스레드 3개 동시),
   두 번째 이후는 정본 단일-writer 계약대로 exit 75 로 끝난다.

**이 change 가 하는 것**: 두 사실을 회귀 시험으로 못 박는다(픽스처 이름 ⊆ matcher, Codex 유효 등록 = wrapper
1개·raw `gbrain serve` 금지). 설정 파일은 바꾸지 않는다.

**이 change 가 못 하는 것(사람 결정 필요)**: 여러 Codex 스레드(또는 Codex+Claude)가 동시에 열릴 때 뒤의
스레드에 뜨는 `MCP startup incomplete (failed: gbrain)` 경고는 이 스펙의 범위 안에서 없앨 수 없다.
없애는 길은 모두 이 스펙이 금지하거나 새 결정을 요구한다 —
(a) wrapper 의 busy 동작 변경(스펙 `Preserve project ownership` 이 보존을 요구), (b) Codex 등록 비활성화
(스펙 "정확히 하나의 enabled 등록"과 충돌), (c) HTTP broker/공유 백엔드(기억 `EP-TOS-20260730-004` 가 별도
change 로 지정). proposal 의 "exactly once" 문구를 "스레드당 최대 1회"로 읽을지, 후속 change 를 열지는
Manager·사람이 정한다.

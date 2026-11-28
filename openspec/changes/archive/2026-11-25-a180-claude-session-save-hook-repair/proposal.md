# a180 — Claude 세션 저장 훅 수리

> **절차 기록.** 위험도는 Small(개발 도구·시험, `docs/WORKFLOW.md` 「위험도 분류」)이라 change 가 필수는 아니다.
> 사용자가 2026-11-04 정식 절차를 선택해 만들었다. 구현은 2026-11-04 `/doctor` 후속으로 이 문서보다 **먼저**
> 작업 트리에 들어갔고(미커밋), 계약·증거 문서는 사후에 썼다. 이 역순을 숨기지 않고 `review.md` 에 적는다.

## Why

2026-11-04 전사 집계(19세션, 16.9일)에서 Claude 훅 두 개가 잘못 돌고 있었다.

- `./save-session.sh > /dev/null 2>&1`(PostToolUse `Edit|Write|Bash`)가 실패 1,697회를 냈다(종료 코드 127 1,593회,
  126 102회, 1 2회). TossOS 와 arkos 전사를 합친 수다. 상대경로이고 git 모드가 `100644` 다.
- 성공해도 도구 호출을 막았다. 백업 파일 시각으로 잰 성공 1회가 16·20·31초였고 제한 시간이 없었다.
  저장소 전체 `find` 가 비용 대부분이고, `openspec/changes/archive` 에 파일이 281,664개 있다.
- 요약 파일이 1.5MB 라 이어받기에 쓰기 어려웠고, 수동 메모는 저장할 때마다 지워졌다.
- `python3 tools/sdd-history/agent_save_hook.py` 도 같은 상대경로 형태로 차단 오류 17회를 냈다.

## What Changes

- 두 훅을 `"$CLAUDE_PROJECT_DIR"` 기준 경로로 부른다. `save-session.sh` 는 `bash` 로 직접 실행한다.
- PostToolUse 저장을 `async` 로 돌리고 `--hook` 인자를 넘긴다. `--hook` 은 다른 저장이 돌면 건너뛰고 5분 안에는 다시 저장하지 않는다.
- 턴 종료 저장을 위해 `Stop` 훅(`async`, `--hook` 없음)을 추가한다. 간격 제한으로 빠진 마지막 상태를 남긴다.
- 인자 없는 실행(수동·턴 종료)은 진행 중인 저장을 최대 120초 기다려 쓰고, 끝내 못 쓰면 종료 코드 1을 낸다.
- 요약의 `git status` 를 200줄로 자르고 생략 수를 적는다. 최근 파일 탐색은 `openspec/changes/archive` 로 들어가지 않는다.
- 수동 메모를 다음 요약으로 옮긴다.
- 저장 중 git 조회가 index 를 쓰지 않게 한다. 비동기로 돌리면서 생기는 `index.lock` 경합을 막는다.
- 요약을 임시 파일에 쓴 뒤 교체하고, 실패는 종료 코드와 `.ai-context/save-failures.log` 로 알린다.
- 시험 `tools/sdd-history/test_claude_save_session.py` 와 변이 하네스를 추가한다.

같은 작업 트리의 규칙 5 문구 수정(`.claude/CLAUDE.md`, `.codex/agents.md`의 `(위 절차)` 교체)은 문서 수정이라
change 불필요 대상이고 이 계약의 요구사항이 아니다. 커밋 범위를 정할 때 함께 다룬다.

## Non-goals

- 주문·위험·원장·인증 등 트레이딩 코드 변경 없음. Go 코드 변경 없음.
- Codex 저장기(`.codex/hooks/save_session.py`, `.codex/hooks.json`) 변경 없음.
- arkos 훅 수정 없음. arkos 받은편지함으로 별도 요청했다(메시지 `01a1a69d-758e-7c01-96bb-a91022da91e8`).
- 요약 파일에 대화 일부가 평문으로 남는 동작은 바꾸지 않는다. 기존 동작이며 사람 결정 대상이다(`design.md` 「위험」).

## Capabilities

### New Capabilities

- `claude-session-save`: Claude 세션 요약 훅의 호출 위치, 비차단 실행, 동시 실행·간격 제한, 요약 크기와 수동 메모 보존.

### Modified Capabilities

없음.

## Impact

- 변경 파일: `save-session.sh`, `.claude/settings.json`, `tools/sdd-history/test_claude_save_session.py`(신규).
- 런타임 영향: Claude Code 세션의 훅 실행 방식뿐이다. 엔진·콘솔·주문 경로 영향 없음.
- 안전 불변식 §0: 해당 경로를 건드리지 않는다.

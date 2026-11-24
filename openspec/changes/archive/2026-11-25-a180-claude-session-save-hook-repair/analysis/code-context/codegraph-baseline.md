# CodeGraph baseline

- 도구: codegraph 1.6.0
- 기준 HEAD: `232d6b86c024dfd71cc8ffe6fd0fbba342b3e9ec`
- 색인 상태: `python3 tools/sdd/sdd_sync.py --no-gbrain` 뒤 `check_index_freshness.py` 가 `CodeGraph hard-evidence index matches the worktree`, rc 0

## 질의

| 명령 | 결과 |
|---|---|
| `codegraph impact agent_save_hook` | 영향 1건: `tools/sdd-history/test_sdd_history.py` 의 import(9행) |
| `codegraph affected save-session.sh tools/sdd-history/agent_save_hook.py --filter 'test_*'` | `tools/sdd-history/test_sdd_history.py` 1건 |
| `codegraph callers agent_save_hook.run` | 심볼 없음 |
| `codegraph query save-session` | 관련 파일만 나오고 셸 스크립트의 호출 관계는 없음 |

## 한계

CodeGraph 는 훅 설정(JSON)이 스크립트를 부르는 관계를 모델링하지 않는다. 설정 결속은 `git grep` 으로 확인했다.

| 위치 | 내용 |
|---|---|
| `.claude/settings.json:29,38,60` | 이 change 가 고친 훅 3곳 |
| `.codex/hooks.json:9` | 같은 `agent_save_hook.py` 상대경로. 이 change 범위 밖 |
| `.codex/hooks/save_session.py:107,208,222` | `save-session.sh` 를 요약에서 제외할 뿐 호출하지 않음 |
| `tools/sdd-history/test_codex_session_save.py:84` | 임시 프로젝트 픽스처 |

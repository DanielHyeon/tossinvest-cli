# CodeGraphContext context

- 도구: CodeGraphContext 0.5.1
- 상태: `codegraphcontext update . --quiet` 가 300초 시한을 넘겨 실패했다. advisory 색인은 낡은 상태다.

## 질의

`codegraphcontext find name agent_save_hook` → 21건. 현재 저장소 외에 옛 격리 worktree(`~/tossos-gate/…`, `~/tossos-work/…`)의
같은 파일이 섞여 나온다. 새 호출자 후보는 없고 CodeGraph 의 `test_sdd_history.py` import 와 같은 결과다.

이 결과는 후보일 뿐이며 편집 근거로 쓰지 않았다.

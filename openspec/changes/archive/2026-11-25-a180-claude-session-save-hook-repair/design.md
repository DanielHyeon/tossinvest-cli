# a180 설계

## 결정

| 결정 | 이유 | 고르지 않은 안 |
|---|---|---|
| `bash "$CLAUDE_PROJECT_DIR/save-session.sh"` | 작업 폴더와 실행 비트에 모두 의존하지 않는다 | 실행 비트 커밋: 공유 index 를 건드리고 상대경로 문제는 남는다 |
| PostToolUse `async` + `--hook` | 도구 호출을 막지 않는다 | 동기 유지: 호출마다 16~31초 대기 |
| `flock -n` 단일 실행 | 비동기 실행이 겹쳐 전체 탐색이 쌓이는 것을 막는다 | 잠금 없음 |
| `--hook` 5분 간격 | 저장 1회가 수 초~수십 초라 호출마다 돌면 디스크를 계속 쓴다 | 더 짧은 간격: 이득 없이 부하만 늘어난다 |
| `Stop` 훅(간격 제한 없음) | 5분 간격이 버린 마지막 상태를 턴 끝에 남긴다 | 지연 저장 예약: 프로세스가 더 복잡해진다 |
| 수동·턴 종료는 `flock -w 120` | 저장 요청이 조용히 사라지면 안 되고, 못 쓰면 실패로 알려야 한다 | `flock -n` + exit 0: 쓰지 않고 성공처럼 끝난다(리뷰 지적) |
| `git status` 200줄 상한 | 미추적 9,905줄이 요약을 MB 단위로 키웠다. 변경 파일이 미추적보다 먼저 나오므로 앞부분을 남긴다 | 전량 유지 |
| `openspec/changes/archive` prune | 탐색 비용 대부분이다. 이름 기반 prune 확대는 측정상 효과가 없었다(1.66초 대 1.39~1.53초) | `.git` 등까지 prune 확대 |
| 수동 메모를 `awk` 로 이전, 제목 전체 일치로 경계 판정 | 메모 안내문을 두고도 저장마다 지웠다. 앞부분 일치는 비슷한 줄에서 메모를 잘랐다 | 백업에서 찾게 두기 |
| `GIT_OPTIONAL_LOCKS=0` + `git -c diff.autoRefreshIndex=false diff` | 비동기 저장의 git 조회가 `index.lock` 을 잡아 동시 `git add`·`commit` 을 실패시켰다(이 change 가 만든 회귀). 실측상 git 2.43 의 `git diff --stat` 은 `GIT_OPTIONAL_LOCKS=0` 을 무시하고 index 를 쓰므로 둘 다 필요하다 | 동기 복귀: 도구 호출을 다시 막는다 |
| 임시 파일 + `mv -f` | 쓰다 실패하면 요약이 잘리고, 다음 저장이 빈 메모를 안내문으로 바꿔 메모를 잃었다 | 제자리 쓰기 |
| `cp` 성공 시에만 백업 회전 | 실패 중에 회전하면 메모 든 백업이 밀려난다 | 무조건 회전 |
| 실패를 `.ai-context/save-failures.log` 에도 기록 | 훅 명령이 표준 오류를 버리므로 자동 경로 실패를 볼 곳이 없었다 | 표준 오류만 사용 |

## 스펙 충돌 검토

`openspec/specs/codex-session-save/spec.md:66` 의 `The change SHALL NOT modify .claude/settings.json or the root save-session.sh`
와 `openspec/specs/gbrain-codex-mcp-startup/spec.md:39` 는 시나리오가 그 change 의 기준선 대조로 한정된다.
Codex 저장기 도입 당시의 범위 조건이므로 이 change 를 막지 않는다고 읽었다. 문구 자체는 고치지 않는다.

`.codex/hooks/save_session.py` 는 `save-session.sh` 와 `.ai-context` 를 요약에서 제외하기만 하며, 이 변경으로 동작이 달라지지 않는다.

## 위험

- **평문 대화 보존.** `save-session.sh` 는 최근 대화 일부와 `MEMORY.md` 를 `.ai-context/session-summary.md` 에 평문으로 쓴다.
  gitignore 대상이지만 안전 부트스트랩 8(시크릿·계좌 정보를 로그에 남기지 않음)과 충돌할 수 있다. 기존 동작이고
  `Stop` 훅으로 저장이 잦아진다. 제거나 마스킹은 사람 결정 대상으로 남긴다.
- **저장 공백.** 턴 진행 중 세션이 끊기면 요약이 보통 5분까지 낡을 수 있고, 저장 실패나 잠금 건너뛰기가 겹치면 더 낡는다.
  `Stop` 훅은 턴이 끝날 때만 돈다. `/exit` 는 저장을 일으키지 않는다. 터미널 CLI 에서는 턴 끝에 시작한 비동기 저장이 `/exit` 뒤에도 끝까지 돌았다(표본 1건, `analysis/verification.md`). VS Code 패널 닫기는 재지 않았다.
- **다른 세션의 대화가 섞임.** 대화 수집은 가장 최근에 쓰인 전사를 읽고 요약 파일도 하나다. 같은 저장소에 세션이 둘이면
  다른 세션의 대화가 들어가고 마지막 저장이 이긴다. 기존 동작이다. 훅 입력의 `transcript_path` 로 고칠 수 있지만 실제 훅에서
  재지 않았고, 이어받기 정본은 `ai-memory` 라 이번 범위에서 고치지 않는다. 이 요약을 단일 세션 기록으로 믿으면 안 된다.
- **자식 프로세스의 잠금 상속.** `find`·`git` 이 잠금 FD 를 물려받으므로 자식이 멈추면 잠금이 안 풀린다. 그동안 `--hook` 은
  건너뛰고 수동·턴 종료 저장은 실패로 기록된다. shim 으로만 재현했고 실제 멈춤은 관측되지 않아 `unsupported` 로 둔다.
- **잘린 정보.** 요약의 `git status` 는 200줄까지이고, archive 안의 최근 수정 파일은 목록에 나오지 않는다.
- **같은 초 백업 덮어쓰기.** 백업 이름이 초 단위 mtime 이라 같은 초에 끝난 저장이 둘이면 백업 하나가 덮인다.
  저장에 수 초가 걸려 드물고 메모는 이전되므로 `unsupported` 로 둔다.
- **남는 임시 파일.** SIGKILL 이면 `trap` 이 돌지 않아 `.session-summary.*` 가 남을 수 있다. gitignore 대상이고 판정에 쓰이지 않는다.
- **기록되지 않는 실패.** 실패 로그는 잠금 대기 초과, 임시 파일 생성, 쓰기, 교체 실패만 다룬다. 백업 `cp` 실패와 git 조회 실패는
  종료 코드 0 으로 요약을 쓴다. git 이 전부 실패하면 git 구역이 비어 깨끗한 트리처럼 보인다. `.ai-context` 가 파일이면 `--hook` 은 조용히
  건너뛰고 수동 실행은 잠금 문구로 실패한다. 진단용 요약이라 `unsupported` 로 둔다.
- **파일 권한.** `mktemp` 때문에 요약과 백업이 POSIX 파일시스템에서 `0600` 이 된다(이전 `0644`). 평문 대화라 이 쪽이 낫다. 이 저장소의 fuseblk 에서는 권한 비트가 적용되지 않는다.
- **미검증 원인.** 저장소 루트로 기록된 작업 폴더에서도 127 이 났다. `cd` 뒤 작업 폴더가 바뀐 탓으로 추정하지만 재현하지 못했다.
  절대경로는 이 추정과 무관하게 실패 경로를 없앤다. 새 세션 전사에서 실패 수를 재서 확인해야 한다.

## 후속 (이 change 밖)

- `.codex/hooks.json:9` 도 `python3 tools/sdd-history/agent_save_hook.py` 상대경로다. Codex 실패 수는 재지 않았다.
- `.codex/hooks/save_session.py:188` 의 `run_git` 도 비동기이고 `timeout=3` 이다. 이 저장소의 `git status` 가 약 2.9초라 강제 종료되면
  `index.lock` 이 남을 수 있다. 읽기만 했고 재현하지 않았다.
- arkos 의 같은 훅과 `--mode hard` 게이트 동작은 arkos 세션 몫이다.

# a180 검증 기록 (2026-11-04)

기준 HEAD `232d6b86c`. 변경은 작업 트리에 미커밋 상태다.

## 실행한 명령과 결과

| 명령 | 결과 |
|---|---|
| `python3 -m unittest discover -s tools/sdd-history -p 'test_claude_save_session.py'` | 17건 OK |
| `python3 -m unittest discover -s tools/sdd-history -p 'test_codex_session_save.py'` | 7건 OK |
| `python3 -m unittest discover -s tools/sdd-history -p 'test_sdd_history.py'` | 15건 OK (`agent_save_hook` 유일 소비자) |
| `python3 openspec/changes/a180-claude-session-save-hook-repair/analysis/harness/mutate_save_session.py` | rc 0, 대조군 통과, 변이 31개 CAUGHT(스크립트 24, 설정 7). 입증 범위는 하네스에 적은 변이뿐이다. 설정 시험은 파일 값만 고정하고 실제 훅 동작은 재지 않는다 |
| `openspec validate a180-claude-session-save-hook-repair --strict` | valid |
| `python3 tools/pm/generate_master_tracker.py --check` | current |
| `python3 tools/sdd/check_agent_config_sync.py` | synchronized |
| `python3 scripts/memory_index.py check` | valid |
| `python3 tools/logic-map/check_analysis.py --change a180-claude-session-save-hook-repair` | required 0, rc 0 |

## 실제 저장소에서 잰 값

| 항목 | 값 |
|---|---|
| 수리 전 성공 1회 | 16·20·31초 (백업 파일 시각) |
| 수리 후 저장 1회 | 6.86초, 19.17초 (백그라운드) |
| `--hook` 건너뛰기 | 0.010~0.012초 |
| 요약 크기 | 1.5MB → 약 53~69KB |

`index.lock` 경합의 근거는 실제 저장소 측정이 아니라 `review.md` 「2차」의 `/tmp` 픽스처 재현이다. 실제 `.git/index` mtime 은 수리 전에도 바뀌지 않아 수리를 가르지 못한다.

이 값들은 표본이 작고 다른 작업과 겹친 시간대의 측정이다.

## tasks 4.1 실측 (2026-11-05)

측정 세션 `799b3922`(확장 2.1.316), 비교 세션 `4ebb1161`. HEAD `232d6b86c`. 시각은 KST.

| 항목 | 결과 | 모집단·한계 |
|---|---|---|
| 동기 훅 실패(`agent_save_hook`) | `hook_non_blocking_error`·차단 오류 0 | `4ebb1161` 의 Write·Edit 62건(새 `Stop` 훅이 처음 돈 11-04 21:31 이후) |
| 비동기 훅 실패(저장) | `async_hook_response` 0 | `4ebb1161` 의 Bash·Write·Edit 88건, `799b3922` 는 측정 시점 Bash 14건. 실패 응답은 다음 턴에 도착하므로 마지막 호출의 실패는 보이지 않는다 |
| 도구 호출 지연 | 도구 결과→다음 API 요청 중앙값 0.355초, 최대 1.06초(n=30). 새 세션 중앙값 0.182초, 최대 0.334초(n=5) | 병렬 호출은 마지막 결과만 셌다 |
| 저장과 겹친 표본 | 저장 03:06:35.899~03:06:45.527(9.6초) 동안 다음 API 요청이 결과 0.57초 뒤, 다음 도구 호출이 3.1초 뒤에 나갔다 | 표본 1건 |
| 턴 종료 저장 | `Stop` 직후 저장 완료까지 6초(23:02:50→56), 13초(23:09:01→14), 5초(02:24:45→50) | 백업 시각. 회전으로 3건만 남았다 |
| 실제 저장소 `index.lock` | `strace` 로 잰 저장 1회에서 `index.lock` 접근 0줄(전체 8,154줄), `.git/index` 쓰기 열기 0, sha256 불변, rc 0, 24.07초(`strace` 부하 포함) | 저장 자체가 잠금을 잡지 않음을 잰다. 동시 `git add` 경합은 돌리지 않았다 |
| 전사의 잠금 실패 | `fatal: Unable to create '...index.lock'` 0건 | 두 세션의 도구 결과 전수 |
| `save-failures.log` | 없음 | 잠금 대기 초과·임시 파일·쓰기·교체 실패만 기록하는 파일이다 |

**탐지 방법 대조군.**

- 동기 실패는 `hook_non_blocking_error` 로 남는다. 과거 전사의 `./save-session.sh` 실패가 `1711ae2d` 205건, `53f847a2` 38건, `a7cb125f` 6건 잡혔다.
- 비동기 실패는 `stop_hook_summary.hookErrors` 에 나오지 않는다. 임시 저장소(CLI 2.1.318)에서 없는 스크립트를 `async` 로 부르자 그 턴의 `hookErrors` 는 `[]` 였고, 다음 턴에 `async_hook_response` exitCode 127 이 남았다. 표본 1건이고 측정 세션과 판이 다르다.
- `strace` 필터는 stat-dirty 저장소의 `git status` 에서 `index.lock` 2줄을 잡았다.

**`/exit` 직후 저장.**

- **직전 세션에는 표본이 없다.** `4ebb1161` 은 VS Code 확장 세션이라 03:06:04 `/exit` 가 `/exit isn't available in this environment.` 로 끝났고, 프로세스 4083997 은 측정 시점에도 살아 있었다. 02:24:50 뒤 첫 저장은 이 측정 세션의 03:06:35 저장이다.
- **`/exit` 는 `Stop` 을 일으키지 않는다.** `Stop` 은 턴 끝에만 돈다. 아래 재현에서도 `stop_hook_summary` 는 턴 끝의 1건뿐이었다. 턴 없이 끝낸 세션에는 종료 저장이 없다.
- **터미널 CLI 재현.** 임시 저장소에서 CLI 2.1.318 을 tmux 로 띄워, `Stop` 훅이 시작되고 0.45초 뒤 실제 `/exit` 를 보냈다. CLI 는 시작 1.31초 뒤 끝났고, 20초 대기 후 저장 사본을 부르는 비동기 훅이 끝까지 돌았다. 요약은 CLI 종료 18.74초 뒤에 쓰였고 rc 0, 실패 로그 없음이다. 표본 1건이고 실제 저장소가 아니다.
- **재지 않은 것.** VS Code 패널 닫기처럼 프로세스가 강제로 끝나는 경우.

memory recall 은 `not-applicable` 이다. 이 측정의 입력은 이전 세션 전사 원본과 현재 파일이고, 기억에서 가져온 값을 쓰지 않았다.

## 선택 이유

CodeGraph 영향이 `test_sdd_history.py` 하나이고, Codex 저장기와의 공존은 `test_codex_session_save.py` 가 지킨다.

## 실행하지 않은 범위

- `make gate`, `make test`, `make test-seams`, `make test-race`, `make sdd-check`, `make sdd-test`: 전체 범위라 사용자 명시 요청이 필요하다.
- VS Code 패널 닫기로 끝나는 세션의 저장, 실제 저장소의 동시 `git add` 경합.
- Codex 훅과 arkos 훅.

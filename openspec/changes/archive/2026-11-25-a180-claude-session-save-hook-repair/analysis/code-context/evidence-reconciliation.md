# Evidence reconciliation

- 기준 commit: `232d6b86c024dfd71cc8ffe6fd0fbba342b3e9ec`
- 변경 표면: `save-session.sh`, `.claude/settings.json`, `tools/sdd-history/test_claude_save_session.py`

| 항목 | CodeGraph | CodeGraphContext | 현재 HEAD·실행 | 결론 |
|---|---|---|---|---|
| `agent_save_hook` 소비자 | `test_sdd_history.py` 1건 | 같은 파일 + 옛 worktree 사본 | `git grep` 결과 동일 | 일치. 모듈은 고치지 않고 호출 경로만 바꿈 |
| 훅 설정 결속 | 모델링 안 함 | 모델링 안 함 | `git grep` 으로 확인 | HEAD 기준 |
| `save-session.sh` 호출자 | 없음 | 조회 안 함 | 훅 2곳(`.claude/settings.json:38,60`)과 시험 픽스처 | HEAD 기준 |

불일치로 편집을 막아야 할 항목은 없다.

## 미갱신 advisory

- CodeGraphContext: `update` 300초 초과.
- GBrain: `python3 tools/sdd/gbrain_project.py search save-session` 이 rc 75, `[gbrain-project] busy`. `docs/WORKFLOW.md` 의 busy 예외에 해당한다.
- `make sdd-sync` 대신 `sdd_sync.py --no-gbrain` 을 썼다. GBrain 단계가 체크아웃을 `pull --ff-only` 로 옮긴 전례(2026-10-23 기록)가 있고, 공유 `main` 체크아웃이기 때문이다.

## 기억 회고

- `scripts/memory-recall.sh save-session`: 결과 0건.
- ai-memory `save-session OR session-summary OR .ai-context`: 세션 기록 스니펫뿐이고 결정·함정 페이지 없음.
- Hindsight `tossos-dev-episodes`: 이 훅에 관한 기록 없음.

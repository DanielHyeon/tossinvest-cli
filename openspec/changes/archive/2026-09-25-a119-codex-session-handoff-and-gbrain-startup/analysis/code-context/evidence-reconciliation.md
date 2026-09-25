# a119 증거 조정 (task 2.1)

- base commit: `54004f4478a3f2d87c9b0dd371dd65fdcd7599ff` (`capture_change_base.py`, 2026-09-25)
- CodeGraph: **1.6.0**, 색인 파일 2,122 · 노드 38,426 · 엣지 146,390 (`codegraph status .`)
- CodeGraphContext: 질의 `codegraphcontext find name save_session` → StockOS 경로만 반환
  (공유 색인에 TossOS 결과 없음). 보조 문맥으로 쓸 결과 0 — advisory 계층 **not-applicable**, 결론에 쓰지 않음.
- GBrain: `python3 tools/sdd/gbrain_project.py search "codex hook"` → **exit 75**
  `[gbrain-project] busy: owner pid=… command='gbrain serve'` (다른 Claude 세션이 소유).
  정본 busy 계약대로 advisory 없이 진행 — **not-applicable(busy)**.

## 확인한 변경 표면

| 표면 | 소유 | 이 change 에서 |
| --- | --- | --- |
| `.codex/hooks.json` | Codex | 읽기만. matcher 변경 근거 없음(host-evidence §2.3) |
| `.codex/config.toml` | Codex | 읽기만. 유효 `gbrain` 등록이 이미 1개(host-evidence §3.1) |
| `.codex/hooks/save_session.py` | Codex | 읽기만. 함수 내부 편집 없음 |
| `tools/sdd/gbrain_project.py` | 공유 wrapper | 읽기만. 락·busy 동작 불변(스펙 요구) |
| `.mcp.json`, `.claude/**`, `save-session.sh` | Claude | 손대지 않음 |
| `tools/sdd-history/test_codex_host_event_coverage.py` | 시험(신규) | 추가 |
| `tools/sdd/test_codex_gbrain_registration.py` | 시험(신규) | 추가 |
| `tools/sdd-history/fixtures/`, `tools/sdd/fixtures/` | 살균 픽스처(신규) | 추가 |

## CodeGraph 질의와 결과

| 질의 | 결과 |
| --- | --- |
| `codegraph query "save_session publish run"` | `run` `.codex/hooks/save_session.py:414`, `publish` `:401` |
| `codegraph callers publish` (save_session 한정) | `run` `:414` 하나 |
| `codegraph callers acquire_process_lock` | `main` `tools/sdd/gbrain_project.py:182` 하나 |
| `codegraph query "codex hooks.json matcher test"` | `test_preserves_sdd_hook_and_adds_async_codex_matcher` `tools/sdd-history/test_codex_session_save.py:280` |
| `codegraph affected .codex/hooks/save_session.py tools/sdd/gbrain_project.py --filter '*test*.py'` | `tools/sdd/test_gbrain_project.py`, `test_sdd_doctor.py`, `test_sdd_sync.py` (3) |

## 불일치와 현재 HEAD 로 내린 결론

1. `codegraph affected` 는 `save_session.py` 의 시험을 못 찾는다. 시험이 저장기를 import 하지 않고
   `subprocess` 로 경로 실행하기 때문이다(`test_codex_session_save.py:14` `SCRIPT = ROOT/…`). 현재 HEAD 에서
   직접 읽어 `tools/sdd-history/test_codex_session_save.py` 가 저장기의 격리·수정·원자 저장·락 시험임을 확인함.
2. `.codex/hooks.json`·`.codex/config.toml` 은 코드 노드가 아니어서 CodeGraph 질의 대상이 아니다.
   설정의 효력은 호스트 로더(`codex mcp list --json`)로 잼(host-evidence §3.1).
3. 기존 함수 내부를 바꾸지 않으므로 호출 사슬 영향은 0 이다. 새 시험 파일 둘은 신규 leaf 이다.

## Function Logic Map

Function Logic Map: not-applicable — 이 change 는 Go 함수를 바꾸지 않고, 기존 Python 함수
(`save_session.py`·`gbrain_project.py`·기존 시험 메서드)의 본문도 바꾸지 않는다. 추가되는 것은 신규
시험 파일 두 개와 살균 픽스처 JSON 두 개뿐이다. 분기·early return 을 근거로 삼는 주장도 하지 않는다
(설정 효력 주장은 호스트 로더 측정으로만 한다).

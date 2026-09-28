# a125 review

## Proposal-freeze — 2026-09-29 (적대 Eng 보이스 1 · codex 는 Manager 슬롯 대기열)

- 구성: 게이트 판정 변경이므로 적대 Eng 보이스 1(독립 서브에이전트, 읽기 전용). 첫 실행(2026-09-28)은 세션 재시작으로
  handback 없이 끝났고 재실행했다. codex freeze 는 Manager 슬롯 대기열에 있다(a094 r5 다음).
- **판정: freeze 가능 — P0 없음.** P1 다섯은 아래처럼 design · tasks 에 반영했다(구현 전).

| # | 등급 | 발견 | 처분 |
|---|---|---|---|
| 1 | P2 | 구멍 없음. 기록 보유는 a063 하나이고 a063 에서 빌리는 change 는 0 이다. `SDD_BASE_REF` 는 persisted 하나로 좁아져 더 엄해진다. 비-a063 에 기록이 생기면 옛 도구는 fail-closed, 새 도구는 무시한다(D2 승인 사항). | 수용 |
| 2 | P2 | 델타 정합(scenario 48 = 49 − 6 + 2 + 3). RENAMED+MODIFIED 의 archive 적용 순서는 미검증이다. | scratch archive dry-run 으로 확인(아래 「archive dry-run」) |
| 3 | P2 | 남은 참조: `check_analysis.py` docstring 의 자식 프로세스 자리 수 주장(`execution_baseline.validate`) · README:95 · WORKFLOW:241 · README 247-290 · `fixture_git_env.py:3`. Makefile · CI 는 discover 라 무변경이다. `SDD_PYTHON` 은 `tools/sdd/test_sdd_doctor.py` 가 덮으므로 skip 시험(456) 삭제의 손실은 0 이다. | 2.1 · 2.3 에 포함 |
| 4 | **P1** | *(실측으로 일부 정정 — design D4-2: revision 은 base 가 맞고 결함은 soak_test 두 번들의 E 해시)* 재고정 조건 ① 이 오늘은 **성립하지 않는다.** base 번들 3 은 E(=`676bd4b4^`)의 파일 해시를 적는데, 세 함수는 오늘 존재하므로 P 기준으로는 current 여야 한다. 재고정 뒤에는 `676bd4b4` 가 창 밖이고, `validate_target` 이 base 번들 해시를 안 보므로 어떤 게이트도 ① 을 다시 재지 않는다. | tasks 4.2 를 바꿨다: base 3 을 current 로 재추출 → 특례 없는 도구로 **옛 base P 에서** `check_analysis` → 출력에서 9 함수 이름의 missing · stale · revision 오류 0 을 필터 영수증으로 남김 → 그다음 재고정. proposal 문구도 "재추출 후 해당" 으로 고침 |
| 5 | **P1** | 소급 고지가 사라진다. `retrospective-exception` · "missing original analysis remains debt" 를 적는 유일한 기계 기록이 삭제되는 `execution-baseline.json` 이다. | tasks 4.1 에 추가: a063 `review.md` 영수증에 그 두 사실을 명시하고, a063 `tasks.md` 4.0 에 "a125 에서 폐기" 주석 |
| 6 | **P1** | 재고정 시점. a063 의 4.2~4.4(사람 운영, 수 일~수 주)가 끝나기 전에 형제 Go 가 착지하면 창에 다시 요구되고, current 번들 6 의 파일 단위 해시가 낡는다. 둘째 재고정 · 재추출이 거의 확실하다. | **Manager 결정 (나)(2026-09-29)**: a125 는 ① 영수증까지, 재고정은 a063 게이트 직전에 한 번(tasks 4.3 · design D4-3) |
| 7 | P2 | `c727ad12`(S)는 HEAD 조상이 아니고, `676bd4b4` 가 main 사본이다. `0c563c6c`·`b8f31f27` "[a063]" 은 renumber 전 **다른** a063(현 a069) 것이다. 현재 a063 디렉터리는 `47a7f90a` 에서 태어났고 옛 이름이 없다. | 4.2 영수증에 적는다 |
| 8 | **P1** | "같은 픽스처로" 반전은 모듈 삭제 뒤 ImportError(`adoption.draft`). 반전 시험은 리터럴 a063 id 를 써야 id 우회 변이를 잡는다. | 이미 반영: RED 는 모듈 없이 세운 정적 기록 픽스처(`_a063_fixture`, 기록의 `execution_base` = 실제 픽스처 커밋) · 리터럴 `A063`. design D3 에 명기 |
| 9 | **P1** | 빠진 시험 · 변이: 기록이 FIFO · 디렉터리 · 심링크 · 해독 불가여도 일반 판정. 변이는 기록 `execution_base` 를 base 로 읽기 · id 로 착지 우회 · `_recording_refusal` 거절 복원. 시한 시험의 모듈 목록을 정확한 집합으로. | 1.2 보강(비정규 기록 시험) · 3.2 변이 목록에 추가 · 시한 시험은 디렉터리의 `subprocess` 쓰는 비시험 모듈 전부와 대조 |
| 10 | P2 | `effective_base` 는 `main` 창 줄이 쓴다 — 지우면 안 된다. 반전의 "이관 키 없음" 은 `execution_baseline_adoption` · `adoption_source` 둘로 한정한다. | 이미 그렇게 구현(초안). 반전 시험도 그 둘만 단언 |

미검증(보이스 자신이 적음): archive 적용 순서(→ dry-run), 착지 기록으로 a063 창을 고정할 수 있는지(6), 시험 밖 경로가
change 디렉터리 파일을 전수로 읽는지.

## Proposal-freeze — codex 교차 모델 (2026-09-29)

- 실행: codex-cli 0.154.0 · gpt-6-astra · `codex exec -s read-only --ephemeral --skip-git-repo-check -C <트리>`,
  session `01a0e8c3-0530-7752-9b53-174964602953`, 01:04:49~01:07:54 KST, rc 0. 트리 = HEAD `7b9df40f` 의 `git archive` export.
  프롬프트 `analysis/freeze-review/codex-prompt.md` sha256 `7e59fd3e…` (실행 사본 일치). 원문 `analysis/freeze-review/codex-output.md`.
- **판정: APPROVE-WITH-FIXES — P0 없음.** 현재 모집단에서 판정 · CLI 가 바뀌는 것은 a063 하나. 일반 기준 · 착지 · 빌림 ·
  `SDD_BASE_REF` 검사 유지.

| # | 등급 | 발견 | 처분 |
|---|---|---|---|
| C1 | P1 | "9 함수 오류 0" 이름 필터는 성공을 보장하지 않는다 — 조기 실패는 함수 이름 없이, 해시 · revision 오류는 **번들 디렉터리명**으로 나온다. | tasks 4.2 · design D4-4 를 양성 단언으로(판정 완료 · P/HEAD/창 · 함수↔번들 대응 · 번들 전체 오류 0 · 나머지는 형제 누락만) |
| C2 | P1 | 결정 (나) 이면 게이트 직전 재실측이 필요하다 — 영수증 뒤 10 번째 자기 함수가 생기면 재고정이 그것을 창 밖으로 뺀다. | tasks 4.3 · design D4-3: a063 tasks 에 게이트 직전 재고정 항목(P→당시 HEAD 귀속 재실측 · 번들 재검증 · 분리 커밋 사람 귀속) |
| C3 | P1 | 반전 시험이 착지 우회 · 안내 행동을 다 고정하지 못한다. | 시험 셋 추가: 계산값이 아닌 유효 착지 기록 셋(P · E · 없는 sha) 거절, 착지 뒤 자기 Go 수정 거절, a063 에 `--record-landing` 안내 양성 단언. `reversal-targets.md` TBD 는 GREEN 커밋에서 채운다 |
| C4 | P1 | 불변 주장의 비교 계약이 불명확하다. 논증 핀은 guard 를 출력만 하고, census 는 CLI 를 안 잰다. | design D5 신설(범위 = 현재 모집단, 계약 명시). `ab_absent_record.py` guard **단언**으로. `ab_old_new.py` 로 옛/새(GREEN 초안) A/B 세 모양 — 차이 0(`harness/1.1-ab-old-new-compare.txt`) |
| C5 | P2 | proposal 의 델타 개수 오기 · 델타 「불변 base」와 재고정 절차의 관계 · a063 issues 갱신 누락. | proposal 개수 정정 · spec delta 에 재고정 절차 관계 한 문장 · a063 issues 는 tasks 4.3 |

## archive dry-run (freeze F2 미검증 항목)

scratch 사본에서 `openspec archive a125-… --yes` 를 돌렸다. 결과는 `sdd-workflow: ~ 2 modified · - 3 removed · → 1 renamed` 이고,
옛 요건 이름 넷의 잔존은 0 이다. 새 시나리오 둘과 새 요건 이름이 적용됐다(2026-09-29).

## 3.2 변이 — 2026-09-29 (`harness/mutate.py`, 사본 · pid 디렉터리 · 무변이 대조군 선행)

원본은 `9e63b681` 의 `check_analysis.py` 사본이다(sha256 대조). 무변이 대조군이 **GREEN** 인 것을 먼저 확인했다. 변이마다 치환 대상이
정확히 한 번 있었음을 단언해 "닿았다" 를 확인했고, 시험 묶음은 `TheA063ExceptionIsRetired` · 문맥 재사용 · 시한 시험이다.

| 변이 | 결과 |
|---|---|
| M1 기록의 `execution_base` 를 base 로 읽기 | CAUGHT |
| M2 a063 착지 기록 거절 복원(판정) | CAUGHT |
| M3 a063 기록 명령 거절 복원(`_recording_refusal`) | CAUGHT |
| M4 id 로 착지 우회(검증 없이 수락) | CAUGHT |
| M5 비정규 기록 거절 복원 | CAUGHT |
| M6 `SDD_BASE_REF` 대조 끔 | CAUGHT |
| M7 이관 라벨 복원(출력) | CAUGHT |
| M8 옛 문맥 키 되살림 | CAUGHT |

**8/8 CAUGHT.** 원장은 `harness/3.2-mutate.log` · `3.2-mutate.json` 이다.

- `not-applicable`: `_verdict` 의 prefetch `except ValueError: sources = []` 는 변이 대상에서 뺐다.
  - 이 갈래는 착지가 있을 때만 서는데, 착지는 같은 `evidence` 로 `_select_pinning` 을 먼저 통과해야 생긴다. 그래서 남는 경우는 두 호출 사이에 심링크 풀이가 바뀌는 경합 하나뿐이다.
  - 이 갈래를 지우는 변이는 판정이 같은 최적화 생략이라 살아남을 것이 확실하다.
  - 갈래는 방어로 남긴다. `_verdict` 가 `_judged` 의 try 밖에서 불리기 때문이다(gstack 리뷰 P2-3).

## gstack 코드 리뷰 — 2026-09-29 (독립 서브에이전트, 커밋 `9e63b681`)

- **APPROVE-WITH-FIXES — P0 · P1 없음.**
- 확인한 것:
  - 지운 키 · 상수 · 인자를 읽거나 넘기는 코드가 0 이다.
  - 비-a063 의 창 · 착지 · 빌림 · `SDD_BASE_REF` 문장이 글자까지 같다.
  - 통째로 지운 시험 하나(`SDD_PYTHON` skip 시험)가 지키던 동작은 `tools/sdd/test_sdd_doctor.py:73-76` 이 덮는다.
- P2 넷:
  1. `check()` docstring 두 자리가 지운 모듈을 현재 판정 입력으로 열거했다 → "a125 에서 삭제" 를 달았다.
  2. 시험 파일 주석(:31)이 지운 모듈을 가리켰다 → 과거형으로 고쳤다.
  3. 도달하지 않는 방어 갈래 → 위 변이 절에 `not-applicable` 사유를 적었다.
  4. 3.x 미체크 → 이번 로트로 닫는다.
- 참고: 핀 시험의 `ADOPTION_WORDS` 는 밑줄 표기 `execution_baseline` 을 세지 않는다. docstring 의 과거 기록 두 자리가 그 표기로 남는 것은 의도다.

## 3.3 · 4.x 상태 — 2026-09-29

- 편집 후 Python FLM 7 재추출(`5440edad` 판, 분기 차이는 편집 전 목록 + `resolve_base` 정정 한 줄로 설명됨) — `analysis/python-function-logic/*/function-logic-map.md`.
- `make sdd-test`(고정 워크트리 `a125-probe@76d0816a`): 스크립트 15 · logic-map 503 · sdd 76 · sdd-history 29 · pm 16 · deploy 18 **전부 OK**(10m52s) — `harness/3.3-sdd-test.txt`.
- `make sdd-check`(공유 트리): **FAIL — 환경 부하**. `codegraph status .` 가 15 초 탐침 시한을 넘었다(같은 명령 단독 25 초, load 6.3). `make sdd-sync` 도 advisory `codegraphcontext update` 300 초 시한으로 incomplete. 코드 판정 실패가 아니라 탐침 시한 — 두 번 재시도해 같다(`harness/3.3-sdd-check-attempt.txt`). 게이트 ⑥ 에서 격리 워크트리로 다시 잰다. 그 전까지 3.3 은 미체크.
- 4.1~4.3: a063 전환 커밋 `76d0816a` · 영수증 PASS(`analysis/a063-receipt.md`).

## 완료 게이트 — 격리 워크트리 `TossOS-worktrees/a125-gate`

- 준비: `make sdd-infra`(워크트리 로컬 `.sdd/.venv`) → `make sdd-sync` rc 2 — advisory 만 incomplete(`codegraphcontext index` 300 초 시한,
  gbrain 은 공유 홈의 YAML 실패 6127 건 — 이 change 와 무관). CodeGraph(hard evidence) 는 `codegraph init .` 로 세웠다.
- 1차 `make gate` @ `e62522ce` — rc 2, **② 에서 5.1 한 줄만 미완료**(예정된 멈춤, `harness/5.1-gate1.log`). 5.1 체크 → tracker 재생성 → 2차.

## 사람 승인 base 재고정 — 2026-09-29

- **Function Logic Map: not-applicable** — 이 change 의 자기 Go 편집은 0 이다(Python 게이트 도구 · 문서만). Python 함수 증거는
  `analysis/python-function-logic` 에 있다.
- 사유: 2차 gate(`6baafd5c`) ⑤ 가 base `c2eec627` → 워킹트리 창의 **형제** Go 함수 9 개를 요구했다(`harness/5.1-gate2.log`). a125 는
  current 번들이 0 이라 착지로 좁힐 수 없다.
- 절차 조건(`docs/WORKFLOW.md` 「사람 승인 base 재고정」):
  1. 귀속 실측: base 뒤 a125 디렉터리를 만진 커밋 14 개의 `.go` 편집은 0 이고, `_self_repair_commits` 도 `[]` 이다. 옛 디렉터리명은 없다.
  2. 승인: 사용자 일괄 승인(2026-09-28, a063 수리안 = 특례 폐기 → 일반 경로)과 Manager 승인(2026-09-29, 이 재고정)이다.
  3. `base-commit.txt` 를 단독 커밋했고, 영수증은 그 커밋 메시지다.

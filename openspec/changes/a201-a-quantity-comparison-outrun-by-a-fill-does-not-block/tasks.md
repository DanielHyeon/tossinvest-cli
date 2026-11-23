# a201 tasks — 체결이 지나간 뒤의 낡은 수량 비교의 재판정

위험 등급 High-risk. 단계를 건너뛰지 않는다. 각 체크는 그 산출물과 같은 커밋에서 한다.

## 1. 등록 · 증거 (1단계)

- [x] 1.1 change 디렉터리 등록: `.openspec.yaml` · proposal · design · tasks · Story `STORY-TOS-a201.yaml` · `_registry.yaml` · tracker 재생성
- [x] 1.2 `base-commit.txt`(change 루트 — `tools/sdd/capture_change_base.py`)
- [x] 1.3 사전 증거를 현재 HEAD 로 재확인해 `analysis/evidence-phase0.md` 로(scratch `a201-evidence/` — AST 재추출, 달라진 곳 표) + 운영 집계(Manager 제공)
- [x] 1.4 **실제 `RunOnce` 재현**: 체결 커밋을 주기 안 정해진 지점(로컬 읽기 뒤 · `Positions` 읽기 뒤 · 조정 tx 직전)에 끼워 넣는 방법을 찾고, W2 · W3(두 종류)를 `RunOnce` 로 만들어 편집 전 코드에서 QM 행 1 이 여러 주기 뒤에도 미해제임을 확인(원문 `analysis/repro/`). 끼워 넣을 방법이 없으면 멈춘다
- [x] 1.5 design 전제 확인(거짓이면 멈춤): (a) `RunOnce` 가 수렴 결과에서 「조정이 반영된 종목」을 알 수 있다(반환값의 필드 — 없으면 무엇이 있는가) (b) 종목별 체결 워터마크와 전역 최대 체결 id 를 `RunOnce` 가 부를 수 있다(계좌 조건 없는 종목 워터마크가 다른 계좌의 체결에 움직이는가 — 단일 계좌 전제의 근거) (c) 그 종목의 로컬 보유 수량을 다시 읽을 기존 함수와 `Compare` 의 일치 판정 함수(공개 여부 · 호출 자리) (d) `Observe` 의 생산 호출자 수 (e) 비교(diff)에서 항목을 뺀 사본을 만들 때 `Observe` 가 의존하는 다른 필드(digest · as-of · 승격 키)가 그 항목과 묶여 있는가 (f) D5 — 수량만 뺐을 때 missing 블록이 서는가, a192 경로로 몇 주기에 풀리는가
- [x] 1.6 CodeGraph hard evidence(버전 기재): `RunOnce` · `ConvergeQuantities` · `Observe` · `blocksFor` · `LocalStateFromJournal` · `Compare` 의 호출자와 호출 **자리 수**
- [x] 1.7 편집 표면 · 고정 장치 표(`analysis/edit-surface.md`) — 편집할 기존 함수(목표 하나), 구조 시험·census 적중, 진행 중 change(a199 — 같은 파일의 다른 함수, a200, a197, a087)와의 겹침
- [x] 1.8 재판정과 `Observe` 사이의 구간에 끼는 호출 열거(잔여 창의 폭 — design §4)
- [x] 1.9 정본 delta: reconciliation ADDED(재판정 규칙 — design D3·D4·D8, §2 경계). :53 · :77 · :79 · :91 · a192 Requirement 의 문장과 대조해 모순이면 그 Requirement 를 MODIFIED(생성 스크립트 + 역치환 단언). ADDED 본문 500자 이하. `openspec validate --strict` valid
- [x] 1.10 편집할 기존 함수마다 FLM/BTM pre-edit 번들
- [x] 1.11 claim check(`analysis/claim-check.md`)
- [ ] 1.12 제안 동결 적대 리뷰(codex, 최대 2회차) → `review.md`
- [ ] 1.13 Manager 결정 · freeze

## 2. 구현

- [ ] 2.1 Pre-Edit 선언(High-risk)
- [ ] 2.2 RED — design §3 의 1~6
- [ ] 2.3 GREEN 최소 구현 — 재판정 함수(새 파일) · `RunOnce` 1자리. `Observe` · `ConvergeQuantities` · `Compare` · 해제 규칙 · 스키마 무편집. 주석은 한글로 「왜」
- [ ] 2.4 post-edit 번들 · `check_analysis`
- [ ] 2.5 변이 원장(design §3-7)
- [ ] 2.6 선택 시험 · `-race`(해당) · `make lint` · strict validate · tracker `--check`

## 3. 리뷰 · 검증

- [ ] 3.1 구현 적대 리뷰(별도 문맥, Opus + codex, 최대 2회차) — 초점: 빼면 안 되는 불일치가 빠지는 입력 · 차단이 늦어지는 상한 · 승격 횟수 · 재판정과 기록 사이의 창
- [ ] 3.2 처분 반영
- [ ] 3.3 gstack `review` 1회
- [ ] 3.4 Manager 검증

## 4. 종결

- [ ] 4.1 착지 기록 → `origin/main` merge(rebase 금지) → 완료 게이트(격리 worktree) → archive → Story 경로 → push

## 5. 종결 뒤 (사람 승인 — 체크박스 아님)

- 운영 반영은 사람이 승인한다. 반영 뒤 확인은 관측으로: 운영 로그의 `stale_local_read` 줄과, 그 뒤 그 종목에 차단 행이 없는 것.

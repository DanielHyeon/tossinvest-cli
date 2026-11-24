# a201 tasks — 체결이 지나간 뒤의 낡은 수량 비교로 선 차단의 체결 credit 해제(v3)

위험 등급 High-risk. 단계를 건너뛰지 않는다. 각 체크는 그 산출물과 같은 커밋에서 한다.
이력: v1(비교에서 빼기) · v2(주기 보류)는 동결 리뷰 반례로 기각 → 보류(2026-11-24, `issues.md` I1) → 같은 날 사용자 지시로 v3(해제 쪽) 재개. v1 · v2 시기의 1단계 산출물은 그대로 증거로 쓴다(`analysis/` — `claim-check.md` · `claim-check-v2.md` · `probe-d5/` · `probe-v2/` 는 기각된 설계의 기록). codex 회차는 v3 부터 새로 센다.

## 1. 등록 · 증거 (1단계)

- [x] 1.1 change 디렉터리 등록: `.openspec.yaml` · proposal · design · tasks · Story `STORY-TOS-a201.yaml` · `_registry.yaml` · tracker 재생성
- [x] 1.2 `base-commit.txt`(change 루트) = `f9b818a0e`. v3 재개 때 `origin/main`(`d66536d9b`, a200 착지) merge — 자기 production Go 커밋 0 이라 base 유지. a200 의 `ProjectPosition` 은 착지 기록으로 창 밖에 둠(종결 때)
- [x] 1.3 사전 증거 재확인(`analysis/evidence-phase0.md`) + 운영 집계 + v3 증거 조사(`analysis/v3-evidence/` — 별도 팀메이트, AST 20 · 탐침 · CodeGraph 1.6.0)
- [x] 1.4 실제 `RunOnce` 재현(`analysis/repro/` — 「현재 동작」 6: W2 · W3-a · tx 직전 · W3-b 합성 · 앞 크레딧 위 W2 · 판 중단 뒤 B)
- [x] 1.5 v3 전제 확인 (q1)~(q7)(`analysis/claim-check-v3.md`) + v3 모양 탐침(`analysis/v3-probe/` — scratch 사본, 제거)
- [x] 1.6 CodeGraph hard evidence(1.6.0 — `analysis/code-context/` · `analysis/v3-evidence/codegraph-*.txt`)
- [x] 1.7 편집 표면 · 고정 장치 표(`analysis/edit-surface-v3.md`) — 편집할 기존 함수 넷, 구조 시험 · census 적중, 진행 중 change 겹침
- [x] 1.8 정본 delta: reconciliation ADDED 1(체결 credit, 본문 410자) + MODIFIED 3(생성 스크립트 `analysis/modified-delta/generate.py` — 치환 1회 적중 · 역치환 · Scenario 이름 보존 단언, 원문 `run.txt`). `openspec validate --strict` valid
- [x] 1.9 편집할 기존 함수마다 FLM/BTM pre-edit 번들(RunOnce · Observe · persist · ValidReconcileReleaseCause), `check_analysis`
- [ ] 1.10 제안 동결 적대 리뷰(codex, v3 회차 새로 셈 — 최대 2회차) → `review.md`
- [ ] 1.11 Manager freeze(이번에는 Manager 가 최종 문서를 읽고 확정)

## 2. 구현

- [ ] 2.1 `origin/main` 재확인(a199 착지 시 merge — `reconcileloop.go` 가 바뀌면 RunOnce pre-edit 번들 재추출) · Pre-Edit 선언(High-risk)
- [ ] 2.2 RED — design §3 의 1~11(1단계 재현 6 을 v3 계약으로 뒤집는 것이 첫 줄) — 원문 `analysis/tdd/red.txt`
- [ ] 2.3 GREEN 최소 구현 — 새 파일 둘 · `RunOnce` 2 자리 · `Observe` 1 문장 · `persist` 원인 갈래 · `ValidReconcileReleaseCause` case 1. 주석은 한글로 「왜」 — 원문 `analysis/tdd/green.txt`
- [ ] 2.4 post-edit 번들 · `check_analysis`(Observe 는 옛/새 ast 정렬 재번호)
- [ ] 2.5 변이 원장(design §3-12)
- [ ] 2.6 선택 시험 · `-race`(해당) · `make lint` · strict validate · tracker `--check`

## 3. 리뷰 · 검증

- [ ] 3.1 구현 적대 리뷰(별도 문맥, Opus + codex, 최대 2회차) — 초점: V8 공시 밖의 틀린 해제 · credit 수명의 구멍 · 그 주기 Observe 무변 · a192 결속 행
- [ ] 3.2 처분 반영
- [ ] 3.3 gstack `review` 1회
- [ ] 3.4 Manager 검증

## 4. 종결

- [ ] 4.1 착지 기록 → `origin/main` merge(rebase 금지) → 완료 게이트(격리 worktree) → archive → Story 경로 → push

## 5. 종결 뒤 (사람 승인 — 체크박스 아님)

- 운영 반영은 사람이 승인한다. 반영 뒤 확인은 관측으로: 다음에 같은 경합이 났을 때 그 종목 차단 행이 `FILL_APPLIED` 로 풀리는 것.

# a199 tasks — 낡은 capability 증명은 진입만 막고 청산은 막지 않는다

위험 등급 High-risk. 단계를 건너뛰지 않는다. 각 체크는 그 산출물과 같은 커밋에서 한다.

## 1. 등록 · 증거 (1단계)

- [x] 1.1 change 디렉터리 등록: `.openspec.yaml` · proposal · design · tasks · `specs/engine-safety/spec.md`(spec-delta-intent 대로 조립, MODIFIED 는 생성 스크립트 + 역치환 단언) · Story `STORY-TOS-a199.yaml` · `_registry.yaml` · tracker 재생성. `openspec validate --strict` valid
- [x] 1.2 `base-commit.txt`(change 루트 — `tools/sdd/capture_change_base.py` 가 쓰는 자리; 초안의 `analysis/` 는 1단계 사실 정정)
- [x] 1.3 사전 증거를 현재 HEAD 로 재확인해 `analysis/evidence-phase0.md` 로 옮김(scratch 의 `a199-evidence/` — 줄 번호·AST 분기 재추출, 달라진 곳 표)
- [x] 1.4 CodeGraph hard evidence(버전 기재): `verifyGate` · `runInterlock` · `NewContext` · `runEngineRun` · `ExitObserver` 생성 · `ReconcileDriver` 생성 · `judgeAttestation` 의 호출자와 호출 **자리 수**
- [x] 1.5 **`Automation.Verified` 를 읽는 자리 전수**(자리 수, 파일:줄, 진입 쪽/청산 쪽/보고 분류) — D3 의 편집 표면을 여기서 확정. `EntryPermitted` 도 같이
- [x] 1.6 계좌 일치 비교의 공개 진입점 유무(`internal/attest`) — D1a 의 구현 자리 확정. `Verify` 본문 무편집으로 되는지
- [x] 1.7 `EntryGate` 래치 API(Block · 사유 상수 · 닫힌 사유 목록이 정본·시험에 있는지) — D4 의 자리와 새 사유 이름 확정. 래치를 세울 함수(`NewContext` 안인지 `runInterlock` 안인지)와 census 영향
- [x] 1.8 편집 표면 · 고정 장치 표(`analysis/edit-surface.md`): 편집할 기존 함수 목록, a166 census 키 적중(`NewContext` 예상 1), 소스를 여는 시험, 감독 루프 순서 시험, 봉인 시험, 진행 중 change(a063 · verify-execution-capability · a105 · a107)와의 파일 겹침
- [x] 1.9 편집할 기존 함수마다 FLM/BTM pre-edit 번들(`tools/logic-map`) — 뒤집히는 기존 시험 함수 포함
- [x] 1.10 claim check: proposal · design 의 분기 주장을 AST 열거와 대조(`analysis/claim-check.md`)
- [x] 1.11 제안 동결 적대 리뷰(작성자와 다른 계열 — codex, 최대 2회차) → `review.md`
- [x] 1.12 Manager 결정 · freeze — **Manager freeze 2026-11-24**(review.md 「Manager freeze」 — codex 2회차 P1 2건 처분: Q-A 수용 · Q-B 기각 · Q-C 수용). base 재고정 승인(review.md 「Manager 결정」 1) — 실행은 2단계 착수 첫 단계(a198 착지 `7fec501f1` merge 뒤)

## 2. 구현

- [ ] 2.1 Pre-Edit 선언(High-risk): 바꾸는 것 · 바꾸지 않는 것 · 되돌리는 법
- [ ] 2.2 RED — design §3 의 1~11 (실제 게이트웨이 · 실제 원장 · 실제 감사 로그 · 실제 런타임 조립으로; production 무편집에서 실패 원문 `analysis/tdd/red.txt`). 뒤집히는 기존 시험 함수는 design §3-6 의 여섯뿐 — 그 밖이 깨지면 멈춤
- [ ] 2.3 GREEN 최소 구현 — 닫힌 집합 판정 함수 하나(인터록과 감시가 같이 부름, 형식 하한 · 빈 endpoint 항목 · 만료 시 계좌 확인 포함) · `AutomationStatus` 새 표지 · `verifyGate`/`runInterlock`(래치는 runInterlock) · 청산 쪽 소비자 넷 · `logGateDecision` 보호 수명 문장 · 래치 · 감사 action · 기동 출력 · 감시(문장 · 기동 표지 조건 알림). 새 설정 키·새 운영 모드·새 outbox 타입 금지. 주석은 한글로 「왜」
- [ ] 2.4 a166 census 키 갱신(`NewContext` 를 편집했으면 그 키 하나, a192 `census-repair` 절차) — 다른 키 무변 단언
- [ ] 2.5 인터록 머리 주석 정정(D9) · `cmd/tossctl/soak.go` 의 낡은 출력 문구(565 · 577 「clause 9」 · 도움말 220)를 새 귀결로(Manager 결정 11)
- [ ] 2.6 post-edit 번들 재생성 · `check_analysis`
- [ ] 2.7 변이 원장(design §3-12, 사본에서 · 시작 sha 단언 · 무변이 대조군 GREEN · 닿음 먼저)
- [ ] 2.8 선택 시험 · `-race`(해당 시험) · `make lint` · strict validate · tracker `--check` — 원문 `analysis/verify/`. 미실행 범위 기록

## 3. 리뷰 · 검증

- [ ] 3.1 구현 적대 리뷰(별도 문맥, Opus + codex 둘째 목소리, 최대 2회차) — 초점: 진입이 열리는 경로 · 닫힌 집합 밖이 내려가는 입력 · 판정 둘 · 재기동 우회
- [ ] 3.2 처분 반영
- [ ] 3.3 gstack `review` 1회
- [ ] 3.4 Manager 검증(diff · 시험 재실행)

## 4. 종결

- [ ] 4.1 착지 기록(`--record-landing`) → `origin/main` merge(rebase 금지) → 완료 게이트(격리 worktree) → archive → Story 경로 → push

## 5. 종결 뒤 (사람 승인 — 체크박스 아님)

- 운영 반영(`make image CHANGE=a199-…`)은 사람이 승인한다. 반영 전까지 운영 엔진의 동작은 지금과 같다(낡으면 기동 거절).
- 반영 뒤 확인: 낡은 attestation 으로 재시작하는 실측은 **하지 않는다**(일부러 낡게 만들지 않는다). 다음에 자연히 낡았을 때의 감사 action · 알림 문장으로 확인한다.

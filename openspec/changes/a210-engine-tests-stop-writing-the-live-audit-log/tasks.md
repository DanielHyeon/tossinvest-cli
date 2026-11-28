# a210 tasks

> 착수 2026-11-28(base `8db0463b7`, 브랜치 `test/a210-engine-tests-stop-writing-the-live-audit-log`). 출처: a209 배포 뒤 audit 확인(Manager).
> 사용자 지시 2026-11-28: Manager 는 OpenSpec 작성 · 검토, Opus 구현, 구현자별 별도 적대적 리뷰 → gstack 리뷰 → Manager 검증 → Opus 커밋 · push · archive.

## 1. 계약과 증거

- [x] 1.1 Story `STORY-TOS-a210` 을 registry 와 알맞은 Feature 에 등록, proposal · design · tasks 확정(Manager 초안), `openspec validate <id> --strict`.
  - 2026-11-28 구현자: Story 는 FEAT-TOS-001(a202 시험 전용 선례). strict validate 통과. `base-commit.txt` 의 짧은 sha 를 같은 커밋의 40 자리로 바꿈(check_analysis 가 짧은 sha 를 거절, `capture_change_base.py` 는 이미 있으면 덮지 않음).
- [x] 1.2 기억 회고, CodeGraph(`codegraph affected … --filter '*_test.go'`) · a166 census 확인, design §1 사실 F1~F6 재확인을 `analysis/code-context/` 에 원문으로.
  - 결과 `analysis/code-context/facts.md`: F1~F5 맞음. **차이** — base 의 엔진 시험은 기본 audit 경로에 닿지 않음(엔진 기동 18 자리 전부 `isolate(t)`), 실제 누수 시험은 `cmd/tossctl`.
- [x] 1.3 D2 census — 엔진 시험 패키지 안에서 `TOSSOS_DATA_DIR`·`XDG_*`·`HOME` 을 바꾸는 시험 목록.
  - 결과 `analysis/d2-census.txt`: 엔진 3 자리(알림 토픽 1 · 자격 증명 2, 경로 env 0), `cmd/tossctl` 은 경로 env 를 전부 `t.TempDir()` 또는 상대경로(거절 시험)로 줌 — 상자를 벗어나는 자리 0.

## 2. 구현 (RED → GREEN)

- [x] 2.1 RED: base 에서 D3 측정(가짜 HOME 상자) — 엔진 패키지가 상자에 audit.log 를 만듦 + 다른 후보 패키지 결과 → `analysis/red/`.
  - 결과(`analysis/measurement.md`): **엔진 패키지 0(두 태그) — 설계 가정과 다름.** 상자에 audit.log 를 만든 패키지는 `cmd/tossctl`(두 태그 각 1), 가려낸 시험 `TestA166EngineDecoratorCannotRemoveRequiredBudget`(`analysis/red-bisect/`) — 리뷰 F3 측정으로 태그 쪽 둘째 출처(`engine_account_seq_recovery_test.go` 5 시험)를 더 찾음(`analysis/review-f3/`). 나머지 후보 16 개 0.
- [x] 2.2 GREEN: D1 `TestMain`(새 파일) + D3 로 찾은 패키지 수리 → 같은 측정에서 상자 파일 0 → `analysis/green/`.
  - 결과: 엔진 새 파일 `a210_isolated_home_test.go`(TestMain), `cmd/tossctl` 새 파일 `a210_isolated_home_test.go` + 기존 `TestMain` 한 줄(FLM 번들 `analysis/function-logic/cmd-tossctl--testmain/`, check_analysis rc 0). 두 패키지 두 태그 모두 상자 파일 0. GREEN 1 회차 뒤 자식 프로세스가 남긴 빈 임시 디렉터리를 보고 부모 상자 재사용으로 고쳐 2 회차로 다시 잼(`analysis/measurement.md`).
- [x] 2.3 D5 거짓 행 목록(시각 · 설정 이름만) → `analysis/contamination.md`.
  - 결과: 컨테이너 이후 거짓 행 묶음 35(되돌림 완전 34 · 부분 1), 컨테이너 이전 판정 불가 1.

## 3. 검증 · 리뷰

- [x] 3.1 고친 패키지 시험 전체(무태그 · 해당하면 `tossos_testseams`) · `make lint` · 대조(TestMain 지운 사본에서 audit.log 재생성) → `analysis/verify/`.
  - 결과: 고친 두 패키지 전체가 상자 안에서 두 태그 모두 rc 0(`analysis/green/*.log` — GREEN 측정이 곧 회귀 실행), `make lint` rc 0(`verify/make-lint.log`). 대조: 엔진은 base 누수 0 이라 임시 시험(커밋 안 함)으로 — TestMain 있음 0 / 치움 1(`verify/control-probe-*`). `cmd/tossctl` 은 RED(1) → GREEN(0) 이 대조. GOCACHE · GOENV 고정값 = go 기본값(`verify/go-dir-pin.txt`).
- [x] 3.2 적대적 리뷰(구현자와 별도 에이전트) → 처분 → `review.md`.
  - 결과: 발견 F1~F8, Manager 처분대로 F1(표지 검증) · F5(`testenv.Env` 사용) · F6(shortcut 주석) · F7(FLM 줄 번호) 수리, F2 · F3 · F8 문구 정정, F3 측정에서 태그 쪽 둘째 출처 발견(`analysis/review-f3/`). 회귀 GREEN 3 회차 · `make lint` 재실행. 재리뷰(blocker 0): N2 상대경로 표지 거절 수리(RED → GREEN, `analysis/rereview-n2/`), N1 · 부모 갈래 nit 는 한계로 기록(`review.md`).
- [ ] 3.3 gstack review 1 회 → 처분 → `review.md`. 비례 원칙 `not-applicable` 사유(FLM · 변이 원장 · 다중 적대 리뷰).

## 4. 종결 (Manager · 사용자 지시 대기)

- [ ] 4.1 Manager 독립 검증.
- [ ] 4.2 착지 기록 → origin/main merge → 완료 게이트(사용자 명시 요청 시) → archive → Story 경로 · PM sync.

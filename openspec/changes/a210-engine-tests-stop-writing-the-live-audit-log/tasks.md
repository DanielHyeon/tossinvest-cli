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

- [ ] 2.1 RED: base 에서 D3 측정(가짜 HOME 상자) — 엔진 패키지가 상자에 audit.log 를 만듦 + 다른 후보 패키지 결과 → `analysis/red/`.
- [ ] 2.2 GREEN: D1 `TestMain`(새 파일) + D3 로 찾은 패키지 수리 → 같은 측정에서 상자 파일 0 → `analysis/green/`.
- [x] 2.3 D5 거짓 행 목록(시각 · 설정 이름만) → `analysis/contamination.md`.
  - 결과: 컨테이너 이후 거짓 행 묶음 35(되돌림 완전 34 · 부분 1), 컨테이너 이전 판정 불가 1.

## 3. 검증 · 리뷰

- [ ] 3.1 고친 패키지 시험 전체(무태그 · 해당하면 `tossos_testseams`) · `make lint` · 대조(TestMain 지운 사본에서 audit.log 재생성) → `analysis/verify/`.
- [ ] 3.2 적대적 리뷰(구현자와 별도 에이전트) → 처분 → `review.md`.
- [ ] 3.3 gstack review 1 회 → 처분 → `review.md`. 비례 원칙 `not-applicable` 사유(FLM · 변이 원장 · 다중 적대 리뷰).

## 4. 종결 (Manager · 사용자 지시 대기)

- [ ] 4.1 Manager 독립 검증.
- [ ] 4.2 착지 기록 → origin/main merge → 완료 게이트(사용자 명시 요청 시) → archive → Story 경로 · PM sync.

# a133 Tasks

Story: `STORY-TOS-a133` · High-risk 인접(alert 전달 결과 → 진입 래치 · 운영 모드 승격의 입력) — 알림 바이트만 바꾸는 방향.

## 0. 계약·증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a133.yaml`, `_registry.yaml`, `FEAT-TOS-009` 역링크) + tracker 재생성
- [x] 0.2 proposal · design(발행 경로 전수 D1 · 치환 지점 D2 · outbox 결정 D4) · spec 델타 · tasks, `openspec validate a133-an-alert-carries-no-account --strict --no-interactive` rc=0
- [x] 0.3 `capture_change_base.py` 로 base 고정
- [x] 0.4 CodeGraph hard evidence(`analysis/code-context/`) — 편집 대상의 정의·호출자·영향, `affected --filter '*_test.go'`, CGC 대조 · 운영 outbox 실측(수만)
      — codegraph 1.6.0, `affected --filter` 1127 · CGC 색인 5023 파일. 불일치 R1~R6(`evidence-reconciliation.md`), 실측 16 행 전부 DELIVERED(`analysis/verify/outbox-census.md`)
- [x] 0.5 Function Logic Map + Branch Test Map(편집 전) — `Ntfy.Publish` · `renderFields`
      — 번들 2(편집 전). BTM 이 인용한 신설 시험 3 은 1.1 에서 생김(그 전 `check_analysis` 는 미존재 시험 인용으로 rc 1 — 예상)
- [x] 0.6 proposal-freeze 적대 리뷰(분리된 컨텍스트) + Pre-Edit 선언 → `review.md`
      — APPROVE-WITH-FIXES(P1 1 · P2 5 · P3 7), 전부 처분(review.md 표) · Manager D4 ① 잔재 수용

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — 경계 · 경로별 · 필드 버림 · 구조 시험이 base 에서 실패함을 확인(`analysis/red/`)
      — obs 7 · engine 2 실패(누출 · 구조), cmd 1 은 base 통과 핀. 엔진 패키지 전체 기준선(격리 worktree, 잘림 없음): 실패는 a133 2개뿐(`red-engine-full.md`)
- [x] 1.2 GREEN — `redactForTransport`(새 leaf) · `Ntfy.Publish` 첫 문장 · `renderFields` 계좌 키 건너뜀
      — 리뷰 반영 시험 추가(F-3 리플렉션 · F-4 키 다섯 · F-10 숫자/Tags · F-2 구조 핀 별칭/⑤ · F-9 B4/B8)
- [x] 1.3 REFACTOR + 편집 뒤 AST 재추출 · FLM/BTM 갱신 · `check_analysis.py --change` 통과
      — Publish 9 분기 불변(+3줄) · renderFields 2 → 3(새 B2, 옛 B2 → B3). check_analysis 는 격리 worktree 영수증(2.2)

## 2. 검증

- [x] 2.1 변이(사본, 무변이 대조군 GREEN 먼저) — design D6 목록, 최소 3 CAUGHT(치환 우회 · 발행 경로 하나 누락 · 필드 버림 삭제), 원장 `analysis/mutation/`
      — round1 10/11(M3 빌드 실패 — 하네스 결함) → round2 11/11 → 구현 리뷰 변이 4 추가 → **round3 CONTROL GREEN · 15/15 CAUGHT**
- [x] 2.2 `make lint` · 관련 패키지 `go test -race -timeout 45m` · `make test-seams` · openspec validate
      — 격리 worktree 영수증 `analysis/verify/receipts.md`. test-seams 의 실패 1(`internal/position` `TestExitEligibilityHasOneSpelling`)은 base 의 a130 파일 원인 — a133 무관
- [ ] 2.3 `make sdd-sync` · `make sdd-check`
- [x] 2.4 구현 리뷰(분리된 리뷰 패스, 적대적) → `review.md`
      — APPROVE-WITH-FIXES(P2 1 · P3 7) 전부 수용 · I-1 은 `redact.String` 고정점 수리(FLM 편집 전)

## 3. 사람 · Manager

- [ ] 3.1 Manager `make gate CHANGE=a133-an-alert-carries-no-account` + 독립 검증 + 푸시

### 게이트 뒤 절차 (체크박스 아님 — a108 선례)

1. archive(Opus) → Story `openspec.path` 를 아카이브 경로로 → tracker 재생성 → PM `--check`.
2. 배포 — a129 · a131 · a130 과 한 번에, 사람이 KST 05:00~09:00 에 호스트 `tossctl` 과 이미지를 함께 교체(a131 design 「배포」 1 · a133 design 「배포」).
3. 교체 뒤 사람이 콘솔 「알림 테스트」 한 번 + 다음 critical 알림 본문에 `account:` 줄 없음 · `pos-[redacted]` 확인.
4. 운영 절차(D4 ① — Manager 결정 2026-10-13 잔재 수용): 계좌 자격 증명을 바꾸기 전에 `alert_outbox` PENDING 0 확인(아니면 먼저 전달 · 정리). 코드로 막지 않음.

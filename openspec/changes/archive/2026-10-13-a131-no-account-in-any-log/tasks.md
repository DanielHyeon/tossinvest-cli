# a131 Tasks

Story: `STORY-TOS-a131` · High-risk(인터록 증거의 계좌 결속 경로 · 엔진 기동 로깅) — 로그·기록 바이트만 바꾸는 방향.

## 0. 계약·증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a131.yaml`, `_registry.yaml`, `FEAT-TOS-009` 역링크) + tracker 재생성
- [x] 0.2 proposal · design(유출 자리 전수 표 D1 — `analysis/harness/account_census.go`) · spec 델타 · tasks, `openspec validate a131-no-account-in-any-log --strict --no-interactive` rc=0
- [x] 0.3 `capture_change_base.py` 로 base 고정
- [x] 0.4 CodeGraph hard evidence(`analysis/code-context/`) — 편집 대상의 정의·호출자·피호출자, `affected --filter '*_test.go'`
      — codegraph 1.6.0, `affected --filter` 1119 줄. CodeGraphContext 색인 0 · `make sdd-sync` 의 CGC update 300s 타임아웃(advisory, a129 선례) — census 로 대신
- [x] 0.5 Function Logic Map + Branch Test Map(편집 전) — 편집하는 기존 함수 전부(13 번들, 작성기 `analysis/harness/flm_write.py`)
- [x] 0.6 proposal-freeze 적대 리뷰(분리된 컨텍스트) + Pre-Edit 선언 → `review.md`
      — APPROVE-WITH-FIXES(P1 4 · P2 9 · P3 7), 전부 GREEN 전에 design 2판 · spec 2판 · 시험으로 처리(review.md 표)

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — 센티널 · 왕복 · 구조 · 정리 명령 시험(design D7)이 base 에서 실패함을 확인(`analysis/red/`)
      — red-1(base + 시험만: 빌드 실패) · red-2(새 leaf 만: 행동 실패 9 + 구조 2). 2판 시험(R-1~R-4)은 같은 판에서 추가 — 변이 원장이 각 가드를 잼
- [x] 1.2 GREEN — `internal/redact`(규칙 · Handler · Writer · 봉인), `obs.NewLogger`, `official.accountsLocked` 등록, `runEngineRun`·`main` Writer,
      `audit.Record`, soak·검증 기록 쓰기·읽기, `runSoakAttest` 계좌 줄 삭제, 정리 명령 `scrub-account-records`
- [x] 1.3 REFACTOR + 편집 뒤 AST 재추출 · FLM/BTM 갱신 · `check_analysis.py --change` 통과
      — 번들 28(생산 함수 15 + 기대값을 바꾼 기존 시험 함수 13), `evidence complete`. 새 함수는 이미 번들이 있는 함수 뒤로 옮겨 이웃 함수 오검출 0

## 2. 검증

- [x] 2.1 변이(사본, 무변이 대조군 GREEN 먼저) — design D7 목록, 최소 4 CAUGHT(필터 우회 · 싱크 누락 · 마스킹 삭제 · 봉인 대신 평문 포함), 원장 `analysis/mutation/`
      — 최종 round4: CONTROL GREEN · **31/31 CAUGHT**(round1 23/25 + 하네스 변이 컴파일 오류 2 → round2 25/25 → 구현 리뷰 변이 6 추가 → round3 29/31 미도달 2 → round4)
- [x] 2.2 `make lint` · 관련 패키지 `go test -race` · `make test-seams` · openspec validate
      — 영수증 `analysis/verify/receipts.md`(engine 패키지 race 2502s rc 0 · test-seams 의 a112 도구 2 실패는 미추적 파일 원인 — 커밋 뒤 rc 0)
- [x] 2.3 `make sdd-sync` · `make sdd-check`
      — sdd-sync: CodeGraph sync 완료 · CGC update 300s 타임아웃(advisory) · GBrain busy(advisory). sdd-check rc 0(최종 커밋 뒤 재실행, receipts)
- [x] 2.4 구현 리뷰(분리된 리뷰 패스, 적대적) → `review.md`
      — APPROVE-WITH-FIXES(P1 1 · P2 3 · P3 11), 처리 표 review.md(I-14 기각 · I-15 Manager 판단)

## 3. 사람 · Manager

- [x] 3.1 Manager `make gate CHANGE=a131-no-account-in-any-log` + 독립 검증 + 푸시
      — 2026-10-13 Manager 독립 검증: 커밋 6613e3ba·f269c615·dfa348ff 의 a130 파일 0 · 트레일러 확인 · redact/obs/audit 시험 직접 실행 rc 0.
      position_id 를 `pos-[redacted]` 로 가리는 결정(a129 화이트리스트 뒤집음) 수용 — 사용자 결정 「계좌정보는 로그에 남기지 않음」, pos- 는 소금 없는 해시라 대입으로 되돌릴 수 있음; 상관은 symbol·intent_id 로. 게이트 = 2026-10-13 `make gate` **11/11 PASS**(sdd-sync 는 CGC 300s 타임아웃 advisory, CodeGraph 최신 — 6단계 sdd-check 통과).

### 게이트 뒤 절차 (체크박스 아님 — a108 선례)

1. archive(Opus) → Story `openspec.path` 를 아카이브 경로로 → tracker 재생성 → PM `--check`.
2. 배포 — 사람이 KST 05:00~09:00 에 호스트 `tossctl`(tossos-attest.timer)과 이미지(`make image CHANGE=a131-no-account-in-any-log`)를 **함께** 교체(design 「배포」 1).
3. 정리 — 사람이 `scrub-account-records` 미리보기 → soak 정지 → `--apply` → soak 재시작 → 다음 attest 타이머 결과 확인.
4. 과거 engine.log 처리 — Manager · 사람(범위 밖).

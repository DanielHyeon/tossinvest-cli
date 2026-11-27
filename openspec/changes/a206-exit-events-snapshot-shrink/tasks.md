# a206 tasks

> 사용자 목표: 운영 저널 파일 크기 축소(주문 원장 보존). 1단계는 조사 + 제안(생산 코드 0), 2단계는 정비 명령 구현.
> 2단계는 proposal 「사람이 결정할 것」이 정해지고 READY 가 된 뒤 시작. 운영 저널 쓰기 · 엔진 정지 · 명령 실행은 사람만. High-risk(원장).

## 1. 조사 · 계약(1단계)

- [x] 1.1 기억 회고: `scripts/memory-recall.sh` · Hindsight `tossos-dev-episodes` · ai-memory — 이 주제의 직접 선례 0(결과 없음 기록)
- [x] 1.2 리더 열거: CodeGraph 1.6.0 callers · `codegraph affected --filter '*_test.go'` · `rg` 전수 → `analysis/code-context/evidence-reconciliation.md`
- [x] 1.3 분기 주장 AST 근거 11 개(`analysis/ast-evidence/claims.md` C1~C7)
- [x] 1.4 운영 저널 읽기 전용 사본 집계 · 시나리오 S0~S4 VACUUM 실측(`analysis/a206_measure.py` · `.out.txt`), 사본 삭제
- [x] 1.5 생산 리더를 시나리오 사본에 돌려 결과 비교(`analysis/a206_reader_check.go.txt` · `.out.txt`), 임시 프로그램 · 사본 삭제
- [x] 1.6 proposal · design · position-ledger spec 델타, Story `STORY-TOS-a206` → FEAT-TOS-009, `openspec validate --strict` · PM `--check`
- [x] 1.7 경량 리뷰 1회(`review.md`)

## 2. 구현(2단계 — READY 뒤)

- [x] 2.1 사람 결정 1~3 반영(대상 경과 일수 · legacy 표시 유지 여부 · 실행 창 · 백업 위치), `base-commit.txt` 고정
  — 2026-11-27 Manager 전달: 범위 = completed CLOSED 전부(경과 일수 없음, 연결 · 훅 · a207 압축 행 · 앵커 제외), 표시 = legacy_event(v37 없음),
  백업 = `<journal>.pre-a206-<UTC>.bak`, 실행 창 = 사람이 나중에, B = 밖, `.bak` 정리 = 사람이 완료(proposal 「사람 결정」). origin/main(a205 · a207) 병합 뒤
  `base-commit.txt` = `657141baa`.
- [x] 2.2 RED: design D6 의 시험 다섯(대상 술어 · 정비 뒤 리더 · 멱등 · 엔진 잠금 거절 · dry-run 쓰기 0). 대상 술어 시험에 a207 압축 행 ·
  앵커 제외(design D1 의 SQL, a207 design D8)와 그 제외가 두 번째 실행에서도 같은 행을 남기는지(멱등)를 넣음
- [x] 2.3 GREEN: 새 파일에 journal 정비 함수(D1 · D2 한 트랜잭션) · `tossctl` 명령(D3 · D4) — `tossctl engine exit-evidence-compact`
- [x] 2.4 `check_analysis.py --change a206-exit-events-snapshot-shrink` 로 기존 함수 편집 0 확인(또는 번들) — 등록 · 표지 집합 · ReadOnly 허용 메서드 집합
  때문에 required 3, 번들 3 으로 「evidence complete」 rc 0
- [x] 2.5 대상 시험 · `make lint` · 사본 저널에 명령 dry-run / apply 실측(운영 저널 아님) — `analysis/a206_copy_apply.out.txt`
  (293,183,488 → 68,947,968 B, 비운 행 28,025, integrity ok, 리더 오류 0, 두 번째 dry-run 대상 0)
- [x] 2.6 gstack review(작성자와 다른 계열) 최대 2회차 — Codex r1 P2 3(2 수리 · 1 기각) · P3 1 수리, r2 지적 0(`review.md`)

## 3. 사람 승인 뒤(이 change 밖에서 사람이 시작)

- [ ] 3.1 두 시장 닫힌 창에서 엔진 · 콘솔 정지 → 명령 실행 → 기동 → 읽기 전용 확인
- [ ] 3.2 백업 검증 뒤 처분(사람)
- [ ] 3.3 main 병합 · archive

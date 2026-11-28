# a208 tasks

## 0. 제안 · 설계(생산 코드 변경 0)

- [x] 0.1 기억 recall(Hindsight `tossos-dev-episodes` · ai-memory) — 직접 관련 선례 없음, 기존 편입 RUNNER 사본(StockOS A168 이식)만 확인
- [x] 0.2 StockOS 원문 재확인 — `analysis/stockos-reference.md`
- [x] 0.3 CodeGraph(1.6.0, `--filter '*_test.go'`) · CodeGraphContext · 증거 조정 — `analysis/code-context/`
- [x] 0.4 인용 함수 AST 29 개 + 주장 C1~C16 — `analysis/ast-evidence/`
- [x] 0.5 proposal · design · spec delta · Story 등록, `openspec validate --strict`
- [x] 0.6 사람 결정 Q1~Q6 기록(design 10, 2026-11-28) — Q1 유예 없음 · Q2 RUNNER 사본 · Q3 설정 파일 토글. 설계 개정 1(유예 · 평가기 편집 · 롤백 제약 삭제), strict validate
- [ ] 0.7 proposal-freeze 리뷰(적대적 Eng 포함) → `review.md`

## 1. 구현 준비(READY 뒤)

- [ ] 1.0 CodeGraph 편집 표면 재확인(`codegraph affected <files> --filter '*_test.go'`) · a166 census(`shapes.json`)에 편집 함수가 없는지 `rg`
- [ ] 1.1 design 9 의 편집 예정 함수마다 `analysis/function-logic/` 번들(ast · FLM · BTM · risk) 을 **편집 전에** 작성, `check_analysis.py` 로 TODO 0 확인
- [ ] 1.2 Pre-Edit 선언(docs/WORKFLOW.md 형식)을 `review.md` 에 기록

## 2. RED → GREEN(High-risk)

- [ ] 2.1 T1~T4 · T7~T10 RED 관측(실패 출력 기록) — T5 · T6 은 기존 경로 끝-끝 확인(design 8.2)
- [ ] 2.2 `config.Adoption` · `rawAdoption` 필드 · `mergeAdoption` · `LoadRawEngineAdoption` · `SaveEngineAdoption` · `handleSettingsSave` → T7~T9 GREEN
- [ ] 2.3 `adoptOne` 정책 선택 · 알림 문구(D1 · D5) → T1~T6 GREEN
- [ ] 2.4 `recordGateSettings` 끝 행 + a163 시험의 「마지막 행」 단언 갱신 → T10 · a163 · a074 · a197 GREEN

## 3. 검증

- [ ] 3.1 선택 시험(design 8.3) + `make lint` — 실행 명령 · 결과 기록, 미실행 범위 명시
- [ ] 3.2 변이 원장 M1~M12(`analysis/mutation-ledger.md`) — 전부 CAUGHT
- [ ] 3.3 FLM 번들을 `revision: current` 로 갱신, `check_analysis.py --change a208-adoption-floor-only-exit`
- [ ] 3.4 gstack review(적대적 패스, 가능하면 다른 계열, 2 회차 상한) → `review.md`

## 4. 운영 문서 · 종결

- [ ] 4.1 `docs/operations.md` 에 토글 켜는 법(설정 키 · 기동 audit 확인)과 「켠 뒤 효과는 이후 편입에만」을 적음
- [ ] 4.2 `docs/operations.md` 롤백 절에 design 7 의 재배포 확인(옛 이미지 콘솔 저장이 키를 지움)을 적음
- [ ] 4.3 착지 기록 → 완료 게이트(사용자 명시 요청 시) → archive → Story 경로

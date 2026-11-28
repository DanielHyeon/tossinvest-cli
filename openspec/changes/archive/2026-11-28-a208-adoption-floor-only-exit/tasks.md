# a208 tasks

## 0. 제안 · 설계(생산 코드 변경 0)

- [x] 0.1 기억 recall(Hindsight `tossos-dev-episodes` · ai-memory) — 직접 관련 선례 없음, 기존 편입 RUNNER 사본(StockOS A168 이식)만 확인
- [x] 0.2 StockOS 원문 재확인 — `analysis/stockos-reference.md`
- [x] 0.3 CodeGraph(1.6.0, `--filter '*_test.go'`) · CodeGraphContext · 증거 조정 — `analysis/code-context/`
- [x] 0.4 인용 함수 AST 29 개 + 주장 C1~C16 — `analysis/ast-evidence/`
- [x] 0.5 proposal · design · spec delta · Story 등록, `openspec validate --strict`
- [x] 0.6 사람 결정 Q1~Q6 기록(design 10, 2026-11-28) — Q1 유예 없음 · Q2 RUNNER 사본 · Q3 설정 파일 토글. 설계 개정 1(유예 · 평가기 편집 · 롤백 제약 삭제), strict validate
- [x] 0.7 proposal-freeze 리뷰(적대적 Eng 포함) → `review.md` — Manager(Fable) 2026-11-28 PASS(조건 1: 롤백 핀까지 a208 포함 이미지로 올린 뒤에만 토글 ON), 구현 뒤 기록(사유 review.md)

## 1. 구현 준비(READY 뒤)

- [x] 1.0 CodeGraph 편집 표면 재확인(`codegraph affected <files> --filter '*_test.go'`) · a166 census(`shapes.json`)에 편집 함수가 없는지 `rg`
- [x] 1.1 design 9 의 편집 예정 함수마다 `analysis/function-logic/` 번들(ast · FLM · BTM · risk) 을 **편집 전에** 작성, `check_analysis.py` 로 TODO 0 확인
- [x] 1.2 Pre-Edit 선언(docs/WORKFLOW.md 형식)을 `review.md` 에 기록

## 2. RED → GREEN(High-risk)

- [x] 2.1 T1~T4 · T6~T10 RED 관측(실패 출력 `analysis/tdd/red.txt`) — T3 · T5 는 기대대로 base 에서도 통과(OFF 동일 · 기존 RUNNER 사본 경로)
- [x] 2.2 `config.Adoption` · `rawAdoption` 필드 · `mergeAdoption` · `LoadRawEngineAdoption` · `SaveEngineAdoption` · `handleSettingsSave` → T7~T9 GREEN
- [x] 2.3 `adoptOne` 정책 선택 · 알림 문구(D1 · D5) → T1~T6 GREEN
- [x] 2.4 `recordGateSettings` 끝 행 + a163 시험의 「마지막 행」 단언 갱신 → T10 · a163 · a074 · a197 GREEN

## 2F. F3 재작업 — 사용자 결정 「F3 2」(2026-11-28): 설정 공통 정책의 보호선 표 + 부분익절 0

- [x] 2.5 CodeGraph(정책 ID 대조 자리 전수) · 편집 전 FLM 5 번들 · Pre-Edit(review.md) · RED(`analysis/tdd/red-f3.txt`, 커밋 233897096)
- [x] 2.6 GREEN — 사본 ID · 고정 정체성(`internal/exitpolicy/a208_adopted_floor_only.go`), 해소 갈래(CommonLadderForPosition · LegacyLadderPolicyIdentity),
      원장 수용(record · OpenExitState), 무시 = OFF + audit `ignored`(mergeEngine · recordGateSettings), 알림(adoptOne 인자)
- [x] 2.7 F3 검증 — 선택 시험 · `make lint` · 변이 M13~M23 14/14(`analysis/a208_f3_mutation.py`, `analysis/mutation-ledger.md`) · FLM `revision: current` · check_analysis — 실행 기록 review.md
- [x] 2.8 F3 gstack review 2 보이스(codex + 같은 계열) 1회차 — P0~P2 0, P3 2(문서 1 · 기록 1), 2회차 불필요(생산 코드 무변)

## 3. 검증

- [x] 3.1 선택 시험(design 8.3) + `make lint` — 실행 명령 · 결과 기록, 미실행 범위 명시
- [x] 3.2 변이 원장 M1~M12(`analysis/mutation-ledger.md`) — 전부 CAUGHT
- [x] 3.3 FLM 번들을 `revision: current` 로 갱신, `check_analysis.py --change a208-adoption-floor-only-exit`
- [x] 3.4 gstack review(적대적 패스, 가능하면 다른 계열, 2 회차 상한) → `review.md`

## 4. 운영 문서 · 종결

- [x] 4.1 (F3 개정 반영) `docs/operations.md` 에 토글 켜는 법(설정 키 · 기동 audit 확인)과 「켠 뒤 효과는 이후 편입에만」을 적음
- [x] 4.2 (F3 개정 반영) `docs/operations.md` 롤백 절 — 사본 ID 의 롤백 순서 제약(design 7 개정 2) · 재배포 확인(옛 이미지 콘솔 저장이 키를 지움)
- [x] 4.3 착지 기록 → 완료 게이트(사용자 명시 요청 시) → archive → Story 경로
      — 2026-11-28 사용자 「a208 게이트 승인」. `--record-landing` = `2e2150b57`(커밋 `b7433ddc3`) → `origin/main`(`5ae7c7511`) merge(rebase 아님, 커밋 `c076fdac1`,
      PM 3 파일 충돌만 — registry 합집합 · tracker 재생성) → tracker 재생성 → `make sdd-sync` → `make gate CHANGE=a208-adoption-floor-only-exit` → archive → Story 경로 · PM sync.
      gate 회차별 결과는 `review.md` 「완료 게이트」. Manager 독립 검증은 main 병합 때 Manager 몫.

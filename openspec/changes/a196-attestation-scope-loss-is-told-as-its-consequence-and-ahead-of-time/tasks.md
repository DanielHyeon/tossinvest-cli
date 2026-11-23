# a196 Tasks

> **Normal(관측 전용 보조 실행자·알림).** 기존 Go 함수 본문을 고치므로 함수마다 FLM 번들이 필요하다(gate ⑤). 변이 원장은 design §5 의 6개로 한정.
> 실계좌 주문을 내는 자동 시험 금지, `tossctl verify run`·`soak attest` 를 시험·검증에서 실행 금지. 픽스처는 합성 값만.
> 전체 시험·전체 게이트는 종결 단계에서만. attestation 파일 형식·`internal/obs` 타입 열거·원장 스키마 무변경.
> 리뷰는 최대 2회차, 리뷰가 낸 새 가드·새 정책은 Manager 결정 전 구현 금지.

## 1. 계약과 증거 (production Go·시험 편집 0)

- [ ] 1.1 Story `STORY-TOS-a196`(FEAT-TOS-003) 등록 확인, `base-commit.txt` 고정.
- [ ] 1.2 CodeGraph hard evidence(버전 명기): `judgeAttestation`·`binding`·`stageKey`·`attestationWatch.prepare`·`eventFor`·`checkOnce`, `soak` 의 발급 함수와 `acceptSupervised` — `analysis/code-context/`.
- [ ] 1.3 `tools/logic-map` AST 를 `analysis/cited-ast/` 에 현재 HEAD 로 뽑고 proposal 표 여섯 줄을 대조한다(특히 key 구성과 「발급자가 같은 유효 기간 값을 두 곳에 쓴다」). 틀리면 문서를 고치고 Manager 에게 보고.
- [ ] 1.4 편집 표면 세기: 고칠 기존 함수 목록 확정, a135 시험 중 뒤집히는 단언을 가진 시험 함수 수, D2 전제 시험을 둘 패키지(import 순환 확인), census·구조 고정 유무.
- [ ] 1.5 편집 대상 기존 함수들의 FLM/BTM 번들(pre-edit 판).
- [ ] 1.6 `openspec validate a196-… --strict` 통과, proposal-freeze 리뷰(다른 계열, 최대 2회) — `review.md`.
- [ ] 1.7 Manager 1단계 검증·freeze.

## 2. Pre-Edit

- [ ] 2.1 Pre-Edit 선언을 `review.md` 에 확정.

## 3. RED → GREEN

- [ ] 3.1 RED: `internal/app/engine/a196_attestation_deadline_internal_test.go` — design §5 의 1~9, D2 전제 시험(10).
- [ ] 3.2 GREEN: 기한 계산 새 함수 + `judgeAttestation`·`stageKey`/`prepare`·`eventFor` 편집.
- [ ] 3.3 뒤집히는 a135 시험을 번들과 함께 고치고(design §5-11), 편집 뒤 번들 AST·FLM·BTM 재생성.

## 4. 검증

- [ ] 4.1 변이 원장(`analysis/mutation/`) — design §5 의 변이 6개.
- [ ] 4.2 선택 시험 원문, `make lint`, `openspec validate --strict`, `check_analysis.py --change a196-…`. 미실행 범위 기록.
- [ ] 4.3 구현 적대 리뷰(다른 컨텍스트) → gstack review 1회 — `review.md`.

## 5. 종결

- [ ] 5.1 Manager 검증(diff·시험 원문 재실행).
- [ ] 5.2 완료 게이트(격리 worktree) → archive → Story 경로 갱신 → push.
- [ ] 5.3 (사람 승인) 운영 반영, 엔진 생존 감시 타이머 설치(design §4-1) — 결과는 `review.md` 에 시각·상태만.

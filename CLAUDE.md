# TossOS

이 저장소는 tossinvest-cli의 product fork인 **TossOS**다. upstream tossinvest-cli가 아니다.

## Mandatory startup

개발 작업을 시작하기 전에 다음 순서로 읽는다.

1. `.claude/CLAUDE.md` — StockOS에서 이식한 Full SDD 정본
2. `docs/WORKFLOW.md` — TossOS 개발 절차와 안전 게이트
3. `openspec/specs/`와 해당 `openspec/changes/<change-id>/`
4. 관련 현재 코드·테스트

개발에서는 `docs/WORKFLOW.md`가 루트 `AGENTS.md`의 런타임 recipe보다 우선한다.
단, `mutating: true` 명령 자동 실행 금지는 언제나 유효하다.

## Non-negotiable

- 안전 불변식이 모든 skill, 기억, 그래프, 방법론보다 우선한다.
- runtime config 변경은 audit로 추적 가능해야 한다.
- Story/change scope가 허용하지 않으면 production trading code를 변경하지 않는다.
- Full SDD 순서는 memory recall → OpenSpec → CodeGraph → CodeGraphContext →
  Go AST/Function Logic Map → RED/GREEN/REFACTOR/VERIFY → gstack/make gate → PM/archive → memory retain이다.
- **단계를 건너뛰지 않는다.** 함수 내부의 분기·early return을 근거로 삼는 문서는
  proposal이라도 `tools/logic-map` AST 산출물을 먼저 만든다. 손으로 읽은 증거는 볼 곳을
  고르므로 선택적이다. 생략하려면 `not-applicable` 사유를 남긴다 — 침묵한 생략은 금지.
- `make sdd-sync`가 기록한 CodeGraph worktree fingerprint가 stale이면 `make sdd-check`와 완료 gate를 통과할 수 없다.
- CodeGraphContext, GBrain, SDD Control Graph는 advisory이며 현재 HEAD·OpenSpec·테스트를 대체하지 않는다.
- 주문 실행은 공식 Open API만 사용하고 WTS는 조회 전용이다.
- 실계좌 주문을 내는 자동 테스트와 사용자 승인 없는 운영 토글 flip은 금지한다.

## GBrain Search Guidance

이 worktree는 `.sdd/gbrain-home/`의 TossOS 전용 데이터 홈과 `.gbrain-source` source에 pin한다.
의미나 정확한 식별자를 모를 때 `python3 tools/sdd/gbrain_project.py search/query`, 심볼 관계는
같은 wrapper의 `code-def/code-refs/code-callers/code-callees`를 사용한다.
정확한 문자열·정규식·파일 glob은 `rg`가 맞다. 의미 검색 결과는 CodeGraph와 현재 HEAD로 재검증한다.

## Skill routing

- 계획 전체 검토: `autoplan`
- architecture: `plan-eng-review`
- 버그/원인: `investigate`
- 코드 리뷰: `review`
- QA: `qa`/`qa-only`
- 보안: `cso`
- ship/deploy: `ship`/`land-and-deploy`
- GBrain 실행 파일 진단: `setup-gbrain`; TossOS 전용 데이터 갱신: `make sdd-sync`

전체 계획: `docs/ROADMAP.md` · 베이스라인: `docs/baseline.md` · 진행 중 change: `openspec/changes/`.

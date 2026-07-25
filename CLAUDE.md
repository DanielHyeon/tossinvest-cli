# TossOS

이 저장소는 tossinvest-cli의 product fork인 **TossOS**(자동매매 제품)다. upstream tossinvest-cli가 아니다.

**개발 작업을 시작하기 전에 반드시 [docs/WORKFLOW.md](docs/WORKFLOW.md)를 먼저 읽어라** — SDD 사이클, 리뷰 게이트, 완료 게이트(`make gate`), 불변 규칙이 거기 있다. 개발 작업에서는 WORKFLOW.md가 AGENTS.md보다 우선한다 (AGENTS.md는 tossctl 런타임 운용 규칙).

- **최상위 안전 불변식**(docs/WORKFLOW.md §0)이 모든 방법론에 우선한다: 승인 없는 LIVE 주문 금지, 손절 즉시성 약화 금지, 사이징·손절 변경은 보수 방향만, 운영 토글은 사람 승인.
- 전체 계획: [docs/ROADMAP.md](docs/ROADMAP.md) · 베이스라인: [docs/baseline.md](docs/baseline.md)
- 진행 중 change: `openspec/changes/` (openspec CLI 1.4.x, `make validate`)
- 주문 실행은 공식 Open API만, WTS는 조회 전용. 실계좌 주문을 내는 자동 테스트 금지.

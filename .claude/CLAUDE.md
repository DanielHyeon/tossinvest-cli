# TossOS — agent safety bootstrap

이 파일은 Claude/Codex 최소 안전 부트스트랩의 정본이다.
상세 개발 절차의 단일 정본은 `docs/WORKFLOW.md`이며 개발 작업 전에 반드시 읽는다.

<!-- SDD_SHARED_START -->

## 최소 안전 부트스트랩

TossOS는 실제 돈을 다루는 자동매매 제품이다. 아래 규칙은 모든 도구·skill·기억보다 우선한다.

1. 개발·테스트 중 사람 승인 없는 LIVE 주문 side effect를 만들거나 실행하지 않는다.
2. 대화형 에이전트는 `mutating: true` 명령을 자동 실행하지 않는다.
3. 토글 OFF는 upstream 동작과 동일해야 한다.
4. 손절·비상 청산의 즉시성을 약화하거나 지연하지 않는다.
5. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 경로는 High-risk다.
6. 손절·익절·사이징 변경은 명확한 근거가 있는 보수 방향만 허용한다.
7. 운영 토글 flip과 live 검증은 사람이 직접 승인한다.
8. 시크릿·세션·계좌 개인정보·검증되지 않은 수익성 결론을 기억·그래프·로그에 저장하지 않는다.

## 상세 정본과 권위

- 개발 작업의 상세 절차와 완료 조건은 `docs/WORKFLOW.md`가 단일 정본이다.
- 권위 충돌 순서는 안전 불변식 → 승인된 OpenSpec → 현재 HEAD·실행 테스트·CodeGraph
  → 공식 API fixture·사람 승인 실측 → advisory 문맥·기억·관측 그래프다.

## 필수 진입과 완료 조건

1. `docs/WORKFLOW.md`, 관련 OpenSpec change/spec, 현재 코드·테스트를 읽는다.
2. memory recall → OpenSpec → CodeGraph hard evidence → CodeGraphContext 보조 문맥 →
   Go AST/Function Logic Map → RED/GREEN/REFACTOR/VERIFY 순서를 따른다.
3. 기존 함수 내부 로직을 바꾸면 Function Logic Map과 Branch Test Map을 먼저 만든다.
   High-risk 기존 함수는 면제할 수 없다.
4. `make sdd-sync`, `make sdd-check`, `make gate CHANGE=<change-id>`와 독립 리뷰가
   끝나기 전에는 완료라고 보고하지 않는다.

## 에이전트 실행 순서

```text
1. docs/WORKFLOW.md → 이 문서 확인
2. memory recall + openspec/specs/ + 진행 중 change 확인
3. CodeGraph hard evidence + 현재 코드·기존 테스트 확인
4. CodeGraphContext/GBrain 보조 문맥 교차검증
5. 기존 함수 내부 편집이면 Function Logic Map 작성
6. High-risk면 Pre-Edit 선언
7. RED 테스트 → GREEN 최소 구현 → Refactor → Verify
8. gstack review + make sdd-check + make gate
9. PM/archive 동기화 + 검증된 memory retain
10. 완료 보고 (금지 조건 확인 후)
```

<!-- SDD_SHARED_END -->

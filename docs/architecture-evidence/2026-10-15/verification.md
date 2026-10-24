# 아키텍처 문서 검증 기록

- 일자: 2026-10-15 (Asia/Seoul)
- 문서: [전체 아키텍처](../../architecture.md)
- 범위: 현재 워킹트리의 정적 구조·계약 설명. 운영 상태·배포 일치·실거래 성공 판정 제외.

## 확인 항목

1. `go list -json ./cmd/tossctl ./internal/...` 성공. Linux 기본 빌드의 111개 패키지에서 내부 직접 import와 빌드 파일 목록을 수집했다.
2. 패키지 의존도 그림의 직접 import 22개를 수집 목록과 대조했다.
3. 저장소 `tools/logic-map`을 빌드하여 핵심 함수 12개의 AST를 생성했다. 함수별 source SHA-256과 실제 파일을 대조했다.
4. 분석 입력 836개(기본 빌드 Go 소스 및 주요 설정·문서)의 해시를 기록하고 문서 작성 후 다시 대조했다. 기록에는 소스 해시만 있으며 자격증명·운영 설정 값은 없다.
5. 문서의 상대 링크 대상 존재 여부와 `git diff --check`를 검사했다.
6. 별도 검토 에이전트가 생산 dispatch 경로, Effective의 권한 경계, 기동/보호 조건, 콘솔/API 원장 소유권, activation·lease·위험 예약 순서를 현재 소스와 대조했다. 검토 범위에서 핵심 설명 오류 없음. 검토 당시 없었던 본 verification.md를 생성한 뒤 링크 검사를 다시 수행한다.

## CodeGraph와 현재 코드 대조

CodeGraph 상태는 미반영 추가 2개·수정 4개를 보고했다. 공유 워킹트리의 인덱스를 갱신하지 않고 상태·호출자 조회 결과를 저장했다. 호출자 조회에는 인터페이스 선언과 테스트도 섞이므로 생산 호출 증거로 단독 사용하지 않았다. `strategy_entry_supervisor.go`의 조립 → `strategy_dispatch_cycle.go`의 dispatch → `execgw/strategy_gateway.go`의 PlaceClaimedStrategy 연결은 현재 파일·AST·Go import로 확인했다.

CodeGraphContext/GBrain 조회는 not-applicable: 이번 산출물은 새로운 동작·변경 영향 설계가 아니라 현재 소스 구조의 설명이며, 의미 검색 결과를 근거로 추가 채택하지 않았다.

## 적용하지 않은 절차

- production 코드 변경, 신규 기능·운영 변경: 없음.
- 함수 로직 수정용 Pre-Edit/RED-GREEN, 전체 회귀·race·LIVE 시험: not-applicable. 문서 변경이며 주문 코드·동작은 수정하지 않음. 분기 설명의 근거로 필요한 AST는 생성함.
- 신규 OpenSpec 기능 change, PM Story/archive, `make gate CHANGE=...`: not-applicable. 기존 기능 계약 변경·릴리스 승인을 수행하는 작업이 아님.
- 전역 SDD 동기화·gate 통과: 수행하지 않음. 다른 미완료 작업의 워킹트리를 이 문서의 완료 증거로 포괄하지 않음.
- Mermaid 실제 렌더 검증: 수행하지 않음. fence·링크와 의존 edge를 정적으로 확인함.

이 기록은 설명 문서 검증이다. 제품 전체 gate·배포·주문 안전성 시험의 통과 선언이 아니다. baseline의 미커밋 변경은 다른 작업의 변경이며 수정·되돌림하지 않았다.

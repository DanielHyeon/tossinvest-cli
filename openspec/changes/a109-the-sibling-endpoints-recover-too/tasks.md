# Tasks — a109-the-sibling-endpoints-recover-too

**등록만 먼저(2026-08-14). 착수 조건: a108 land.** A1 실측(a108 review.md §1 F3·F4)이 입력.

## 0. 착수 전 조건

- [ ] 0.1 a108 land 확인, base 재고정 + `make sdd-sync`.
- [ ] 0.2 세 transport의 FLM 생성(발행·회수·Close 함수) + Pre-Edit 선언.
- [ ] 0.3 endpoint별 강등 가능성 판정(조회 전용인가 명령 표면인가) — proposal Non-goals의
  질문을 설계로 확정.

## 1. RED — 사고급 모양의 재현

- [ ] 1.1 세 endpoint 각각: pre-chmod socket 잔재(umask 077)에서 현재 기동이 영구
  거부됨을 고정(A1 절차 재현).
- [ ] 1.2 alert control: 수락 중인 socket 위에 두 번째 서버가 올라서는 현재 동작을 고정.

## 2. GREEN

- [ ] 2.1 socket 발행 stage+rename 통일, 잔재 회수 전체성(검증-사망 perm 완화 포함).
- [ ] 2.2 산-주인 처리(거부 또는 flock 명문화 — 0.3의 판정대로).
- [ ] 2.3 뮤테이션 원장.

## 2b. a108 gstack 리뷰가 이관한 항목

- [ ] 2b.1 httpapi의 소비자측 lazy 재-dial(rate-limit, reader nil일 때 stat+dial 1회) 또는
  감독 순서(engine-ready 후 httpapi) 고정 — 회복 부팅마다 전략 화면이 수동 재시작 전까지
  소실되는 기본값을 닫는다(a108 red-team). 엔진측 in-process 재시도 금지(D2-2)는 유지.
- [ ] 2b.2 "Close가 flock 보유 중에 돈다" defer 순서 불변식 핀(a108 Claude 적대 3라운드 —
  지금은 defer 등록 순서 하나에 얹혀 있고 테스트가 없다).
- [ ] 2b.3 롤백 전 잔재 제거 절차를 세 endpoint에도 일반화(a108 D5-3).

## 3. 게이트

- [ ] 3.1 영향 패키지 `go test -race`(app/engine은 `-timeout 25m` 이상) + `go vet`.
- [ ] 3.2 `openspec validate --all --strict` → `make sdd-sync` → `make sdd-check` →
  `make gate CHANGE=a109-the-sibling-endpoints-recover-too`.
- [ ] 3.3 review.md + PM 동기화.

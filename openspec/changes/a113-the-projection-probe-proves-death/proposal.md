# a113 — projection probe도 사망을 증명한다

> **상태: 등록만 먼저 했다(2026-08-16).** a109 A1 적대 리뷰가 실측으로 연 change다
> (a109 issues I1). **착수 선행 조건은 없다** — a108 확정 코드의 소유권 문제로 a109가
> 형제 기계만 고쳤을 뿐, 계약(chmod-then-probe)은 이미 spec에 있다.

## Why

a109 §1-fix F1은 형제 endpoint 회수의 owner-write 사망 추정을 chmod-then-probe로
바꿨다. 그 판정의 근거는 A1 P1-A의 결정적 재현이다: **쓰기 비트가 깎인 수락 중인
socket을 죽었다고 읽고 지운다**(a109 review.md §1). 그 절의 원형인 a108
`internal/strategyprojectionrpc/transport_unix.go`의 `projectionSocketAccepts`에는
같은 추정이 **그대로 남아 있다**(a109 issues.md I1 — a109 pre-edit-gate T1의 선언된
무변경 표면이라 손대지 않았다).

노출은 형제보다 좁다: 그 endpoint의 최종 이름은 이미 a108 D1 의례(stage+rename)를
지나 0600으로만 나타나므로, 추정이 틀리는 상태에 이르려면 외부에서 chmod가 일어나야
한다. 그래도 병의 모양은 동일하고, engine-safety spec의 "산 주인의 endpoint는
탈취되지 않는다"는 이 표면에도 적용된다.

## What Changes

- `projectionSocketAccepts`의 owner-write 추정 절을 삭제하고 형제와 같은
  chmod-then-probe(0600 복원 → dial → 재-Lstat SameFile 검증)로 교체한다.
- a109 F1-N1 뮤테이션의 원형판(추정 절 재도입 시 실패하는 핀)을 이식한다.
- 이 밖의 a108 회수·발행 의례는 무변경이다.

## Impact

- engine-safety spec: 사망 검증 요구를 "권한 비트 추정 금지"로 명시(MODIFIED).
- 코드: `internal/strategyprojectionrpc/transport_unix.go` 한 함수 + 테스트.
- High-risk: 엔진 boot 경로의 회수 기계 — Function Logic Map 필수(착수 task 0).

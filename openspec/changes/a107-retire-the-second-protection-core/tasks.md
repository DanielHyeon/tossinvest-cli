# Tasks — a107-retire-the-second-protection-core

**이 change는 a100 tasks 6.5의 이행으로 등록만 먼저 했다(2026-08-13). 착수 조건: a100 land.**

## 0. 착수 전 조건

- [ ] 0.1 a100이 land되어 있는지 확인한다 — 같은 봉인 파일을 편집하므로 순서를 겹치지 않는다.
- [ ] 0.2 `python3 tools/sdd/capture_change_base.py --change a107-retire-the-second-protection-core`로
  base를 재고정하고 `make sdd-sync`를 돌린다.
- [ ] 0.3 소비자 추적 — `internal/protection` 각 파일의 non-test importer를 CodeGraph로 열거하고
  제거/유지 목록을 확정한다. proposal의 유지 목록(domain.go)은 등록 시점 추정이며 이 열거가 정본이다.
- [ ] 0.4 편집할 기존 함수·봉인 테스트의 Function Logic Map을 만들고 Pre-Edit 선언을 남긴다.

## 1. 제거

- [ ] 1.1 `controller.go`·`repository.go`와 이 둘만 검증하는 테스트를 제거한다.
- [ ] 1.2 `protection_sagas`·`protection_mutation_attempts` 스키마 생성 경로가 어떤 바이너리에서도
  실행되지 않음을 확인한 뒤 제거한다.
- [ ] 1.3 domain 타입 소비자(`execgw`·`app/engine`)가 무변경으로 컴파일되고 기존 테스트가 GREEN이다.

## 2. 봉인 재조정

- [ ] 2.1 죽은 심볼을 참조하던 봉인만 줄인다. a100 4.2의 import walk 확장은 유지한다.
- [ ] 2.2 "프로덕션 보호 core는 하나다" 정적 가드 — 자체 영속 저장소를 갖는 보호 상태 core의
  부활을 거부한다.

## 3. 게이트

- [ ] 3.1 영향 패키지 `go test -race` + `go vet`.
- [ ] 3.2 `openspec validate --all --strict`.
- [ ] 3.3 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=a107-retire-the-second-protection-core`.
- [ ] 3.4 `review.md` 작성 + PM 동기화(STORY-TOS-a107 acceptance 대조).

# Tasks — a113-the-projection-probe-proves-death

**a109 issues I1 의 이행. 등록 2026-08-16, 착수 2026-09-25(freeze 리뷰 P1-4 로 design D1/D2 기준 재작성).**

## 0. 착수 전 조건

- [x] 0.1 base 재고정(`tools/sdd/capture_change_base.py --change a113-…`) — `54004f44`.
- [x] 0.2 `projectionSocketAccepts`·`reclaimStaleControlDirectory`·`verifyStaleSocketShape`·`Dial` 의
  Function Logic Map 을 proposal/design 갱신 **전에** 만든다(`analysis/function-logic/` 4 번들).
- [x] 0.3 CodeGraph hard evidence·CGC·reconciliation(`analysis/code-context/`).
- [x] 0.4 design.md 작성 + proposal-freeze 독립 적대 리뷰(review.md §0) + 판결 반영
  (spec delta 를 projection 요구로 이전, chmod 순서, 모드 핀).

## 1. 구현

- [x] 1.1 RED: 쓰기 비트가 깎인 산 socket 위 `Start` 가 거부·보존·0600·"still alive" 를 요구하는
  `TestTheReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped` + 순수 probe 판정 표 행 +
  회수 전용 probe 판정 표·바뀐 파일·미검증 이름 테스트.
- [x] 1.2 GREEN: `projectionSocketAccepts` 의 owner-write 절 삭제 · 새 파일 `transport_probe_unix.go`
  의 `staleProjectionSocketAccepts` · `verifyStaleSocketShape` 의 `(os.FileInfo, error)` 반환(단일 stat) ·
  회수 B11·B12 두 줄 · 주석 정정.
- [x] 1.3 뮤테이션 N1~N6 적용·격추 측정(원장 `mutation-ledger.md`), N6 은 AST 핀.
- [x] 1.4 FLM 구현 후 재최신화(AST 재생성·분기 재번호) + `check_analysis.py` rc 0.
- [x] 1.5 검증: `go test -race -count=1 ./internal/strategyprojectionrpc/...`, cmd/tossctl 재부착 회귀,
  `make vet`, `make lint`, `make test`.
- [x] 1.6 구현 후 gstack `review`(또는 독립 서브에이전트) → review.md §1.
- [ ] 1.7 착지 기록(`--record-landing`) 커밋.

## 2. 게이트 (Manager)

- [ ] 2.1 `make gate CHANGE=a113-the-projection-probe-proves-death` 후 archive.

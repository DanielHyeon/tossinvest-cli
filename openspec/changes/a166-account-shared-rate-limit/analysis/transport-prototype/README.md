# 로컬 수명주기 실험 — 2026-10-31

Go 표준 라이브러리만 쓰는 가짜 연결 시험. 운영 전송 코드·네트워크·브로커 호출 없음. 기존 저장소와 Go 표준 기능을 우선 조사한 search-first 결과를 사용했고 의존성 추가 없음.

- RED: 임시 사본에서 `revoke`의 mutex 획득과 해제를 제거한 뒤 `GO111MODULE=off go test <임시 lifecycle_test.go> -run TestRevokeWaitsForWriteAndBlocksLaterWrites -count=1` → FAIL `revoke returned during write`.
- GREEN: `GO111MODULE=off go test -race -count=1 ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` → `ok`, 1.065s.
- 시험은 진행 중 write 동안 revoke가 반환하지 않고, revoke 뒤 추가 write를 거부하며, 늦게 등록된 연결을 닫고, 모델의 잠금 해제가 revoke 반환 뒤에 일어나는지 확인한다. 30ms 대기는 스케줄링을 완전히 제어하지 못하므로 이 단일 실행은 증명이 아니다.
- 실제 TLS, net/http, flock, 프로세스 사망, priority, 파일 무결성, network 및 모든 wire 시도는 시험하지 않았다. 이 결과는 실제 3/s 보장이나 production 준비 증거가 아니다.

# a134 verify receipts (2026-10-14, 구현 팀메이트 Opus, 공유 체크아웃)

| 명령 | 결과 |
|---|---|
| `/usr/local/go/bin/gofmt -l internal/enginewatch cmd/tossctl internal/app/engine internal/obs` | 빈 출력, rc 0 |
| `go vet ./...` · `rtk proxy go vet ./internal/enginewatch/ ./cmd/tossctl/ ./internal/app/engine/ ./internal/obs/` | 문제 0, rc 0 |
| `rtk proxy go test -count=1 -v ./internal/enginewatch/` | 14/14 PASS |
| `rtk proxy go test -count=1 ./internal/obs/ ./internal/enginelock/` | ok · ok (a133 구조 핀 포함 — 새 Publisher 구현 타입 0) |
| `rtk proxy go test -count=1 ./internal/app/engine/` | ok (310.6s) |
| `rtk proxy go test -count=1 -run 'TestA134\|TestLeafCommandsHaveSourceAnnotation\|TestMutatingAnnotationOnTradeCommands\|TestShortDescriptionsAreEnglish' -v ./cmd/tossctl/` | 7/7 PASS |
| `rtk proxy go test -count=1 ./cmd/tossctl/` | FAIL 1 — `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(a098 시험, **병행 change a135 의 미커밋 편집**: 시험이 `attestation-watch` 보조 실행자를 기대하는데 그 배선은 작업 트리에 없음). a134 파일 무관 — 그 밖 전부 통과 |
| `openspec validate a134-a-dead-engine-is-noticed --strict --no-interactive` | valid, rc 0 |
| `systemd-analyze --user verify deploy/systemd/tossos-heartbeat.{timer,service}` | 단위 문법 오류 0(마운트의 실행 비트 경고만 — git 모드는 100644) |

## 변이(경량, 비례 원칙 — 의무 아님)

사본 없이 `internal/enginewatch/watch.go` 를 백업 → 변이 → `go test ./internal/enginewatch/` → 백업 복원(sha256 앞 16자 일치 확인).

| # | 변이 | 결과 |
|---|---|---|
| M1 | 전송 **전에** 상태 기록(순서 역전) | CAUGHT(`TestA134AFailedSendDoesNotAdvanceTheState`) |
| M2 | 부재 외 stat 오류를 「없음」으로 접음 | CAUGHT(`TestA134AnUnreadableMarkerIsNotAbsence`) |
| M3 | 중복 억제 제거(끊김 지속에도 전송) | CAUGHT(`TestA134OneAlertPerOutageAndOnePerRecovery`) |
| M4 | Publisher nil 검사 제거 | CAUGHT(`TestA134NoPublisherFailsLoudly`) |

복원 확인: `restored OK`.

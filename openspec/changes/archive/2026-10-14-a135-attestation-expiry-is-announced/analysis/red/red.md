# RED (a135, 구현 전 · base 421bf53c + 시험 파일 + `internal/obs/event.go` 이벤트 상수 2개만)

## internal/app/engine — 판정 · 실행자 부재로 빌드 실패
```text
$ go test -count=1 -run TestA135 ./internal/app/engine/
# github.com/JungHoonGhae/tossinvest-cli/internal/app/engine [github.com/JungHoonGhae/tossinvest-cli/internal/app/engine.test]
internal/app/engine/a135_attestation_watch_internal_test.go:67:13: undefined: attestationStage
internal/app/engine/a135_attestation_watch_internal_test.go:71:60: undefined: stageHealthy
internal/app/engine/a135_attestation_watch_internal_test.go:72:30: undefined: stageHealthy
internal/app/engine/a135_attestation_watch_internal_test.go:73:37: undefined: stageNotice
internal/app/engine/a135_attestation_watch_internal_test.go:74:42: undefined: stageNotice
internal/app/engine/a135_attestation_watch_internal_test.go:75:37: undefined: stageNotice
internal/app/engine/a135_attestation_watch_internal_test.go:76:44: undefined: stageNotice
internal/app/engine/a135_attestation_watch_internal_test.go:77:37: undefined: stageImminent
internal/app/engine/a135_attestation_watch_internal_test.go:78:40: undefined: stageImminent
internal/app/engine/a135_attestation_watch_internal_test.go:165:91: undefined: attestationWatch
internal/app/engine/a135_attestation_watch_internal_test.go:78:40: too many errors
(빌드 실패 — FAIL [build failed])
```

## cmd/tossctl — 생산 조립에 감시 실행자가 없음
```text
$ go test -count=1 -run TestProductionRuntimeStartsExactlyOneAlertDeliverer ./cmd/tossctl/
--- FAIL: TestProductionRuntimeStartsExactlyOneAlertDeliverer (0.00s)
    a098_the_engine_registers_a_sender_test.go:29: production auxiliary names=[alert-delivery normal-alert-relay] want=[alert-delivery normal-alert-relay attestation-watch] — 밀린 알림을 보내는 주체가 없다
FAIL
FAIL	github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl	0.012s
FAIL
```

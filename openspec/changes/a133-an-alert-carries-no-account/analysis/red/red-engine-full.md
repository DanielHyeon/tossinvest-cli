# 엔진 패키지 기준선 — 격리 worktree(02660f3d + a133 시험 4 파일), 잘림 없이

`GOFLAGS=-trimpath rtk proxy go test ./internal/app/engine/ -count=1`

```text
--- FAIL: TestA133TheDelivererRedeliversWithoutTheAccount (0.21s)
--- FAIL: TestA133TheDelivererStillLatchesOnRefusal (0.24s)
--- FAIL: TestA111ObserverUsesClockLeaseHelpersForTheUseLease (0.00s)
    a111_flat_exit_observation_test.go:1010: parse github.com/JungHoonGhae/tossinvest-cli/internal/app/engine/exitloop.go: open github.com/JungHoonGhae/tossinvest-cli/internal/app/engine/exitloop.go: no such file or directory
--- FAIL: TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration (0.00s)
    a111_flat_exit_observation_test.go:1182: read github.com/JungHoonGhae/tossinvest-cli/internal/app/engine/exitloop.go: open github.com/JungHoonGhae/tossinvest-cli/internal/app/engine/exitloop.go: no such file or directory
FAIL
FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/app/engine	334.520s
FAIL
rc=1
```

- TestA111 둘은 `-trimpath` 의 산물(소스 경로로 exitloop.go 를 여는 시험) — trimpath 없이 같은 worktree 에서 둘만 다시 돌림 → `ok`.
- 결론: base 에서 실패하는 엔진 시험은 a133 신설 2개뿐.

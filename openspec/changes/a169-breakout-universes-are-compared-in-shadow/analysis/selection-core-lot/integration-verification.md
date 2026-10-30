# 통합 검증

대상 merge 8def9cc7acea41f21771a3549194580c961cdaa6, origin/main 2d67c226ed175a4bd53500f8bc668f46f5909f9b. 후보 b43569c7 대비 selection source 및 test diff 없음.

Go 1.25, GOPROXY=off, loopback만 있는 network namespace에서 `./tools/a140-breakout-signal-shadow`의 `^(TestA169(Selection.*|SecurityType.*)|TestEffectivePolicyIsCanonicalAndExcludesRuntimeSource)$` 선택 normal 및 `-race -tags tossos_testseams` PASS. 원문은 integration-normal.log.gz 및 integration-race.log.gz. Scoped OpenSpec 1.4.1 strict, PM tracker check, git diff --check PASS.

범위는 비연결 U0/U1 순수 계산. 전체 시험·게이트, 운영·실측·활성화·archive 미실행. 기존 policy 시험은 semantic 검증이며 저장된 digest byte golden이 아니다.

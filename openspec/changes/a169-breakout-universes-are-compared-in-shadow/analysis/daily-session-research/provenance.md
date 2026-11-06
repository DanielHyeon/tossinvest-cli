# 출처 및 도구 확인 — 2026-11-06

`../daily-session-research.md`는 `/tmp/a169-daily-claim-research.md` 원문 복사. 작성 시 source checkout `/mnt/D/Axipient/workspace/TossOS`, HEAD `ca7a3f4f77f597ed529752d9b89dd81bcd8a4b59`; 보존 대상 checkout HEAD `232d6b86c024dfd71cc8ffe6fd0fbba342b3e9ec`. 두 checkout의 다음 파일 SHA-256 비교 결과 일치:

- `internal/accountbudget/files_linux.go` d2c6c8bf5a8df35e406fe7c2720770b0a00eb4cfe9c05a3dc4830aec2624b8ff
- `tools/a140-breakout-signal-shadow/shadow.go` 62c2fe84d04026f1f398a01cb97b270880c82f1f355ad0a6b2841ccb59e62de1
- `tools/logic-map/extract_go_ast.go` 3b885ecdf4fa95f38f5f03831d8cc2cbb6ae2daaa25725790e8b531cc1117a11
- `internal/accountbudget/observation_linux.go` 777ff9a4b68a2d5f5c4a2c3fb62e4d83a4cf500cff46bb16ddd5753ae298abfe
- `internal/accountbudget/budget_linux.go` 54b6a9c50d3646e2ad72c1f12af74bfa7bd97b29a2f6e282bbf3ada4b9225675
- `internal/enginelock/enginelock.go` e615b0c66ce3ecb71540cb96f5bcf81ed0240530af5c1aac430a577e0c582b38
- `internal/runlock/runlock.go` f2f2ab6914b62ca95d76d2381e8d0e406fa2e8448583384a8606e5f8edd6ec02
- `internal/candidate/store.go` 1622bb2352f1359e83b95d94e7ba46cc2746dcee5210fc2e63aa59b0b795a52c
- `internal/journal/journal.go` f7bf598890e191466894eb5c95adf16baf41bd06ac308988eec60051542a7abc
- `tools/a140-breakout-signal-shadow/ledger.go` 44523e91f5508b762d79b8fe3ad171d0436d746109856ffa0523fb16f3736626
- `go.mod` 6d21ba03888e9bd2154a7e6f175cee2a3969f081fb2ae343ebd17409d98f9ef6
- `go.sum` 85d477a7bc9ba7f1454012e1b7980f785667bee51dc6b1f202b890f427e5e145

동일 원본 checkout에서 아래 명령의 전체 JSON 출력 보존(생성 당시 이전 출력과 source hash 동일):

- `go run ./tools/logic-map --file internal/accountbudget/files_linux.go --func checkedDir` → `checkedDir.ast.json`; 파일 SHA-256 `3ab2427cadae354f064d1127cf70c92db273b258f0a791b9e1053b8d082d8016`
- `go run ./tools/logic-map --file internal/accountbudget/files_linux.go --func checkedFile` → `checkedFile.ast.json`; `64cbf41a1723a0cfbefc9ca4812109fa4b245f2e5c783a464e70a922c0caacda`
- `go run ./tools/logic-map --file tools/a140-breakout-signal-shadow/shadow.go --func probe.startSession` → `probe.startSession.ast.json`; `7e696a9f7e6baaa073f0beb0133a2cce636d12e13ae9cc238bf9e0f9e2c89cf9`

도구 availability: CodeGraph 1.6.0, ast-grep 0.45.0, Go 1.26.5, ai-memory 3.2.2, ripgrep 15.2.0 실행 가능. 이 연구에서 CodeGraph 그래프 검색/색인 freshness는 확인하지 않았으며 CodeGraph 호출 증거가 아니다. ast-grep도 실행하지 않았다. Go AST는 분기·호출 구조만 나타내고 조건의 의미 및 안전 결론은 해당 원문 대조에 기반한다. 분석 과정 테스트·운영 파일·계좌·시크릿·Git mutation 없음. 연구는 source 변경/런타임 검증이 아니다.

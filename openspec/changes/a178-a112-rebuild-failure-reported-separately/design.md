# a178 design

## 1. 지형 (HEAD `526650e50`, codegraph 1.6.0 원문 + logic-map AST)
| 대상 | 위치 | 비고 |
|---|---|---|
| 시한 | `tools/a112-mb-us-source/main.go:19` `overallBudget = 120 * time.Second`, `:82` `collectorCtx` | spec 101행 SHALL — 바꾸지 않는다 |
| 호출 | `main.go:116` `verifyPrescribedBinary(collectorCtx, pre, deps.builder)` → 실패 시 `holdf("prescribed build identity: %v", err)` | HOLD 문구 앞부분 유지 |
| 판정 함수 | `identity.go:470-495` `verifyPrescribedBinary` — AST 분기 B1(:471) 입력 없음, B2(:475) 메모리 builder, B3(:478) Go 스냅숏 결속, B4(:481) 배포 스냅숏 결속, B5(:484) rebuild 전 변경, **B6(:488) rebuild 오류 ∨ 명령 불일치 ∨ digest 불일치**, B7(:491) rebuild 뒤 변경 | B6 하나만 셋으로 나눈다 |
| 실제 builder | `identity.go:155` `systemBuildRunner.rebuild(ctx, …)` — `exec.CommandContext(ctx, …)` 들이 ctx 시한을 따른다 | 시한 초과 오류가 여기서 올라온다 |
| 시험 | `main_test.go:1099` `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch`(err != nil 만 단언), e2e `:1005`·`:1080` 근처 | 문구를 고정한 시험 0(grep) |

## 2. 설계
B6 을 순서 있는 세 검사로 나눈다(앞의 것이 먼저 판정):
```go
rebuilt, command, err := builder.rebuild(ctx, snapshot.RepoRoot, goBinary)
if err != nil {
    return fmt.Errorf("prescribed rebuild failed: %w", err)
}
if command != prescribedBuildCommand {
    return errors.New("prescribed rebuild used an unexpected command")
}
if digestBytes(rebuilt) != snapshot.ExecutableSHA256 {
    return errors.New("running executable does not match the prescribed rebuild")
}
```
- 판정 불변: 세 조건 중 하나라도 참이면 오류(전과 같은 OR). 성공 경로(셋 다 거짓)도 같다.
- 시한 초과 시 상위 HOLD 문구: `a112 m-b1 measurement HOLD: prescribed build identity: prescribed rebuild failed: … context deadline exceeded`.
- 오류 원문에 경로·명령 출력이 들어갈 수 있다. 이 도구의 HOLD 는 운영자 터미널과 시험 로그에만 간다(receipt 에 오류 원문을 쓰는지는 1.1 에서 확인하고, 쓰면 그 자리는 바꾸지 않는다 — 이 change 는 반환 오류만 바꾼다).

## 3. 시험 (RED 먼저)
- `TestVerifyPrescribedBinaryReportsRebuildFailureSeparately`: 가짜 builder 가 `context.DeadlineExceeded` 를 돌려줌 → `errors.Is(err, context.DeadlineExceeded)` 참, 문구에 "prescribed rebuild failed", "does not match" 없음.
- `TestVerifyPrescribedBinaryReportsUnexpectedCommandSeparately`: 명령만 다름 → "unexpected command".
- 기존 `…RejectsExecutableHashMismatch` 는 문구 단언을 더해 digest 갈래를 고정(시험 함수 편집 → 번들 대상이면 번들).
- 가짜 builder 가 명령을 돌려주는지(현재 `fakeBuildRunner`)는 1.1 에서 확인.

## 4. 검증
`go test ./tools/a112-mb-us-source/ -count=1`(무거운 e2e 포함 — 다른 무거운 실행이 없을 때) · `go vet ./tools/a112-mb-us-source/` · `gofmt -l tools/a112-mb-us-source` · `check_analysis.py --change a178-…`.

# Branch Test Map: `a112MBUSReferenceAllowed`

행 번호는 편집 뒤 `ast.json` 기준. RED/GREEN 은 `../../tdd/red-2.1.txt`·`green-2.2.txt`, 변이는 `../../mutation/`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 491행 `tools/a112-mb-us-source/` 파일 허용 | `TestA112MBUSStaticNoProductCallerOrForbiddenPath`(순회) | 그 분기는 통과, 시험은 v2_token_test.go 위반으로 FAIL | yes(PASS) |
| B2 | 494행 `internal/official/a112_mbus_*_test.go` 허용 | `TestA112MBUSStaticNoProductCallerOrForbiddenPath`(순회) | 위와 같음 | yes |
| return true | 498행 `v2_token_test.go` 정확 경로 허용 | `TestA112MBUSStaticNoProductCallerOrForbiddenPath`(순회; 편집 전에는 false 라 RED), `TestA189MBUSGuardAdmitsOnlyTheExactA166RuntimeTestFile` 허용 사례; 변이 g4(되돌림) CAUGHT | yes(RED 원인) | yes |
| return false | 498행 그 밖의 경로 | 픽스처 `TestA112MBUSResolvedASTGuard*` 3개, `TestA189…` 거절 사례 4개; 변이 g1(허용 밖 시험 파일의 참조)·g2(접두사)·g3(접미사) CAUGHT | 통과 | yes |

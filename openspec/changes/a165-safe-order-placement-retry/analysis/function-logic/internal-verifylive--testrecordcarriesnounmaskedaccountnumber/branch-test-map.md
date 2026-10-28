# Branch Test Map: TestRecordCarriesNoUnmaskedAccountNumber

- Source: `internal/verifylive/record_test.go` (156-182, 편집 뒤). AST branches 6. 시험 함수 자신이 대상이라 「Test」 열은 이 함수와 변이 판정.
- 편집 전 → 뒤 분기 대응(AST 열거 대조, 손 재번호 아님): B1·B2·B3 같음 · 옛 B4(원 바이트 `8901`) → 새 B4(원 바이트 마스크 전체) + 새 B5(풀린 키·문자열 값의 `8901`) · 옛 B5 → 새 B6.
- 측정(2026-10-28, `internal/verifylive`):
  - 편집 전 `go test -count=3000 -run '^TestRecordCarriesNoUnmaskedAccountNumber$'` → 5/3000 실패(옛 B4, 일치 자리 전부 digest 안).
  - 편집 뒤 같은 명령 → **3000/3000 ok**(120.224s). 패키지 전체 `go test -count=1 ./internal/verifylive/` ok · `go vet`(무태그·`-tags tossos_testseams`) ok · gofmt 0.
  - 변이(생산 `record.go` 를 잠시 바꾸고 원복, 원복 sha 일치 확인): M0 무변이 대조군 ok · M1 `Recorder.Append` 봉인 생략 → FAIL(172 B4 · 176 B5 `account_ref=*********8901`) ·
    M2 `Title` 에 마스크 꼬리 4자 덧붙임 → FAIL(176 B5 `title=… tail 8901`, B4 는 안 걸림 — `*` 없음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 실행 오류도 기록은 검사 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a | yes |
| B2 | 기록 읽기 실패 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a(시험 기반 경로) | yes |
| B3 | 평문 계좌 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a(편집 안 함) | yes |
| B4 | 마스크 전체 평문 | `TestRecordCarriesNoUnmaskedAccountNumber` | yes — 변이 M1 | yes(3000/3000) |
| B5 | 마스크 꼬리 평문(키·문자열 값) | `TestRecordCarriesNoUnmaskedAccountNumber` | yes — 변이 M1 · M2 | yes(3000/3000) |
| B6 | 열린 기록에 마스크 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a(편집 안 함) | yes |

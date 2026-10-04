# 보이스 ③ — 증거 · 시험 품질 (a112 8.5)

공통 브리프: 같은 디렉터리의 `brief.md` 를 **먼저 전부** 읽고 그 안전 규칙을 지켜라. 네 배역은 아래 하나다.

**배역: 시험과 증거가 주장을 정말 받치는지 본다.** 변이는 사본 또는 `go test -overlay` 로만(실제 저장소 편집 금지).
1. 변이 충분성: 원장 `analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`(13) · `lot-2.3/mutation-2.3.tsv`(10)에 **없는** 변이를 스스로 만들어 살아남는 것을 찾아라 —
   예: `failedFields` 목록에서 항 하나 빼기(그 항의 거절이 사라짐), 서술자 검사 항 순서 바꾸기, B5/B6 다시 합치기의 다른 철자, collectMarket 의 관문 대입을 다른 철자로 옮기기
   (census 인식기가 `gate = loader.familyGateFor(...)` 대입 · `fail(...)` 호출 · 합성 리터럴만 센다 — 헬퍼 · 다른 변수 이름 · 클로저로 우회 가능한가).
2. census 회피: `a112_proposal_closure_carriage_test.go` 의 AST census 와 `a112_activation_error_fields_test.go` 의 표가 새 갈래 · 새 결속 필드 추가를 정말 강제하는가.
3. `1_200_000` 리터럴 두 자리(2.3) — 변이 핀(CF-03/04 · CF-10)이 골든 `rvol_counterfactual_ppm[0]` 결속을 대신할 만한가.
4. B11(계보 충돌) · B14(미해결 선택)를 census 로만 재는 것이 타당한가 — 입력으로 닿게 만들 수 있는가.
5. 편집 뒤 번들(`analysis/function-logic/` 의 loadproductionfamilyactivation · validateproductionfamilyactivation · decode… · failedfields · collectmarket · evaluatefresh)의 BTM 행이
   인용한 시험이 그 분기를 실제로 지나는가(커버리지로 확인).
출력 형식은 브리프 「출력」.

# Evidence reconciliation (a189, 1단계)

- 기준 커밋: `7d1e3163a70487ae7c387ce06fd5c73a2bcfa7d3`.
- CodeGraph(1.6.0) 호출자 열거와 현재 HEAD 의 `grep`·AST 가 일치함: `a112MBUSReferenceAllowed` 호출 자리는 a112_mbus_static_test.go:424 한 곳(소스 직접 확인), 두 도구의 가드 시험은 호출자 0.
- M-B0 이름을 가진 Go 파일 전수(`grep -rln A112MBUS(Result|HoldError|Candle|Orderbook|Calendar)`): 6개. 그중 허용 규칙 밖은 `internal/accountbudgetruntime/v2_token_test.go` 하나 — base 실패 원문의 위반 목록과 같음.
- D2 열거(AST)와 base 실패 원문이 일치함: 허용 목록 밖은 import 2경로와 `official.Option` 뿐이고 예산 패키지 선택자는 지금 어느 목록에도 추적되지 않음(그래서 실패 원문에 안 나옴).
- CodeGraphContext: 안 봤음(`codegraphcontext-context.md`). 불일치 없음.

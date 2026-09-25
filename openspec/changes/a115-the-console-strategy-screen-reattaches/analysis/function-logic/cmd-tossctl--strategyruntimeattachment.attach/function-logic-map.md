# Function Logic Map: `strategyRuntimeAttachment.attach`

- Source: `cmd/tossctl/httpapi_strategy_attach.go` (:158–163) · AST: `ast.json`
> 인용 전용 번들(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다.

## Branches — 0개 (AST). 잠금 아래 reader·attached(=live)·failed(=!live) 대입 + seat++.
부팅 1회 해석의 세 값(nil·sentinel·client) 전부가 유효한 출발점 — design D1 의 「부팅 해석을 그대로 받는다」의 근거.

## Safety conclusion
- 무편집. High-risk impact: no.

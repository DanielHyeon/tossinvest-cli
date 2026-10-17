# Function Logic Map: `newProductionStrategyFirstLegAuthorityLoader`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go`
- Source SHA-256: `c29e90e2a1e9f04e531a1cc000caa4bcc6a97a0cd796a73845a97dd23e1de7ca`
- Signature: `newProductionStrategyFirstLegAuthorityLoader(params=8, results=1)`
- Source range: `277:1`–`286:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 6.2 봉인 리뷰 수리, 2026-10-01). 편집 전 AST: `analysis/measurements/lot-6.2-seal/pre-edit-fix/newproductionstrategyfirstlegauthorityloader.ast.json`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 입력: 시계 · 원장 · Guardian · 조립의 일정 · 제안 · 위험 · 환율 · 계좌 권한 쌍. 결과: 1차 레그 권한 loader.
- **불변식(6.2 리뷰 codex #1 · 보이스 A #1):** loader 가 드는 제안 쌍은 호출자 배열과 **떼어 낸 사본**이다 — 봉인의 재유도 원본이 dispatch 쪽
  사본의 제자리 원소 교체로 바뀌지 않는다.

## Branches and early returns

- 분기 없음. Exact AST return nodes: `284:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | happy path | 277:1 | 모든 권한을 담은 loader 를 돌려준다 — 제안 쌍만 `detachedStrategyProposalPair` 로 떼어 냄 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `detachedStrategyProposalPair` | 285:14 |

## State mutations and fallbacks

- 상태 쓰기 없음. 새 배열 할당(제안 쌍 두 시장의 entries).

## Safety conclusion

- High-risk impact: yes(1차 레그 권한의 대조 원본). 편집은 공유를 끊는 방향만 — 판정 조건 불변, 입력이 같으면 결과가 같다.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리 2차: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 codex 재확인 #2 T: 주석만 바뀜(본문 불변)

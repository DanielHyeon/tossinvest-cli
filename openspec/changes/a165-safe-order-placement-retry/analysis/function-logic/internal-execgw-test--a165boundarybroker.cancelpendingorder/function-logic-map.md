# Function Logic Map: `a165BoundaryBroker.CancelPendingOrder`

- Source: `internal/execgw/a165_cancel_refusal_boundary_test.go` (40-51)
- Revision: D5 4판 test pre-edit, `source_sha256` `70ec4addd02ca8554550d44194212fde8293419e9940383111f6cd8b1c9f8a6e`
- AST evidence: `ast.json` — AST 분기 2, 반환 2, 호출 2.

첫 호출만 `firstDelay`만큼 실제로 sleep한 뒤 AP 409를 돌려주고, 두 번째 호출은 caller가 받은 context deadline 여부를 기록한 뒤 ACK를 돌려줌. Fake Clock은 이 sleep에서 전진하지 않음. 이번 변경은 fake 구현을 바꾸지 않고, 이 관찰값을 D5 Clock 판정 회귀에 사용함.

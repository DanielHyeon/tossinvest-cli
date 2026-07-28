# Branch Test Map: `optionsSeamInterfaces`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `Options` 필드 순회 | 필드 26개 | — | yes |
| B2 | 인터페이스로 해석되는 필드만 집합에 | 인터페이스 seam 6개 + `len(seams) == 0` positive control | yes | yes |

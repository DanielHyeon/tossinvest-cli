# Branch Test Map: `resolveDeclared`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 사슬 순회 | 인터페이스 seam 6(1홉) + 별칭 사슬 변이(2홉 이상) | yes | yes |
| B2 | 이름이 아닌 선언에 도달 | func 타입 seam 8 | — | yes |
| B3 | 미선언 이름 또는 순환 | 순환 별칭 변이 | — | yes |

# Function Logic Map: `_isolated_tree` (Python, a122 task 7.5.34 — 새 함수)

`ast.after-7.5.34.json` — 분기 **4** · raise 3 · 반환 1.

`entries` 만 담은 트리를 격리한 저장소에 세운다: 임시 인덱스(`update-index -z --index-info`) → `write-tree --missing-ok`.

| 분기 | 뜻 |
|---|---|
| B1 | 항목마다 `<모드> <oid>\t<경로>\0` |
| B2 | `update-index` 실패 → 결함 |
| B3 | `write-tree` 실패 → 결함 |
| B4 | `write-tree` 의 답이 온전한 oid 가 아니다 → 결함 |

`--missing-ok`: 두 쪽이 같은 oid 인 항목은 diff 가 내용을 안 읽으므로 쓰지 않는다. **두 호출 모두** `INDEX_WRITE_PINS`
(fsmonitor 끔 · split index 끔 · 훅 끔)를 받는다 — `write-tree` 도 캐시 트리를 인덱스에 다시 쓰므로 split index 면
`sharedindex.*` 를 실제 `.git` 에 썼다(7.5.34 에서 `update-index` 에만 고정을 준 첫 판본을 시험이 잡았다).

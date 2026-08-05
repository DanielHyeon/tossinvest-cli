# Branch Test Map: `TestTheReloadCellAndTheMetaTagAreOneFact`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 셸을 그리는 화면을 모두 돈다 — 재로드 5개, 무재로드 나머지 | `TestTheReloadCellAndTheMetaTagAreOneFact` | no (a080 이전부터 통과) | yes |
| B2 | meta 태그와 셀의 on/off가 갈라진다 | `TestTheReloadCellAndTheMetaTagAreOneFact` | no (a080 이전부터 통과) | yes |
| B3 | 재로드 없는 화면이 `걸려 있지 않음`을 안 쓴다 | `TestTheReloadCellAndTheMetaTagAreOneFact` | no (a080 이전부터 통과) | yes |
| B4 | 재로드 화면의 셀에서 주기를 읽는다. 숫자가 없으면 즉시 중단 | `TestTheReloadCellAndTheMetaTagAreOneFact` | yes — 변이 M-F2 | yes |
| B5 | 읽어낸 주기가 선언된 주기와 다르다 | `TestTheReloadCellAndTheMetaTagAreOneFact` + 변이 M-F2 | **yes** | yes |

## 변이 M-F2 — 이 change가 여는 눈

리뷰어가 실행한 변이를 그대로 재현했다. `templates.go`의 재로드 셀을 고정값으로
바꾼다.

```
-{{if .Refresh}}{{.RefreshSeconds}}초마다{{else}}
+{{if .Refresh}}15초마다{{else}}
```

**편집 후 (RED, 실측):**

```
[FAIL] TestTheReloadCellAndTheMetaTagAreOneFact
   status_strip_test.go:240: /dashboard: the strip says 15s and the meta tag uses 5s
   status_strip_test.go:240: /positions: the strip says 15s and the meta tag uses 5s
```

**편집 전에는 이 변이가 통과했다.** 옛 단정은
`strings.Contains(cell, strconv.Itoa(screen.reload)+"초마다")`였고
`screen.reload == 5`에 대해 `strings.Contains("15초마다", "5초마다")`는 참이다.
`/signals`(15초)와 두 라인 화면(5초)이 선행 숫자 하나만 다르기 때문에, 그 화면의
주기로 템플릿을 고정해도 세 화면이 모두 통과했다. 이것은 실험이 필요한 주장이
아니라 `strings.Contains`의 정의다.

a080 이전에는 라인 화면이 `30초마다`였고 테이블의 어떤 값도 다른 값의 접미사가
아니었으므로 이 눈멀음이 드러나지 않았다. **주기를 바꾼 것이 이 결함을 도달
가능하게 만들었다** — a081에서 본 것과 같은 형태다.

변이는 되돌렸고(`templates.go` 원본 복원) 패키지 711건 전부 통과를 재확인했다.

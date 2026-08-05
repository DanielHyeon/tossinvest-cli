# Branch Test Map: `overviewPage.RefreshSeconds`

> **개정 2026-08-05 (review.md F5).** 앞 판은 철회된 스크립트 설계와 없는 task를
> 참조했다. `positionsPage.RefreshSeconds`의 개정과 같은 이유다.

AST `branches: null` — 분기가 없다. `positionsPage.RefreshSeconds`와 같은 이유로
계약별 표다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — 분기 없는 유일 경로. 엔진 관측 주기를 초로 환산해 5를 반환한다 | 3.1 | yes | yes |
| C1 | 렌더된 meta 태그가 `content="5"` | 3.1 | yes | yes |
| C2 | 출처 분리 — `holdingsTTL`을 바꿔도 이 값이 따라 바뀌지 않는다 | 3.1 | yes | yes |
| C3 | 스트립 reload 셀과 meta 태그가 계속 한 사실이다 | 3.3 | no | yes |
| C4 | `positionsPage.RefreshSeconds`와 같은 값을 렌더한다 | 3.2 | yes | yes |
| C5 | shadowing 미발생 — `overviewPage`가 자기 메서드를 계속 갖는다 | 3.4 | no | yes |
| C6 | 브로커 호출 **0회** — 어떤 주기에서 재로드해도 `/dashboard`는 브로커를 부르지 않는다 | 2.3 | no | yes |
| C7 | 엔진 부하 — 주기가 6배가 되어도 엔진 프로세스 읽기는 간격당 1벌을 넘지 않는다 | a081 2.2, 2.4 | yes | yes |

C6이 `positionsPage`의 것과 다르다. 저쪽은 "TTL당 1회를 넘지 않는다"이고 이쪽은
"0회"다 — `peek` 계약(design D2)이며 화면에 문구로도 쓰여 있다. 갱신 주기를 5초로
내리는 것이 이 0을 깨뜨리지 않는다는 것이 이 change에서 가장 쉽게 깨질 수 있는
성질이므로 별도 테스트로 고정한다.

C7은 `/dashboard`에서 특히 중요하다. 브로커에 0원인 이 화면이 **엔진에는 공짜가
아니었다** — `decoratePositionRows`가 렌더마다 엔진을 두 번 읽었고, "브로커 호출
0회"라는 검증이 그것을 가려 왔다. a081이 그 부분을 고친다.

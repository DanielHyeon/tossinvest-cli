# a175 P3 프리셋 표·관련 화면 격리 측정

- 대상: `293a19f` 이후 라운드 2 작업트리.
- 환경: 임시 `GOCACHE`·`GOFLAGS=-trimpath` 빌드, 빈 HOME·빈 config-dir·자격증명 0, loopback `127.0.0.1:38352`, `--bind` 미사용.
- 방법: a173 `shoot.cjs`와 기존 Playwright만 사용해 GET 이동만 수행함. 버튼·프리셋·preview·apply·rollback·계산을 포함한 클릭은 0회임.

| 화면 | 뷰포트 | 페이지 높이 | nav | 본문 | 표 scroll/client | 페이지 scroll/client | 판정 |
|---|---|---:|---:|---:|---:|---:|---|
| `/optimization` | 1280px | 1328 | 667 | 518 | — | 1280/1280 | 관련 화면 목록 제목·평문 링크 3개 |
| `?category=exit-protection` | 1280px | 3479 | 667 | 2929 | 879/879 | 1280/1280 | 프리셋 표 가로 overflow 없음 |
| `?category=exit-protection` | 390px | 5233 | 518 | 3972 | 576/342 | 390/390 | 표 안 가로 스크롤은 의도된 모바일 동작 |
| `?category=position-management` | 1280px | 1217 | 667 | 386 | — | 1280/1280 | 링크 1·KPI 0 유지 |
| `?category=strategy-runtime` | 1280px | 1217 | 667 | 337 | — | 1280/1280 | 두 시장 열·기술 계약 접기 유지 |
| `?category=performance-history` | 1280px | 1217 | 667 | 354 | — | 1280/1280 | 감사 이력 영역 유지 |

round 1의 내비 710px와 비교하면, 관련 링크가 category card가 아닌 목록으로 바뀌며 이번 1280px 내비는 667px가 됐음. 여백·설명을 측정 목표에 맞춰 누르지 않았음.

격리 콘솔 PID `2395197`은 측정 셸 종료 시 종료했고, 38351·38352 listener 부재를 `ss`로 확인함. 임시 측정 디렉터리는 후속 기록 전용 값을 옮긴 뒤 삭제함.

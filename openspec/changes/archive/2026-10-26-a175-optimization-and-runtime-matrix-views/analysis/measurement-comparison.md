# a175 사전·사후 화면 측정

측정 도구: a173 `shoot.cjs` GET 전용 하네스. 기준선은 `baseline-20261026T1019Z.json`, 최종값은 `after-review-20261026T1130Z.json`임. 두 테마의 값이 같아 light만 표기함.

| 화면 | 1280px 전 | 1280px 후 | 390px 전 | 390px 후 | 판정 |
|---|---:|---:|---:|---:|---|
| `/optimization` | 1684 | 1371 | 2696 | 1954 | 개요 카드 삭제·KPI 범위 제한으로 감소 |
| `?category=position-management` | 1218 | 1261 | 1858 | 1596 | 목표 700 초과. 링크·설명·여백을 누르지 않음 |
| `?category=strategy-runtime` | 1218 | 1261 | 1791 | 1597 | 목표 700 초과. 링크 2개·안전 설명을 유지함 |
| `?category=candidate-filters` | 3950 | 2073 | 5823 | 3907 | 목표 1200 초과. Help 설명만 접고 preview 계약·KR·US 값·사유·미측정·CAS는 접거나 삭제하지 않음 |
| `?category=exit-protection` | 3445 | 3424 | 5898 | 5083 | 프리셋 표 행렬화, 390px는 허용된 표 스크롤 |
| `/strategy-runtime` | 3165 | 1499 | 3275 | 2398 | KR·US 2열·모바일 1열로 감소 |
| `/signals` | 900 | 900 | 941 | 941 | P2 동일 공백 격리 상태(`discovery_unreadable`) 재측정. 시장 행 없는 오류 상태의 공통 설명을 제거해 base와 같음 |
| `?category=performance-history` | 1218 | 1261 | 1796 | 1588 | 감사 이력 표 유지, 목표 적용 대상 아님 |

## 잘림·격리 확인

- 프리셋 비교표는 1280px와 390px 스크린샷에서 `.table-scroll` 영역 안에 있으며, 390px의 가로 스크롤은 의도된 표 동작임. 페이지 본문 잘림은 관찰하지 않음.
- 후보 필터·runtime의 모바일 화면은 값을 줄이거나 숨기지 않고 줄바꿈함. 후보 필터 높이 목표 초과는 그대로 기록함.
- 최종 촬영은 빈 HOME·빈 config·자격증명 0의 `127.0.0.1:38329` 격리 콘솔에서 GET만 수행함. `after-review-20261026T1130Z-run.txt`에 PID `1944443` 종료와 포트 해제가 기록됨.

## P2 동일 상태 보정

- 기존 3243px/3640px 신호 결과는 900px/965px 기준선과 availability 상태가 달라 전후 비교 근거가 될 수 없었음. 동일 빈 HOME/config를 매번 같은 경로명으로 다시 만든 base·HEAD 비교로 행을 보정함.
- `/signals`의 h1·주요 section 수·cause code와 중계 카테고리별 `.optimization-content`·`.optimization-nav` 높이는 `p2-isolated-measurement.md`에 기록함. 중계 카테고리 데스크톱 전체 높이 증가는 본문이 아니라 내비 513→710px가 결정함.

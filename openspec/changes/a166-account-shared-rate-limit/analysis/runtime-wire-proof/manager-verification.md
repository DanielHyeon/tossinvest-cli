# runtime 연결 계측 시험 Manager 판정

2026-10-31. `16b17a470108f8e19ea7da948d0791ed5262a3c3`의 test-only helper와 external 시험 원문을 손으로 읽고 독립 Go 1.25 race 원문을 확인했다. 선택 시험 5.137초 PASS, 두 자식의 총 8개 관측, token/accounts 각 2회, 최소 인접 표본 377.871267ms를 확인했다. 작성자의 간격 제거·계측 제거 변이는 각각 반증을 제공한다. 추가 P0/P1 발견 없음으로 시험 추가분의 부분 병합을 승인한다.

private guard의 포인터를 reflect/unsafe로 읽되 client/transport/authority를 교체하지 않고 요청 전에 관측 callback만 설치한다. helper는 `_test.go`에만 있어 제품 바이너리의 API를 넓히지 않는다. 관측 표본은 callback의 CLOCK_MONOTONIC이며 그 뒤 파일 I/O와 스케줄링이 있다. 따라서 instrumentation 아래 동작의 증거로 한정한다. 비계측 실제 Write의 정확한 시각·브로커 도착률·운영 namespace는 증명하지 않는다. 기존 소유권 전체 수명/350ms 규칙의 소스·단위 증거와 이 통합 표본을 구분한다.

Terra는 독립 원문과 본 판정을 보존하고 main 통합 후 scoped strict/PM/diff 검사로 non-force commit/merge/push할 수 있다. 제품 코드·운영 설정 변경은 없으며 a166 전체 완료·archive는 아니다. 전체 시험/gate·실계좌·배포 미실행. gstack은 수동 범위 검토, impeccable N/A — test/documentation-only.

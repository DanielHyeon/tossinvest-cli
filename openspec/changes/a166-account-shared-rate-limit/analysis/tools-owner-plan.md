# 독립 도구의 budget 수명 — 검토안

기준 `686842b0889f257389b1ab99fcc31cf5f1dea674`. `/tmp/a166-tools-lifetime.md`는 source 검색·손읽기 지도이며 실행·배포 증명이 아니다. Manager는 a112 cached-token reader, a130 method closure, a087 probe opener와 각 run/exit 경계를 대조했다. 마지막 절의 독립 검토 처분을 반영해 이 문서 범위 구현을 승인했으며, 각 완성 로트는 별도 독립 검증 후 게시한다.

## 공통 설정과 소유

승인된 두 환경변수 판정은 root의 private parser와 같아야 한다. `accountbudgetruntime`에 부수효과 없는 작은 launch-config parser로 추출해 root와 tools가 공유한다. 두 항목의 값을 한 번만 snapshot하고 기존 absent/empty/unknown/off+manifest/required missing 행렬을 보존한다. 기존 단일변수 FromEnv를 조용히 재해석하지 않는다. 새 service·registry·프로세스 간 Go 객체 공유는 없다.

각 tool 실행은 factory 하나를 소유한다. 기존 실제 official client 생성 지점에서만 지연 Open하고, 그 실행의 모든 재생성 client와 method closure가 같은 factory를 빌린다. dry-run/입력 거절/읽을 대상 없음 등 client가 필요 없던 경로에 budget 파일 접근을 추가하지 않는다. factory 실패를 raw constructor나 빈 성공으로 바꾸지 않는다. 기존 profile/credential/token/계좌 선택 규칙을 보존한다. 서로 다른 프로세스는 동일 trusted manifest·고정 파일 ID를 검증한다. 실행 중 env를 재조회해 OFF로 전환하지 않는다.

main의 defer만으로 Close를 처리하지 않는다. 각 execution/run helper가 모든 method closure와 요청 사용을 마친 뒤 Close하고, 그 후 main이 원래 exit code로 종료한다. 기존 작업 오류와 Close 오류를 함께 보존하며 성공 code=0이었던 작업의 Close 실패는 실패로 드러낸다. 이미 정해진 비영 code(예: rate-limit 2)는 Close 실패 때문에 0으로 바꾸지 않는다. signal context·기존 per-request timeout을 보존하고 무기한 Background 대기를 새로 넣지 않는다. 클라이언트를 반환하는 helper의 내부에서 factory를 미리 닫지 않는다.

## 도구별 범위

- a112-l1c-quote-probe: 호가·가격 읽기와 기존 30초 timeout을 유지하며 auxiliary context를 두 읽기와 그 token 사슬에 전달한다.
- a130-outcome-backfill: `openCandles`의 bound method가 마지막으로 쓰일 때까지 owner를 유지한다. 읽기는 auxiliary다. 원장의 preview/confirm/execute 및 토큰 캐시 공유 정책은 그대로이며 운영 원장에 도구를 실행하지 않는다.
- a128-jev-shadow-probe, a128-d2-receipts, a140-breakout-signal-shadow: 공식 측정 GET은 auxiliary, 외부 judge HTTP는 공식 budget 대상이 아니다. 각 dry-run·receipt·ledger·종료 code를 유지하며 토큰 교환도 같은 요청 context와 budget을 사용한다.
- a112-mb-us-source: 빈 credentials와 사전 cached token으로 수행하는 기존 single-GET/no OAuth/no Accounts/no retry 계약을 유지한다. factory.NewClient 생성만으로 OAuth/Accounts를 실행하지 않으며 A112MBUS의 별도 guard clone에 같은 budget을 결속한다. 새 expiry 갱신이나 인증 절차를 추가하지 않는다.
- a087-market-sell-probe: 일반 사람 측정 요청이므로 normal이다. SELL·token이라는 이유로 critical로 올리지 않는다. main에서 required를 명시한 경우에만 기존 required budgetClient/budgetRequired 경계를 실제 factory와 결속한다. 별도 POST transport도 같은 budget이며, 기존 OFF는 유지한다. preview·사람 confirm·POST 1회·receipt·Sent 의미·주문 유형을 바꾸지 않는다. 이 개발은 실제 probe 실행이나 a087 활성화 승인이 아니다.

앞의 auxiliary는 전체 3/s cap의 예외가 아니다. 기존 `tryAux`는 queue/wire가 점유돼 있으면 기다리지 않고 즉시 거절한다. 따라서 required 측정 도구는 혼잡 시 기존보다 실패/HOLD가 늘 수 있고 token 교환도 같은 정책이다. 이 비용을 숨기거나 자동 retry로 바꾸지 않는다. 도구별 기존 오류·exit code·receipt 분류를 유지하고, queue 점유와 wire 점유 각각에서 wire 0, token 단일 시도, a112 cached-only/no-retry, OFF 무변화를 시험한다. 도구 내부 실제 bootstrap 호출은 소스와 local fixture로 확인하며 이름만으로 token/Accounts가 없다고 주장하지 않는다. 운영 환경의 모든 참여자 포함 여부는 별도 사람 검증이다.

## 수용과 제한

기존 함수별 사전 AST/FLM/BTM, parser의 root/tools 동일 행렬, required 실패 bootstrap wire0, 성공 factory 재사용, 실제 반환 method 사용 뒤 Close, 성공·오류·취소와 기존 exit code, dry-run 파일 접근0, a112 no OAuth/Accounts, probe bootstrap+POST 같은 budget·원문 body·단일 전송을 시험한다. fake credentials와 격리 파일만 사용하며, 모든 네트워크 시험·변이는 외부 egress 없는 user/network namespace에서 local TLS·explicit transport로 수행한다. 세부 default client의 독립 transport를 global DefaultTransport 교체로 막았다고 주장하지 않는다.

검증은 정확 선택 normal/tag/race·변이·수동 gstack·독립 리뷰다. 전체 tests/gate·실제 도구 실행·live 주문·운영 timer/env/배포 변경은 범위 밖이다. search-first는 기존 runtime factory·root parser·run helper와 stdlib context/errors 재사용이며 새 의존성은 없다. impeccable N/A — CLI/backend wiring-only; 출력 계약은 기존 fixture로 확인한다. grill-me/grilling은 정해진 cap·대기·OFF·운영 경계를 재질문하지 않고 별도 실행 파일의 종료와 부트스트랩 실패를 명시한다.

## 독립 검토 처분

Terra 독립 소스 검토 `/tmp/a166-tools-owner-review.md`의 auxiliary P2를 위 수용 조건으로 반영했다. Manager가 `budget_linux.go`의 Auxiliary 분기와 tryAux의 queue/wire 비차단 처리를 직접 대조했다. 나머지 설정·수명·cached-only·probe 경계는 조건부 적합이다. 이 문서 범위의 구현을 승인한다. root parser 추출은 console writer와 파일 소유를 분리하며, 별도 parser가 남아 정책을 달리하지 않는 구조 시험을 요구한다. 완성 후보마다 독립 검증 후에만 병합한다.

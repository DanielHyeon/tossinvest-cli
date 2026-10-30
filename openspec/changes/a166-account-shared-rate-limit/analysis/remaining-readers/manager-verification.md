# Manager 검증 — reader 배선

대상 `41c801e78e5b98d221ca2d4627e60daf3bd643d8`. 이 후보의 제한된 병합·푸시를 승인한다. a166 전체 완료·archive·운영 활성화 판정은 아니다.

Manager는 생산 diff를 손으로 읽었다. verify-reconcile의 최초 Accounts와 seq 고정 client는 root factory를 공유하며 profile 전용 자격·단일 계좌·seq 검사를 유지한다. OpenAPI test/status, init, doctor, console onboarding은 root를 전달하며 factory 실패를 성공으로 감추지 않는다. 기존 nil-root helper는 호환 경로다. AST/FLM 산출물은 작성자 증거이며 Manager의 이번 대조는 수동 소스 검토다.

별도 Terra가 Go 1.25의 정확한 새 시험 6개를 normal 및 tagged race로 재실행했다. Manager는 `/tmp/a166-reader-independent-normal.log`, `/tmp/a166-reader-independent-tagged-race.log`, `/tmp/a166-reader-independent-netns.log` 원문을 열어 PASS, 2.450초/3.559초, loopback-only network namespace와 route 없음 기록을 확인했다. 독립 리뷰는 생성자/root 우회 변이 4개 RED 및 gzip 원문 14개 SHA 일치를 확인했고 P0/P1 발견 없음으로 판정했다. local TLS와 explicit transport를 사용했으며 외부 차단의 근거는 namespace다.

확인하지 않은 범위: console 전체 수명·verify factory·다른 호출자·전체 tests/gate·브로커 운영. CodeGraph hard-evidence 미초기화 한계와 앞서 기록한 test-boundary incident는 소급 해소되지 않는다. UI 기능 변경 없음, impeccable N/A — backend/CLI factory wiring; 기존 출력 오류 전파만 확인했다. 기존 factory 재사용, 신규 의존성 없음. 병합 시 a128 등 타 작업을 보존하고 PM/OpenSpec/diff 범위 검사만 수행한다.

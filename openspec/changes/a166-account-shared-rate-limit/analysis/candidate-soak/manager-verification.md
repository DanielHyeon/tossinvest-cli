# Manager 검증 — 후보 패널과 soak client

후보 `14d4bec7b`, 시험 경계 보완 `164db491eeaa0ef07c2f56ee3fa2f1365040819c`의 제한 병합·푸시 승인.

Manager가 생산 diff를 손으로 읽었다. candidate/soak 공식 client는 실제 caller context와 root factory를 사용한다. candidate factory 실패 뒤 store release, 필수 공식 rankings·선택적 WTS, soak run의 signal context와 기존 읽기를 보존한다. soak attestation의 base URL 메모는 기존 best-effort이며 factory 오류를 빈 문자열로 처리한다. 이 메모를 required 가용성 증거로 사용하지 않는다.

처음 독립 검증은 local TLS fixture에 exact DialContext pin이 없어서 네트워크 시험을 보류했다. 후속 delta에서 생성 listener의 tcp 주소만 허용하고 다른 주소를 직접 거절하는 단언을 넣었다. 생산 변경은 없다. 독립 Terra의 Go1.25 5개 선택 normal 2.456초/tagged race 3.495초 PASS. Manager는 `/tmp/a166-candidate-soak-independent-review.md` 및 final-normal/final-race 원문과 pin delta를 확인했다. 사전 AST 6개 source SHA는 독립 검토자가 대조했다.

기존 영수증에 exact dial 경계 준수를 소급하지 않는다. 초기 RED는 signature compile 실패이고 실제 행동 반례는 사후 factory 우회 변이다. 로컬 TLS fixture는 OAuth 성공 후 GET500을 반환하므로 정상 응답 데이터 전체 경로는 미검증이다. 네트워크 시험은 loopback-only/no-route namespace다. 실제 CLI·브로커·운영·전체 tests/gate는 미실행이며 자동 전체 gstack 인증이 아닌 수동 범위 review다.

impeccable: N/A — backend factory wiring-only. 기존 root factory/stdlib 재사용, 신규 의존성 없음. v2 fixture migration은 후속 runtime 로트다. 최신 main과 순차 통합·PM/OpenSpec/diff 검사 후 게시하고 main a128 자료를 보존한다. 전체 a166 archive/활성화 승인은 아니다.

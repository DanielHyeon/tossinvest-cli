# Manager 검증 — cached source와 수동 probe 배선

후보 `8fb2653554587e8a50193af345b099960e55526d`의 제한 범위 병합·푸시 승인. 실제 probe 실행이나 거래 권한 승인이 아니다.

Manager가 cli.go/official_reader.go와 a087 probe 생산 diff를 손으로 읽었다. a112는 빈 credentials와 기존 cached token을 사용하고 기존 one-shot GET·15초·OAuth/Accounts/재시도 금지 경로를 유지한다. factory는 run 종료 뒤 os.Exit 전에 닫고 read context는 Auxiliary다. a087은 기존 preview/TTY/확인/영수증 경로 가드 후 factory를 연다. required 오류는 raw fallback으로 전환하지 않으며 별도 POST는 같은 Budget와 sealed client에 묶인다. 사람 probe는 매도라는 이유로 Critical이 되지 않고 Normal이다. 기존 Sent/영수증·POST 1회 계약은 유지한다.

독립 Terra의 정확한 7개 Go 1.25 선택 normal/tagged race가 통과했다. a112 0.777초/1.796초, a087 0.796초/1.808초. Manager는 `/tmp/a166-tools-last-independent-review.md`와 normal/tag-race 원문을 확인했다. 사후 AST 6개 source SHA와 원문 gzip 12개 해시는 독립 검토자가 대조했다. 초기 TLS authority 거절·unsupported ServerName false-green 및 소비된 body 재사용 때문에 POST2가 되지 않은 변이는 제외하고 보완한 RED와 POST2 변이를 별도 보존했다.

네트워크는 loopback-only/no-route namespace, 명시 local TLS와 exact DialContext다. a112 test-only reflection/unsafe는 constructor 소유 transport의 dial/root만 변경하고 base/flag/pointer 봉인을 바꾸지 않는다. 이는 격리 fixture이며 운영 authority 검증이 아니다. a087 Normal은 AST 구조 검증이고 동적 wire priority는 미검증이다. 실제 prepareProbeBudget 자격증명 성공 경로, Close 오류 직접 주입, 실제 프로세스 exit는 미검증이다. 실제 도구·계좌·원장·브로커·전체 tests/gate·운영은 실행하지 않았다.

impeccable: N/A — backend CLI wiring-only. 기존 runtime factory 및 표준 라이브러리 재사용, 신규 의존성 없음. scoped 수동 gstack 결과이며 자동 전체 pipeline 인증은 아니다. 최신 main과 순차 통합·PM/OpenSpec/diff 검사 후 게시하며 v2 fixture migration은 별도 로트에 남긴다. a166 완료나 archive로 표시하지 않는다.

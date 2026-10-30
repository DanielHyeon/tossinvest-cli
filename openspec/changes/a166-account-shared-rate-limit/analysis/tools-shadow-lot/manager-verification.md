# Manager 검증 — JEV·a140 도구 배선

대상 `78d777cfe`. source 범위는 두 도구의 factory 생성·Auxiliary context·실행 수명과 Close 오류 보존이다. 이 로트의 병합·푸시를 승인하며 실제 측정·a169 재개를 승인하지 않는다.

Manager가 생산 diff와 실제 client 생성 이후 probe 구성·실행 원문을 손으로 읽었다. 두 env는 한 번 읽고 기존 자격/토큰 경로를 사용한다. run 반환까지 같은 factory가 살아 있으며 named error와 Close 오류를 결합한다. JEV의 기존 retry loop, WTS 별도 읽기와 judge HTTP는 바꾸지 않았다. a140의 store/ledger 수명과 calendar/probe 실행은 보존했다. Auxiliary 혼잡 거절은 새 retry 없이 기존 처리에 전달된다. JEV는 기존 loop 때문에 해당 fixture에서 context deadline으로 끝나며 a140은 unavailable로 끝난다.

독립 Terra가 Go 1.25에서 각 도구의 required 누락·OFF local token 오류·busy wire0·synthetic Close 오류 4개를 normal/tagged race로 통과했다. Manager는 `/tmp/a166-shadow-independent-go125-normal.log`, `/tmp/a166-shadow-independent-go125-tag-race.log` 원문과 loopback-only namespace/route 없음 기록을 확인했다. normal JEV 2.395초/a140 0.953초, tagged race 3.411초/2.055초다. 독립 검토는 양쪽 AST source hash와 required/priority/Close 변이 RED를 대조했다.

전체 성공 출력·JEV retry 횟수·실제 프로세스 exit code·실제 filesystem Close 장애는 이 fixture로 증명하지 않는다. 실제 운영 도구 실행·브로커·전체 tests/gate·archive·활성화는 미실행이다. 초기 잘못된 fixture 경로의 실패는 의미 있는 RED와 분리한 기록을 유지한다. impeccable N/A — CLI/backend-only, 기존 runtime과 context/errors 재사용·신규 의존성 없음. PM/OpenSpec/diff 범위 확인 후 a128의 기존 운영 ledger/state 등 타 변경을 보존하여 게시한다.

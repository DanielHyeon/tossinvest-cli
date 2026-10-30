# Manager 검증 — console 종료와 재실행

후보 `406d8b68e`, test/doc delta `dafd3f06b5863ede93ea940c287e872070464de9`. 이 제한 로트의 병합·푸시를 승인한다. verify factory·HTTP API·console broker 후속 배선과 전체 a166 완료는 아니다.

Manager가 생산 diff를 손으로 읽었다. 기존 finishedFlag가 아니라 실제 run.done으로 이전 run 교체와 settle 완료를 판단한다. HTTP Shutdown을 기존 5초 상한으로 먼저 수행하고 실제 worker done을 기존 2분 상한으로 기다린다. server/shutdown/worker 오류를 보존하고 오류 시 relaunch하지 않는다. 정상 경로의 callback은 root owner.Close 성공 뒤에만 기존 path/argv의 exec를 호출한다. Shutdown 실패 뒤 connection Close를 handler 완료로 오인하지 않으며 해당 경로는 재실행하지 않는다. Budget.Close도 미완 worker 증명의 대체물이 아니다.

독립 Terra가 Go 1.25의 정확한 새 시험 7개 normal/tagged race와 기존 port-release/worker-shutdown 시험 2개의 race를 통과했다. 추가 handler-order 시험은 normal 0.050초/tagged race 1.081초 통과했고 Shutdown→Close 변이를 검출했다. Manager는 `/tmp/a166-console-independent-{normal,tag-race,existing-race}.log` 및 `/tmp/a166-console-handler-delta-independent-{normal,tag-race}.log` 원문을 확인했다. 네트워크 fixture는 loopback-only/no-route namespace다.

추가 fixture의 owner-close와 exec는 callback 내 모의 순서이며 실제 cmd callback은 선행 owner-error/fake-exec fixture와 소스 대조를 합친 근거다. 실제 console·exec·운영·전체 tests/gate는 실행하지 않았다. 기존 branch-test-map의 미실행 표기는 실제 선택 검증 상태로 수정했다. UI 화면 변경 없음, impeccable N/A — backend lifecycle only; 브라우저 화면 검증은 하지 않았다. render/accessibility PASS로 확대하지 않는다. 기존 context/http/errors 재사용, 신규 의존성 없음. parser/tools의 동시 변경과 main a128 파일을 보존하며 PM/OpenSpec/diff 범위 검사 후 게시한다. archive는 아직 하지 않는다.

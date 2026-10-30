# a166 마지막 tools 독립 검토

대상 8fb2653554587e8a50193af345b099960e55526d, 기준 e5c9d0256. 지정 변경 범위에서 P0/P1 미발견. 이는 a166 전체 완료·실거래 권한·운영 승인 판정이 아니다.

Go1.25 선택 normal/tagged race: a112 TestA166MBCachedOff/RequiredMissing/BusyNoWire/RequiredOneGET, a087 TestA166ProbeRequiredUsesNormalPriority/RequiredMissingAfterLocalGuards/FactoryLocalTokenAndPost 모두 PASS. 원문 /tmp/a166-tools-last-independent-normal.log, /tmp/a166-tools-last-independent-tag-race.log. 각 실행은 별도 user+network namespace, loopback만 활성·route 없음. local TLS fixture는 고정 listener 주소 이외 dial을 거절한다. 실제 도구·브로커·전체 test/gate는 실행하지 않았다.

소스 및 선택 fixture: a112는 빈 credential·기존 cached token client를 factory에서 만들고 reader에 Auxiliary context를 준다. 기존 one-shot helper의 GET 1회·15초·OAuth/Accounts/재시도 금지는 해당 local fixture와 원문 소스로 구분해 확인했다. test-only unsafe/reflection은 constructor 소유 transport의 dial/root만 수정하여 봉인 pointer를 유지한다. 이는 격리 fixture이고 실제 운영 환경의 금전 권한 검증이 아니다.

a087 execute는 preview/TTY/confirm/receipt local guard 뒤 factory를 연다. required 오류는 wire0, local fixture는 OAuth 1회·POST 1회. Normal은 구조 AST 시험이며 wire별 동적 priority 검증은 아니다. 실제 prepareProbeBudget의 자격증명 성공 경로, Close 오류 직접 주입, cached token 없음/만료의 추가 조합 및 실제 프로세스 exit 경로는 미검증이다.

post AST 6개 source SHA 일치, writer 영수증 gzip SHA 12개 일치. 초기 TLS authority 거절 및 unsupported ServerName false-green과 사용한 body 재전송으로 POST 1회였던 초기 중복 변이는 유효 증거로 세지 않는다. 후속 pinned RED 및 재생성 body 중복 변이는 실제 POST 2회가 되어 실패한 원문을 확인했다. Go 코드·test fixture 수정 없음.

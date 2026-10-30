# outcome/d2 owner 사전 분석

기준 34ec59fe7643b3a20b680b663199f354c228f992. a130 run/openCandles 및 a128 d2 run 사전 AST; 기존 root/runtime parser와 factory 재사용, 신규 의존성 없음. a130 method closure 마지막 사용까지 factory 유지·run 반환에서 Close 오류 결합. lots가 없으면 factory 접근0이며 preview라도 lots가 있으면 조회한다. d2 dry-run은 factory 접근0, 그 외 기존 run timeout 5분(`runBudget`)·rate-limit exit code 보존. 두 tool의 읽기+token은 Auxiliary이며 busy는 즉시 거절. RED: local isolated required missing 오류에서 첫 wire0; 변경 이후 정상 OFF, 부하, Close와 결과 보존 fixture 필요. 원문에 실데이터 접근하지 않는다. UI N/A CLI/backend.

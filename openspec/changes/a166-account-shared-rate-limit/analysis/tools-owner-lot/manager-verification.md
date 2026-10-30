# Manager 검증 — parser와 L1c 첫 도구

후보 `10b3a96dd11035cdb59383e9c843bea086b47840`, delta `6b9ff9f60`. 이 로트의 병합·푸시를 승인한다. 나머지 여섯 도구·전체 a166 완료 판정은 아니다.

Manager가 생산 diff를 손으로 읽었다. root의 2변수 판정은 순수 runtime parser로 동일하게 옮겨졌고 legacy FromEnv는 바뀌지 않았다. L1c는 입력/자격 확인 후 factory를 열어 기존 두 GET과 token context에 Auxiliary를 전달하고, 모든 반환에서 Close 오류를 작업 오류와 결합한다. OFF factory는 기존 official.New를 사용한다. delta는 동일 errors.Join 표현을 시험 가능한 작은 helper로 추출했다. source의 30초 budget과 기존 render 호출을 보존한다.

독립 Terra 1이 parser/root matrix와 L1c의 required 누락·첫 GET 오류/무재시도·auxiliary 혼잡 wire0를 Go 1.25 normal/tagged race로 재실행했다. 독립 Terra 2는 delta의 두 시험을 같은 계열에서 normal 0.018초/tagged race 1.047초로 통과했다. Manager는 `/tmp/a166-tools-independent-{normal,tag-race}.log`, `/tmp/a166-l1c-delta-independent-{normal,tag-race}.log` 원문을 확인했다. local TLS·loopback-only network namespace/route 없음이며 실제 probe가 아니다. AST source hash와 변이 RED는 독립 검토자가 대조했다.

출력 fixture는 AAPL 포함·token과 두 GET의 3회 호출을 확인하며 전체 출력 golden은 아니다. deadline fixture는 양수이면서 30초 이하를 확인하고 30초 상수 자체는 소스 대조다. Close 오류는 synthetic seam이고 실제 filesystem 장애가 아니다. 선행 행동 RED 누락과 뒤의 원본 복원 RED 절차는 review.md 기록을 유지한다. 전처리되지 않은 기본 endpoint 반례는 외부 route가 없는 namespace에 차단됐으며 local TLS fixture로 소급 설명하지 않는다.

impeccable N/A — CLI/backend 배선, UI 변경 없음. 기존 runtime/context/errors 재사용, 신규 의존성 없음. 자동 전체 gstack/전체 tests/gate·운영·추가 도구·a087 활성화·a169 측정은 미실행. PM/OpenSpec/diff 범위 검사 후 기존 a128 변경을 보존하여 종결 커밋한다. archive는 아직 하지 않는다.

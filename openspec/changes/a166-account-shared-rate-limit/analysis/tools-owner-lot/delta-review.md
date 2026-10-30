# L1c 추가 검증

기준 10b3a96dd. 기존 `post-runWithOptions.json`이 delta 전 AST이며 `post-runWithOptions-delta.json`은 이후 원문이다. 실제 OFF local TLS token + 두 GET 성공, 출력 및 세 요청 context의 deadline이 30초 안에 있음을 확인한다. 정확 loopback 주소를 강제하는 DialContext와 OS lo-only namespace를 사용했다. `finishL1C`는 기존 defer의 오류 결합을 작은 함수로 추출하여 작업 오류와 Close 오류 양쪽 보존 및 성공 뒤 Close 오류 노출을 synthetic Close 함수로 검증한다. 실제 파일시스템 Close 오류를 유발한 시험은 아니다.

`a166-tools-l1c-close-red.log`는 helper 도입 전 컴파일 RED, `a166-tools-l1c-close-mutant.log`는 Close 오류를 버리는 overlay 변이의 assertion RED다. 선택 normal/tagged race는 delta-* 원문. 이전 세 fixture는 이 delta의 영향이 없어 재실행하지 않았다. 기능 변경은 Close 결합 표현의 추출뿐이며 추가 tools/운영/실제 probe는 미구현·미실행. scoped 수동 gstack 및 독립 검토가 필요하다.

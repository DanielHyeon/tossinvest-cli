# 마지막 두 도구 제한 리뷰

기준 e5c9d0256, 사전 FLM/BTM은 `pre-edit.md`, 변경 전후 AST는 이 디렉터리 JSON. 관련 gstack 항목을 수동으로 점검했다. 전체 자동 gstack 또는 외부 모델 리뷰는 실행하지 않았다. impeccable N/A: CLI/backend. 저장소의 runtime factory, 공식 guard와 표준 라이브러리만 사용, 새 의존성 없음.

a112: 빈 credential의 cached-token A112 client를 factory로 생성하고 반환 전 Close. 세 reader에 auxiliary context, 기존 single GET·15초 제한·no OAuth/no Accounts/no retry. OFF 및 required에서 고정된 local TLS GET 각각 1회, required 누락 거부, held critical 중 wire0·빠른 auxiliary 거부. test-only unsafe/reflection은 constructor 소유 transport의 dial/root만 local fixture로 변경하고 base URL/authority 봉인은 바꾸지 않는다. production 코드는 reflection을 사용하지 않는다.

a087: preview 및 local confirm/receipt 가드 이후에만 factory 생성. OFF는 기존 opener, required는 같은 factory Budget과 official client 및 별도 POST guard; Normal priority. fixture는 token 1회 및 POST 1회, missing manifest 거절, confirmation 실패에서 factory 호출0. 우선순위 시험은 AST 구조 시험이며 wire별 우선순위를 직접 관측한 시험은 아니다. 정상 required 테스트는 실행 경계에서 test-injected factory를 쓰므로 실제 `prepareProbeBudget`의 자격증명 성공 경로는 직접 실행하지 않았다. Close 오류 직접 주입, cached-token 없음/만료 새 시나리오 및 preview 전체 조합은 미검증이다. 기존 요구된 각 상태와 반환을 전체 범위로 PASS 주장하지 않는다.

Go1.25 선택 normal·tagged race 원문, 올바른 pinned RED와 변이 원문은 gzip으로 보존. a112 첫 TLS fixture는 authority 사전거부로 false-green, 첫 busy fixture는 unsupported ServerName 사전거부로 false-green: 모두 제외하고 대체 시험을 수행했다. a087 첫 duplicate 변이는 사용한 body를 재사용해 wire1로 실패했으므로 실제 중복 증거가 아니다; body 재생성한 두 번째 변이는 wire2로 실패한다. 실계좌·운영·도구 직접 실행·전체 tests/gate·아카이브 미실행. 독립 검토 및 Manager 판정 전 게시 금지.

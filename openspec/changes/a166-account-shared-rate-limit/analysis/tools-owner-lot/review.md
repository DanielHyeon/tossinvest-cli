# tools owner 첫 로트 검토

root 2변수 parser를 runtime의 순수 ParseLaunchConfig로 이관하고 root가 위임한다. a112 l1c의 client 사용 기간에만 factory를 열고 Close 오류를 반환하며 두 GET/token context에 Auxiliary를 준다. 나머지 여섯 tools는 미구현이며 이 기록은 전체 caller/운영 완료가 아니다. 독립 Manager 검증 전이다.

분기·검증: parser absent/off/required/empty/unknown 등은 TestLaunchConfigMatrix와 기존 TestA166RootBudgetEnvMatrix; l1c 필수 manifest 오류는 `TestA166L1CRequiredMissingManifestFailsBeforeWire`, local TLS 정상 token 후 GET 오류의 1회 종료는 `TestA166L1CUsesLocalTLSAndStopsOnReadError`, 타 critical lease 보유 시 auxiliary 즉시거절과 local server 0은 `TestA166L1CAuxiliaryBusySendsNothing`. pre/post AST 및 pre-edit.md 보존. 로컬 no-route namespace normal/tagged race 영수증은 a166-tools-lot1-*; fake credential만 사용, live broker 없음.

절차 보정: l1c 기존 함수에 대한 행동 RED 전에 초안을 편집했다. 내 source patch만 보존하고 HEAD 원본을 복원해 missing-required 반례의 baseline RED를 별도 lo-only namespace에서 실행한 뒤 patch를 재적용해 GREEN을 확인했다. 사전 행동 RED 순서를 지켰다고 주장하지 않는다. baseline/mutant는 fixture 옵션이 없는 기본 endpoint를 호출하려 시도했으나 namespace에 외부 route가 없어 연결되지 않았다. local TLS endpoint 고정 fixture와는 별도 증거다. parser는 새 API compile RED→GREEN. parser 우회 변이는 a166-tools-l1c-mutant.log RED.

수동 gstack 범위: 입력 거절·env matrix·factory 소유/Close·OFF/required 분기·token/GET priority 및 local egress 경계를 손 읽기했다. 자동 전체 리뷰/독립 리뷰는 미실행. 전체 tests/gate, 다른 tools, 실제 probe·live·운영·deployment는 미실행. auxiliary 혼잡은 immediate error이지 대기가 아니다.

# candidate·soak factory 부분 후보

기준 42009abb4. candidate panel과 soak wiring의 공식 client를 root factory로 만들고 각 호출의 context를 전달한다. Candidate 공식 rankings 필수·WTS optional, soak record와 읽기 경로, factory 실패 시 candidate store 정리를 유지한다. soak attestation의 base URL 노트는 기존 best-effort 경로이며 오류를 빈 노트로 다룬다; 네트워크를 보내지 않지만 이 노트가 factory 가용성 증거는 아니다.

사전 source/AST/분기 분석은 preedit/. 컴파일 RED→GREEN 및 required 설정 오류·취소, local TLS token/GET 전 budget 경합, candidate 실패 시 store release, OFF 선택 회귀는 Go1.25 선택 normal/tagged race를 lo-only/no-route namespace에서 확인했다. candidate와 soak 각각 factory 우회 변이는 local fixture에서 RED다. 영수증 원문은 receipts.md. local TLS 요청의 fake server는 정상 OAuth 후 의도적으로 GET 500을 반환하므로 정상 응답 데이터의 전 경로를 검증했다고 주장하지 않는다.

수동 gstack 범위: ctx 및 factory 실패 전파, official/WTS 선택·store 정리·token 경로, attestation best-effort 경계를 검토했다. CodeGraph worktree 인덱스 부재, 자동 mixed-model gstack·독립 검증은 미실행이다. 신규 의존성 없음; UI 재설계 없음(impeccable N/A). 전체 tests/gate, 실 broker·운영·archive 및 a166 전체 완료는 범위 밖이다.

초기 RED는 새 context signature의 컴파일 RED다. 행동 반례는 구현 후 factory 우회 변이로 별도 재현했고 이를 편집 전 행동 RED라고 주장하지 않는다. 임시 required fixture는 현재 v1 factory 계약을 사용하며 향후 runtime v2 통합 시 별도 갱신이 필요하다.

독립 검토 P1: 기존 local TLS fixture는 lo-only namespace와 local base를 사용했지만 DialContext의 exact destination 고정이 없었다. 그 이전 영수증을 완전한 승인 경계 PASS로 확대하지 않는다. 후속 delta는 생성한 listener의 tcp 주소만 허용하고 다른 주소를 명시 거부하는 fixture를 추가했다. 생산 코드 변경 없이 Go1.25 선택 normal/tagged race 재실행; 원문은 pin-receipts.md에 보존한다. 독립 재검증 전이다.

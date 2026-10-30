# candidate/soak 사전 분석

기준 42009abb4; 함수별 AST와 source SHA는 옆 JSON. CodeGraph worktree 인덱스 없음, 아래는 원문 손읽기.

| 경로 | 현재 분기 | 변경·보존 |
|---|---|---|
| candidateWiring | store 생성 후 panel 실패면 release; 빈 panel 거절 | 현재 ctx를 panel 생성에 전달, 정리와 빈 panel 정책 유지 |
| buildCandidatePanel | 공식 자격 필수, WTS session은 optional | 첫 token 전 root factory; WTS 실패 강등 정책 유지; ctx 취소 전파 |
| runSoakRun/buildSoakWiring | signal child ctx가 있으나 factory는 context 없음 | child ctx로 root factory; token 경로와 survey·record 유지 |
| runSoakAttest/soakSurveyedBase | 기존 best-effort base URL 노트, 실패 시 빈 문자열, 네트워크 전송 없음 | attestation ctx 사용; 기존 메타데이터 실패 정책 유지 |

BTM: required 잘못된 설정·취소 wire0, local TLS 첫 token/공식 호출 결속, OFF/WTS optional 기존 선택, candidate store 실패 정리, soak 기존 wiring/record와 factory 실패 전파. 모든 network fixture는 lo-only 별도 namespace+local TLS pin, 가짜 자격만 사용. 실제 broker·전체 tests/gate·운영 금지.

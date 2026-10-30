# 수동 gstack review / 시험 범위

기준 5f681f2a5, 기존 RoundTrip 본문 사전 AST/FLM/BTM. gstack review SKILL.md와 checklist의 critical 및 informational 항목을 변경분에 수동 적용: 상태 내구성·lock·sent·body 소유·에러·concurrency·분기·시간경계를 검토했다. SQL/LLM/화면은 해당 없음. 업그레이드·telemetry·자동 specialist는 실행하지 않았으며 이 검토는 독립 review를 대체하지 않는다. 신규 구현 시 기존 store/clock와 표준 라이브러리 재사용, 새 의존성 없음. impeccable N/A — backend-only.

수정 후 Go 1.25 선택 race 10.013초, Go 1.26 선택 normal 8.954초. observed 시험 및 기존 transport 소유/취소/drain/template 시험만 실행했다. lo-only user/net namespace, route 없음, 새 TLS fixture는 DialContext가 생성된 listener 주소만 허용; fake credentials만 사용한다. raw gzip은 원문과 SHA-256 대조한다. `netns-final-race.log`의 한 실패는 Begin 저장 중 취소 결과를 정상 완료로 잘못 가정한 초기 시험의 RED이며, leaf의 sequence 0·orphan 계약을 반영한 수정 후 위 최종 실행에서 PASS했다. poll 재검사·429 collector·Finish·취소 분리 제거의 네 변이는 각각 대상 시험 RED. 변경 파일의 테스트는 종료된 실행과 일치한다.

미배선 low-level helper는 deployment identity나 모든 생산 wire의 관측을 증명하지 않는다. pending은 TLS handshake 전에 기록하므로 HTTP 첫 Write 증거가 아니며 crash 때 false gap이 가능하다. 실제 filesystem 전원 장애·실 broker·runtime v2·token/Retrier/caller 연결·전체 test/gate·운영·archive는 미실행이다. callback은 동시 안전하고 신속히 반환해야 하며 2초는 event lock 대기만 제한한다. 독립 검토 대기.

최종 코드에서 poll skip과 동시 취소 오류가 collector에서 사라지는 결함을 수동 검토로 발견해 errors.Join으로 수정했다. v2 로그는 수정 후 선택 시험이며 별도 body-race는 429 header 기록 뒤 body의 UnexpectedEOF가 event를 지우지 않음을 검증한다. 이전 오류를 복원한 다섯 번째 변이는 RED다.

# 오프라인 예산 도구의 Manager 검증

2026-10-31, 후보 ef62f4cdd792290987ada38a7830c36548627dc7, base cf0f0919e. **개발 로트 승인**. a166 전체 완료·archive·운영 사용 승인이 아니다.

Manager는 신규 runtime helper 전체와 CLI, 사전 분기/시험 지도, 부분 실패 fixture, freeze3 normal 원문, 세 변이 및 독립 보고서를 대조했다. 테스트 실행은 Terra가 담당했다. root 사전 fd identity와 경로 재검사, 검증된 parent fd에 결속한 O_EXCL/RENAME_NOREPLACE 게시, write/file sync/close/parent sync 순서와 부분 root·게시 manifest 보존은 승인 설계와 일치한다. core Provision의 경로 기반 한계와 동일 UID 적대적 rename을 막는다는 주장은 하지 않는다.

독립 Terra는 Go 1.25 정확 선택 `^TestA166Offline`을 runtime·도구 두 패키지에서 별도 재실행했다. normal 0.260/0.043초, tagged race 1.288/1.057초 PASS. loopback-only/no-route 및 임시 root를 사용했다. 9개 압축 원문과 fixture SHA 일치, 세 변이 RED 원문을 확인했다. 독립 재변이는 수행하지 않았다. 작성자 gstack은 수동 범위 checklist이며 자동 전체 pipeline이 아니다. 기존 함수 변경은 없고 신규 분기 지도와 직접 호출 AST 검사를 사용했다.

독립 P2를 반영했다: parent fsync 실패 뒤 check 성공은 이전 provision을 성공으로 바꾸거나 내구성을 증명하지 않는다. 운영 사용 전 사람 처분이 필요하다. check는 실제 core/event state, boot source, Close를 확인하지만 writer 부재·namespace 동등성·모든 참여자 또는 운영 readiness를 입증하지 않는다. nil context는 runtime Open에서 오류다.

선택 시험의 owner/filesystem은 합성 stat/magic, I/O 장애는 injection이다. 실제 타 UID·filesystem·전원 장애·운영 경로·브로커·전체 tests/gate는 미검증. CodeGraph 새 색인 검증을 주장하지 않는다. impeccable N/A — backend/CLI 문서만 영향. 기존 stdlib·x/sys 및 runtime/provision 재사용, 의존성 추가 없음.

최신 main 대조·scoped OpenSpec/PM/diff 확인 뒤 Terra가 이 로트만 정상 병합·푸시한다. a166 tasks 4.2와 a087/a169 사람 단계는 유지하며 archive하지 않는다. 독립 원문은 게시 커밋에 별도 보존한다.

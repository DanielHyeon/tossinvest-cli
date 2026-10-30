# Manager 검증 — wire 구조 검사와 consumer 합성

후보 `9dab3dda4`, 독립 검토 후보 `072440a28` 및 symlink 수정 `cbd3ae98b`. Manager는 검사 구현·목록·변이 정의와 결과 원문, 독립 검토, consumer 합성 기록을 읽었다. 제품 소스 변경 없는 정적 검증 로트의 게시를 승인한다. 문서 상태 교차표의 독립 확인과 scoped strict/PM/diff 검사는 게시 전 남아 있다.

독립 normal/tagged race에서 발견된 directory symlink 검사 순서 P2는 수정 후 fixture로 해소했다. 최신 consumer 합성은 83 함수 shape·116 selector 위치를 유지하며 Snapshot은 HTTP를 추가하지 않는다. 정확 선택 normal 1.454/1.399초, tagged race 10.264/8.243초에서 구조 검사, 8개 source-copy 변이, symlink 반례와 a165/a166 지정 회귀 3개가 통과했다. 실행은 lo-only/no-route이며 replay fixture는 exact local TLS다. Manager가 직접 시험을 실행하지 않았고 Terra의 고정 후보 독립·합성 증거를 검증했다.

shape 동일성은 본문 변경 감지이며 의미적 값 흐름 증명과 같지 않다. selector reason은 손읽기 분류다. 임의 동적 호출·reflection·새 wrapper·endpoint 전역 변경·운영 참여자의 완전성은 증명하지 않는다. 각 로트의 실제 client 행동 시험과 함께만 해석한다. 전체 tests/gate·실계좌·운영·activation 미실행. archive는 미완료 사람 단계 때문에 보류한다.

최종 문서 독립 검토는 P0/P1 없음. P2(운영 절차 초안의 완료 오인)는 tasks 4.1을 미완료로 유지하고 개발 상태표를 초안으로 고쳐 해소했다. 4.3 scoped strict/PM/diff 결과는 Terra 게시자가 기록한다.

게시 전 범위 검사: OpenSpec 1.4.1 a166 strict, PM tracker 재생성 및 --check, 변경 파일 diff --check PASS. tasks 4.3은 이 검사 범위만 뜻한다. 전체 tests/gate는 실행하지 않는다.

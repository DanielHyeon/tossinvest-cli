# 디렉터리 symlink 검사 delta — 사전 FLM/BTM

기준 `072440a28`. `wiringInspect`는 WalkDir 오류를 반환한 뒤 디렉터리·비 Go·test 파일을 건너뛰고, 그 뒤 Go 파일 symlink만 거절한다. 따라서 확장자 없는 디렉터리 symlink는 순회되지도 거절되지도 않는다. 수정: WalkDir 오류 다음 모든 항목의 symlink를 먼저 거절하고 나머지 필터·목록·shape 경로는 유지. `cmd/internal/tools` 실제 tree에서 symlink 0을 확인했다. BTM: 확장자 없는 디렉터리 symlink → 명시 오류; baseline 정상 tree 및 기존 8개 source-copy 변이 → 종전 결과. 테스트 root는 임시 세 디렉터리와 임시 링크만 사용한다. 이 함수는 test-only; 생산 source·실제 링크는 수정하지 않는다.

임시 디렉터리 symlink 반례는 수정 전 오류 없이 통과해 RED, 변경 후 오류를 반환해 GREEN. Go1.25 선택 normal과 `-race -tags tossos_testseams`에서 baseline·8개 변이·새 반례 PASS. RED/normal/race 원문 gzip과 SHA는 receipts에 보존. 수동 gstack 범위 검토: 필터 순서만 변경하고 실제 tree 및 생산 source를 변경하지 않는다. 전체 test/gate·운영·네트워크 검증 미실행. 기존 83개 AST shape는 손읽은 본문을 변경 감지하는 pin이며 스스로 의미를 증명하지 않는다.

# a166 offline 도구 독립 검토

대상 ef62f4cdd792290987ada38a7830c36548627dc7, base cf0f0919e. 판정: 범위 내 P0/P1 발견 없음. 아래 P2 문서 명료화 권고. 소스 변경·운영 경로 접근 없음.

- 수동 source 검토: offline_linux.go:92-262는 빈 root, NOFOLLOW 부모, inode 재검사, 기존 manifest 거부, 임시 파일 동기화, RENAME_NOREPLACE 및 parent fsync와 실패 상태 보존을 구현한다. CLI main.go는 provision/check 외 client 호출 없음. checkOffline은 Open에 nil context를 전달하고 runtime_linux.go:65-67에서 오류로 반환하므로 nil panic 아님. Open→OpenObserved→OpenContext의 queue state read 및 OpenObservation의 event state read가 수행된다(accountbudget/budget_linux.go:105-157, observation_linux.go:100-119). clock snapshot/Close도 실행한다. active owner 부재·실제 namespace 동등성·운영 readiness까지 증명하지 않는다.
- P2: offline_linux.go:252-261 parent fsync 실패 뒤 manifest는 게시된 상태로 남는다. offline_test.go:132-165는 재-provision 거절을 확인하지만, 이후 별도 check 성공은 당초 fsync 성공을 증명하지 않는다. operator-cutover-draft.md의 부분 실패 보존과 성공 조건은 이에 부합하나, `check 성공이 이전 provision의 dir-sync 실패를 소급 복구/성공으로 전환하지 않는다`는 한 줄과 해당 경우 수동 처분이 필요함을 명시 권고. 실제 fault 후 check 성공 여부는 별도 시험하지 않았고 소스상의 가능성이다.
- 독립 Go 1.25 normal/tagged race: 승인 선택 `^TestA166Offline` 두 패키지 PASS. 원문 `/tmp/a166-offline-independent-normal.log`, `/tmp/a166-offline-independent-tag-race.log`; loopback-only network namespace 내부 lo 외 인터페이스 없음, route 출력 없음. 임시 root fixture만 사용. normal runtime/CLI 0.260/0.043초, tagged race 1.288/1.057초. namespace inode 번호는 재사용될 수 있으므로 실행별 출력 자체를 영수증으로 보존.
- 원문 gzip 9건 sha256이 raw-sha256.txt와 일치, fixture 두 파일 sha256 일치. 세 변이 원문은 기존 후보 시점 root교체·NOREPLACE·clock 제거에 각각 기대 RED를 보임; 독립 새 mutation 실행은 하지 않았다. pre-edit은 신규함수 분기 맵, 기존함수 변경 없음; AST 확인은 fixture의 직접 호출 검사 범위이며 전체 코드 경로 정리 증명 아님.
- owner/filesystem fixture는 synthetic 값, 장애는 injection. 실제 다른 UID·전원 장애·운영 파일시스템·운영 namespace·라이브·전체 tests/gate 미실행. gstack 작성자 수동 범위 검토이며 자동 전체 리뷰 아님. 최신 main 합성·종결 및 운영 승인은 별도.

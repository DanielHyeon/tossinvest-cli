# runtime factory Manager 부분 판정

2026-10-31. 후보 `d2ce75b6229107a8bc5f8d3c9ef1c9ef93a806f1`의 factory·manifest·VerifyIdentity 원문 및 독립 검토를 확인했다. 기존 생산 호출자에 연결되지 않은 내부 라이브러리로 부분 병합을 승인한다. Go 1.25 고정 toolchain의 같은 선택 시험 재확인을 출판 전 조건으로 둔다. 기능 완료·운영 승인·archive는 아니다.

독립 Terra의 선택 normal 2.499초·race 5.584초 원문을 확인했고, 작성자의 normal 2.640초·race 5.567초 기록과 대상 15개 시험을 대조했다. 새 재현 P0/P1은 없다. AST 함수 열거와 손으로 읽은 수명·권한 해석은 독립 기록에서 구별했다. required의 외부 기대 파일 ID 검증, NOFOLLOW 및 strict manifest, 명시 OFF/불완전 mode 거절, 동일 factory의 client 결속, Close 오류 보존을 확인했다.

부재 환경변수→OFF 편의 함수만으로 required launch 누락을 잡았다고 주장하지 않는다. 후속 실제 호출자 배선은 별도 mode/manifest 입력과 배포 참여자 검증으로 이 조건을 다룬다. 기대 inode 비교와 Acquire 재검증은 영구적인 namespace 봉인이 아니다. 신뢰된 manifest를 관리자가 바꾸는 공격과 임의 mount alias를 막는다는 주장은 없다.

두 official client의 bootstrap 차단·동일 Budget 결속 및 순차 자식의 복제 root 거절은 검증됐다. 동시 다중 프로세스의 실제 TLS 첫 plaintext Write를 이 factory를 통해 직접 계수하는 시험은 후속 test-only 로트에서 추가한다. core 시험을 그 직접 증거로 대체하지 않는다. 현재 UI 변경 없음(impeccable N/A). CodeGraph 미초기화, 전체 시험/gate·Alpine 실행·실계좌·운영 namespace·배포는 미검증이다. gstack은 수동 범위 체크리스트와 독립 검토이며 전체 자동 파이프라인 실행이 아니다.

Go 1.25 선택 검증을 통과하면 Terra는 독립 원문·판정을 보존하고 scoped strict/PM/diff 검사 뒤 non-force commit/merge/push할 수 있다. 같은 폴더 상위의 `caller-lifetime-plan.md`는 별도 설계 검토 중이므로 이 출판에 포함하지 않는다.

출판 전 재확인: Go 1.25.0 Linux에서 동일 선택 15개 normal 2.517초·race 5.580초 통과했다. 원문 `a166-runtime-go125-normal.log`, `a166-runtime-go125-race.log`. 실제 Alpine 실행을 뜻하지 않는다.

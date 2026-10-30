# Function Logic Map: `newMCPCmd`

기준: c412c0e55, ast.json은 편집 전 source SHA와 분기를 기록한다.

| 분기 | 현재 동작 | 보존할 불변식 |
|---|---|---|
| 설정·backend·자격 읽기 실패 | 오류 반환 | 서버 생성 전 실패 |
| official eligible | 공식 client와 config-gated trading service 생성 | WTS 선택 시 공식 생성 안 함 |
| 공식 비대상 | 공식 client 없음 | WTS-only 자격 허용 |
| session 없음 | 공식 client 있으면 진행, 둘 다 없으면 실패 | 기존 no-credentials 오류 유지 |
| 서버 실행 | 업데이트 조회 후 command context로 Serve | 기존 입출력·주문 게이트 유지 |

변경 경계: eligible 분기의 client 생성만 root factory helper로 교체하고 오류를 반환한다. required 오류를 WTS-only 성공으로 변환하지 않는다.

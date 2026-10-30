# Function Logic Map: newverifyreconcilereader

기준 5762b161b. AST source SHA는 ast.json. CodeGraph worktree 인덱스 없음; 소스 손읽기.

첫 Accounts 결과에서 빈 행, 계좌 수, 양수 seq를 검증한다. 양수 seq를 새 client에 명시하고 좁은 읽기 wrapper만 반환한다.

변경 시 최초 client와 계좌 지정 client는 같은 root factory를 빌리고 ctx와 credentials/token 경로를 보존한다. Factory 실패를 성공이나 WTS fallback으로 바꾸지 않는다.

# Function Logic Map: buildverifyreconcilereader

기준 5762b161b. AST source SHA는 ast.json. CodeGraph worktree 인덱스 없음; 소스 손읽기.

프로필 자격만 읽으며 환경 자격을 사용하지 않는다. 읽기 실패·자격 없음은 기존 refusal로 반환하고 단일 계좌 reader를 생성한다.

변경 시 최초 client와 계좌 지정 client는 같은 root factory를 빌리고 ctx와 credentials/token 경로를 보존한다. Factory 실패를 성공이나 WTS fallback으로 바꾸지 않는다.

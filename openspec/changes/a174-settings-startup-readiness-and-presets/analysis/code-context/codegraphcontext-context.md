# CodeGraphContext 보조 문맥 — a174

- 실행: `codegraphcontext version`, `codegraphcontext stats .`
- 결과: `CodeGraphContext 0.5.1`은 이 새 worktree를 저장소로 찾지 못해 보조 문맥을 만들지 못함.
- 처분: 도구 설치·전역 색인·main 체크아웃을 변경하지 않는다. 현재 worktree의 CodeGraph 1.6.0 질의, 현재 소스, 기존 시험과 FLM으로 교차검증한다.

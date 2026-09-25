import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 부팅 문구 원복(「dormant로 뜬다」).
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '"전략 화면은 붙기 전까지 도달 불가를 표시하고, 엔진이 돌아오면 콘솔 재시작 없이 다시 붙는다.\\n"',
    '"전략 화면은 dormant로 뜬다. 엔진이 돌아오면 콘솔 재시작 없이 다시 붙는다.\\n"')

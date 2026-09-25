import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 부팅이 못 붙으면 nil 을 돌려준다(「언제나 non-nil」 위반, 리뷰 P2-4).
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\tgo pumpConsoleStrategyRuntime(attachment)\n\treturn attachment',
    '\tgo pumpConsoleStrategyRuntime(attachment)\n\tif reader, _, _ := attachment.state(); reader == nil {\n\t\treturn nil\n\t}\n\treturn attachment')

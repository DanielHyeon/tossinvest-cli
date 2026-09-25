import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 펌프 제거.
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\tgo pumpConsoleStrategyRuntime(attachment)\n', '')

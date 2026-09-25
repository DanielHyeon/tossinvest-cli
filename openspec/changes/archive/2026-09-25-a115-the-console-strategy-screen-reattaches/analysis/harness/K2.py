import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# sentinel 대신 nil: dial 실패를 부재로 돌려준다.
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\t\treturn unavailableStrategyRuntime{cause: dialErr}, false',
    '\t\treturn nil, false')

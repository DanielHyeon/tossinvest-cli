import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 재시도 경고를 버리지 않는다.
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\t\t\treturn resolveConsoleStrategyRuntime(attemptCtx, engineDir, io.Discard)',
    '\t\t\treturn resolveConsoleStrategyRuntime(attemptCtx, engineDir, errOut)')

import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# (리뷰 R17) 배선이 `if engineDir != ""` 밖으로 나간다.
p = d + '/cmd/tossctl/console.go'
sub(p, '\t\tstrategyRuntime = consoleStrategyRuntimeReaderFor(ctx, engineDir, cmd.ErrOrStderr())\n', '')
sub(p, '\tvar strategyRuntime console.MultiMarketStrategyRuntimeReader\n',
    '\tvar strategyRuntime console.MultiMarketStrategyRuntimeReader = consoleStrategyRuntimeReaderFor(ctx, engineDir, cmd.ErrOrStderr())\n')

import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# (리뷰 R6) 부팅 live 플래그를 버린다 — 언제나 attached=false 로 출발.
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\tattachment.attach(resolveConsoleStrategyRuntime(ctx, engineDir, errOut))',
    '\tboot, _ := resolveConsoleStrategyRuntime(ctx, engineDir, errOut)\n\tattachment.attach(boot, false)')

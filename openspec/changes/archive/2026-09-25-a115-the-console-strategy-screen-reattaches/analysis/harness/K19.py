import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# (리뷰 R2b) 부팅 해석이 lastTry 를 찍는다 — 간격이 있으면 첫 wake 가 간격만큼 막힌다.
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\tattachment.attach(resolveConsoleStrategyRuntime(ctx, engineDir, errOut))',
    '\tattachment.attach(resolveConsoleStrategyRuntime(ctx, engineDir, errOut))\n\tif attachment.interval > 0 {\n\t\tattachment.lastTry = time.Now()\n\t}')

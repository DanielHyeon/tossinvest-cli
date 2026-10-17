import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 부팅 nil 접힘 재도입: 부팅 해석이 live 가 아니면 자리를 비운다(편집 전 B35 의 모양).
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\tattachment.attach(resolveConsoleStrategyRuntime(ctx, engineDir, errOut))',
    '\tboot, live := resolveConsoleStrategyRuntime(ctx, engineDir, errOut)\n\tif !live {\n\t\tboot = nil\n\t}\n\tattachment.attach(boot, live)')

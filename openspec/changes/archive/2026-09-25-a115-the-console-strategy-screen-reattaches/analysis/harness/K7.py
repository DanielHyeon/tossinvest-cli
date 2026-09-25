import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 부팅 dial 재도입: runConsole 이 직접 한 번 dial 한다(결과는 버려도 구조 핀이 잡아야 한다).
p = d + '/cmd/tossctl/console.go'
sub(p, '\t\tstrategyRuntime = consoleStrategyRuntimeReaderFor(ctx, engineDir, cmd.ErrOrStderr())',
    '\t\t_, _ = strategyprojectionrpc.Dial(ctx, strategyprojectionrpc.DescriptorPath(engineDir))\n'
    '\t\tstrategyRuntime = consoleStrategyRuntimeReaderFor(ctx, engineDir, cmd.ErrOrStderr())')
sub(p, '\t"github.com/JungHoonGhae/tossinvest-cli/internal/soak"\n',
    '\t"github.com/JungHoonGhae/tossinvest-cli/internal/soak"\n'
    '\t"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"\n')

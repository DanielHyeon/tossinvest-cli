import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
# 부팅 1회 lifecycle dial 을 runConsole 안에 되살린다(굳은 client 가 wrapper 를 덮는다)
sub(d+'/console.go','''	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
''','''	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
''')
sub(d+'/console.go','''		positionPolicyCommander = consolePositionPolicyCommanderFor(ctx, engineDir, cmd.ErrOrStderr())''','''		positionPolicyCommander = consolePositionPolicyCommanderFor(ctx, engineDir, cmd.ErrOrStderr())
		if early, dialErr := positionpolicyrpc.Dial(ctx, positionpolicyrpc.DescriptorPath(engineDir)); dialErr == nil {
			positionPolicyCommander = &consolePositionPolicyCommander{lifecycle: early}
		}''')

import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/console_lifecycle_attach.go','''var consolePositionPolicyRedialInterval = 30 * time.Second''','''var consolePositionPolicyRedialInterval = 0 * time.Second''')

import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/console_lifecycle_attach.go','''	if client == nil {
		return nil, seat, errPositionPolicyLifecycleDetached
	}''','''	if client == nil {
		return nil, seat, fmt.Errorf("%w: %w", exitquarantine.ErrUnwired, errPositionPolicyLifecycleDetached)
	}''')

import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/console_lifecycle_attach.go','''	if err == nil || endpointAnswered(err) {
		a.failed = false
		announce := !a.attached
		a.attached = true''','''	if err == nil || endpointAnswered(err) {
		a.failed = false
		announce := false''')

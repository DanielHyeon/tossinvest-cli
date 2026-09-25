import sys; sys.path.insert(0, __import__('os').path.dirname(__file__)); from apply import sub
d=sys.argv[1]
sub(d+'/console_lifecycle_attach.go','''	if err != nil || client == nil {
		// 실패는 침묵이고''','''	if err != nil || client == nil {
		a.client = nil
		// 실패는 침묵이고''')

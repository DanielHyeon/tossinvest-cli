//go:build !unix

package main

// fileOwner 는 uid 개념이 없는 플랫폼에서 확인을 건너뜀.
func fileOwner(string) (int, bool) { return 0, false }

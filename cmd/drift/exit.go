package main

type exitCodeErr struct {
	code int
	msg  string
}

func (e *exitCodeErr) Error() string { return e.msg }

func newExitCode(code int, msg string) error {
	return &exitCodeErr{code: code, msg: msg}
}

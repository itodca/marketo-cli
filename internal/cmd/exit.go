package cmd

type exitError struct {
	code int
}

func (err *exitError) Error() string {
	return ""
}

func (err *exitError) ExitCode() int {
	if err == nil || err.code <= 0 {
		return 1
	}
	return err.code
}

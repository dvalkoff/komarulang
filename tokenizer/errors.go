package tokenizer

import (
	"fmt"
	"strings"
)

type TokenizerError struct {
	File    string
	Line    int
	Message string
	Cause   error
}

func (e TokenizerError) Error() string {
	if e.Cause != nil {
		return strings.Join(
			[]string{
				fmt.Sprintf("Error at line: %v:%v. error: %v", e.File, e.Line, e.Message),
				fmt.Sprintf("Cause: %v", e.Cause),
			},
			"\n",
		)
	}
	return fmt.Sprintf("Error at line: %v:%v. error: %v", e.File, e.Line, e.Message)
}

func (e TokenizerError) Unwrap() error {
	return e.Cause
}

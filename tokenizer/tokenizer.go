package tokenizer

import (
	"bufio"
	"io"
	"strings"

	"github.com/dvalkoff/komarulang/tokenizer/token"
)

type Tokenizer struct {
	File string
}

func (t Tokenizer) Scan(reader io.Reader) ([]token.Token, error) {
	scanner := bufio.NewScanner(reader)
	currentLine := 0
	tokens := make([]token.Token, 0)
	for ; scanner.Scan(); currentLine++ {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		lineTokens, err := t.ScanLine(line, currentLine)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, lineTokens...)
	}
	if err := scanner.Err(); err != nil {
		return nil, TokenizerError{
			File: t.File,
			Line: currentLine,
			Message: "Tokenization failed",
			Cause: err,
		}
	}
	tokens = append(tokens, token.GetEOF())
	return tokens, nil
}

func (t Tokenizer) ScanLine(line string, lineNumber int) ([]token.Token, error) {
	lineTokens := make([]token.Token, 0)
	splittedSeq := strings.SplitSeq(line, " ") // TODO: replace with character by character (1+1 won't work)
	for potentialToken := range splittedSeq {
		token, err := token.GetToken(potentialToken)
		if err != nil {
			return nil, TokenizerError{
				File: t.File,
				Line: lineNumber,
				Message: "Failed to recognize token",
				Cause: err,
			}
		}
		lineTokens = append(lineTokens, token)
	}
	return lineTokens, nil
}

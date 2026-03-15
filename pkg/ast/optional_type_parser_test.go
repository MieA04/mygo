package ast

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
)

type collectErrorListener struct {
	*antlr.DefaultErrorListener
	errCount int
}

func (l *collectErrorListener) SyntaxError(_ antlr.Recognizer, _ interface{}, _ int, _ int, _ string, _ antlr.RecognitionException) {
	l.errCount++
}

func parseTypeForTest(input string) int {
	is := antlr.NewInputStream(input)
	lexer := NewMyGoLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := NewMyGoParser(stream)
	listener := &collectErrorListener{DefaultErrorListener: &antlr.DefaultErrorListener{}}
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)
	parser.TypeType()
	return listener.errCount
}

func TestTypeTypeSupportsOptionalSuffix(t *testing.T) {
	cases := []string{
		"int?",
		"User?",
		"Map<string,int>?",
		"fn(int):int?",
	}
	for _, input := range cases {
		if errs := parseTypeForTest(input); errs != 0 {
			t.Fatalf("parse %q failed with %d errors", input, errs)
		}
	}
}

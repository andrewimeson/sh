package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"mvdan.cc/sh/v3/syntax"
)

func main() {
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	p := syntax.NewParser(syntax.RecoverErrors(5))
	f, err := p.Parse(bytes.NewReader(src), "")
	if err != nil {
		panic(err)
	}

	last := lastStmt(f)
	compl := extractCompletion(last, string(src))
	fmt.Printf("%#v\n", compl)
}

func lastStmt(node syntax.Node) *syntax.Stmt {
	var last *syntax.Stmt
	// TODO: keep the first incomplete statement, too
	syntax.Walk(node, func(node syntax.Node) bool {
		if stmt, _ := node.(*syntax.Stmt); stmt != nil {
			last = stmt
		}
		return true
	})
	return last
}

func extractCompletion(stmt *syntax.Stmt, src string) compResult {
	syntax.DebugPrint(os.Stderr, stmt)
	if len(stmt.Redirs) > 0 {
		lastRedir := stmt.Redirs[len(stmt.Redirs)-1]
		// fmt.Println(lastRedir.Word.Pos())
		// fmt.Fprintln(os.Stderr, lastRedir.Word.Pos())
		if lastRedir.Word.Pos().IsRecovered() {
			return compResult{"redirect", nil}
		}
	}
	if call, _ := stmt.Cmd.(*syntax.CallExpr); call != nil {
		flat := flatWords(call.Args, src)
		if stmt.End().Offset() < uint(len(src)) {
			// Starting a new word.
			flat = append(flat, "")
		}
		return compResult{"args", flat}
	}

	// other commands, e.g. "if foo; then bar; fi"
	return compResult{}
}

func flatWords(words []*syntax.Word, src string) []string {
	var flat []string
	for _, word := range words {
		start := word.Pos().Offset()
		end := word.End().Offset()
		flat = append(flat, src[start:end])
	}
	return flat
}

type compResult struct {
	Type   string
	Tokens []string
}

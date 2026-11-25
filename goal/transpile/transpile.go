// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/base/logx"
)

// TranspileLine is the main function for parsing a single line of goal input,
// returning a new transpiled line of code that converts Exec code into corresponding
// Go function calls.
func (st *State) TranspileLine(code string) string {
	if len(code) == 0 {
		return code
	}
	if strings.HasPrefix(code, "#!") {
		return ""
	}
	toks := st.TranspileLineTokens(code)
	paren, brace, brack := toks.BracketDepths()
	preBrace := st.BraceDepth
	st.ParenDepth += paren
	st.BraceDepth += brace
	st.BrackDepth += brack
	// a type or var, const, import declaration ends at the depths at which
	// it started, which are not 0 inside a function
	if st.TypeDepth > 0 && st.BraceDepth <= st.typeBrace {
		st.TypeDepth = 0
	}
	// the fields of a struct or the methods of an interface type literal
	// over lines, as of a variable, are Go, as for a type declaration
	if st.TypeDepth == 0 && st.BraceDepth > preBrace-1 && opensTypeLiteral(toks) {
		st.TypeDepth = 1
		st.typeBrace = st.BraceDepth - 1
	}
	if st.DeclDepth > 0 && st.ParenDepth <= st.declParen && st.BraceDepth <= st.declBrace {
		st.DeclDepth = 0
	}
	return st.finishLine(toks)
}

// opensTypeLiteral returns whether the given tokens of a line have a struct
// or interface type literal that is not closed on the line.
func opensTypeLiteral(toks Tokens) bool {
	for i := len(toks) - 2; i >= 0; i-- {
		if (toks[i].Tok == token.STRUCT || toks[i].Tok == token.INTERFACE) && toks[i+1].Tok == token.LBRACE {
			return toks[i+1:].RightMatching() < 0
		}
	}
	return false
}

// finishLine returns the code of the given transpiled tokens of a line,
// recording their column map.
func (st *State) finishLine(toks Tokens) string {
	logx.PrintlnDebug("depths: ", st.ParenDepth, st.BraceDepth, st.BrackDepth, st.TypeDepth, st.DeclDepth)
	code, cols := toks.CodeMap()
	st.lastCols = cols
	return code
}

// TranspileLineTokens returns the tokens for the full line
func (st *State) TranspileLineTokens(code string) Tokens {
	if code == "" {
		return nil
	}
	toks := TokensFromString(code)
	n := len(toks)
	if n == 0 {
		return toks
	}
	if st.MathMode {
		if len(toks) >= 2 {
			if toks[0].Tok == token.ILLEGAL && toks[0].Str == "#" && toks[1].Tok == token.ILLEGAL && toks[1].Str == "#" {
				st.MathMode = false
				return nil
			}
		}
		return st.TranspileMath(toks, code, true)
	}
	if toks[0].Tok == token.COMMENT && n == 1 {
		return toks // a comment line does not change goContinues
	}
	cont := st.goContinues
	st.goContinues = false
	if cont || goElement(toks) {
		st.forgetDefines(toks)
		logx.PrintlnDebug("go:   continued expression or list element")
		return st.goLine(st.TranspileGo(toks, code), toks)
	}
	ewords, err := ExecWords(code)
	if err != nil {
		if goScans(code) { // valid Go, as with a rune '\'' or a [ closed on another line
			st.forgetDefines(toks)
			logx.PrintlnDebug("go:   not shell words")
			return st.goLine(st.TranspileGo(toks, code), toks)
		}
		st.addWordsError(err, 0)
		return nil
	}
	st.forgetDefines(toks)
	logx.PrintlnDebug("\n########## line:\n", code, "\nTokens:", len(toks), "\n", toks.String(), "\nWords:", len(ewords), "\n", ewords)

	if toks[0].Tok == token.TYPE {
		if st.TypeDepth == 0 {
			st.typeBrace = st.BraceDepth
		}
		st.TypeDepth++
	}
	if toks[0].Tok == token.IMPORT || toks[0].Tok == token.VAR || toks[0].Tok == token.CONST {
		if st.DeclDepth == 0 {
			st.declParen, st.declBrace = st.ParenDepth, st.BraceDepth
		}
		st.DeclDepth++
	}
	// logx.PrintlnDebug("depths: ", st.ParenDepth, st.BraceDepth, st.BrackDepth, st.TypeDepth, st.DeclDepth)
	if st.TypeDepth > 0 || st.DeclDepth > 0 {
		logx.PrintlnDebug("go:   type / decl defn")
		return st.goLine(st.TranspileGo(toks, code), toks)
	}

	if envPrefixed(ewords) {
		logx.PrintlnDebug("exec: NAME=value command")
		return st.TranspileExec(ewords, exec.NoCapture)
	}

	t0 := toks[0]
	_, t0pn := toks.Path(true) // true = first position
	en := len(ewords)

	f0exec := (t0.Tok == token.IDENT && ExecWordIsCommand(ewords[0]))
	if f0exec && n > 1 && toks[1].Tok == token.COLON { // go Ident: static initializer
		f0exec = false
	}
	if f0exec && sendStmt(toks) { // go send: ch <- x
		f0exec = false
	}
	if f0exec && n > 1 && (toks[1].Tok == token.DEFINE || toks[1].Tok == token.ASSIGN) { // set := x
		f0exec = false
	}

	switch {
	case t0.Tok == token.ILLEGAL:
		if t0.Str == "#" {
			logx.PrintlnDebug("math #")
			if toks[1].Tok == token.ILLEGAL && toks[1].Str == "#" {
				st.MathMode = true
				return st.TranspileMath(toks[2:], code, true)
			}
			return st.TranspileMath(toks[1:], code, true)
		}
		return st.TranspileExec(ewords, exec.NoCapture)
	case t0.Tok == token.LBRACE && (n == 1 || lastTok(toks) == token.COMMA):
		logx.PrintlnDebug("go:   { block or composite literal element")
		return st.goLine(st.TranspileGo(toks, code), toks)
	case t0.Tok == token.LBRACE:
		logx.PrintlnDebug("go:   { } line")
		return st.TranspileGoRange(toks, code, 1, n-1)
	case t0.Tok == token.LBRACK:
		logx.PrintlnDebug("exec: [ ] line")
		return st.TranspileExec(ewords, exec.NoCapture) // it processes the [ ]
	case t0.Tok == token.IDENT && t0.Str == "command":
		st.lastCommand = toks[1].Str // 1 is the name -- triggers AddCommand
		toks = toks[2:]              // get rid of first
		toks.Insert(0, token.IDENT, "goalrun.AddCommand")
		toks.Insert(1, token.LPAREN)
		toks.Insert(2, token.STRING, `"`+st.lastCommand+`"`)
		toks.Insert(3, token.COMMA)
		toks.Insert(4, token.FUNC)
		toks.Insert(5, token.LPAREN)
		toks.Insert(6, token.IDENT, "args")
		toks.Insert(7, token.ELLIPSIS)
		toks.Insert(8, token.IDENT, "string")
		toks.Insert(9, token.RPAREN)
		toks.AddTokens(st.TranspileGo(toks[11:], code)...)
	case t0.IsGo():
		if t0.Tok == token.GO {
			if !toks.Contains(token.LPAREN) {
				logx.PrintlnDebug("exec: go command")
				return st.TranspileExec(ewords, exec.NoCapture)
			}
		}
		logx.PrintlnDebug("go    keyword")
		return st.goLine(st.TranspileGo(toks, code), toks)
	case lastTok(toks) == token.INC || lastTok(toks) == token.DEC:
		logx.PrintlnDebug("go    ++ / --")
		return st.goLine(st.TranspileGo(toks, code), toks)
	case t0pn > 0: // path expr
		logx.PrintlnDebug("exec: path...")
		return st.TranspileExec(ewords, exec.NoCapture)
	case t0.Tok == token.STRING:
		logx.PrintlnDebug("exec: string...")
		return st.TranspileExec(ewords, exec.NoCapture)
	case f0exec && en == 1:
		logx.PrintlnDebug("exec: 1 word")
		return st.TranspileExec(ewords, exec.NoCapture)
	case !f0exec: // exec must be IDENT
		logx.PrintlnDebug("go:   not ident")
		return st.goLine(st.TranspileGo(toks, code), toks)
	case f0exec && en > 1 && ewords[0] != "set" && toks.IsAssignExpr():
		logx.PrintlnDebug("go:   assignment or defn")
		return st.goLine(st.TranspileGo(toks, code), toks)
	case f0exec && en > 1 && ewords[0] != "set" && toks.IsAssignExpr():
		logx.PrintlnDebug("go:   assignment or defn")
		return st.goLine(st.TranspileGo(toks, code), toks)
	case f0exec: // now any ident
		logx.PrintlnDebug("exec: ident..")
		return st.TranspileExec(ewords, exec.NoCapture)
	default:
		logx.PrintlnDebug("go:   default")
		return st.goLine(st.TranspileGo(toks, code), toks)
	}
	return toks
}

// goLine records whether the given tokens of a line of Go code continue
// on the next line (see [State.goContinues]), returning the given code.
func (st *State) goLine(code, toks Tokens) Tokens {
	st.goContinues = continues(toks)
	return code
}

// continues returns whether the given tokens of a line of Go code end in
// a token after which the expression or list continues on the next line:
// a binary operator, assignment, comma, ( or [.
func continues(toks Tokens) bool {
	switch t := lastTok(toks); {
	case t == token.CASE: // a list of cases over lines
		return true
	case t.IsOperator() && t != token.RPAREN && t != token.RBRACK && t != token.RBRACE && t != token.LBRACE &&
		t != token.SEMICOLON && t != token.COLON && t != token.INC && t != token.DEC && t != token.ELLIPSIS:
		return true
	}
	return false
}

// lastTok returns the last token of the given tokens that is not a comment,
// or [token.ILLEGAL] if there is none.
func lastTok(toks Tokens) token.Token {
	for i := len(toks) - 1; i >= 0; i-- {
		if toks[i].Tok != token.COMMENT {
			return toks[i].Tok
		}
	}
	return token.ILLEGAL
}

// goElement returns whether the given tokens of a line are elements of a
// Go list over lines, as of a composite literal: they start with a string
// that is followed by a comma, : or closer, or end with a comma, which a
// command that starts with a string does not.
func goElement(toks Tokens) bool {
	if len(toks) < 2 || toks[0].Tok != token.STRING {
		return false
	}
	switch toks[1].Tok {
	case token.COMMA, token.COLON, token.RPAREN, token.RBRACK, token.RBRACE:
		return true
	}
	return lastTok(toks) == token.COMMA
}

// TranspileGoRange returns transpiled tokens assuming Go code,
// for given start, end (exclusive) range of given tokens and code.
// In general the positions in the tokens applies to the _original_ code
// so you should just keep the original code string. However, this is
// needed for a specific case.
func (st *State) TranspileGoRange(toks Tokens, code string, start, end int) Tokens {
	codeSt := toks[start].Pos - 1
	codeEd := token.Pos(len(code))
	if end <= len(toks)-1 {
		codeEd = toks[end].Pos - 1
	}
	return st.TranspileGo(toks[start:end], code[codeSt:codeEd])
}

// TranspileGo returns transpiled tokens assuming Go code.
// Unpacks any encapsulated shell or math expressions.
func (st *State) TranspileGo(toks Tokens, code string) Tokens {
	n := len(toks)
	if n == 0 {
		return toks
	}
	if st.FuncToVar && toks[0].Tok == token.FUNC { // reorder as an assignment
		if len(toks) > 1 && toks[1].Tok == token.IDENT {
			toks[0] = toks[1]
			toks.Insert(1, token.DEFINE)
			toks[2] = &Token{Tok: token.FUNC}
			n = len(toks)
		}
	}
	gtoks := make(Tokens, 0, len(toks)) // return tokens
	for i := 0; i < n; i++ {
		tok := toks[i]
		switch {
		case tok.Tok == token.ILLEGAL:
			et := toks[i:].ModeEnd()
			if et > 0 {
				if tok.Str == "#" {
					gtoks.AddTokens(st.TranspileMath(toks[i+1:i+et], code, false)...)
				} else {
					gtoks.AddTokens(st.TranspileExecTokens(toks[i+1:i+et+1], code, exec.CaptureOutput)...)
				}
				i += et
				continue
			} else {
				gtoks = append(gtoks, tok)
			}
		case tok.Tok == token.LBRACK && i > 0 && toks[i-1].Tok == token.IDENT: // index expr
			ixtoks := toks[i:]
			rm := ixtoks.RightMatching()
			if rm < 3 {
				gtoks = append(gtoks, tok)
				continue
			}
			idx := st.TranspileGoNDimIndex(toks, code, &gtoks, i-1, rm+i)
			if idx > 0 {
				i = idx
			} else {
				gtoks = append(gtoks, tok)
			}
		default:
			gtoks = append(gtoks, tok)
		}
	}
	return gtoks
}

// TranspileExecString returns transpiled tokens assuming Exec code,
// from a string, with the given bool indicating whether [Output] is needed.
func (st *State) TranspileExecString(str string, output exec.OutputMode) Tokens {
	if len(str) <= 1 {
		return nil
	}
	ewords, err := ExecWords(str)
	if err != nil {
		st.addWordsError(err, 0)
	}
	return st.TranspileExec(ewords, output)
}

// addWordsError adds the given error of [ExecWords] for the code at the
// given byte offset of the source line, at its position in it.
func (st *State) addWordsError(err error, offset int) {
	if we, ok := errors.AsType[*WordsError](err); ok {
		st.addSourceError(offset+we.Pos+1, we.Msg)
		return
	}
	st.AddError(err)
}

// TranspileExecTokens returns transpiled tokens assuming Exec code,
// from given tokens, with the given bool indicating
// whether [Output] is needed.
func (st *State) TranspileExecTokens(toks Tokens, code string, output exec.OutputMode) Tokens {
	nt := len(toks)
	if nt == 0 {
		return nil
	}
	str := code[toks[0].Pos-1 : toks[nt-1].Pos-1]
	return st.TranspileExecString(str, output)
}

// TranspileExec returns transpiled tokens assuming Exec code,
// with the given bools indicating the type of run to execute.
func (st *State) TranspileExec(ewords []string, output exec.OutputMode) Tokens {
	etoks := st.transpileExec(ewords, output)
	if !output {
		return etoks
	}
	// The output of a pipe or list of commands ($a | b$, $a; b$) is that of
	// the last one (unlike bash, which outputs those of all of a list), but
	// they are statements, so they are wrapped in a func that returns it,
	// which is an expression.
	last := -1
	for i, tk := range etoks {
		if tk.Tok == token.SEMICOLON {
			last = i
		}
	}
	if last < 0 {
		return etoks
	}
	wtoks := make(Tokens, 0, len(etoks)+10)
	wtoks.AddMulti(token.FUNC, token.LPAREN, token.RPAREN)
	wtoks.Add(token.IDENT, "string")
	wtoks.Add(token.LBRACE)
	wtoks.AddTokens(etoks[:last+1]...)
	wtoks.Add(token.RETURN)
	wtoks.AddTokens(etoks[last+1:]...)
	wtoks.AddMulti(token.RBRACE, token.LPAREN, token.RPAREN)
	return wtoks
}

// transpileExec returns the transpiled tokens of [State.TranspileExec],
// as a list of statements.
func (st *State) transpileExec(ewords []string, output exec.OutputMode) Tokens {
	n := len(ewords)
	if n == 0 {
		return nil
	}
	etoks := make(Tokens, 0, n+5) // return tokens
	var execTok *Token
	bgJob := false
	noStop := false
	if ewords[0] == "[" {
		ewords = ewords[1:]
		n--
		noStop = true
	}
	startExec := func() {
		bgJob = false
		etoks.Add(token.IDENT, "goalrun")
		etoks.Add(token.PERIOD)
		out := output == exec.CaptureOutput
		switch {
		case out && noStop:
			execTok = etoks.Add(token.IDENT, "OutputErrOK")
		case out && !noStop:
			execTok = etoks.Add(token.IDENT, "Output")
		case !out && noStop:
			execTok = etoks.Add(token.IDENT, "RunErrOK")
		case !out && !noStop:
			execTok = etoks.Add(token.IDENT, "Run")
		}
		etoks.Add(token.LPAREN)
	}
	endExec := func() {
		if bgJob {
			execTok.Str = "Start"
		}
		etoks.DeleteLastComma()
		etoks.Add(token.RPAREN)
	}

	startExec()

	cmdStart := true // before the command, for NAME=value words
	for i := 0; i < n; i++ {
		f := ewords[i]
		if cmdStart {
			if name, value, ok := envAssignWord(f); ok && i < n-1 {
				etoks.Add(token.IDENT, "goalrun.EnvVar")
				etoks.Add(token.LPAREN)
				etoks.Add(token.STRING, strconv.Quote(name))
				etoks.Add(token.COMMA)
				etoks.AddTokens(execWordTokens(value)...)
				etoks.Add(token.RPAREN)
				etoks.Add(token.COMMA)
				continue
			}
			cmdStart = f == "["
		}
		switch {
		case f == "{": // embedded go
			if n < i+3 {
				st.addSourceError(1, "missing closing } of the Go expression in the command")
			} else {
				gstr := ewords[i+1]
				etoks.AddTokens(st.TranspileGo(TokensFromString(gstr), gstr)...)
				etoks.Add(token.COMMA)
				i += 2
			}
		case f == "[":
			noStop = true
		case f == "]": // solo is def end
			// just skip
			noStop = false
		case f == "&":
			bgJob = true
		case f[0] == '|':
			execTok.Str = "Start"
			etoks.Add(token.IDENT, AddQuotes(f))
			etoks.Add(token.COMMA)
			endExec()
			etoks.Add(token.SEMICOLON)
			etoks.AddTokens(st.transpileExec(ewords[i+1:], output)...)
			return etoks
		case f == ";":
			endExec()
			etoks.Add(token.SEMICOLON)
			etoks.AddTokens(st.transpileExec(ewords[i+1:], output)...)
			return etoks
		default:
			etoks.AddTokens(execWordTokens(f)...)
			etoks.Add(token.COMMA)
		}
	}
	endExec()
	return etoks
}

// execWordTokens returns the tokens of the Go expression of the given
// word of a command. Words with a $ outside of single quotes have their
// environment variables expanded when they run (see goal.Goal.Expand), but
// not the values of Go expressions, which are passed as is.
func execWordTokens(f string) Tokens {
	var etoks Tokens
	if f == "" {
		etoks.Add(token.STRING, `""`)
		return etoks
	}
	expand := strings.Contains(f, "$") && f[0] != '\''
	if expand {
		etoks.Add(token.IDENT, "goalrun.Expand")
		etoks.Add(token.LPAREN)
	}
	switch {
	case f[0] == '"' || f[0] == '`':
		etoks.Add(token.STRING, f)
	case len(f) > 1 && f[0] == '\'' && f[len(f)-1] == '\'': // literal, as in bash
		etoks.Add(token.STRING, strconv.Quote(f[1:len(f)-1]))
	default:
		etoks.Add(token.IDENT, AddQuotes(f)) // mark as an IDENT but add quotes!
	}
	if expand {
		etoks.Add(token.RPAREN)
	}
	return etoks
}

// envAssignRE matches a NAME=value word of a command.
var envAssignRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// envAssignWord returns the name and value of the given word of a command
// if it is a NAME=value assignment of an environment variable, as in
// bash: FOO=bar cmd. Its value must be quoted if it has any of ( [ { , in
// it, which are those of Go code, as in x=f(a, b).
func envAssignWord(f string) (name, value string, ok bool) {
	loc := envAssignRE.FindStringIndex(f)
	if loc == nil {
		return "", "", false
	}
	name, value = f[:loc[1]-1], f[loc[1]:]
	if value != "" && value[0] != '"' && value[0] != '\'' && value[0] != '`' && strings.ContainsAny(value, "([{,") {
		return "", "", false
	}
	return name, value, true
}

// envPrefixed returns whether the given words of a line are a command with
// NAME=value assignments of environment variables before it, as in bash:
// FOO=bar cmd args.
func envPrefixed(ewords []string) bool {
	i := 0
	for i < len(ewords) {
		if _, _, ok := envAssignWord(ewords[i]); !ok {
			break
		}
		i++
	}
	if i == 0 || i >= len(ewords) {
		return false
	}
	w := ewords[i]
	return !strings.ContainsAny(w[:1], "{;|&<>") && ExecWordIsCommand(w)
}

// sendStmt returns whether the given tokens of a line start with a Go send
// statement: an operand of names and selectors, followed by <-.
func sendStmt(toks Tokens) bool {
	for i, t := range toks {
		switch t.Tok {
		case token.IDENT, token.PERIOD:
		case token.ARROW:
			return i > 0
		default:
			return false
		}
	}
	return false
}

// goScans returns whether the given code has only valid Go tokens.
func goScans(code string) bool {
	return goScanErrors(code) == nil
}

// goScanErrors returns the messages of the errors of scanning the given
// code as Go tokens.
func goScanErrors(code string) []string {
	src := []byte(code)
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	var errs []string
	s.Init(file, src, func(_ token.Position, msg string) { errs = append(errs, msg) }, scanner.ScanComments)
	for {
		if _, tok, _ := s.Scan(); tok == token.EOF {
			return errs
		}
	}
}

// isGenericInstance returns whether the given tokens from the [ at lb to
// the ] at rb, with the given commas between the args, are the type args
// of a generic function or type, not the indexes of a tensor: when it is
// called or has a composite literal ( or { after it, or any of its args
// is a type, as a predeclared type (int) or a pointer, slice, map, func,
// chan, struct or interface type.
func isGenericInstance(toks Tokens, lb, rb int, commas []int) bool {
	if rb+1 < len(toks) && (toks[rb+1].Tok == token.LPAREN || toks[rb+1].Tok == token.LBRACE) {
		return true
	}
	starts := append([]int{lb}, commas...)
	for _, c := range starts {
		if c+1 >= rb {
			continue
		}
		t := toks[c+1]
		switch t.Tok {
		case token.MUL, token.LBRACK, token.MAP, token.FUNC, token.CHAN, token.STRUCT, token.INTERFACE:
			return true
		case token.IDENT:
			if predeclaredTypes[t.Str] && (c+2 >= rb || toks[c+2].Tok == token.COMMA || toks[c+2].Tok == token.RBRACK) {
				return true
			}
		}
	}
	return false
}

// predeclaredTypes are the names of the predeclared types of Go.
var predeclaredTypes = map[string]bool{
	"any": true, "bool": true, "byte": true, "comparable": true, "complex64": true, "complex128": true,
	"error": true, "float32": true, "float64": true, "int": true, "int8": true, "int16": true,
	"int32": true, "int64": true, "rune": true, "string": true, "uint": true, "uint8": true,
	"uint16": true, "uint32": true, "uint64": true, "uintptr": true,
}

// TranspileGoNDimIndex processes an ident[*] sequence of tokens,
// translating it into a corresponding tensor Value or Set expression,
// if it is a multi-dimensional indexing expression which is not valid in Go,
// to support simple n-dimensional tensor indexing in Go (not math) mode.
// For 2 to 5 indexes it uses the fixed-arity methods, such as
//
//	x[i, j, k]      -> x.Value3D(int(i), int(j), int(k))
//	x[i, j, k] = v  -> x.Set3D(v, int(i), int(j), int(k))
//	x[i, j, k] += v -> x.SetAdd3D(v, int(i), int(j), int(k))
//	&x[i, j, k]     -> x.ValuePtr3D(int(i), int(j), int(k))
//
// and the variadic Value, Set etc for more indexes.
// Gets the current sequence of toks tokens, where the ident starts at idIdx
// and the ] is at rbIdx. It puts the results in gtoks generated tokens.
// Returns a positive index to resume processing at, if it is actually an
// n-dimensional expr, and -1 if not, in which case the normal process resumes.
func (st *State) TranspileGoNDimIndex(toks Tokens, code string, gtoks *Tokens, idIdx, rbIdx int) int {
	// Find the commas between the indexes, skipping nested expressions.
	var commas []int
	for i := idIdx + 2; i < rbIdx; i++ {
		tk := toks[i]
		if tk.Tok == token.COMMA {
			commas = append(commas, i)
		}
		if tk.Tok == token.LPAREN || tk.Tok == token.LBRACE || tk.Tok == token.LBRACK {
			rp := toks[i:rbIdx].RightMatching()
			if rp > 0 {
				i += rp
			}
		}
	}
	if len(commas) == 0 { // not multidim
		return -1
	}
	if isGenericInstance(toks, idIdx+1, rbIdx, commas) {
		return -1
	}

	// The address of the element is taken with ValuePtr, without the &
	// that precedes the receiver expression, already in gtoks.
	isPtr := false
	if rs := receiverStart(*gtoks); rs > 0 && (*gtoks)[rs-1].Tok == token.AND {
		isPtr = true
		*gtoks = slices.Delete(*gtoks, rs-1, rs)
	}

	// An assignment after the ] sets the element, with Set or an operator
	// version of it, such as SetAdd.
	isSet := false
	stok := token.ILLEGAL
	n := len(toks)
	hasComment := false
	if toks[n-1].Tok == token.COMMENT {
		hasComment = true
		n--
	}
	if !isPtr && n-rbIdx > 1 {
		ntk := toks[rbIdx+1].Tok
		if ntk == token.ASSIGN || (ntk >= token.ADD_ASSIGN && ntk <= token.QUO_ASSIGN) {
			isSet = true
			stok = ntk
		}
	}
	fun := "Value"
	switch {
	case isPtr:
		fun = "ValuePtr"
	case isSet:
		fun = "Set" + map[token.Token]string{token.ADD_ASSIGN: "Add", token.SUB_ASSIGN: "Sub", token.MUL_ASSIGN: "Mul", token.QUO_ASSIGN: "Div"}[stok]
	}
	if nd := len(commas) + 1; nd <= 5 {
		fun += fmt.Sprintf("%dD", nd)
	}

	// The arguments are the value set, if any, and the indexes, each
	// converted to int, as tensor indexes may be of any integer type.
	gtoks.Add(token.PERIOD)
	gtoks.Add(token.IDENT, fun)
	gtoks.Add(token.LPAREN)
	if isSet {
		gtoks.AddTokens(st.TranspileGo(toks[rbIdx+2:n], code)...)
		gtoks.Add(token.COMMA)
	}
	sti := idIdx + 2
	for _, cp := range append(commas, rbIdx) {
		if sti > idIdx+2 {
			gtoks.Add(token.COMMA)
		}
		gtoks.Add(token.IDENT, "int")
		gtoks.Add(token.LPAREN)
		gtoks.AddTokens(st.TranspileGo(toks[sti:cp], code)...)
		gtoks.Add(token.RPAREN)
		sti = cp + 1
	}
	gtoks.Add(token.RPAREN)
	if !isSet {
		return rbIdx
	}
	if hasComment {
		gtoks.AddTokens(toks[len(toks)-1])
	}
	return len(toks)
}

// receiverStart returns the index in toks of the start of the primary
// expression that ends toks, such as x, a.b.x, f(y).x or a[i].x.
func receiverStart(toks Tokens) int {
	i := len(toks) - 1
	for i >= 0 {
		// Skip back over a bracketed part, such as the arguments of a call
		// or an index, to the operand that it applies to.
		switch toks[i].Tok {
		case token.RPAREN, token.RBRACK:
			if lb := leftMatching(toks[:i+1]); lb > 0 {
				i = lb - 1
				continue
			}
			return i + 1
		case token.IDENT:
		default:
			return i + 1
		}

		// An identifier continues a selector chain if a period precedes it.
		if i >= 1 && toks[i-1].Tok == token.PERIOD {
			i -= 2
			continue
		}
		return i
	}
	return 0
}

// leftMatching returns the index of the bracket that matches the bracket
// that ends toks, or -1 if there is none.
func leftMatching(toks Tokens) int {
	rb := toks[len(toks)-1].Tok
	lb := token.LPAREN
	if rb == token.RBRACK {
		lb = token.LBRACK
	}
	depth := 0
	for i := len(toks) - 2; i >= 0; i-- {
		switch toks[i].Tok {
		case lb:
			if depth == 0 {
				return i
			}
			depth--
		case rb:
			depth++
		}
	}
	return -1
}

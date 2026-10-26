// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"cogentcore.org/core/base/stack"
	"cogentcore.org/lab/goal/transpile/mparser"
	"cogentcore.org/lab/tensor"
)

// TranspileMath does math mode transpiling. fullLine indicates code should be
// full statement(s).
func (st *State) TranspileMath(toks Tokens, code string, fullLine bool) Tokens {
	if fullLine {
		st.math.endBlocks(st.BraceDepth)
	}
	return st.transpileMath(toks, code, fullLine, st.BraceDepth)
}

// transpileMath does math mode transpiling of the given tokens, which are
// in the block of the given brace depth.
func (st *State) transpileMath(toks Tokens, code string, fullLine bool, depth int) Tokens {
	nt := len(toks)
	if nt == 0 {
		return nil
	}
	if fullLine {
		switch toks[0].Tok {
		case token.RBRACE:
			return st.mathCloseLine(toks, code, depth)
		case token.CASE, token.DEFAULT:
			return st.mathCaseLine(toks, code, depth)
		case token.IDENT: // a label on its own line
			if nt >= 2 && toks[1].Tok == token.COLON && (nt == 2 || nt == 3 && toks[2].Tok == token.COMMENT) {
				return toks
			}
		}
	}

	str := mathSource(toks, code)
	mp := mathParse{state: st, toks: toks, code: code, depth: depth}
	// mp.trace = true

	mods := mparser.AllErrors // | Trace

	if fullLine {
		ewords, err := ExecWords(str)
		if len(ewords) > 0 {
			if cmd, ok := tensorfsCommands[ewords[0]]; ok {
				mp.ewords = ewords
				err := cmd(&mp)
				if err != nil {
					fmt.Println(ewords[0]+":", err.Error())
					return nil
				} else {
					return mp.out
				}
			}
		}

		stmts, err := mparser.ParseLine(str, mods)
		if err != nil {
			fmt.Println("line code:", str)
			fmt.Println("parse err:", err)
		}
		if len(stmts) == 0 {
			return toks
		}
		mp.stmtList(stmts)
	} else {
		ex, err := mparser.ParseExpr(str, mods)
		if err != nil {
			fmt.Println("expr:", str)
			fmt.Println("parse err:", err)
		}
		mp.expr(ex)
	}

	if !(mp.idx == len(toks) || mp.idx == len(toks)-1) { // -1 is for comment at end
		fmt.Println(code)
		fmt.Println(mp.out.Code())
		fmt.Printf("parsing error: index: %d != len(toks): %d\n", mp.idx, len(toks))
	}

	return mp.out
}

// mathSource returns the source code of the given tokens.
func mathSource(toks Tokens, code string) string {
	last := toks[len(toks)-1]
	n := len(last.Str)
	if n == 0 { // operators and other tokens without their text
		n = len(last.Tok.String())
	}
	return code[toks[0].Pos-1 : int(last.Pos)-1+n]
}

// mathCloseLine transpiles a line of the given block depth that starts with
// a closing brace, such as "} else if x > 0 {". The rest of it is in the
// enclosing block.
func (st *State) mathCloseLine(toks Tokens, code string, depth int) Tokens {
	out := Tokens{toks[0]}
	rest := toks[1:]
	if len(rest) > 0 && rest[0].Tok == token.ELSE {
		out = append(out, rest[0])
		rest = rest[1:]
		if len(rest) > 0 && rest[0].Tok == token.IF {
			return append(out, st.transpileMath(rest, code, true, depth-1)...)
		}
	}
	return append(out, rest...) // "{", or the end of an expression, such as ")"
}

// mathCaseLine transpiles a case or default line of a switch statement
// whose block is at the given depth: the cases of a switch without a tag
// are conditions (see [mathParse.cond]), and the others expressions.
// Statements after the colon are transpiled as in the case block.
func (st *State) mathCaseLine(toks Tokens, code string, depth int) Tokens {
	colon := topLevelIndex(toks, 0, token.COLON)
	if colon < 0 {
		return toks
	}
	out := Tokens{toks[0]}
	if toks[0].Tok == token.CASE {
		tagless := st.math.taglessSwitch(depth)
		start := 1
		for start < colon {
			end := topLevelIndex(toks[:colon], start, token.COMMA)
			if end < 0 {
				end = colon
			}
			out = append(out, st.mathCaseExpr(toks[start:end], code, depth, tagless)...)
			if end < colon {
				out = append(out, toks[end])
			}
			start = end + 1
		}
	}
	out = append(out, toks[colon])
	if colon+1 < len(toks) {
		out = append(out, st.transpileMath(toks[colon+1:], code, true, depth)...)
	}
	return out
}

// mathCaseExpr transpiles an expression of a case, as a condition if cond.
func (st *State) mathCaseExpr(toks Tokens, code string, depth int, cond bool) Tokens {
	if len(toks) == 0 {
		return nil
	}
	mp := mathParse{state: st, toks: toks, code: code, depth: depth}
	ex, err := mparser.ParseExpr(mathSource(toks, code), mparser.AllErrors)
	if err != nil {
		return toks
	}
	if cond {
		mp.cond(ex)
	} else {
		mp.anyExpr(ex)
	}
	return mp.out
}

// topLevelIndex returns the index of the first token of the given kind at
// or after start that is not within parentheses, brackets or braces,
// or -1 if there is none.
func topLevelIndex(toks Tokens, start int, tok token.Token) int {
	depth := 0
	for i := start; i < len(toks); i++ {
		switch toks[i].Tok {
		case token.LPAREN, token.LBRACK, token.LBRACE:
			depth++
		case token.RPAREN, token.RBRACK, token.RBRACE:
			depth--
		case tok:
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// funcInfo is info about the function being processed
type funcInfo struct {
	tensor.Func

	// current arg index we are processing
	curArg int

	// nargs is the number of args of the call.
	nargs int
}

// mathParse has the parsing state, active only during a parsing pass
// on one specific chunk of code and tokens.
type mathParse struct {
	state   *State
	code    string   // code string
	toks    Tokens   // source tokens we are parsing
	ewords  []string // exec words
	idx     int      // current index in source tokens -- critical to sync as we "use" source
	out     Tokens   // output tokens we generate
	trace   bool     // trace of parsing -- turn on to see alignment
	inArray bool     // we are in an array

	// depth is the brace depth of the block of the code.
	depth int

	// stack of function info -- top of stack reflects the current function
	funcs stack.Stack[*funcInfo]
}

// returns the current argument for current function
func (mp *mathParse) curArg() *tensor.Arg {
	cfun := mp.funcs.Peek()
	if cfun == nil {
		return nil
	}
	if cfun.curArg < len(cfun.Args) {
		return cfun.Args[cfun.curArg]
	}
	return nil
}

func (mp *mathParse) nextArg() {
	cfun := mp.funcs.Peek()
	if cfun == nil || len(cfun.Args) == 0 {
		// fmt.Println("next arg no fun or no args")
		return
	}
	n := len(cfun.Args)
	if cfun.curArg == n-1 {
		carg := cfun.Args[n-1]
		if !carg.IsVariadic {
			fmt.Println("math transpile: args exceed registered function number:", cfun)
		}
		return
	}
	cfun.curArg++
}

func (mp *mathParse) curArgIsTensor() bool {
	carg := mp.curArg()
	if carg == nil {
		return false
	}
	return carg.IsTensor
}

func (mp *mathParse) curArgIsInts() bool {
	carg := mp.curArg()
	if carg == nil {
		return false
	}
	return carg.IsInt && carg.IsVariadic
}

// curArgIsInt returns whether the current arg is an int, and if so,
// whether it is the only arg of a variadic ...int parameter, which then
// takes a slice of ints.
func (mp *mathParse) curArgIsInt() (isInt, ints bool) {
	carg := mp.curArg()
	if carg == nil || !carg.IsInt {
		return false, false
	}
	cfun := mp.funcs.Peek()
	return true, carg.IsVariadic && cfun.nargs == len(cfun.Args)
}

// startFunc is called when starting a new function.
// empty is "dummy" assign case using Inc.
// optional noLookup indicates to not lookup type and just
// push the name -- for internal cases to prevent arg conversions.
func (mp *mathParse) startFunc(name string, noLookup ...bool) *funcInfo {
	fi := &funcInfo{}
	sname := name
	if name == "" || name == "tensor.Tensor" {
		sname = "tmath.Inc" // one arg tensor fun
	}
	if len(noLookup) == 1 && noLookup[0] {
		fi.Name = name
	} else {
		if tf, err := tensor.FuncByName(sname); err == nil {
			fi.Func = *tf
		} else {
			fi.Name = name
		}
	}
	mp.funcs.Push(fi)
	if name != "" {
		mp.out.Add(token.IDENT, name)
	}
	return fi
}

func (mp *mathParse) endFunc() {
	mp.funcs.Pop()
}

// addToken adds output token and increments idx
func (mp *mathParse) addToken(tok token.Token) {
	mp.out.Add(tok)
	if mp.trace {
		ctok := &Token{}
		if mp.idx < len(mp.toks) {
			ctok = mp.toks[mp.idx]
		}
		fmt.Printf("%d\ttok: %s \t replaces: %s\n", mp.idx, tok, ctok)
	}
	mp.idx++
}

func (mp *mathParse) addCur() {
	if len(mp.toks) > mp.idx {
		mp.out.AddTokens(mp.toks[mp.idx])
		mp.idx++
		return
	}
	fmt.Println("out of tokens!", mp.idx, mp.toks)
}

// astPos returns the position in the parsed source of source token i.
func (mp *mathParse) astPos(i int) token.Pos {
	return mp.toks[i].Pos - mp.toks[0].Pos + 1
}

// copyTo copies the source tokens before the given position to the output.
func (mp *mathParse) copyTo(end token.Pos) {
	for mp.idx < len(mp.toks) && mp.astPos(mp.idx) < end {
		mp.addCur()
	}
}

// copyNode copies the source tokens of the given node (and of any before
// it that are not yet used) to the output, as plain Go code.
func (mp *mathParse) copyNode(n ast.Node) {
	mp.copyTo(n.End())
}

// copySemi copies a semicolon that is the current source token.
func (mp *mathParse) copySemi() {
	if mp.idx < len(mp.toks) && mp.toks[mp.idx].Tok == token.SEMICOLON {
		mp.addCur()
	}
}

func (mp *mathParse) stmtList(sts []ast.Stmt) {
	for _, st := range sts {
		mp.stmt(st)
		mp.copySemi()
	}
}

func (mp *mathParse) stmt(st ast.Stmt) {
	if st == nil {
		return
	}
	switch x := st.(type) {
	case *ast.BadStmt:
		fmt.Println("bad stmt!")

	case *ast.DeclStmt:
		mp.declStmt(x)

	case *ast.LabeledStmt:
		mp.addCur() // label
		mp.addCur() // :
		mp.stmt(x.Stmt)

	case *ast.ExprStmt:
		mp.expr(x.X)

	case *ast.SendStmt:
		mp.expr(x.Chan)
		mp.addToken(token.ARROW)
		mp.expr(x.Value)

	case *ast.IncDecStmt:
		mp.incDecStmt(x)

	case *ast.AssignStmt:
		switch x.Tok {
		case token.DEFINE:
			mp.defineStmt(x)
		default:
			mp.assignStmt(x)
		}

	case *ast.GoStmt:
		mp.addToken(token.GO)
		mp.callExpr(x.Call)

	case *ast.DeferStmt:
		mp.addToken(token.DEFER)
		mp.callExpr(x.Call)

	case *ast.ReturnStmt:
		mp.addToken(token.RETURN)
		mp.exprList(x.Results)

	case *ast.BranchStmt:
		mp.addToken(x.Tok)
		mp.ident(x.Label)

	case *ast.BlockStmt:
		mp.block(x)

	case *ast.IfStmt:
		mp.addToken(token.IF)
		if x.Init != nil {
			mp.stmt(x.Init)
			mp.addToken(token.SEMICOLON)
		}
		mp.cond(x.Cond)
		mp.block(x.Body)
		if x.Else != nil {
			mp.addToken(token.ELSE)
			mp.stmt(x.Else)
		}

	case *ast.ForStmt:
		mp.forStmt(x)

	case *ast.RangeStmt:
		mp.rangeStmt(x)

	case *ast.SwitchStmt:
		mp.addToken(token.SWITCH)
		if x.Init != nil {
			mp.stmt(x.Init)
			mp.copySemi()
		}
		if x.Tag != nil {
			mp.anyExpr(x.Tag)
		}
		mp.state.math.addSwitch(mp.depth+1, x.Tag == nil)
		mp.copyNode(x.Body)

	case *ast.TypeSwitchStmt:
		mp.copyNode(x)
		mp.state.math.addSwitch(mp.depth+1, false)

	default: // select etc: plain Go
		mp.copyNode(x)
	}
}

// block transpiles a block, which can be only its opening brace, at the
// end of a line.
func (mp *mathParse) block(b *ast.BlockStmt) {
	if b == nil {
		return
	}
	mp.addToken(token.LBRACE)
	mp.depth++
	mp.stmtList(b.List)
	mp.depth--
	if b.Rbrace.IsValid() {
		mp.addToken(token.RBRACE)
	}
}

// forStmt transpiles a for loop, whose init and post statements are plain
// Go, so that its variables are Go values, and whose condition is a
// condition (see [mathParse.cond]).
func (mp *mathParse) forStmt(x *ast.ForStmt) {
	mp.addToken(token.FOR)
	if x.Init != nil {
		if as, ok := x.Init.(*ast.AssignStmt); ok && as.Tok == token.DEFINE {
			for i, l := range as.Lhs {
				if id, ok := l.(*ast.Ident); ok {
					typ := ""
					if len(as.Lhs) == len(as.Rhs) {
						typ, _, _ = mp.goType(as.Rhs[i])
					}
					mp.state.math.declare(id.Name, mp.depth+1, mathGo, defaultType(typ))
				}
			}
		}
		mp.goStmt(x.Init)
	}
	mp.copySemi()
	if x.Cond != nil {
		mp.cond(x.Cond)
	}
	mp.copySemi()
	if x.Post != nil {
		if mp.isGoStmt(x.Post) {
			mp.goStmt(x.Post)
		} else {
			mp.stmt(x.Post)
		}
	}
	mp.block(x.Body)
}

// rangeStmt transpiles a range loop. A range over a Go int is plain Go.
// Otherwise, it is over [tensor.Iter], which ranges over an int, or the
// indexes and float64 values of a tensor or a slice of numbers, so that
// its variables are Go values.
func (mp *mathParse) rangeStmt(x *ast.RangeStmt) {
	mp.addToken(token.FOR)
	mp.copyTo(x.X.Pos()) // key, value, := and range
	goRange := mp.isGoExpr(x.X)
	if lit, ok := x.X.(*ast.BasicLit); ok && lit.Kind == token.INT {
		goRange = true
	}
	if goRange {
		mp.goExprOut(x.X)
	} else {
		mp.out.Add(token.IDENT, "tensor.Iter")
		mp.out.Add(token.LPAREN)
		mp.anyExpr(x.X)
		mp.out.Add(token.RPAREN)
	}
	if x.Tok == token.DEFINE {
		valueType := "float64" // of tensor.Iter
		if goRange {
			valueType = "" // of a Go slice etc
		}
		if id, ok := x.Key.(*ast.Ident); ok {
			mp.state.math.declare(id.Name, mp.depth+1, mathGo, "int")
		}
		if id, ok := x.Value.(*ast.Ident); ok {
			mp.state.math.declare(id.Name, mp.depth+1, mathGo, valueType)
		}
	}
	mp.block(x.Body)
}

// declStmt transpiles a var, const or type declaration, which is plain Go.
// The variables of basic types, and of untyped Go values, are Go values.
func (mp *mathParse) declStmt(ds *ast.DeclStmt) {
	gd, ok := ds.Decl.(*ast.GenDecl)
	if !ok || gd.Tok == token.TYPE || gd.Tok == token.IMPORT {
		mp.copyNode(ds)
		return
	}
	for _, sp := range gd.Specs {
		vs, ok := sp.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, nm := range vs.Names {
			kind, typ := mathUnknown, ""
			switch {
			case vs.Type != nil:
				kind, typ = typeKind(vs.Type)
			case i < len(vs.Values):
				if vt, g, _ := mp.goType(vs.Values[i]); g {
					kind, typ = mathGo, defaultType(vt)
				}
			}
			mp.state.math.declare(nm.Name, mp.depth, kind, typ)
		}
		for _, v := range vs.Values {
			mp.copyTo(v.Pos())
			mp.goExprOut(v)
		}
	}
	mp.copyNode(ds)
}

// basicTypes are the Go basic types, whose values math mode handles.
var basicTypes = map[string]bool{
	"bool": true, "string": true, "int": true, "int8": true, "int16": true,
	"int32": true, "int64": true, "uint": true, "uint8": true, "uint16": true,
	"uint32": true, "uint64": true, "uintptr": true, "byte": true, "rune": true,
	"float32": true, "float64": true,
}

// typeKind returns the kind of the values of the given type expression:
// Go for basic types, with the type, tensor for tensor.Tensor and
// tensor.Values, and unknown for others.
func typeKind(typ ast.Expr) (mathKind, string) {
	switch x := typ.(type) {
	case *ast.Ident:
		if basicTypes[x.Name] {
			return mathGo, x.Name
		}
	case *ast.SelectorExpr:
		if pkg, ok := x.X.(*ast.Ident); ok && pkg.Name == "tensor" && (x.Sel.Name == "Tensor" || x.Sel.Name == "Values") {
			return mathTensor, ""
		}
	}
	return mathUnknown, ""
}

// funcLit transpiles a function literal, which is plain Go. If its body
// is open at the end of the line, its parameters are declared in it.
func (mp *mathParse) funcLit(fl *ast.FuncLit) {
	mp.copyNode(fl)
	if fl.Body == nil || fl.Body.Rbrace.IsValid() || fl.Type.Params == nil {
		return
	}
	for _, fd := range fl.Type.Params.List {
		kind, typ := typeKind(fd.Type)
		for _, nm := range fd.Names {
			mp.state.math.declare(nm.Name, mp.depth+1, kind, typ)
		}
	}
}

// isGoStmt returns whether the given statement only involves Go values:
// an increment of a Go variable, or an assignment of Go values to them.
func (mp *mathParse) isGoStmt(s ast.Stmt) bool {
	switch x := s.(type) {
	case *ast.IncDecStmt:
		id, ok := x.X.(*ast.Ident)
		return ok && mp.state.math.isGo(id.Name)
	case *ast.AssignStmt:
		if x.Tok == token.DEFINE {
			return false
		}
		for i, l := range x.Lhs {
			id, ok := l.(*ast.Ident)
			if !ok || !(id.Name == "_" || mp.state.math.isGo(id.Name)) {
				return false
			}
			if len(x.Lhs) != len(x.Rhs) {
				continue
			}
			rt, g, _ := mp.goType(x.Rhs[i])
			if !g {
				return false
			}
			if _, ok := commonType(mp.state.math.lookup(id.Name).typ, rt); !ok && id.Name != "_" {
				return false
			}
		}
		for _, r := range x.Rhs {
			if g, _ := mp.goExpr(r); !g {
				return false
			}
		}
		return true
	}
	return false
}

// goStmt transpiles a statement as plain Go, in which only the
// conversions are transpiled (see [mathParse.goExprOut]).
func (mp *mathParse) goStmt(s ast.Stmt) {
	if as, ok := s.(*ast.AssignStmt); ok {
		for _, r := range as.Rhs {
			mp.copyTo(r.Pos())
			if fl, ok := r.(*ast.FuncLit); ok {
				mp.funcLit(fl)
			} else {
				mp.goExprOut(r)
			}
		}
	}
	mp.copyNode(s)
}

func (mp *mathParse) incDecStmt(x *ast.IncDecStmt) {
	fn := "Inc"
	if x.Tok == token.DEC {
		fn = "Dec"
	}
	if _, isIdent := x.X.(*ast.Ident); !isIdent && mp.isGoExpr(x.X) { // an element or field of a Go value
		mp.copyNode(x)
		return
	}
	switch lx := x.X.(type) {
	case *ast.Ident:
		switch mp.state.math.kind(lx.Name) {
		case mathGo:
			mp.copyNode(x)
			return
		case mathUnknown:
			mp.varAssign(x.X, x.Tok, nil)
			return
		}
	case *ast.SelectorExpr, *ast.StarExpr:
		mp.varAssign(x.X, x.Tok, nil)
		return
	}
	mp.startFunc("tmath." + fn)
	mp.out.Add(token.LPAREN)
	mp.expr(x.X)
	mp.addToken(token.RPAREN)
	mp.endFunc()
}

func (mp *mathParse) expr(ex ast.Expr) {
	if ex == nil {
		return
	}
	if mp.curArgIsTensor() && mp.needsToTensor(ex) {
		mp.toTensor(ex)
		return
	}
	if isInt, ints := mp.curArgIsInt(); isInt && mp.needsInts(ex) {
		if ints {
			mp.out.Add(token.IDENT, "tensor.ToInts")
		} else {
			mp.out.Add(token.IDENT, "tensor.ToInt")
		}
		mp.out.Add(token.LPAREN)
		mp.anyExpr(ex)
		mp.out.Add(token.RPAREN)
		if ints {
			mp.out.Add(token.ELLIPSIS)
		}
		return
	}
	if mp.isGoExpr(ex) {
		mp.goExprOut(ex)
		return
	}
	switch x := ex.(type) {
	case *ast.BadExpr:
		fmt.Println("bad expr!")

	case *ast.Ident:
		mp.ident(x)

	case *ast.UnaryExpr:
		mp.unaryExpr(x)

	case *ast.Ellipsis:
		cfun := mp.funcs.Peek()
		if cfun != nil && cfun.Name == "tensor.AnySlice" {
			mp.out.Add(token.IDENT, "tensor.Ellipsis")
			mp.idx++
		} else {
			mp.addToken(token.ELLIPSIS)
		}

	case *ast.StarExpr:
		mp.addToken(token.MUL)
		mp.expr(x.X)

	case *ast.BinaryExpr:
		mp.binaryExpr(x)

	case *ast.BasicLit:
		mp.basicLit(x)

	case *ast.FuncLit:
		mp.funcLit(x)

	case *ast.ParenExpr:
		mp.addToken(token.LPAREN)
		mp.expr(x.X)
		mp.addToken(token.RPAREN)

	case *ast.SelectorExpr:
		mp.selectorExpr(x)

	case *ast.IndexExpr:
		mp.indexExpr(x)

	case *ast.IndexListExpr:
		if x.X == nil { // array literal
			mp.arrayLiteral(x)
		} else {
			mp.indexListExpr(x)
		}

	case *ast.SliceExpr:
		mp.sliceExpr(x)

	case *ast.CallExpr:
		mp.callExpr(x)

	default: // composite literals, type assertions, types: plain Go
		mp.copyNode(x)
	}
}

// anyExpr transpiles an expression in a position that takes any value,
// such as the arg of a Go function, which does not convert it.
func (mp *mathParse) anyExpr(ex ast.Expr) {
	mp.startFunc("", true)
	mp.expr(ex)
	mp.endFunc()
}

// toTensor transpiles an expression that is not known to be a tensor
// where a tensor is needed, converting it with [tensor.ToTensor].
func (mp *mathParse) toTensor(ex ast.Expr) {
	mp.out.Add(token.IDENT, "tensor.ToTensor")
	mp.out.Add(token.LPAREN)
	mp.anyExpr(ex)
	mp.out.Add(token.RPAREN)
}

// cond transpiles a condition, of an if, for or a case of a switch without
// a tag: a Go condition as such, and others with [tensor.ToBool],
// which works for tensors and Go values.
func (mp *mathParse) cond(ex ast.Expr) {
	if g, _ := mp.goExpr(ex); g {
		mp.goExprOut(ex)
		return
	}
	mp.out.Add(token.IDENT, "tensor.ToBool")
	mp.out.Add(token.LPAREN)
	mp.anyExpr(ex)
	mp.out.Add(token.RPAREN)
}

// needsToTensor returns whether the given expression, where a tensor is
// needed, is not known to be one, and needs [tensor.ToTensor].
func (mp *mathParse) needsToTensor(ex ast.Expr) bool {
	switch x := ex.(type) {
	case *ast.Ident:
		if _, ok := consts[x.Name]; ok {
			return false
		}
		return !mp.state.math.isTensor(x.Name)
	case *ast.ParenExpr:
		return mp.needsToTensor(x.X)
	case *ast.SelectorExpr:
		_, isProp := numpyProps[x.Sel.Name]
		return !isProp
	case *ast.CallExpr:
		return !mp.callIsTensor(x)
	case *ast.BinaryExpr, *ast.UnaryExpr:
		return mp.isGoExpr(ex)
	case *ast.CompositeLit, *ast.TypeAssertExpr, *ast.StarExpr:
		return true
	}
	return false
}

// needsInts returns whether the given expression, where an int is needed,
// is not known to be one, and needs [tensor.ToInt] or [tensor.ToInts].
func (mp *mathParse) needsInts(ex ast.Expr) bool {
	switch x := ex.(type) {
	case *ast.BasicLit:
		return false
	case *ast.IndexListExpr:
		return x.X != nil // not an array literal
	case *ast.ParenExpr:
		return mp.needsInts(x.X)
	case *ast.UnaryExpr:
		if _, ok := x.X.(*ast.BasicLit); ok {
			return false
		}
	}
	return !mp.isGoExpr(ex)
}

// goType returns the Go type of the given expression, and whether it is a
// Go value, made only of known Go variables, literals and conversions (see
// [mathNames]), and whether it has any such variable or conversion. The
// type is "" if not known, and an untyped type for a constant, such as
// "untyped int". An operator on values of types that Go does not allow it
// on, such as an int and a float64, is not a Go value.
func (mp *mathParse) goType(ex ast.Expr) (typ string, isGo, hasName bool) {
	switch x := ex.(type) {
	case *ast.Ident:
		if n := mp.state.math.lookup(x.Name); n.kind == mathGo {
			return n.typ, true, true
		}
		if x.Name == "true" || x.Name == "false" {
			return "untyped bool", true, false
		}
	case *ast.BasicLit:
		return literalTypes[x.Kind], true, false
	case *ast.ParenExpr:
		return mp.goType(x.X)
	case *ast.UnaryExpr:
		switch x.Op {
		case token.SUB, token.ADD, token.XOR:
			return mp.goType(x.X)
		case token.NOT:
			_, g, h := mp.goType(x.X)
			return "bool", g, h
		}
	case *ast.BinaryExpr:
		if x.Op == token.ILLEGAL { // @ matmul
			return "", false, false
		}
		if _, isPow := x.Y.(*ast.StarExpr); isPow && x.Op == token.MUL {
			return "", false, false
		}
		tx, gx, hx := mp.goType(x.X)
		ty, gy, hy := mp.goType(x.Y)
		if !gx || !gy {
			return "", false, false
		}
		switch x.Op {
		case token.LAND, token.LOR:
			return "bool", true, hx || hy
		case token.SHL, token.SHR:
			return tx, true, hx || hy
		}
		typ, ok := commonType(tx, ty)
		if !ok {
			return "", false, false
		}
		switch x.Op {
		case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
			typ = "bool"
		}
		return typ, true, hx || hy
	case *ast.SelectorExpr: // a field or method of a Go value
		if _, g, h := mp.goType(x.X); g && h {
			return "", true, true
		}
	case *ast.IndexExpr: // an element of a Go value
		if _, g, h := mp.goType(x.X); !g || !h {
			return "", false, false
		}
		il, ok := x.Index.(*ast.IndexListExpr)
		if !ok {
			return "", false, false
		}
		for _, ix := range il.Indices {
			if _, g, _ := mp.goType(ix); !g {
				return "", false, false
			}
		}
		return "", true, true
	case *ast.CompositeLit, *ast.FuncLit:
		return "", true, true
	case *ast.CallExpr:
		if isConversion(x) {
			return x.Fun.(*ast.Ident).Name, true, true
		}
		// A call of a Go function value or method, or len or cap of a
		// Go value, with Go args.
		switch f := x.Fun.(type) {
		case *ast.Ident:
			if f.Name == "len" || f.Name == "cap" {
				if len(x.Args) != 1 {
					return "", false, false
				}
				_, g, h := mp.goType(x.Args[0])
				return "int", g && h, true
			}
			if !mp.state.math.isGo(f.Name) {
				return "", false, false
			}
		default:
			if _, g, h := mp.goType(f); !g || !h {
				return "", false, false
			}
		}
		for _, a := range x.Args {
			if _, g, _ := mp.goType(a); !g {
				return "", false, false
			}
		}
		return "", true, true
	}
	return "", false, false
}

// goExpr returns whether the given expression is a Go value, and whether
// it has any Go variable or conversion (see [mathParse.goType]).
func (mp *mathParse) goExpr(ex ast.Expr) (isGo, hasName bool) {
	_, g, h := mp.goType(ex)
	return g, h
}

// isGoExpr returns whether the given expression is a Go value with at least
// one Go variable or conversion, so that its operators are Go operators.
// Expressions of only literals are tensors.
func (mp *mathParse) isGoExpr(ex ast.Expr) bool {
	g, h := mp.goExpr(ex)
	return g && h
}

// goExprOut transpiles a Go expression as such, except for conversions
// (see [mathParse.conversion]).
func (mp *mathParse) goExprOut(ex ast.Expr) {
	switch x := ex.(type) {
	case *ast.ParenExpr:
		mp.addCur()
		mp.goExprOut(x.X)
		mp.addCur()
	case *ast.UnaryExpr:
		mp.addCur()
		mp.goExprOut(x.X)
	case *ast.BinaryExpr:
		mp.goExprOut(x.X)
		mp.addCur()
		mp.goExprOut(x.Y)
	case *ast.CallExpr:
		if isConversion(x) {
			mp.conversion(x)
			return
		}
		mp.goExprsOut(x.Args)
		mp.copyNode(x)
	case *ast.IndexExpr:
		if il, ok := x.Index.(*ast.IndexListExpr); ok {
			mp.goExprsOut(il.Indices)
		}
		mp.copyNode(x)
	case *ast.FuncLit:
		mp.funcLit(x)
	default:
		mp.copyNode(x)
	}
}

// goExprsOut transpiles the given Go expressions, such as the args of a
// call, with the source between them.
func (mp *mathParse) goExprsOut(exs []ast.Expr) {
	for _, ex := range exs {
		mp.copyTo(ex.Pos())
		mp.goExprOut(ex)
	}
}

// isConversion returns whether the given call is a conversion to a Go
// basic type, such as float64(x).
func isConversion(x *ast.CallExpr) bool {
	id, ok := x.Fun.(*ast.Ident)
	return ok && basicTypes[id.Name] && len(x.Args) == 1
}

// conversion transpiles a conversion to a Go basic type: of a Go value as
// such, and of others with the To function of [tensor] for the type,
// such as [tensor.ToFloat64], which works for tensors and Go values.
func (mp *mathParse) conversion(x *ast.CallExpr) {
	typ := x.Fun.(*ast.Ident).Name
	arg := x.Args[0]
	if g, _ := mp.goExpr(arg); g {
		mp.addCur() // type
		mp.addCur() // (
		mp.goExprOut(arg)
		mp.addCur() // )
		return
	}
	fn, wrap := "", typ
	switch typ {
	case "float64":
		fn, wrap = "tensor.ToFloat64", ""
	case "float32":
		fn = "tensor.ToFloat64"
	case "int":
		fn, wrap = "tensor.ToInt", ""
	case "bool":
		fn, wrap = "tensor.ToBool", ""
	case "string":
		fn, wrap = "tensor.ToString", ""
	default: // other ints
		fn = "tensor.ToInt"
	}
	if wrap != "" {
		mp.out.Add(token.IDENT, wrap)
		mp.out.Add(token.LPAREN)
	}
	mp.out.Add(token.IDENT, fn)
	mp.idx++ // type
	mp.addToken(token.LPAREN)
	mp.anyExpr(arg)
	mp.addToken(token.RPAREN)
	if wrap != "" {
		mp.out.Add(token.RPAREN)
	}
}

func (mp *mathParse) exprList(ex []ast.Expr) {
	n := len(ex)
	if n == 0 {
		return
	}
	if n == 1 {
		mp.expr(ex[0])
		return
	}
	for i := range n {
		mp.expr(ex[i])
		if i < n-1 {
			mp.addToken(token.COMMA)
		}
	}
}

func (mp *mathParse) argsList(ex []ast.Expr) {
	n := len(ex)
	if cfun := mp.funcs.Peek(); cfun != nil {
		cfun.nargs = n
	}
	if n == 0 {
		return
	}
	if n == 1 {
		mp.expr(ex[0])
		return
	}
	for i := range n {
		// cfun := mp.funcs.Peek()
		// if i != cfun.curArg {
		// 	fmt.Println(cfun, "arg should be:", i, "is:", cfun.curArg)
		// }
		mp.expr(ex[i])
		if i < n-1 {
			mp.nextArg()
			mp.addToken(token.COMMA)
		}
	}
}

func (mp *mathParse) binaryExpr(ex *ast.BinaryExpr) {
	if ex.Op == token.ILLEGAL { // @ = matmul
		mp.startFunc("matrix.Mul")
		mp.out.Add(token.LPAREN)
		mp.expr(ex.X)
		mp.out.Add(token.COMMA)
		mp.idx++
		mp.expr(ex.Y)
		mp.out.Add(token.RPAREN)
		mp.endFunc()
		return
	}

	fn := ""
	switch ex.Op {
	case token.ADD:
		fn = "Add"
	case token.SUB:
		fn = "Sub"
	case token.MUL:
		fn = "Mul"
		if un, ok := ex.Y.(*ast.StarExpr); ok { // ** power operator
			ex.Y = un.X
			fn = "Pow"
		}
	case token.QUO:
		fn = "Div"
	case token.REM:
		fn = "Mod"
	case token.EQL:
		fn = "Equal"
	case token.LSS:
		fn = "Less"
	case token.GTR:
		fn = "Greater"
	case token.NEQ:
		fn = "NotEqual"
	case token.LEQ:
		fn = "LessEqual"
	case token.GEQ:
		fn = "GreaterEqual"
	case token.LOR:
		fn = "Or"
	case token.LAND:
		fn = "And"
	default:
		fmt.Println("binary token:", ex.Op)
	}
	mp.startFunc("tmath." + fn)
	mp.out.Add(token.LPAREN)
	mp.expr(ex.X)
	mp.out.Add(token.COMMA)
	mp.idx++
	if fn == "Pow" {
		mp.idx++
	}
	mp.nextArg()
	mp.expr(ex.Y)
	mp.out.Add(token.RPAREN)
	mp.endFunc()
}

func (mp *mathParse) unaryExpr(ex *ast.UnaryExpr) {
	if lit, isbl := ex.X.(*ast.BasicLit); isbl {
		if mp.curArgIsTensor() && (ex.Op == token.SUB || ex.Op == token.ADD) {
			mp.idx++ // the sign goes in the literal
			mp.tensorLit(lit, ex.Op.String())
			return
		}
		mp.addToken(ex.Op)
		mp.expr(ex.X)
		return
	}
	fn := ""
	switch ex.Op {
	case token.NOT:
		fn = "Not"
	case token.SUB:
		fn = "Negate"
	case token.ADD:
		mp.idx++
		mp.expr(ex.X)
		return
	default: // * goes to StarExpr, & and <- are Go
		mp.addToken(ex.Op)
		mp.anyExpr(ex.X)
		return
	}
	mp.startFunc("tmath." + fn)
	mp.addToken(token.LPAREN)
	mp.expr(ex.X)
	mp.out.Add(token.RPAREN)
	mp.endFunc()
}

// defineStmt transpiles a := definition. It defines Go variables if it is
// Go code: of several variables, of a value that only Go has, such as a
// composite literal, or of a Go expression (see [mathParse.goExpr]).
// Otherwise, it defines a tensor.
func (mp *mathParse) defineStmt(as *ast.AssignStmt) {
	if mp.isGoDefine(as) {
		mp.goStmt(as)
		for i, l := range as.Lhs {
			id, ok := l.(*ast.Ident)
			if !ok {
				continue
			}
			kind, typ := mathUnknown, ""
			rhs := as.Rhs[0]
			if len(as.Lhs) == len(as.Rhs) {
				rhs = as.Rhs[i]
				typ, _, _ = mp.goType(rhs)
			}
			if mp.isGoValue(rhs) {
				kind = mathGo
			}
			mp.state.math.declare(id.Name, mp.depth, kind, defaultType(typ))
		}
		return
	}
	firstStmt := mp.idx == 0
	mp.exprList(as.Lhs)
	mp.addToken(as.Tok)
	if mp.needsToTensor(as.Rhs[0]) {
		mp.toTensor(as.Rhs[0])
	} else {
		mp.startFunc("tensor.Tensor")
		mp.out.Add(token.LPAREN)
		mp.exprList(as.Rhs)
		mp.out.Add(token.RPAREN)
		mp.endFunc()
	}
	for _, l := range as.Lhs {
		if id, ok := l.(*ast.Ident); ok {
			mp.state.math.declare(id.Name, mp.depth, mathTensor, "")
		}
	}
	if firstStmt && mp.state.MathRecord {
		nvar, ok := as.Lhs[0].(*ast.Ident)
		if ok {
			mp.out.Add(token.SEMICOLON)
			mp.out.Add(token.IDENT, "tensorfs.Record("+nvar.Name+",`"+nvar.Name+"`)")
		}
	}
}

// isGoDefine returns whether the given := definition defines Go variables.
func (mp *mathParse) isGoDefine(as *ast.AssignStmt) bool {
	if len(as.Lhs) != 1 || len(as.Rhs) != 1 {
		return true
	}
	switch r := ast.Unparen(as.Rhs[0]).(type) {
	case *ast.CompositeLit, *ast.FuncLit, *ast.TypeAssertExpr:
		return true
	case *ast.UnaryExpr:
		if r.Op == token.AND || r.Op == token.ARROW {
			return true
		}
	case *ast.CallExpr:
		if id, ok := r.Fun.(*ast.Ident); ok && (id.Name == "make" || id.Name == "new") {
			return true
		}
	}
	return mp.isGoExpr(as.Rhs[0])
}

// isGoValue returns whether the given expression, the value of a Go
// definition, is a Go value: a Go expression, or a value that only Go has,
// such as a composite literal.
func (mp *mathParse) isGoValue(ex ast.Expr) bool {
	switch r := ast.Unparen(ex).(type) {
	case *ast.TypeAssertExpr:
		return true
	case *ast.UnaryExpr:
		if r.Op == token.AND || r.Op == token.ARROW {
			return true
		}
	case *ast.CallExpr:
		if id, ok := r.Fun.(*ast.Ident); ok && (id.Name == "make" || id.Name == "new") {
			return true
		}
	}
	g, _ := mp.goExpr(ex)
	return g
}

// assignFuncs are the names of the tmath functions of the assignment
// operators.
var assignFuncs = map[token.Token]string{
	token.ASSIGN:     "Assign",
	token.ADD_ASSIGN: "AddAssign",
	token.SUB_ASSIGN: "SubAssign",
	token.MUL_ASSIGN: "MulAssign",
	token.QUO_ASSIGN: "DivAssign",
	token.REM_ASSIGN: "ModAssign",
}

// assignStmt transpiles an assignment: of Go values as such, to a tensor
// variable or element as a tensor, and to other variables, which can be of
// any kind, with [tensor.SetVar] or [tmath.AddAssignVar] etc.
func (mp *mathParse) assignStmt(as *ast.AssignStmt) {
	fn := assignFuncs[as.Tok]
	if mp.isGoStmt(as) || len(as.Lhs) != 1 || len(as.Rhs) != 1 || fn == "" {
		mp.goStmt(as)
		return
	}
	if _, isIdent := as.Lhs[0].(*ast.Ident); !isIdent && mp.isGoExpr(as.Lhs[0]) { // an element or field of a Go value
		if g, _ := mp.goExpr(as.Rhs[0]); g {
			mp.goStmt(as)
		} else {
			mp.varAssign(as.Lhs[0], as.Tok, as.Rhs[0])
		}
		return
	}
	switch x := as.Lhs[0].(type) {
	case *ast.Ident:
		switch {
		case x.Name == "_":
			mp.copyTo(as.Rhs[0].Pos())
			mp.anyExpr(as.Rhs[0])
			return
		case mp.state.math.isTensor(x.Name):
			if as.Tok == token.ASSIGN {
				mp.copyTo(as.Rhs[0].Pos())
				mp.startFunc("")
				mp.exprList(as.Rhs)
				mp.endFunc()
				return
			}
		default:
			mp.varAssign(x, as.Tok, as.Rhs[0])
			return
		}
	case *ast.SelectorExpr, *ast.StarExpr:
		mp.varAssign(x, as.Tok, as.Rhs[0])
		return
	}
	mp.startFunc("tmath." + fn)
	mp.out.Add(token.LPAREN)
	mp.exprList(as.Lhs)
	mp.out.Add(token.COMMA)
	mp.idx++
	mp.nextArg()
	mp.exprList(as.Rhs)
	mp.out.Add(token.RPAREN)
	mp.endFunc()
}

// varAssign transpiles an assignment of the given operator (or ++ / --
// for a nil rhs) to a variable that can be either a tensor or a Go value,
// with [tensor.SetVar] or [tmath.AddAssignVar] etc.
func (mp *mathParse) varAssign(lhs ast.Expr, op token.Token, rhs ast.Expr) {
	fn := "tensor.SetVar"
	switch op {
	case token.INC:
		fn = "tmath.AddAssignVar"
	case token.DEC:
		fn = "tmath.SubAssignVar"
	case token.ASSIGN:
	default:
		fn = "tmath." + assignFuncs[op] + "Var"
	}
	mp.out.Add(token.IDENT, fn)
	mp.out.Add(token.LPAREN)
	mp.out.Add(token.AND)
	mp.copyNode(lhs)
	mp.out.Add(token.COMMA)
	mp.idx++ // the operator
	if rhs == nil {
		mp.out.Add(token.INT, "1")
	} else {
		mp.anyExpr(rhs)
	}
	mp.out.Add(token.RPAREN)
}

func (mp *mathParse) basicLit(lit *ast.BasicLit) {
	if mp.curArgIsTensor() {
		mp.tensorLit(lit, "")
		return
	}
	mp.out.Add(lit.Kind, lit.Value)
	if mp.trace {
		fmt.Printf("%d\ttok: %s literal\n", mp.idx, lit.Value)
	}
	mp.idx++
}

// tensorLit transpiles a literal as a scalar tensor, with the given sign.
func (mp *mathParse) tensorLit(lit *ast.BasicLit, sign string) {
	switch lit.Kind {
	case token.INT:
		mp.out.Add(token.IDENT, "tensor.NewIntScalar("+sign+lit.Value+")")
		mp.idx++
	case token.FLOAT:
		mp.out.Add(token.IDENT, "tensor.NewFloat64Scalar("+sign+lit.Value+")")
		mp.idx++
	case token.STRING:
		mp.out.Add(token.IDENT, "tensor.NewStringScalar("+lit.Value+")")
		mp.idx++
	default:
		mp.addCur()
	}
}

// funWrap is a function wrapper for simple numpy property / functions
type funWrap struct {
	fun  string // function to call on tensor
	wrap string // code for wrapping function for results of call
}

// nis: NewIntScalar, niv: NewIntFromValues, etc
var numpyProps = map[string]funWrap{
	"ndim":  {"NumDims()", "nis"},
	"len":   {"Len()", "nis"},
	"size":  {"Len()", "nis"},
	"shape": {"Shape().Sizes", "niv"},
	"T":     {"", "tensor.Transpose"},
}

// tensorFunc outputs the wrapping function and whether it needs ellipsis
func (fw *funWrap) wrapFunc(mp *mathParse) bool {
	ellip := false
	wrapFun := fw.wrap
	switch fw.wrap {
	case "nis":
		wrapFun = "tensor.NewIntScalar"
	case "nfs":
		wrapFun = "tensor.NewFloat64Scalar"
	case "nss":
		wrapFun = "tensor.NewStringScalar"
	case "niv":
		wrapFun = "tensor.NewIntFromValues"
		ellip = true
	case "nfv":
		wrapFun = "tensor.NewFloat64FromValues"
		ellip = true
	case "nsv":
		wrapFun = "tensor.NewStringFromValues"
		ellip = true
	default:
		wrapFun = fw.wrap
	}
	mp.startFunc(wrapFun, true) // don't lookup -- don't auto-convert args
	mp.out.Add(token.LPAREN)
	return ellip
}

func (mp *mathParse) selectorExpr(ex *ast.SelectorExpr) {
	fw, ok := numpyProps[ex.Sel.Name]
	if !ok {
		mp.anyExpr(ex.X)
		mp.addToken(token.PERIOD)
		mp.out.Add(token.IDENT, ex.Sel.Name)
		mp.idx++
		return
	}
	ellip := fw.wrapFunc(mp)
	mp.expr(ex.X)
	if fw.fun != "" {
		mp.addToken(token.PERIOD)
		mp.out.Add(token.IDENT, fw.fun)
		mp.idx++
	} else {
		mp.idx += 2
	}
	if ellip {
		mp.out.Add(token.ELLIPSIS)
	}
	mp.out.Add(token.RPAREN)
	mp.endFunc()
}

func (mp *mathParse) indexListExpr(il *ast.IndexListExpr) {
	// fmt.Println("slice expr", se)
}

func (mp *mathParse) indexExpr(il *ast.IndexExpr) {
	if _, ok := il.Index.(*ast.IndexListExpr); ok {
		mp.basicSlicingExpr(il)
	}
}

func (mp *mathParse) basicSlicingExpr(il *ast.IndexExpr) {
	iil := il.Index.(*ast.IndexListExpr)
	fun := "tensor.AnySlice"
	mp.startFunc(fun)
	mp.out.Add(token.LPAREN)
	mp.expr(il.X)
	mp.nextArg()
	mp.addToken(token.COMMA) // use the [ -- can't use ( to preserve X
	mp.exprList(iil.Indices)
	mp.addToken(token.RPAREN) // replaces ]
	mp.endFunc()
}

func (mp *mathParse) sliceExpr(se *ast.SliceExpr) {
	if se.Low == nil && se.High == nil && se.Max == nil {
		mp.out.Add(token.IDENT, "tensor.FullAxis")
		mp.idx++
		return
	}
	mp.out.Add(token.IDENT, "tensor.Slice")
	mp.out.Add(token.LBRACE)
	prev := false
	if se.Low != nil {
		mp.out.Add(token.IDENT, "Start:")
		mp.expr(se.Low)
		prev = true
		if se.High == nil && se.Max == nil {
			mp.idx++
		}
	}
	if se.High != nil {
		if prev {
			mp.out.Add(token.COMMA)
		}
		mp.out.Add(token.IDENT, "Stop:")
		mp.idx++
		mp.expr(se.High)
		prev = true
	}
	if se.Max != nil {
		if prev {
			mp.out.Add(token.COMMA)
		}
		mp.idx++
		if se.Low == nil && se.High == nil {
			mp.idx++
		}
		mp.out.Add(token.IDENT, "Step:")
		mp.expr(se.Max)
	}
	mp.out.Add(token.RBRACE)
}

func (mp *mathParse) arrayLiteral(il *ast.IndexListExpr) {
	kind := inferKindExprList(il.Indices)
	if kind == token.ILLEGAL {
		kind = token.FLOAT // default
	}
	// todo: look for sub-arrays etc.
	typ := "float64"
	fun := "Float64"
	switch kind {
	case token.FLOAT:
	case token.INT:
		typ = "int"
		fun = "Int"
	case token.STRING:
		typ = "string"
		fun = "String"
	}
	if mp.inArray || mp.curArgIsInts() {
		mp.idx++ // opening brace we're not using
		mp.exprList(il.Indices)
		mp.idx++ // closing brace we're not using
		return
	}
	var sh []int
	mp.inArray = true
	mp.arrayShape(il.Indices, &sh)
	if len(sh) > 1 {
		mp.startFunc("tensor.Reshape")
		mp.out.Add(token.LPAREN)
	}
	mp.startFunc("tensor.New" + fun + "FromValues")
	mp.out.Add(token.LPAREN)
	mp.out.Add(token.IDENT, "[]"+typ)
	mp.addToken(token.LBRACE)
	mp.exprList(il.Indices)
	mp.addToken(token.RBRACE)
	mp.out.AddMulti(token.ELLIPSIS, token.RPAREN)
	mp.endFunc()
	if len(sh) > 1 {
		mp.out.Add(token.COMMA)
		nsh := len(sh)
		for i, s := range sh {
			mp.out.Add(token.INT, fmt.Sprintf("%d", s))
			if i < nsh-1 {
				mp.out.Add(token.COMMA)
			}
		}
		mp.out.Add(token.RPAREN)
		mp.endFunc()
	}
	mp.inArray = false
}

func (mp *mathParse) arrayShape(ex []ast.Expr, sh *[]int) {
	n := len(ex)
	if n == 0 {
		return
	}
	*sh = append(*sh, n)
	for i := range n {
		if il, ok := ex[i].(*ast.IndexListExpr); ok {
			mp.arrayShape(il.Indices, sh)
			return
		}
	}
}

// nofun = do not accept a function version, just a method
var numpyFuncs = map[string]funWrap{
	"array":    {"tensor.NewFromValues", ""},
	"zeros":    {"tensor.NewFloat64", ""},
	"full":     {"tensor.NewFloat64Full", ""},
	"ones":     {"tensor.NewFloat64Ones", ""},
	"rand":     {"tensor.NewFloat64Rand", ""},
	"arange":   {"tensor.NewIntRange", ""},
	"linspace": {"tensor.NewFloat64SpacedLinear", ""},
	"reshape":  {"tensor.Reshape", ""},
	"copy":     {"tensor.Clone", ""},
	"get":      {"tensorfs.Get", ""},
	"set":      {"tensorfs.Set", ""},
	"setcp":    {"tensorfs.SetCopy", ""},
	"flatten":  {"tensor.Flatten", "nofun"},
	"squeeze":  {"tensor.Squeeze", "nofun"},
}

func (mp *mathParse) callExpr(ex *ast.CallExpr) {
	if isConversion(ex) {
		mp.conversion(ex)
		return
	}
	switch x := ex.Fun.(type) {
	case *ast.Ident:
		if fw, ok := numpyProps[x.Name]; ok && fw.wrap != "nofun" {
			mp.callPropFun(ex, fw)
			return
		}
		mp.callName(x.Name, "")
	case *ast.SelectorExpr:
		fun := x.Sel.Name
		if fw, ok := numpyFuncs[fun]; ok {
			mp.callPropSelFun(ex, x.X, fw)
			return
		}
		pkg, ok := x.X.(*ast.Ident)
		if !ok {
			mp.callOther(ex)
			return
		}
		mp.callName(fun, pkg.Name)
	default:
		mp.callOther(ex)
		return
	}
	mp.argsList(ex.Args)
	if ex.Ellipsis.IsValid() {
		mp.addCur()
	}
	mp.addToken(token.RPAREN)
	mp.endFunc()
}

// callOther transpiles a call of a function that is not named, such as
// a method of an expression or a function value, whose args can be any
// values.
func (mp *mathParse) callOther(ex *ast.CallExpr) {
	mp.startFunc("", true)
	mp.expr(ex.Fun)
	mp.addToken(token.LPAREN)
	mp.argsList(ex.Args)
	if ex.Ellipsis.IsValid() {
		mp.addCur()
	}
	mp.addToken(token.RPAREN)
	mp.endFunc()
}

// callIsTensor returns whether the given call returns a tensor: a numpy
// style function or property, or a registered tensor function that returns
// one (see [tensor.Funcs]).
func (mp *mathParse) callIsTensor(ex *ast.CallExpr) bool {
	switch x := ex.Fun.(type) {
	case *ast.Ident:
		if isConversion(ex) {
			return false
		}
		if fw, ok := numpyProps[x.Name]; ok && fw.wrap != "nofun" {
			return true
		}
		if _, ok := numpyFuncs[x.Name]; ok {
			return true
		}
		return funcReturnsTensor("tmath."+x.Name) || funcReturnsTensor("tmath."+strings.ToUpper(x.Name[:1])+x.Name[1:])
	case *ast.SelectorExpr:
		if _, ok := numpyFuncs[x.Sel.Name]; ok {
			return true
		}
		if pkg, ok := x.X.(*ast.Ident); ok {
			return funcReturnsTensor(pkg.Name + "." + x.Sel.Name)
		}
	}
	return false
}

// funcReturnsTensor returns whether the registered tensor function of the
// given name returns a tensor.
func funcReturnsTensor(name string) bool {
	fn, err := tensor.FuncByName(name)
	if err != nil {
		return false
	}
	ft := reflect.TypeOf(fn.Fun)
	return ft.NumOut() > 0 && ft.Out(0).Implements(reflect.TypeFor[tensor.Tensor]())
}

// this calls a "prop" function like ndim(a) on the object.
func (mp *mathParse) callPropFun(cf *ast.CallExpr, fw funWrap) {
	ellip := fw.wrapFunc(mp)
	mp.idx += 2
	mp.exprList(cf.Args) // this is the tensor
	mp.addToken(token.PERIOD)
	mp.out.Add(token.IDENT, fw.fun)
	if ellip {
		mp.out.Add(token.ELLIPSIS)
	}
	mp.out.Add(token.RPAREN)
	mp.endFunc()
}

// this calls global function through selector like: a.reshape()
func (mp *mathParse) callPropSelFun(cf *ast.CallExpr, ex ast.Expr, fw funWrap) {
	mp.startFunc(fw.fun)
	mp.out.Add(token.LPAREN) // use the (
	mp.expr(ex)
	mp.idx += 2
	if len(cf.Args) > 0 {
		mp.nextArg() // did first
		mp.addToken(token.COMMA)
		mp.argsList(cf.Args)
	} else {
		mp.idx++
	}
	mp.addToken(token.RPAREN)
	mp.endFunc()
}

// callName starts the call of a named function, of the given package if
// non-empty, up to its opening paren: a numpy style function, a registered
// tensor function (see [tensor.Funcs]), including those of tmath, which
// are used without a package and lowercase, or another Go function.
func (mp *mathParse) callName(funName, pkgName string) {
	if fw, ok := numpyFuncs[funName]; ok {
		mp.startFunc(fw.fun)
		mp.addToken(token.LPAREN) // use the (
		mp.idx++                  // paren too
		return
	}
	nameToks := 1
	name := funName
	var err error // validate name
	if pkgName != "" {
		nameToks = 3
		name = pkgName + "." + funName
		_, err = tensor.FuncByName(name)
	} else { // non-package qualified names are _only_ in tmath! can be lowercase
		_, err = tensor.FuncByName("tmath." + funName)
		if err != nil {
			upName := strings.ToUpper(funName[:1]) + funName[1:] // first letter uppercased
			if _, err = tensor.FuncByName("tmath." + upName); err == nil {
				funName = upName
			}
		}
		if err == nil { // registered, must be in tmath
			name = "tmath." + funName
		}
	}
	mp.startFunc(name)
	if err != nil { // not a registered tensor function: no arg conversions
		mp.funcs.Peek().Args = nil
	}
	mp.idx += nameToks
	mp.addToken(token.LPAREN)
}

// basic ident replacements
var consts = map[string]string{
	"newaxis": "tensor.NewAxis",
	"pi":      "tensor.NewFloat64Scalar(math.Pi)",
}

func (mp *mathParse) ident(id *ast.Ident) {
	if id == nil {
		return
	}
	if cn, ok := consts[id.Name]; ok {
		mp.out.Add(token.IDENT, cn)
		mp.idx++
		return
	}
	mp.addCur()
}

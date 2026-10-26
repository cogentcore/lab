// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"go/token"
	"slices"
	"strings"
)

// mathNames is what math mode knows about the variables and blocks of the
// code so far, from its syntax only, as math mode transpiles one line at a
// time without type information.
//
// A variable is known to be a plain Go value if math mode declared it in a
// way that makes it one: the variables of a for loop header, var
// declarations of basic types, := of a conversion like int(x) or of an
// expression of other Go values, and basic-typed function parameters.
// Operators on Go values stay Go operators. A variable is known to be a
// tensor if math mode defined it with := as one. Other variables, such as
// those of Go code outside of math mode, can be either, and are converted
// at run time (see [tensor.ToTensor] and the other To functions).
//
// Each name has the brace depth of the block that it is declared in, and
// is forgotten at the end of that block.
type mathNames struct {
	names []mathName

	// switches are the switch statements whose blocks are open.
	switches []mathSwitch
}

// mathKind is the kind of value of a variable, as math mode knows it.
type mathKind int

const (
	// mathUnknown is a variable that can be a tensor or a Go value.
	mathUnknown mathKind = iota

	// mathGo is a plain Go value.
	mathGo

	// mathTensor is a tensor.
	mathTensor
)

// mathName is a variable name and its kind of value.
type mathName struct {
	name  string
	depth int
	kind  mathKind

	// typ is the Go basic type of a Go variable, if known, such as int.
	typ string
}

// mathSwitch is a switch statement whose block is open.
type mathSwitch struct {
	depth int

	// tagless is whether it has no tag, so that its cases are conditions.
	tagless bool
}

// endBlocks forgets the names and switches of the blocks that are deeper
// than the given brace depth, which have ended.
func (mn *mathNames) endBlocks(depth int) {
	mn.names = slices.DeleteFunc(mn.names, func(n mathName) bool { return n.depth > depth })
	mn.switches = slices.DeleteFunc(mn.switches, func(s mathSwitch) bool { return s.depth > depth })
}

// lookup returns the innermost declaration of the given name, which is
// of unknown kind for one that math mode has not seen declared.
func (mn *mathNames) lookup(name string) mathName {
	for i := len(mn.names) - 1; i >= 0; i-- {
		if mn.names[i].name == name {
			return mn.names[i]
		}
	}
	return mathName{name: name}
}

// kind returns the kind of the innermost declaration of the given name.
func (mn *mathNames) kind(name string) mathKind { return mn.lookup(name).kind }

// isGo returns whether the given name is a known Go variable.
func (mn *mathNames) isGo(name string) bool { return mn.kind(name) == mathGo }

// isTensor returns whether the given name is a known tensor variable.
func (mn *mathNames) isTensor(name string) bool { return mn.kind(name) == mathTensor }

// declare records the given name as declared at the given depth with the
// given kind, and Go type if known, replacing a declaration at the same
// depth. Declaring it as unknown hides a declaration of an outer block.
func (mn *mathNames) declare(name string, depth int, kind mathKind, typ string) {
	if name == "_" {
		return
	}
	mn.names = slices.DeleteFunc(mn.names, func(n mathName) bool { return n.name == name && n.depth == depth })
	mn.names = append(mn.names, mathName{name: name, depth: depth, kind: kind, typ: typ})
}

// addSwitch records a switch whose block is at the given depth.
func (mn *mathNames) addSwitch(depth int, tagless bool) {
	mn.switches = append(mn.switches, mathSwitch{depth: depth, tagless: tagless})
}

// taglessSwitch returns whether the switch whose block is at the given
// depth has no tag.
func (mn *mathNames) taglessSwitch(depth int) bool {
	for i := len(mn.switches) - 1; i >= 0; i-- {
		if mn.switches[i].depth == depth {
			return mn.switches[i].tagless
		}
	}
	return false
}

// forgetDefines records the names that the given line of Go code, outside
// of math mode, defines with :=, as unknown, as they can be of any kind,
// except for a name defined as a literal, such as n := 10, which is Go.
func (st *State) forgetDefines(toks Tokens) {
	for i, t := range toks {
		if t.Tok == token.DEFINE {
			kind, typ := mathUnknown, ""
			if i == 1 {
				if lt := literalType(toks[2:]); lt != "" {
					kind, typ = mathGo, lt
				}
			}
			for _, nt := range toks[:i] {
				if nt.Tok == token.IDENT {
					st.math.declare(nt.Str, st.BraceDepth, kind, typ)
				}
			}
			return
		}
		if t.Tok != token.IDENT && t.Tok != token.COMMA {
			return
		}
	}
}

// literalType returns the Go type of the value of the given tokens if they
// are a number, string or char literal, with an optional sign and comment,
// and otherwise "".
func literalType(toks Tokens) string {
	if len(toks) > 0 && (toks[0].Tok == token.SUB || toks[0].Tok == token.ADD) {
		toks = toks[1:]
	}
	if len(toks) == 2 && toks[1].Tok == token.COMMENT {
		toks = toks[:1]
	}
	if len(toks) != 1 {
		return ""
	}
	return defaultType(literalTypes[toks[0].Tok])
}

// literalTypes are the untyped types of the literals of each token.
var literalTypes = map[token.Token]string{
	token.INT:    "untyped int",
	token.FLOAT:  "untyped float",
	token.STRING: "untyped string",
	token.CHAR:   "untyped rune",
}

// defaultType returns the default type of an untyped type, which is the
// type of a variable defined as such a value, and other types as such.
func defaultType(typ string) string {
	switch typ {
	case "untyped int":
		return "int"
	case "untyped float":
		return "float64"
	case "untyped string":
		return "string"
	case "untyped rune":
		return "rune"
	case "untyped bool":
		return "bool"
	}
	return typ
}

// isUntyped returns whether the given type is that of an untyped constant.
func isUntyped(typ string) bool { return strings.HasPrefix(typ, "untyped ") }

// isFloatType returns whether the given type is a float type.
func isFloatType(typ string) bool { return typ == "float64" || typ == "float32" }

// isNumberType returns whether the given type is a number type.
func isNumberType(typ string) bool {
	return basicTypes[typ] && typ != "bool" && typ != "string"
}

// commonType returns the type of the result of a Go arithmetic operator on
// values of the given types, and whether Go allows the operator on them,
// as far as math mode knows: types that it does not know ("") are taken
// to be allowed, and untyped constants are converted to the other type,
// to which a float constant can only be converted if it is a float type.
func commonType(a, b string) (string, bool) {
	switch {
	case a == b:
		return a, true
	case a == "" || b == "":
		return "", true
	case isUntyped(a) && isUntyped(b): // different untyped constants: numbers only
		if !isNumberType(defaultType(a)) || !isNumberType(defaultType(b)) {
			return "", false
		}
		if a == "untyped float" || b == "untyped float" {
			return "untyped float", true
		}
		return a, true
	case isUntyped(b):
		return a, untypedFits(b, a)
	case isUntyped(a):
		return b, untypedFits(a, b)
	}
	return "", false
}

// untypedFits returns whether a constant of the given untyped type can be
// converted to the given type.
func untypedFits(untyped, typ string) bool {
	switch untyped {
	case "untyped int", "untyped rune":
		return isNumberType(typ)
	case "untyped float":
		return isFloatType(typ)
	case "untyped string":
		return typ == "string"
	case "untyped bool":
		return typ == "bool"
	}
	return false
}

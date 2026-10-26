// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"go/ast"
	"go/token"
)

// kindBool is the kind of the bool constants, true and false, which are
// identifiers rather than literals.
const kindBool = token.IDENT

// inferKindExpr infers the basic Kind level type from given expression
func inferKindExpr(ex ast.Expr) token.Token {
	if ex == nil {
		return token.ILLEGAL
	}
	switch x := ex.(type) {
	case *ast.BadExpr:
		return token.ILLEGAL

	case *ast.Ident:
		// todo: get type of object is not possible!
		if x.Name == "true" || x.Name == "false" {
			return kindBool
		}

	case *ast.UnaryExpr:
		return inferKindExpr(x.X)

	case *ast.BinaryExpr:
		ta := inferKindExpr(x.X)
		tb := inferKindExpr(x.Y)
		if ta == tb {
			return ta
		}
		if ta != token.ILLEGAL {
			return ta
		} else {
			return tb
		}

	case *ast.BasicLit:
		return x.Kind // key grounding

	case *ast.FuncLit:

	case *ast.ParenExpr:
		return inferKindExpr(x.X)

	case *ast.SelectorExpr:

	case *ast.TypeAssertExpr:

	case *ast.IndexListExpr:
		if x.X == nil { // array literal
			return inferKindExprList(x.Indices)
		} else {
			return inferKindExpr(x.X)
		}

	case *ast.SliceExpr:

	case *ast.CallExpr:

	}
	return token.ILLEGAL
}

// inferKindExprList infers the basic Kind level type of the given
// expressions, such as the elements of an array literal: the first one
// that is known, except that float takes precedence over int.
func inferKindExprList(ex []ast.Expr) token.Token {
	kind := token.ILLEGAL
	for _, x := range ex {
		t := inferKindExpr(x)
		switch {
		case t == token.FLOAT && kind == token.INT:
			kind = t
		case kind == token.ILLEGAL:
			kind = t
		}
	}
	return kind
}

// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"testing"

	_ "cogentcore.org/lab/stats/metric"
	_ "cogentcore.org/lab/stats/stats"
	_ "cogentcore.org/lab/tensor/tmath"
	"github.com/stretchr/testify/assert"
)

type exIn struct {
	i string
	e string
}

type wexIn struct {
	i     string
	isErr bool
	e     []string
}

// these are more general tests of full-line statements of various forms
func TestExecWords(t *testing.T) {
	tests := []wexIn{
		{`ls`, false, []string{`ls`}},
		{`cat "be"`, false, []string{`cat`, `"be"`}},
		{`cat "be`, true, []string{`cat`, `"be`}},
		{`cat "be a thing"`, false, []string{`cat`, `"be a thing"`}},
		{`cat "{be \"a\" thing}"`, false, []string{`cat`, `"{be \"a\" thing}"`}},
		{`cat {vals[1:10]}`, false, []string{`cat`, `{`, `vals[1:10]`, `}`}},
		{`cat {myfunc(vals[1:10], "test", false)}`, false, []string{`cat`, `{`, `myfunc(vals[1:10],"test",false)`, `}`}},
		{`cat vals[1:10]`, false, []string{`cat`, `vals[1:10]`}},
		{`cat vals...`, false, []string{`cat`, `vals...`}},
		{`[cat vals...]`, false, []string{`[`, `cat`, `vals...`, `]`}},
		{`[cat vals...]; ls *.tsv`, false, []string{`[`, `cat`, `vals...`, `]`, `;`, `ls`, `*.tsv`}},
		{`cat vals... | grep -v "b"`, false, []string{`cat`, `vals...`, `|`, `grep`, `-v`, `"b"`}},
		{`cat vals...>&file.out`, false, []string{`cat`, `vals...`, `>&`, `file.out`}},
		{`cat vals...>&@0:file.out`, false, []string{`cat`, `vals...`, `>&`, `@0:file.out`}},
		{`./"Cogent Code"`, false, []string{`./"Cogent Code"`}},
		{`Cogent\ Code`, false, []string{`Cogent Code`}},
		{`./Cogent\ Code`, false, []string{`./Cogent Code`}},
		// comments have no words, and their quotes are not quotes
		{`mlr = 1 // don't use sig deriv`, false, []string{`mlr`, `=`, `1`}},
		{`// the neuron's index`, false, nil},
		{`ls // the "a`, false, []string{`ls`}},
		{`curl http://example.com/a`, false, []string{`curl`, `http://example.com/a`}},
		{`echo "// don't"`, false, []string{`echo`, `"// don't"`}},
	}
	for _, test := range tests {
		o, err := ExecWords(test.i)
		assert.Equal(t, test.e, o)
		if err != nil {
			if !test.isErr {
				t.Error("should not have been an error:", test.i)
			}
		} else if test.isErr {
			t.Error("was supposed to be an error:", test.i)
		}
	}
}

// Paths tests the Path() code
func TestPaths(t *testing.T) {
	// logx.UserLevel = slog.LevelDebug
	tests := []exIn{
		{`fmt.Println("hi")`, `fmt.Println`},
		{"./goal -i", `./goal`},
		{"main.go", `main.go`},
		{"cogent/", `cogent/`},
		{`./"Cogent Code"`, `./\"Cogent Code\"`},
		{`Cogent\ Code`, ``},
		{`./Cogent\ Code`, `./Cogent Code`},
		{"./ios-deploy", `./ios-deploy`},
		{"ios_deploy/sub", `ios_deploy/sub`},
		{"C:/ios_deploy/sub", `C:/ios_deploy/sub`},
		{"..", `..`},
		{"../another/dir/to/go_to", `../another/dir/to/go_to`},
		{"../an-other/dir/", `../an-other/dir/`},
		{"https://google.com/search?q=hello%20world#body", `https://google.com/search?q=hello%20world#body`},
	}
	for _, test := range tests {
		toks := TokensFromString(test.i)
		p, _ := toks.Path(false)
		assert.Equal(t, test.e, p)
	}
}

// these are more general tests of full-line statements of various forms
func TestTranspile(t *testing.T) {
	// logx.UserLevel = slog.LevelDebug
	tests := []exIn{
		{"ls", `goalrun.Run("ls")`},
		{"mlr = 1 // don't use sig deriv", `mlr = 1 // don't use sig deriv`},
		{"$ls -la$", `goalrun.Run("ls", "-la")`},
		{"ls -la", `goalrun.Run("ls", "-la")`},
		{"chmod +x file", `goalrun.Run("chmod", "+x", "file")`},
		{"ls --help", `goalrun.Run("ls", "--help")`},
		{"ls go", `goalrun.Run("ls", "go")`},
		{"cd go", `goalrun.Run("cd", "go")`},
		{`var name string`, `var name string`},
		{`name = "test"`, `name = "test"`},
		{`echo {name}`, `goalrun.Run("echo", name)`},
		{`echo "testing"`, `goalrun.Run("echo", "testing")`},
		{`number := 1.23`, `number := 1.23`},
		{`res1, res2 := FunTwoRet()`, `res1, res2 := FunTwoRet()`},
		{`res1, res2, res3 := FunThreeRet()`, `res1, res2, res3 := FunThreeRet()`},
		{`println("hi")`, `println("hi")`},
		{`fmt.Println("hi")`, `fmt.Println("hi")`},
		{`for i := 0; i < 3; i++ { fmt.Println(i, "\n")`, `for i := 0; i < 3; i++ { fmt.Println(i, "\n")`},
		{"for i, v := range $ls -la$ {", `for i, v := range goalrun.Output("ls", "-la") {`},
		{"x := $ls | grep go$", `x := func()string { goalrun.Start("ls", "|"); return goalrun.Output("grep", "go") } ()`},
		{"x := $cd sub; ls$", `x := func()string { goalrun.Output("cd", "sub"); return goalrun.Output("ls") } ()`},
		{`// todo: fixit`, `// todo: fixit`},
		{"$go build$", `goalrun.Run("go", "build")`},
		{"{go build()}", `go build()`},
		{"go build", `goalrun.Run("go", "build")`},
		{"go build()", `go build()`},
		{"go build &", `goalrun.Start("go", "build")`},
		{"[mkdir subdir]", `goalrun.RunErrOK("mkdir", "subdir")`},
		{"set something hello-1", `goalrun.Run("set", "something", "hello-1")`},
		{"set something = hello", `goalrun.Run("set", "something", "=", "hello")`},
		{`set something = "hello"`, `goalrun.Run("set", "something", "=", "hello")`},
		{`set something=hello`, `goalrun.Run("set", "something=hello")`},
		{`set "something=hello"`, `goalrun.Run("set", "something=hello")`},
		{`set something="hello"`, `goalrun.Run("set", "something=\"hello\"")`},
		{`add-path /opt/sbin /opt/homebrew/bin`, `goalrun.Run("add-path", "/opt/sbin", "/opt/homebrew/bin")`},
		{`cat file > test.out`, `goalrun.Run("cat", "file", ">", "test.out")`},
		{`cat file | grep -v exe > test.out`, `goalrun.Start("cat", "file", "|"); goalrun.Run("grep", "-v", "exe", ">", "test.out")`},
		{`cd sub; pwd; ls -la`, `goalrun.Run("cd", "sub"); goalrun.Run("pwd"); goalrun.Run("ls", "-la")`},
		{`cd sub; [mkdir sub]; ls -la`, `goalrun.Run("cd", "sub"); goalrun.RunErrOK("mkdir", "sub"); goalrun.Run("ls", "-la")`},
		{`cd sub; mkdir names[4]`, `goalrun.Run("cd", "sub"); goalrun.Run("mkdir", "names[4]")`},
		{"ls -la > test.out", `goalrun.Run("ls", "-la", ">", "test.out")`},
		{"ls > test.out", `goalrun.Run("ls", ">", "test.out")`},
		{"ls -la >test.out", `goalrun.Run("ls", "-la", ">", "test.out")`},
		{"ls -la >> test.out", `goalrun.Run("ls", "-la", ">>", "test.out")`},
		{"ls -la >& test.out", `goalrun.Run("ls", "-la", ">&", "test.out")`},
		{"ls -la >>& test.out", `goalrun.Run("ls", "-la", ">>&", "test.out")`},
		{"ls | grep ev", `goalrun.Start("ls", "|"); goalrun.Run("grep", "ev")`},
		{"@1 ls -la", `goalrun.Run("@1", "ls", "-la")`},
		{"git switch main", `goalrun.Run("git", "switch", "main")`},
		{"git checkout 123abc", `goalrun.Run("git", "checkout", "123abc")`},
		{"go get cogentcore.org/core@main", `goalrun.Run("go", "get", "cogentcore.org/core@main")`},
		{"ls *.go", `goalrun.Run("ls", "*.go")`},
		{"ls ??.go", `goalrun.Run("ls", "??.go")`},
		{`fmt.Println("hi")`, `fmt.Println("hi")`},
		{"goal -i", `goalrun.Run("goal", "-i")`},
		{"./goal -i", `goalrun.Run("./goal", "-i")`},
		{"cat main.go", `goalrun.Run("cat", "main.go")`},
		{"cd cogent", `goalrun.Run("cd", "cogent")`},
		{"cd cogent/", `goalrun.Run("cd", "cogent/")`},
		{"echo $PATH", `goalrun.Run("echo", "$PATH")`},
		{`ls My\ Documents`, `goalrun.Run("ls", "My Documents")`},
		{`ls a\{b\}\\c`, `goalrun.Run("ls", "a{b}\\c")`},
		{`ls 'My Documents'`, `goalrun.Run("ls", "My Documents")`},
		{`ls 'a\b"c'`, `goalrun.Run("ls", "a\\b\"c")`},
		{`ls "a\"b"`, `goalrun.Run("ls", "a\"b")`},
		{`"./Cogent Code"`, `goalrun.Run("./Cogent Code")`},
		{`./"Cogent Code"`, `goalrun.Run("./\"Cogent Code\"")`},
		{`Cogent\ Code`, `goalrun.Run("Cogent Code")`},
		{`./Cogent\ Code`, `goalrun.Run("./Cogent Code")`},
		{`ios\ deploy -i`, `goalrun.Run("ios deploy", "-i")`},
		{"./ios-deploy -i", `goalrun.Run("./ios-deploy", "-i")`},
		{"ios_deploy -i tree_file", `goalrun.Run("ios_deploy", "-i", "tree_file")`},
		{"ios_deploy/sub -i tree_file", `goalrun.Run("ios_deploy/sub", "-i", "tree_file")`},
		{"C:/ios_deploy/sub -i tree_file", `goalrun.Run("C:/ios_deploy/sub", "-i", "tree_file")`},
		{"ios_deploy -i tree_file/path", `goalrun.Run("ios_deploy", "-i", "tree_file/path")`},
		{"ios-deploy -i", `goalrun.Run("ios-deploy", "-i")`},
		{"ios-deploy -i tree-file", `goalrun.Run("ios-deploy", "-i", "tree-file")`},
		{"ios-deploy -i tree-file/path/here", `goalrun.Run("ios-deploy", "-i", "tree-file/path/here")`},
		{"cd ..", `goalrun.Run("cd", "..")`},
		{"cd ../another/dir/to/go_to", `goalrun.Run("cd", "../another/dir/to/go_to")`},
		{"cd ../an-other/dir/", `goalrun.Run("cd", "../an-other/dir/")`},
		{"curl https://google.com/search?q=hello%20world#body", `goalrun.Run("curl", "https://google.com/search?q=hello%20world#body")`},
		{"func splitLines(str string) []string {", `splitLines := func(str string)[]string {`},
		{"type Result struct {", `type Result struct {`},
		{"var Jobs *table.Table", `var Jobs *table.Table`},
		{"type Result struct { JobID string", `type Result struct { JobID string`},
		{"type Result struct { JobID string `width:\"60\"`", "type Result struct { JobID string `width:\"60\"`"},
		{"func RunInExamples(fun func()) {", "RunInExamples := func(fun func()) {"},
		{"ctr++", "ctr++"},
		{"ctr--", "ctr--"},
		{"stru.ctr++", "stru.ctr++"},
		{"stru.ctr--", "stru.ctr--"},
		{"meta += ln", "meta += ln"},
		{"var data map[string]any", "var data map[string]any"},
		// non-math-mode tensor indexing, with fixed-arity methods up to 5D:
		{"x = a[1,f(2,3)]", `x = a.Value2D(int(1), int(f(2, 3)))`},
		{"x = a[1]", `x = a[1]`},
		{"x = a[f(2,3)]", `x = a[f(2, 3)]`},
		{"a[1,2] = 55", `a.Set2D(55, int(1), int(2))`},
		{"a[1,2] = 55 // and that is good", `a.Set2D(55, int(1), int(2)) // and that is good`},
		{"a[1,2] += f(2,55)", `a.SetAdd2D(f(2, 55), int(1), int(2))`},
		{"a[1,2] *= f(2,55)", `a.SetMul2D(f(2, 55), int(1), int(2))`},
		{"a[1,2] -= b[3,4]", `a.SetSub2D(b.Value2D(int(3), int(4)), int(1), int(2))`},
		{"Data[idx, Integ] = integ", `Data.Set2D(integ, int(idx), int(Integ))`},
		{"Data[Idxs[idx, 25], Integ] = integ", `Data.Set2D(integ, int(Idxs.Value2D(int(idx), int(25))), int(Integ))`},
		{"Layers[NeuronIxs[NrnLayIndex, ni]].GatherSpikes(&Ctx[0], ni, di)", `Layers[NeuronIxs.Value2D(int(NrnLayIndex), int(ni))].GatherSpikes( & Ctx[0], ni, di)`},
		{"vm := Neurons[ni, di, Vm]", `vm := Neurons.Value3D(int(ni), int(di), int(Vm))`},
		{"atomic.AddInt32(&PoolsInt[pi, di, FBsRawInt], v)", `atomic.AddInt32(PoolsInt.ValuePtr3D(int(pi), int(di), int(FBsRawInt)), v)`},
		{"f(&ss.Net.Pools[pi, di])", `f(ss.Net.Pools.ValuePtr2D(int(pi), int(di)))`},
		{"x = ss.Net.Neurons[ni, di]", `x = ss.Net.Neurons.Value2D(int(ni), int(di))`},
		{"x = ls[li].Pools[pi, di]", `x = ls[li].Pools.Value2D(int(pi), int(di))`},
		{"x = a[1,2,3,4,5]", `x = a.Value5D(int(1), int(2), int(3), int(4), int(5))`},
		{"x = a[1,2,3,4,5,6]", `x = a.Value(int(1), int(2), int(3), int(4), int(5), int(6))`},
		{"a[1,2,3,4,5,6] /= 2", `a.SetDiv(2, int(1), int(2), int(3), int(4), int(5), int(6))`},
	}

	st := NewState()
	for _, test := range tests {
		o := st.TranspileLine(test.i)
		assert.Equal(t, test.e, o)
	}
}

// tests command generation
func TestCommand(t *testing.T) {
	// logx.UserLevel = slog.LevelDebug
	tests := []exIn{
		/*		{
							`command list {
								ls -la args...
								}`,
							`goalrun.AddCommand("list", func(args ...string) {
				goalrun.Run("ls", "-la", "args...")
				})`},
						{
							`	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{
						Label: "Reset RunLog",
					})
				`,
							`ss.GUI.AddToolbarItem(p, egui.ToolbarItem {
				Label: "Reset RunLog",
				} )
				`},
		*/
		{
			`#!/usr/bin/env goal
command build {
`,
			`
goalrun.AddCommand("build", func(args ...string) {
`},
	}

	for _, test := range tests {
		st := NewState()
		st.TranspileCode(test.i)
		o := st.Code()
		assert.Equal(t, test.e, o)
	}
}

// Use this for testing the current thing working on.
func TestCur(t *testing.T) {
	// logx.UserLevel = slog.LevelDebug
	tests := []exIn{
		{"# totalTime := 100", `totalTime := tensor.Tensor(tensor.NewIntScalar(100))`},
		{"# driver := zeros(100)", `driver := tensor.Tensor(tensor.NewFloat64(100))`},
		// {"type Servers keylist.List[int, *Server]", `type Servers keylist.List[int, *Server]`},
		// {"`string literal over lines todo add \n`", `todo`},
	}
	st := NewState()
	st.MathRecord = false
	for _, test := range tests {
		o := st.TranspileLine(test.i)
		assert.Equal(t, test.e, o)
	}
}

// Use this for testing the current thing working on -- multi-line code.
func TestCurCode(t *testing.T) {
	// logx.UserLevel = slog.LevelDebug
	tests := []exIn{
		{`	b := 5
		//gosl:wgsl
	// a := 2
	//gosl:end`,
			`b := 5
//gosl:wgsl
// a := 2
//gosl:end`},
	}
	for _, test := range tests {
		st := NewState()
		st.MathRecord = false
		st.TranspileCode(test.i)
		o := st.Code()
		assert.Equal(t, test.e, o)
	}
}

func TestMath(t *testing.T) {
	// logx.UserLevel = slog.LevelDebug
	tests := []exIn{
		{"# x := 1", `x := tensor.Tensor(tensor.NewIntScalar(1))`},
		{"# x := a + 1", `x := tensor.Tensor(tmath.Add(tensor.ToTensor(a), tensor.NewIntScalar(1)))`},
		{"# x = x * 4", `x = tmath.Mul(x, tensor.NewIntScalar(4))`},
		{"# a = x + y", `tensor.SetVar( & a, tmath.Add(x, tensor.ToTensor(y)))`},
		{"# a := x ** 2", `a := tensor.Tensor(tmath.Pow(x, tensor.NewIntScalar(2)))`},
		{"# a = -x", `a = tmath.Negate(x)`},
		{"# x @ a", `matrix.Mul(x, a)`},
		{"# x += 1", `tmath.AddAssign(x, tensor.NewIntScalar(1))`},
		{"# a := [1,2,3,4]", `a := tensor.Tensor(tensor.NewIntFromValues([]int { 1, 2, 3, 4 }  ...))`},
		{"# a := [1.,2,3,4]", `a := tensor.Tensor(tensor.NewFloat64FromValues([]float64 { 1., 2, 3, 4 }  ...))`},
		{"# a := [[1,2], [3,4]]", `a := tensor.Tensor(tensor.Reshape(tensor.NewIntFromValues([]int { 1, 2, 3, 4 }  ...), 2, 2))`},
		{`# U := [[1., 2.], [3., 4.]]`, `U := tensor.Tensor(tensor.Reshape(tensor.NewFloat64FromValues([]float64 { 1., 2., 3., 4. }  ...), 2, 2))`},
		{"# a.ndim", `tensor.NewIntScalar(a.NumDims())`},
		{"# ndim(a)", `tensor.NewIntScalar(a.NumDims())`},
		{"# x.shape", `tensor.NewIntFromValues(x.Shape().Sizes ...)`},
		{"# x.T", `tensor.Transpose(x)`},
		{"# zeros(3, 4)", `tensor.NewFloat64(3, 4)`},
		{"# full(5.5, 3, 4)", `tensor.NewFloat64Full(5.5, 3, 4)`},
		{"# zeros(sh)", `tensor.NewFloat64(tensor.ToInts(sh) ...)`},
		{"# arange(36)", `tensor.NewIntRange(36)`},
		{"# arange(36, 0, -1)", `tensor.NewIntRange(36, 0, -1)`},
		{"# linspace(0, 5, 6, true)", `tensor.NewFloat64SpacedLinear(tensor.NewIntScalar(0), tensor.NewIntScalar(5), 6, true)`},
		{"# reshape(x, 6, 6)", `tensor.Reshape(x, 6, 6)`},
		{"# reshape(x, [6, 6])", `tensor.Reshape(x, 6, 6)`},
		{"# reshape(x, sh)", `tensor.Reshape(x, tensor.ToInts(sh) ...)`},
		{"# reshape(arange(36), 6, 6)", `tensor.Reshape(tensor.NewIntRange(36), 6, 6)`},
		{"# a.reshape(6, 6)", `tensor.Reshape(a, 6, 6)`},
		{"# a[1, 2]", `tensor.AnySlice(a, 1, 2)`},
		{"# a[:, 2]", `tensor.AnySlice(a, tensor.FullAxis, 2)`},
		{"# a[1:3:1, 2]", `tensor.AnySlice(a, tensor.Slice { Start:1, Stop:3, Step:1 } , 2)`},
		{"# a[::-1, 2]", `tensor.AnySlice(a, tensor.Slice { Step:-1 } , 2)`},
		{"# a[:3, 2]", `tensor.AnySlice(a, tensor.Slice { Stop:3 } , 2)`},
		{"# a[2:, 2]", `tensor.AnySlice(a, tensor.Slice { Start:2 } , 2)`},
		{"# a[2:, 2, newaxis]", `tensor.AnySlice(a, tensor.Slice { Start:2 } , 2, tensor.NewAxis)`},
		{"# a[..., 2:]", `tensor.AnySlice(a, tensor.Ellipsis, tensor.Slice { Start:2 } )`},
		{"# a[:, 2] = b", `tmath.Assign(tensor.AnySlice(a, tensor.FullAxis, 2), tensor.ToTensor(b))`},
		{"# a[:, 2] += b", `tmath.AddAssign(tensor.AnySlice(a, tensor.FullAxis, 2), tensor.ToTensor(b))`},
		{"# cos(a)", `tmath.Cos(a)`},
		{"# stats.Mean(a)", `stats.Mean(a)`},
		{"# (stats.Mean(a))", `(stats.Mean(a))`},
		{"# stats.Mean(reshape(a,36))", `stats.Mean(tensor.Reshape(a, 36))`},
		{"# z = a[1:5,1:5] - stats.Mean(ra)", `tensor.SetVar( & z, tmath.Sub(tensor.AnySlice(a, tensor.Slice { Start:1, Stop:5 } , tensor.Slice { Start:1, Stop:5 } ), stats.Mean(tensor.ToTensor(ra))))`},
		{"# metric.Matrix(metric.Cosine, a)", `metric.Matrix(metric.Cosine, a)`},
		{"# a > 5", `tmath.Greater(a, tensor.NewIntScalar(5))`},
		{"# !a", `tmath.Not(a)`},
		{"# a[a > 5]", `tensor.AnySlice(a, tmath.Greater(a, tensor.NewIntScalar(5)))`},
		{"# a[a > 5].flatten()", `tensor.Flatten(tensor.AnySlice(a, tmath.Greater(a, tensor.NewIntScalar(5))))`},
		{"# a[:3, 2].copy()", `tensor.Clone(tensor.AnySlice(a, tensor.Slice { Stop:3 } , 2))`},
		{"# a[:3, 2].reshape(4,2)", `tensor.Reshape(tensor.AnySlice(a, tensor.Slice { Stop:3 } , 2), 4, 2)`},
		{"# a > 5 || a < 1", `tmath.Or(tmath.Greater(a, tensor.NewIntScalar(5)), tmath.Less(a, tensor.NewIntScalar(1)))`},
		{"# fmt.Println(a)", `fmt.Println(a)`},
		{"# }", `}`},
		{"# if a[1,2] == 2 {", `if tensor.ToBool(tmath.Equal(tensor.AnySlice(tensor.ToTensor(a), 1, 2), tensor.NewIntScalar(2))) {`},
		{"# for i := 0; i < 3; i++ {", `for i := 0; i < 3; i++ {`},
		{"# for i, v := range a {", `for i, v := range tensor.Iter(a) {`},
		{`# x := get("item")`, `x := tensor.Tensor(tensorfs.Get("item"))`},
		{`# set("item", x)`, `tensorfs.Set("item", x)`},
		{`# set("item", 5)`, `tensorfs.Set("item", tensor.NewIntScalar(5))`},
		{`fmt.Println(#zeros(3,4)#)`, `fmt.Println(tensor.NewFloat64(3, 4))`},
		{"# totalTime := 100", `totalTime := tensor.Tensor(tensor.NewIntScalar(100))`},
		{"# driver := zeros(100)", `driver := tensor.Tensor(tensor.NewFloat64(100))`},
	}

	st := NewState()
	st.MathRecord = false
	for _, test := range tests {
		o := st.TranspileLine(test.i)
		assert.Equal(t, test.e, o)
	}
}

// TestMathGo tests math mode with plain Go code and Go values, in a
// sequence of lines, as math mode knows the kinds of the variables
// that the lines before declare: n and tau are Go, x is unknown,
// and a is a tensor.
func TestMathGo(t *testing.T) {
	tests := []exIn{
		{"n := 10", `n := 10`},
		{"tau := 0.5", `tau := 0.5`},
		{"x := f()", `x := f()`},
		{"##", ``},
		{"var y float64", `var y float64`},
		{"a := zeros(n)", `a := tensor.Tensor(tensor.NewFloat64(n))`},
		{"y = a[0] * tau", `tensor.SetVar( & y, tmath.Mul(tensor.AnySlice(a, 0), tensor.ToTensor(tau)))`},
		{"y += tau", `y += tau`},
		{"z := a * tau", `z := tensor.Tensor(tmath.Mul(a, tensor.ToTensor(tau)))`},
		{"c := a * x", `c := tensor.Tensor(tmath.Mul(a, tensor.ToTensor(x)))`},
		{"for i := range n {", `for i := range n {`},
		{"a[i] = i * tau", `tmath.Assign(tensor.AnySlice(a, i), tmath.Mul(tensor.ToTensor(i), tensor.ToTensor(tau)))`},
		{"a[i] = float64(i) * tau", `tmath.Assign(tensor.AnySlice(a, i), tensor.ToTensor(float64(i) *tau))`},
		{"j := i * 2 + 1", `j := i *2 + 1`},
		{"g := i * 0.5", `g := tensor.Tensor(tmath.Mul(tensor.ToTensor(i), tensor.NewFloat64Scalar(0.5)))`},
		{"y += i", `tmath.AddAssignVar( & y, i)`},
		{"y += tau * 2", `y += tau *2`},
		{"if i < j {", `if i < j {`},
		{"}", `}`},
		{"if i < y {", `if tensor.ToBool(tmath.Less(tensor.ToTensor(i), tensor.ToTensor(y))) {`},
		{"}", `}`},
		{"}", `}`},
		{"for i, v := range a {", `for i, v := range tensor.Iter(a) {`},
		{"y += v", `y += v`},
		{"}", `}`},
		{"for i := 0; i < n; i++ {", `for i := 0; i < n; i++ {`},
		{"}", `}`},
		{"for y < 10 {", `for y < 10 {`},
		{"}", `}`},
		{"for a[0] < 10 {", `for tensor.ToBool(tmath.Less(tensor.AnySlice(a, 0), tensor.NewIntScalar(10))) {`},
		{"}", `}`},
		{"switch {", `switch {`},
		{"case a[0] > 1:", `case tensor.ToBool(tmath.Greater(tensor.AnySlice(a, 0), tensor.NewIntScalar(1))):`},
		{"case y > 2: y = 0", `case y > 2: y = 0`},
		{"default:", `default :`},
		{"}", `}`},
		{"switch int(y) {", `switch int(y) {`},
		{"case 1, 2:", `case 1, 2:`},
		{"}", `}`},
		{"if y > 1 {", `if y > 1 {`},
		{"} else if a[0] > 1 {", `} else if tensor.ToBool(tmath.Greater(tensor.AnySlice(a, 0), tensor.NewIntScalar(1))) {`},
		{"} else {", `} else {`},
		{"}", `}`},
		{"k := int(a[1]) + 1", `k := tensor.ToInt(tensor.AnySlice(a, 1)) + 1`},
		{"b := float64(k) * 2", `b := float64(k) *2`},
		{"s := []float64{1, 2}", `s := []float64 { 1, 2 }`},
		{"s[0] = 3", `s[0] = 3`},
		{"s[1] = a[0]", `tensor.SetVar( & s[1], tensor.AnySlice(a, 0))`},
		{"s[0]++", `s[0]++`},
		{"w := s[1] * tau", `w := s[1] *tau`},
		{"q := tau", `q := tau`},
		{"q++", `q++`},
		{"x++", `tmath.AddAssignVar( & x, 1)`},
		{"x = a[0]", `tensor.SetVar( & x, tensor.AnySlice(a, 0))`},
		{"x *= 2", `tmath.MulAssignVar( & x, 2)`},
		{"##", ``},
	}
	st := NewState()
	st.MathRecord = false
	for _, test := range tests {
		o := st.TranspileLine(test.i)
		assert.Equal(t, test.e, o, test.i)
	}
}

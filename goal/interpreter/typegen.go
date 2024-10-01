// Code generated by "core generate -add-types -add-funcs"; DO NOT EDIT.

package interpreter

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/goal/interpreter.Config", IDName: "config", Doc: "Config is the configuration information for the goal cli.", Directives: []types.Directive{{Tool: "go", Directive: "generate", Args: []string{"core", "generate", "-add-types", "-add-funcs"}}}, Fields: []types.Field{{Name: "Input", Doc: "Input is the input file to run/compile.\nIf this is provided as the first argument,\nthen the program will exit after running,\nunless the Interactive mode is flagged."}, {Name: "Expr", Doc: "Expr is an optional expression to evaluate, which can be used\nin addition to the Input file to run, to execute commands\ndefined within that file for example, or as a command to run\nprior to starting interactive mode if no Input is specified."}, {Name: "Args", Doc: "Args is an optional list of arguments to pass in the run command.\nThese arguments will be turned into an \"args\" local variable in the goal.\nThese are automatically processed from any leftover arguments passed, so\nyou should not need to specify this flag manually."}, {Name: "Interactive", Doc: "Interactive runs the interactive command line after processing any input file.\nInteractive mode is the default mode for the run command unless an input file\nis specified."}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/goal/interpreter.Interpreter", IDName: "interpreter", Doc: "Interpreter represents one running shell context", Fields: []types.Field{{Name: "Goal", Doc: "the goal shell"}, {Name: "HistFile", Doc: "HistFile is the name of the history file to open / save.\nDefaults to ~/.goal-history for the default goal shell.\nUpdate this prior to running Config() to take effect."}, {Name: "Interp", Doc: "the yaegi interpreter"}}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/goal/interpreter.Run", Doc: "Run runs the specified goal file. If no file is specified,\nit runs an interactive shell that allows the user to input goal.", Directives: []types.Directive{{Tool: "cli", Directive: "cmd", Args: []string{"-root"}}}, Args: []string{"c"}, Returns: []string{"error"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/goal/interpreter.Interactive", Doc: "Interactive runs an interactive shell that allows the user to input goal.", Args: []string{"c", "in"}, Returns: []string{"error"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/goal/interpreter.Build", Doc: "Build builds the specified input goal file, or all .goal files in the current\ndirectory if no input is specified, to corresponding .go file name(s).\nIf the file does not already contain a \"package\" specification, then\n\"package main; func main()...\" wrappers are added, which allows the same\ncode to be used in interactive and Go compiled modes.", Args: []string{"c"}, Returns: []string{"error"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/goal/interpreter.init"})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/goal/interpreter.NewInterpreter", Doc: "NewInterpreter returns a new [Interpreter] initialized with the given options.\nIt automatically imports the standard library and configures necessary shell\nfunctions. End user app must call [Interp.Config] after importing any additional\nsymbols, prior to running the interpreter.", Args: []string{"options"}, Returns: []string{"Interpreter"}})
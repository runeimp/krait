/**
 * Krait Go FlagSet enhancer
 *
 *
 *
 *
 */
package krait

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	LibName    = "Krait"
	LibVersion = "0.1.0"
)

const (
	ContinueOnError     flag.ErrorHandling = flag.ContinueOnError
	ExitOnError         flag.ErrorHandling = flag.ExitOnError
	PanicOnError        flag.ErrorHandling = flag.PanicOnError
	ErrorInvalidCommand                    = "invalid command"
	ErrorNoArguments                       = "no command line arguments"
	ErrorNotParsed                         = "command line has not been parsed"
	OptionBool                             = "bool"
	OptionInt                              = "int"
	OptionString                           = "string"
	OptionUint                             = "uint"
	summeryTitle                           = "COMMAND SUMMERY\n---------------"
	optionTitle                            = "\n\nOPTIONS\n-------"
	// ErrorInvalidSubCommand                    = "invalid subcommand"

// 	appUsage = `
// USAGE: %s [OPTIONS] path/to/sqlite.db

// OPTIONS:
// `
)

/*
- Need to wrap the flag stdlib module

*/

const (
	ArgBareArgument   ArgumentType = "bare-argument"
	ArgOptionArgument ArgumentType = "option-argument"
	ArgOptionAlias    ArgumentType = "option-alias"
	ArgSubCommand     ArgumentType = "sub-command"
)

const usageNoOptions = `Usage: %s [ARGUMENTS]
`

const usageWithOptions = `Usage: %s [OPTIONS] [ARGUMENTS]

OPTIONS
-------
`

type ArgumentType string

type Argument struct {
	Type  ArgumentType
	Value string
}

// FlagSet is the Krait expansion of flag.FlagSet
type FlagSet struct {
	AppLabel      string
	NArgs         int                                  // The number of arguments expected for this subcommand. 0 = none, 1+ = the exact number of expected arguments, -1 = any number of arguments
	args          []string                             // bare arguments
	cmd           string                               // Command name
	CmdFunc       func(fs *FlagSet, args ...string)    // Function to call when the command is active
	Epilogue      string                               // Help epilogue
	flagSet       *flag.FlagSet                        // flag.FlagSet for the krait.FlagSet
	HelpOutput    func(fs *FlagSet, cmdName ...string) // The default help output method
	isParsed      bool                                 // If a command line was parsed yet
	optionAliases map[string]string                    // POSIX or GNU aliases for an option
	Options       map[string]Option                    // Map of options to track
	parent        *FlagSet                             // Parent krait.FlagSet of this krait.FlagSet
	subcmd        string                               // Active sub-command
	subcmdAliases map[string]string                    // Map of valid sub-command aliases
	subcommands   map[string]*FlagSet                  // Map of krait.FlagSet sub-commands
	Summery       string                               // krait.FlagSet sub-command usage summery
	// Root          bool
	// Usage         func()
	// usageTemplate string
}

func (fs *FlagSet) argIsSubcommand(arg string) (result bool) {
	arg = strings.ToLower(arg) // Lowercase because subcommand case shouldn't matter

	// Check subcommands for a match
	for k := range fs.subcommands {
		if arg == k {
			result = true
			break
		}
	}

	if result == false {
		// Check subcommand aliases for a match
		for k := range fs.subcmdAliases {
			if arg == k {
				result = true
				break
			}
		}
	}

	// log.Printf("krait.FlagSet.argIsSubcommand() | %q | arg: %q | result: %t\n", fs.cmd, arg, result)

	return result
}

func (fs *FlagSet) Args() (args []string, err error) {
	if fs.isParsed {
		args = fs.args
	} else {
		err = fmt.Errorf(ErrorNotParsed)
	}
	return args, err
}

func (fs *FlagSet) getCommandList() (list []string) {
	list = append(list, fs.cmd)
	parent := fs.parent

	for parent != nil {
		list = append(list, parent.cmd)
		parent = parent.parent
	}

	list = sliceStringsReverse(list)

	return list
}

// getRoot may not be necessary
func (fs *FlagSet) getRoot() (p *FlagSet) {
	// p = fs.parent
	if p.ParentName() != "nil" {
		p = p.parent
	}

	return p
}

// func (fs *FlagSet) NewFlagSet(subcommand string, cmdFunc func(fs *FlagSet, args ...string), errHandler ...flag.ErrorHandling) *FlagSet {
func (fs *FlagSet) NewFlagSet(subcommand string, errHandler ...flag.ErrorHandling) *FlagSet {
	// log.Printf("krait.FlagSet.HelpOutput() | %q | subcommand: %q\n", fs.cmd, subcommand)
	errorHandler := flag.ExitOnError
	if len(errHandler) > 0 {
		errorHandler = errHandler[0]
	}

	subcommand = strings.ToLower(subcommand) // Lowercase because case shouldn't matter

	ksc := &FlagSet{
		cmd:           subcommand,
		flagSet:       flag.NewFlagSet(subcommand, errorHandler),
		parent:        fs,
		Options:       make(map[string]Option),
		optionAliases: make(map[string]string),
		subcmdAliases: make(map[string]string),
		subcommands:   make(map[string]*FlagSet),
	}
	// ksc.flagSet.Usage = func() {
	// 	fmt.Fprintf(flag.CommandLine.Output(), "\nUSAGE: %s %s\n\n", ksc.flagSet.Name(), ksc.Summery)
	// }
	ksc.flagSet.Usage = func() {
		optionCount := 0
		usage := usageNoOptions

		ksc.flagSet.VisitAll(func(f *flag.Flag) { optionCount++ })

		if optionCount > 0 {
			usage = usageWithOptions
		}

		cmdChain := strings.Join(ksc.getCommandList(), " ")

		// fmt.Fprintf(flag.CommandLine.Output(), usage, ksc.parent.cmd, ksc.cmd)
		fmt.Fprintf(flag.CommandLine.Output(), usage, cmdChain)

		ksc.flagSet.VisitAll(func(f *flag.Flag) {
			optionName := fmt.Sprintf("-%s", f.Name)
			if f.DefValue == "" {
				fmt.Fprintf(flag.CommandLine.Output(), "  %-13s  %s (no default)\n", optionName, f.Usage)
			} else {
				fmt.Fprintf(flag.CommandLine.Output(), "  %-13s  %s (default: %v)\n", optionName, f.Usage, f.DefValue)
			}
		})
		fmt.Println()
	}

	fs.subcommands[subcommand] = ksc

	return ksc
}

func (fs *FlagSet) OptionBool(aliases []string, defaultValue bool, description string) *bool {
	var alias string
	alias, aliases = fs.optionAliasSetup(aliases)
	p := fs.flagSet.Bool(alias, defaultValue, description)
	fs.Options[alias] = Option{
		Type:  OptionBool,
		value: p,
	}
	return p
}

// func (k FlagSet) BoolVar(p *bool, name string, defaultValue bool, description string) {
// 	fs.flagSet.BoolVar(p, name, defaultValue, description)
// }

// func (k FlagSet) Int(name string, defaultValue int, description string) *int {
// 	p := fs.flagSet.Int(name, defaultValue, description)
// 	return p
// }

// func (k FlagSet) IntVar(p *int, name string, defaultValue int, description string) {
// 	fs.flagSet.IntVar(p, name, defaultValue, description)
// }

func (fs *FlagSet) optionAliasSetup(aliasList []string) (alias string, aliases []string) {
	var (
		aliasPrefixed string
		longest       int
	)

	for i, arg := range aliasList {
		if len(arg) >= len(aliasList[longest]) {
			longest = i
		}
	}
	alias = aliasList[longest]

	for i := 0; i < len(aliasList); i++ {
		if i != longest {
			aliases = append(aliases, aliasList[i])
		}
	}

	if len(alias) == 1 {
		aliasPrefixed = "-" + alias
	} else {
		aliasPrefixed = "--" + alias
	}

	for _, a := range aliases {
		// POSIX short options and Multics options
		a = "-" + a
		fs.optionAliases[a] = aliasPrefixed

		// GNU long options
		if len(a) > 1 {
			a = "-" + a
			fs.optionAliases[a] = aliasPrefixed
		}
	}

	return alias, aliases
}

func (fs *FlagSet) OptionInt(aliases []string, dataDefault int, description string) (o *int) {
	var alias string

	log.Printf("krait.FlagSet.OptionInt() | %q | aliases: %q\n", fs.cmd, aliases)
	alias, aliases = fs.optionAliasSetup(aliases)
	log.Printf("krait.FlagSet.OptionInt() | %q | alias: %q\n", fs.cmd, alias)
	log.Printf("krait.FlagSet.OptionInt() | %q | aliases: %q\n", fs.cmd, aliases)
	log.Printf("krait.FlagSet.OptionInt() | %q | dataDefault: %d | description: %q\n", fs.cmd, dataDefault, description)
	o = fs.flagSet.Int(alias, dataDefault, description)
	fs.Options[alias] = Option{Type: OptionInt, value: o}
	// log.Printf("krait.FlagSet.OptionInt() | %q | fs.Options[%q]: %v\n", fs.cmd, alias, fs.Options[alias])
	return o
}

func (fs *FlagSet) ParentName() string {
	if fs == nil || fs.parent == nil {
		return "nil"
	}
	return fs.parent.cmd
}

func (fs *FlagSet) parse(subFS *FlagSet, args []string, pass int) (subcmd string, err error) {
	// log.Printf("krait.FlagSet.parse() | %q | pass: %d | args: %q\n", fs.cmd, pass, args)

	if pass == 1 {
		if fs.cmd != subcmd {
			/*
				The expected command line name is different.
				For now we don't care but could be useful in a later version. ~RuneImp
			*/
		}
		args = args[1:]
		subcmd, err = fs.parse(subFS, args, 2)
	} else {
		if len(args) == 0 {
			// No possibility of a subcommand
			// log.Printf("krait.FlagSet.parse() | %q | pass: %d | subcmd: %q | No Arguments\n", fs.cmd, pass, subcmd)
		} else {
			arg := args[0]

			if fs.argIsSubcommand(arg) {
				// The 1st argument was a subcommand so keep drilling down

				subcmd = strings.ToLower(arg) // Lowercase because subcommand case shouldn't matter
				subFS = fs.subcommands[arg]

				arg, err = fs.parse(subFS, args[1:], pass+1)
				if arg != "" {
					subcmd = arg
				}
			} else {
				// The 1st argument wasn't a subcommand so parse arguments with the flag library
				args = args[1:]
				// log.Printf("krait.FlagSet.parse() | %q | pass: %d | subcmd: %q | Parsing args: %q\n", fs.cmd, pass, subcmd, args)
				err = subFS.flagSet.Parse(args)
			}
		}
	}

	// log.Printf("krait.FlagSet.parse() | %q | pass: %d | subcmd: %q | err: %v\n", fs.cmd, pass, subcmd, err)
	return subcmd, err
}

func (fs *FlagSet) Parse(args []string) (subcmd string, err error) {
	log.Printf("krait.FlagSet.Parse() | %q | args: %q\n", fs.cmd, args)
	// log.Printf("krait.FlagSet.Parse() | %q | fs.flagSet.Name(): %q\n", fs.cmd, fs.flagSet.Name())
	// log.Printf("krait.FlagSet.Parse() | %q | fs.ParentName(): %s\n", fs.cmd, quoteNotNil(fs.ParentName()))

	/*
		## Basic Truth

		* The Parse method just walks the command tree
		* If the 1st argument isn't a subcommand hand off to flag.FlatSet.Parse()
	*/

	// Sanity Check: the args slice should always have the command name for the FlagSet at the very least
	if len(args) == 0 {
		err = fmt.Errorf(ErrorInvalidCommand)
		return subcmd, err
	}

	// Check if the FlagSet is receiving any arguments to parse
	if len(args) == 1 {
		err = fmt.Errorf(ErrorNoArguments)
		if fs.CmdFunc != nil {
			fs.CmdFunc(fs)
		}
		fs.subcmd = subcmd
		return subcmd, err
	}

	subcmd, err = fs.parse(fs, args, 1)

	if subFS, ok := fs.subcommands[subcmd]; ok {
		// log.Printf("krait.FlagSet.Parse() | %q | subcmd: %q | valid: %t\n", fs.cmd, subcmd, ok)
		// log.Printf("krait.FlagSet.Parse() | %q | args: %q\n", fs.cmd, args)

		// Manage option aliases in args before parsing
		for i, a := range args {
			// log.Printf("krait.FlagSet.Parse() | %q | alias fix | a: %q | subFS.optionAliases: %q\n", subFS.cmd, a, subFS.optionAliases)
			for k, v := range subFS.optionAliases {
				// log.Printf("krait.FlagSet.Parse() | %q | alias fix | k: %q\n", subFS.cmd, k)
				if a == k {
					// log.Printf("krait.FlagSet.Parse() | %q | alias fix | v: %q\n", subFS.cmd, v)
					args[i] = v
				}
			}
		}
		// log.Printf("krait.FlagSet.Parse() | %q | args: %q\n", fs.cmd, args)

		// subcmd, err = subFS.Parse(args)
		_, err = subFS.Parse(args)
		// err = subFS.flagSet.Parse(args)
	} else {
		// log.Printf("krait.FlagSet.Parse() | %q | subcmd: %q | valid: false\n", fs.cmd, subcmd)

		err = fs.flagSet.Parse(args)
		// root := fs.getRoot()
		// root.args = append(root.args, fs.flagSet.Args()...)
		// // for _, a := range fs.flagSet.Args() {
		// // 	fs.args = append(fs.args, a)
		// // }

		if fs.CmdFunc != nil {
			fs.CmdFunc(fs, args...)
		}
	}

	// log.Printf("krait.FlagSet.Parse() | %q | subcmd: %q | err: %v\n", fs.cmd, subcmd, err)

	return subcmd, err
}

// Returns true if the FlagSet has been parsed or false otherwise
func (fs *FlagSet) Parsed() bool {
	return fs.isParsed
}

// func (k FlagSet) String(name string, defaultValue string, description string) *string {
// 	p := fs.flagSet.String(name, defaultValue, description)
// 	return p
// }

// func (k FlagSet) StringVar(p *string, name string, defaultValue string, description string) {
// 	fs.flagSet.StringVar(p, name, defaultValue, description)
// }

func (fs *FlagSet) SubCommand() string {
	return fs.subcmd
}

func (fs *FlagSet) SubcommandAlias(subcommand, alias string) {
	// ksc := fs.subcommands[subcommand]
	// fs.subcommands[alias] = ksc

	fs.subcmdAliases[alias] = subcommand
}

func NewFlagSet(name string) (fs *FlagSet) {
	// log.Printf("krait.NewFlagSet() | name: %q\n", name)

	// Root FlagSet
	fs = &FlagSet{
		cmd:           name,
		flagSet:       flag.NewFlagSet(name, flag.ExitOnError),
		HelpOutput:    helpOutput,
		optionAliases: make(map[string]string),
		Options:       make(map[string]Option),
		subcmdAliases: make(map[string]string),
		subcommands:   make(map[string]*FlagSet),
		// Epilogue:      "",
		// Options:       make(map[string]any),
		// Options:       make(map[string]Optional),
		// root:          true,
	}

	fs.flagSet.Usage = func() {
		optionCount := 0
		usage := usageNoOptions

		fs.flagSet.VisitAll(func(f *flag.Flag) { optionCount++ })

		if optionCount > 0 {
			usage = usageWithOptions
		}

		cmdChain := strings.Join(fs.getCommandList(), " ")

		// fmt.Fprintf(flag.CommandLine.Output(), usage, fs.parent.cmd, fs.cmd)
		fmt.Fprintf(flag.CommandLine.Output(), usage, cmdChain)

		fs.flagSet.VisitAll(func(f *flag.Flag) {
			optionName := fmt.Sprintf("-%s", f.Name)
			if f.DefValue == "" {
				fmt.Fprintf(flag.CommandLine.Output(), "  %-13s  %s (no default)\n", optionName, f.Usage)
			} else {
				fmt.Fprintf(flag.CommandLine.Output(), "  %-13s  %s (default: %v)\n", optionName, f.Usage, f.DefValue)
			}
		})
		fmt.Println()
	}

	fs.SubcommandAlias("version", "ver")
	fs.NewFlagSet("version", flag.ExitOnError)
	helpFS := fs.NewFlagSet("help", flag.ExitOnError)
	helpFS.CmdFunc = cmdHelp
	helpFS.HelpOutput = helpOutput
	helpFS.Summery = "Displays this help information"

	return fs
}

type Option struct {
	Type  string
	value any
}

// GetBool returns boolean true or false for a given value. If the value is
// a string it will return false if the the value is a zero length string or
// true otherwise.
func (o Option) GetBool() (result bool, err error) {
	switch o.Type {
	case OptionBool:
		result = o.value.(bool)
	case OptionInt:
		// result = false // result is false by default
		if o.value.(int) != 0 {
			result = true
		}
	case OptionString:
		// result = false // result is false by default
		if o.value.(string) != "" {
			result = true
		}
	}
	return result, err
}

func (o Option) GetInt() (result int, err error) {

	// log.Printf("krait.Option.GetInt() | o.Type: %s | *o.value.(*int): %v (%T)\n", o.Type, *o.value.(*int), o.value)
	switch o.Type {
	case OptionBool:
		result = 0
		if o.value.(bool) {
			result = 1
		}
	case OptionInt:
		result = *o.value.(*int)
	case OptionString:
		result, err = strconv.Atoi(o.value.(string))
	}
	return result, err
}

func (o Option) GetString() (result string, err error) {
	switch o.Type {
	case OptionBool:
		result = "false"
		if o.value.(bool) {
			result = "true"
		}
	case OptionInt:
		result = strconv.Itoa(o.value.(int))
	case OptionString:
		result = o.value.(string)
	}
	return result, err
}

func cmdHelp(fs *FlagSet, args ...string) {
	// log.Printf("krait.cmdHelp() | fs: %v | args: %q\n", fs, args)
	// log.Printf("krait.cmdHelp() | fs.cmd: %q | fs.HelpOutput: %v | args: %q\n", fs.cmd, fs.HelpOutput, args)
	// log.Printf("krait.cmdHelp() | fs.cmd: %q | args: %q\n", fs.cmd, args)
	if fs != nil && fs.HelpOutput != nil {
		fs.HelpOutput(fs, args...)
	}
	os.Exit(0)
}

func helpOutput(fs *FlagSet, args ...string) {
	// log.Printf("krait.helpOutput() | fs: %v | args: %q\n", fs, args)
	// log.Printf("krait.helpOutput() | fs.cmd: %q | args: %q\n", fs.cmd, args)

	cmdName := fs.cmd

	if len(args) > 0 {
		if cmdName == "help" {
			cmdName = args[0]
		}
		if len(args) > 1 {
			args = args[1:]
		} else {
			args = []string{}
		}
	}

	// log.Printf("krait.helpOutput() | cmdName: %q | args: %q\n", cmdName, args)

	if cmdName == "help" {
		rfs := fs.getRoot()
		fmt.Fprintln(flag.CommandLine.Output(), rfs.AppLabel)

		// Output for all commands
		widestCommand := 0

		cmdList := []string{}
		for subCmdName, subCmdFS := range fs.parent.subcommands {
			// log.Printf("krait.FlagSet.HelpOutput() | subCmdName: %q | subCmdFS.flagSet.Name(): %q\n", subCmdName, subCmdFS.flagSet.Name())
			if len(subCmdFS.flagSet.Name()) > widestCommand {
				widestCommand = len(subCmdFS.flagSet.Name())
			}
			cmdList = append(cmdList, subCmdName)
		}
		sort.Strings(cmdList)

		format := fmt.Sprintf("  %%-%ds  %%s\n", widestCommand)
		formatOptionDefault := fmt.Sprintf("  %%-%ds  %%s (default: %%v)\n", widestCommand)
		formatOptionNodefault := fmt.Sprintf("  %%-%ds  %%s (no default)\n", widestCommand)

		// Commands
		// log.Printf("krait.FlagSet.HelpOutput() | format: %q\n", format)
		// log.Printf("krait.FlagSet.HelpOutput() | cmdList: %q\n", cmdList)
		fmt.Fprintln(flag.CommandLine.Output())

		fmt.Fprintln(flag.CommandLine.Output(), summeryTitle)
		for _, cmdName := range cmdList {
			subCmdFS := fs.parent.subcommands[cmdName]
			fmt.Fprintf(flag.CommandLine.Output(), format, subCmdFS.flagSet.Name(), subCmdFS.Summery)
			// subCmdFS.flagSet.Usage()
		}

		// Options
		optionCount := 0
		rfs.flagSet.VisitAll(func(f *flag.Flag) { optionCount++ })

		if optionCount > 0 {
			fmt.Fprintln(flag.CommandLine.Output(), optionTitle)
			rfs.flagSet.VisitAll(func(f *flag.Flag) {
				optionName := fmt.Sprintf("-%s", f.Name)
				if f.DefValue == "" {
					fmt.Fprintf(flag.CommandLine.Output(), formatOptionNodefault, optionName, f.Usage)
				} else {
					fmt.Fprintf(flag.CommandLine.Output(), formatOptionDefault, optionName, f.Usage, f.DefValue)
				}
			})
			fmt.Fprintln(flag.CommandLine.Output())
		}

		// Epilogue
		fmt.Fprintln(flag.CommandLine.Output())
		// log.Printf("krait.FlagSet.HelpOutput() | fs.cmd: %q | fs.AppLabel: %q | fs.Epilogue: %q\n", fs.cmd, fs.AppLabel, fs.Epilogue)
		// log.Printf("krait.FlagSet.HelpOutput() | fs.parent.cmd: %q | fs.parent.Epilogue: %q\n", fs.parent.cmd, fs.parent.Epilogue)
		if fs.Epilogue != "" {
			fmt.Fprintf(flag.CommandLine.Output(), fs.Epilogue)
			fmt.Fprintln(flag.CommandLine.Output())
		} else {
			parent := fs.parent
			for parent != nil {
				// log.Printf("krait.FlagSet.HelpOutput() | parent.cmd: %q\n", parent.cmd)
				if parent.Epilogue != "" {
					fmt.Fprintf(flag.CommandLine.Output(), parent.Epilogue)
					fmt.Fprintln(flag.CommandLine.Output())
					break
				}
				parent = parent.parent
			}
		}
	} else {
		for _, subcmd := range fs.parent.subcommands {
			// log.Printf("krait.FlagSet.HelpOutput() | subcmd.flagSet.Name(): %q\n", subcmd.flagSet.Name())
			if subcmd.flagSet.Name() == cmdName {
				// log.Printf("krait.FlagSet.HelpOutput() | cmdName match: %q\n", cmdName)
				subcmd.flagSet.Usage()
				// flag.PrintDefaults()
				// subcmd.flagSet.PrintDefaults()

				// fmt.Fprintf(flag.CommandLine.Output(), subcmd.Summery)
				break
			}
		}
	}

	fmt.Fprintln(flag.CommandLine.Output())
}

func quoteNotNil(s string) string {
	if s == "nil" {
		return s
	}
	return fmt.Sprintf("%q", s)
}

func sliceStringsReverse(input []string) (output []string) {
	for i := len(input) - 1; i >= 0; i-- {
		output = append(output, input[i])
	}
	// log.Printf("krait.sliceStringsReverse() | input: %q | output: %q\n", input, output)
	return output
}

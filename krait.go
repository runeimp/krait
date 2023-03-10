/**
 * Krait Go FlagSet enhancer
 *
 *
 *
 *
 */
package krait

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	LibName    = "Krait"
	LibVersion = "0.2.0"
)

const (
	ContinueOnError     flag.ErrorHandling = flag.ContinueOnError
	ExitOnError         flag.ErrorHandling = flag.ExitOnError
	PanicOnError        flag.ErrorHandling = flag.PanicOnError
	ErrorInvalidCommand                    = "invalid command"
	ErrorNoArguments                       = "no command line arguments"
	OptionBool                             = "bool"
	OptionFloat                            = "float64"
	OptionInt                              = "int"
	OptionString                           = "string"
	OptionUint                             = "uint"
	summeryTitle                           = "COMMAND SUMMERY\n---------------" // May become editable in future versions
	optionTitle                            = "\n\nOPTIONS\n-------"             // May become editable in future versions
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
	AppLabel          string                               // Application name and version number
	args              []string                             // bare arguments
	NArgs             int                                  // The number of arguments expected for this subcommand. 0 = none, 1+ = the exact number of expected arguments, -1 = any number of arguments
	cmd               string                               // Command name
	CmdFunc           func(fs *FlagSet, args ...string)    // Function to call when the command is active
	DefaultSubCommand string                               // The subcommand to use when none is specified
	Epilogue          string                               // Help epilogue
	flagSet           *flag.FlagSet                        // flag.FlagSet for the krait.FlagSet
	HelpOutput        func(fs *FlagSet, cmdName ...string) // The default help output method
	isParsed          bool                                 // If a command line was parsed yet
	level             int                                  // Sub command level
	optionAliases     map[string]string                    // POSIX or GNU aliases for an option
	Options           map[string]Option                    // Map of options to track
	parent            *FlagSet                             // Parent krait.FlagSet of this krait.FlagSet
	subcmd            string                               // Active sub-command
	subcmdAliases     []map[string]string                  // Slice of Map string of valid sub-command aliases. NOTE: really should be []map[string][]string (level / subcommand / aliases)
	subcommands       []map[string]*FlagSet                // Slice of Map of krait.FlagSet sub-commands. The slice represents the subcommand levels
	Summery           string                               // krait.FlagSet sub-command usage summery
	// Root          bool
	// Usage         func()
	// usageTemplate string
}

func (fs *FlagSet) String() string {
	dict := make(map[string]string)
	dict["AppLabel"] = fs.AppLabel
	dict["NArgs"] = strconv.Itoa(fs.NArgs)
	dict["args"] = strings.Join(fs.args, ", ")
	dict["cmd"] = fs.cmd
	dict["isParsed"] = fmt.Sprintf("%t", fs.isParsed)
	dict["level"] = strconv.Itoa(fs.level)
	dict["subcmd"] = fs.subcmd
	dict["Summery"] = fs.Summery
	// dict["AppLabel"] = fs.AppLabel
	// dict["AppLabel"] = fs.AppLabel

	jsonBytes, err := json.MarshalIndent(dict, "", "	")

	result := ""
	if err != nil {
		result = err.Error()
	} else {
		result = string(jsonBytes)
	}

	return result
}

func (fs *FlagSet) argIsSubcommand(arg string, level int) (subcmd string, result bool) {
	arg = strings.ToLower(arg) // Lowercase because subcommand case shouldn't matter
	// log.Printf("krait.FlagSet.argIsSubcommand() | %q | level: %d | arg: %q\n", fs.cmd, level, arg)

	rfs := fs.getRoot()

	// log.Printf("krait.FlagSet.argIsSubcommand() | %q | level: %d | len(rfs.subcommands): %d\n", fs.cmd, level, len(rfs.subcommands))
	// Check subcommands for a match
	if level < len(rfs.subcommands) {
		for k := range rfs.subcommands[level] {
			if arg == k {
				subcmd = arg
				result = true
				break
			}
		}
	}

	// log.Printf("krait.FlagSet.argIsSubcommand() | %q | level: %d | len(rfs.subcmdAliases): %d | result: %t\n", fs.cmd, level, len(rfs.subcmdAliases), result)
	if result == false && level < len(rfs.subcmdAliases) {
		// Check subcommand aliases for a match
		for alias, sub := range rfs.subcmdAliases[level] {
			// log.Printf("krait.FlagSet.argIsSubcommand() | %q | level: %d | arg: %q | alias: %q | sub: %q\n", fs.cmd, level, arg, alias, sub)
			if arg == alias {
				subcmd = sub
				result = true
				break
			}
		}
	}

	// log.Printf("krait.FlagSet.argIsSubcommand() | %q | level: %d | arg: %q | subcmd: %q | result: %t\n", fs.cmd, level, arg, subcmd, result)

	return subcmd, result
}

func (fs *FlagSet) Args() []string {
	return fs.args
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

func (fs *FlagSet) getDefaultSubCommand(level int, subcmd string) (result string) {
	rfs := fs.getRoot()
	if level < len(rfs.subcommands) {
		if _, ok := rfs.subcommands[level][subcmd]; ok {
			result = rfs.subcommands[level][subcmd].DefaultSubCommand
		}
	}
	return result
}

// getRoot returns the root *FlagSet
func (fs *FlagSet) getRoot() (p *FlagSet) {
	p = fs
	// log.Printf("krait.FlagSet.getRoot() | %q | p: %v\n", fs.cmd, p)

	for p.parent != nil {
		// log.Printf("krait.FlagSet.getRoot() | %q | p: %v\n", fs.cmd, p)
		p = p.parent
	}

	return p
}

func (fs *FlagSet) getSubCommandAliases(subcommand string) (result []string) {
	rfs := fs.getRoot()
	if rfs.subcmdAliases[fs.level] != nil {
		for alias, subcmd := range rfs.subcmdAliases[fs.level] {
			if subcmd == subcommand {
				result = append(result, alias)
			}
		}
	}
	sort.Strings(result)
	return result
}

func (fs *FlagSet) NewFlagSet(subcommand string, errHandler ...flag.ErrorHandling) *FlagSet {
	// log.Printf("krait.FlagSet.NewFlagSet() | %q | subcommand: %q\n", fs.cmd, subcommand)
	errorHandler := flag.ExitOnError
	if len(errHandler) > 0 {
		errorHandler = errHandler[0]
	}

	subcommand = strings.ToLower(subcommand) // Lowercase because case shouldn't matter

	nfs := &FlagSet{
		cmd:           subcommand,
		flagSet:       flag.NewFlagSet(subcommand, errorHandler),
		level:         fs.level + 1,
		parent:        fs,
		Options:       make(map[string]Option),
		optionAliases: make(map[string]string),
		subcmdAliases: make([]map[string]string, 2),
		subcommands:   make([]map[string]*FlagSet, 2),
	}

	// nfs.flagSet.Usage = func() {
	// 	fmt.Fprintf(flag.CommandLine.Output(), "\nUSAGE: %s %s\n\n", nfs.flagSet.Name(), nfs.Summery)
	// }
	nfs.flagSet.Usage = func() {
		optionCount := 0
		usage := usageNoOptions
		longestName := 0
		formatNoDefault := "  %%-%ds  %%s (no default)\n"
		formatWithDefault := "  %%-%ds  %%s (default: %%v)\n"

		nfs.flagSet.VisitAll(func(f *flag.Flag) {
			if len(f.Name) > longestName {
				longestName = len(f.Name)
			}
			optionCount++
		})

		longestName += 3
		formatNoDefault = fmt.Sprintf(formatNoDefault, longestName)
		formatWithDefault = fmt.Sprintf(formatWithDefault, longestName)

		if optionCount > 0 {
			usage = usageWithOptions
		}

		cmdChain := strings.Join(nfs.getCommandList(), " ")

		fmt.Fprintln(flag.CommandLine.Output(), fs.AppLabel)
		fmt.Fprintln(flag.CommandLine.Output())
		// fmt.Fprintf(flag.CommandLine.Output(), usage, nfs.parent.cmd, nfs.cmd)
		fmt.Fprintf(flag.CommandLine.Output(), usage, cmdChain)

		// log.Printf("FlagSet.flagSet.Usage() | nfs.optionAliases: %#v\n", nfs.optionAliases)
		// for k, v := range nfs.optionAliases {
		// 	log.Printf("FlagSet.flagSet.Usage() | nfs.optionAliases[%q]: %s\n", k, v)
		// }

		nfs.flagSet.VisitAll(func(f *flag.Flag) {
			optionName := fmt.Sprintf("-%s", f.Name)
			gnuOptionName := fmt.Sprintf("--%s", f.Name)

			aliases := []string{}
			for k, v := range nfs.optionAliases {
				if v == gnuOptionName {
					aliases = append(aliases, k)
				}
			}
			// log.Printf("FlagSet.flagSet.Usage() | optionName: %q | aliases: %q\n", optionName, aliases)
			// log.Printf("FlagSet.flagSet.Usage() | f.Name: %q\n", f.Name)

			if len(aliases) == 1 && len(aliases[0]) < 3 {
				optionName += ", " + aliases[0]
			}

			if f.DefValue == "" {
				// fmt.Fprintf(flag.CommandLine.Output(), "  %-13s  %s (no default)\n", optionName, f.Usage)
				fmt.Fprintf(flag.CommandLine.Output(), formatNoDefault, optionName, f.Usage)
			} else {
				// fmt.Fprintf(flag.CommandLine.Output(), "  %-13s  %s (default: %v)\n", optionName, f.Usage, f.DefValue)
				fmt.Fprintf(flag.CommandLine.Output(), formatWithDefault, optionName, f.Usage, f.DefValue)
			}
			if len(aliases) > 0 {
				// fmt.Fprintf(flag.CommandLine.Output(), "    %v\n", aliases)
				if len(aliases) > 1 {
					aliasSort(aliases)
					str := ""
					for _, a := range aliases {
						str += ", " + a
					}
					fmt.Println("      aliases: " + str[2:])
				}
			}
		})
		fmt.Println()
	}

	if fs.subcommands[nfs.level] == nil {
		fs.subcommands[nfs.level] = make(map[string]*FlagSet)
	}
	fs.subcommands[nfs.level][subcommand] = nfs

	return nfs
}

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
		if len(a) > 2 {
			a = "-" + a
			fs.optionAliases[a] = aliasPrefixed
		}
	}
	fs.optionAliases["--"+alias] = aliasPrefixed

	// for k, v := range fs.optionAliases {
	// 	log.Printf("krait.FlagSet.optionAliasSetup() | %d | fs.optionAliases[%q]: %q\n", fs.level, k, v)
	// }
	// log.Printf("krait.FlagSet.optionAliasSetup() | %d | alias: %q | aliases: %q\n", fs.level, alias, aliases)

	return alias, aliases
}

func (fs *FlagSet) OptionBool(aliases []string, defaultValue bool, description string) (o *bool) {
	var alias string
	alias, aliases = fs.optionAliasSetup(aliases)
	o = fs.flagSet.Bool(alias, defaultValue, description)
	fs.Options[alias] = Option{
		Type:  OptionBool,
		value: o,
	}
	return o
}

func (fs *FlagSet) OptionFloat(aliases []string, defaultValue float64, description string) (o *float64) {
	var alias string

	// log.Printf("krait.FlagSet.OptionFloat() | %q | aliases: %q\n", fs.cmd, aliases)
	alias, aliases = fs.optionAliasSetup(aliases)
	o = fs.flagSet.Float64(alias, defaultValue, description)
	fs.Options[alias] = Option{Type: OptionFloat, value: o}
	return o
}

func (fs *FlagSet) OptionInt(aliases []string, defaultValue int, description string) (o *int) {
	var alias string

	// log.Printf("krait.FlagSet.OptionInt() | %q | aliases: %q\n", fs.cmd, aliases)
	alias, aliases = fs.optionAliasSetup(aliases)
	// log.Printf("krait.FlagSet.OptionInt() | %q | alias: %q\n", fs.cmd, alias)
	// log.Printf("krait.FlagSet.OptionInt() | %q | aliases: %q\n", fs.cmd, aliases)
	// log.Printf("krait.FlagSet.OptionInt() | %q | defaultValue: %d | description: %q\n", fs.cmd, defaultValue, description)
	o = fs.flagSet.Int(alias, defaultValue, description)
	fs.Options[alias] = Option{Type: OptionInt, value: o}
	// log.Printf("krait.FlagSet.OptionInt() | %q | fs.Options[%q]: %v\n", fs.cmd, alias, fs.Options[alias])
	return o
}

func (fs *FlagSet) OptionUint(aliases []string, defaultValue uint, description string) (o *uint) {
	var alias string

	// log.Printf("krait.FlagSet.OptionInt() | %q | aliases: %q\n", fs.cmd, aliases)
	alias, aliases = fs.optionAliasSetup(aliases)
	o = fs.flagSet.Uint(alias, defaultValue, description)
	fs.Options[alias] = Option{Type: OptionUint, value: o}
	return o
}

func (fs *FlagSet) ParentName() string {
	if fs == nil || fs.parent == nil {
		return "nil"
	}
	return fs.parent.cmd
}

func (fs *FlagSet) parseSubCMD(args []string, level int, subcommand string) (subcmd string, lvl int, err error) {
	// log.Printf("krait.FlagSet.parseSubCMD() | %q | levels: %d:%d | subcommand: %q | args: %q (start)\n", fs.cmd, fs.level, level, subcommand, args)

	if level == 0 {
		if fs.cmd != args[0] {
			/*
				The expected command line name is different than expected.
				For now we don't care but would likely be useful feature in a later version. ~RuneImp
			*/
		}
		subcmd, lvl, err = fs.parseSubCMD(args[1:], 1, args[0]) // NOTE: should probably use fs.cmd instead of args[0] but seems less flexible
	} else {
		arg := subcommand
		argCmd, isSubCMD := fs.argIsSubcommand(arg, level)
		// log.Printf("krait.FlagSet.parseSubCMD() | %q | levels: %d:%d | arg: %q | argCmd: %q | isSubCMD: %t\n", fs.cmd, fs.level, level, arg, argCmd, isSubCMD)

		// NOTE: this logic probably needs lots more work within this method ~RuneImp
		// if isSubCMD == false {
		// 	log.Printf("krait.FlagSet.parseSubCMD() | %q | fs.level: %d | level: %d | fs.DefaultSubCommand: %q\n", fs.cmd, fs.level, level, fs.DefaultSubCommand)
		// 	if fs.getDefaultSubCommand(level) != "" {
		// 		tmp := []string{fs.DefaultSubCommand}
		// 		args = append(tmp, args...)
		// 		argCmd = fs.getDefaultSubCommand(level) // Needs to check the prior subcommand for this level of default subcommand
		// 		// isSubCMD = true
		// 	}
		// }

		if isSubCMD {
			// The 1st argument was a subcommand so keep drilling down

			subcmd = argCmd
			lvl = level
			// log.Printf("krait.FlagSet.parseSubCMD() | %q | levels: %d:%d | args: %q | subcmd: %q | len(args): %d\n", fs.cmd, fs.level, level, args, subcmd, len(args))

			if len(args) > 0 {
				// Recursive check for more subcommands if there are enough arguments to allow for more subcommands
				var (
					argTmp string
					errTmp error
					lvlTmp int
				)
				argTmp, lvlTmp, errTmp = fs.parseSubCMD(args[1:], level+1, args[0])
				// log.Printf("krait.FlagSet.parseSubCMD() | %q | levels: %d:%d | args: %q | argTmp: %q | lvlTmp: %d | errTmp: %v\n", fs.cmd, fs.level, level, args, argTmp, lvlTmp, errTmp)

				if argTmp != "" {
					subcmd = argTmp
					lvl = lvlTmp
				}
				if errTmp != nil {
					err = errTmp
				}
			}
		}
	}

	// log.Printf("krait.FlagSet.parseSubCMD() | %q | levels: %d:%d | subcmd: %q | err: %v (end)\n", fs.cmd, fs.level, level, subcmd, err)
	return subcmd, lvl, err
}

func (fs *FlagSet) Parse(input ...[]string) (subcmd string, err error) {
	args := os.Args
	if len(input) > 0 {
		args = input[0]
	}
	// log.Printf("krait.FlagSet.Parse() | %q | args: %q\n", fs.cmd, args)
	// log.Printf("krait.FlagSet.Parse() | %q | fs.flagSet.Name(): %q\n", fs.cmd, fs.flagSet.Name())
	// log.Printf("krait.FlagSet.Parse() | %q | fs.ParentName(): %s\n", fs.cmd, quoteNotNil(fs.ParentName()))

	/*
		## Basic Truths

		* The Parse method should only be called from the root FlagSet
		* The Parse method just starts the walk down the command line and handles option parsing for the specified subcommand
		* If there are no args at all something went horribly wrong, panic!
		* If there are no arguments beyond the command check fs.DefaultSubCommand
		* If the 1st argument is a subcommand call fs.parseSubCMD()
		* Else hand off to flag.FlatSet.Parse()
	*/

	// Sanity Check: this Parse method should only be called on the root FlagSet
	if fs.level != 0 {
		panic("the Parse method should only be called on the root FlagSet")
	}

	// Sanity Check: the args slice should always have the command name for the base FlagSet at the very least
	if len(args) == 0 {
		err = fmt.Errorf(ErrorInvalidCommand)
		return subcmd, err
	}

	// Check if there is a default subcommand to implement
	if len(args) == 1 {
		err = fmt.Errorf(ErrorNoArguments)
		if fs.DefaultSubCommand != "" {
			args = append(args, fs.DefaultSubCommand)
		}
	}

	level := 0

	subcmd, level, err = fs.parseSubCMD(args[1:], 0, args[0])

	// log.Printf("krait.FlagSet.Parse() | %q | level: %d | subcmd: %q\n", fs.cmd, level, subcmd)

	if subFS, ok := fs.subcommands[level][subcmd]; ok {
		args = args[level+1:]
		// log.Printf("krait.FlagSet.Parse() | %q | level: %d | subcmd: %q | valid: %t\n", fs.cmd, level, subcmd, ok)
		// log.Printf("krait.FlagSet.Parse() | %q | args: %q\n", fs.cmd, args)

		if len(args) > 0 {
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
			// log.Printf("krait.FlagSet.Parse() | %q | level: %d | subFS.flagSet.Parse(args) | args: %q\n", fs.cmd, level, args)

			// subcmd, err = subFS.Parse(args)
			// _, err = subFS.Parse(args)
			err = subFS.flagSet.Parse(args)
			fs.getRoot().args = subFS.flagSet.Args()
		}

		if subFS.CmdFunc != nil {
			if len(args) > 0 {
				// log.Printf("krait.FlagSet.Parse() | %q | level: %d | subFS.CmdFunc(subFS, args...) | args: %q\n", fs.cmd, level, args)
				subFS.CmdFunc(subFS, args...)
			} else {
				// log.Printf("krait.FlagSet.Parse() | %q | level: %d | subFS.CmdFunc(subFS)\n", fs.cmd, level)
				subFS.CmdFunc(subFS)
			}
		}
	} else {
		// log.Printf("krait.FlagSet.Parse() | %q | subcmd: %q | valid: false\n", fs.cmd, subcmd)
	}

	// log.Printf("krait.FlagSet.Parse() | %q | subcmd: %q | err: %v\n", fs.cmd, subcmd, err)

	fs.getRoot().subcmd = subcmd

	return subcmd, err
}

// Parsed returns true if the FlagSet has been parsed or false otherwise
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

// SubCommand returns the name of the active subcommand
func (fs *FlagSet) SubCommand() string {
	return fs.getRoot().subcmd
}

// SubcommandAlias links the supplied alias to the specified subcommand
func (fs *FlagSet) SubcommandAlias(aliases ...string) {
	// log.Printf("krait.FlagSet.SubcommandAlias() | fs.level: %d | fs.cmd: %q | alias: %q\n", fs.level, fs.cmd, aliases)

	rfs := fs.getRoot()

	if rfs.subcmdAliases[fs.level] == nil {
		rfs.subcmdAliases[fs.level] = make(map[string]string)
	}
	for _, alias := range aliases {
		rfs.subcmdAliases[fs.level][alias] = fs.cmd
	}

	// for lvl := range rfs.subcmdAliases {
	// 	for alias, subcmd := range rfs.subcmdAliases[lvl] {
	// 		log.Printf("krait.FlagSet.SubcommandAlias() | lvl: %d | alias: %q | subcmd: %q\n", lvl, alias, subcmd)
	// 	}
	// }
}

func NewFlagSet(name string) (fs *FlagSet) {
	// log.Printf("krait.NewFlagSet() | name: %q\n", name)

	// Root FlagSet
	fs = &FlagSet{
		cmd:               name,
		DefaultSubCommand: "help", // DefaultSubCommand defines a subcommand to use when non is specified on the command line which is "help" default
		flagSet:           flag.NewFlagSet(name, flag.ExitOnError),
		HelpOutput:        helpOutput,
		level:             0,
		optionAliases:     make(map[string]string),
		Options:           make(map[string]Option),
		subcmdAliases:     make([]map[string]string, 2),
		subcommands:       make([]map[string]*FlagSet, 2),
		// Epilogue:      "",
		// Options:       make(map[string]any),
		// Options:       make(map[string]Optional),
		// root:          true,
	}

	/*
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
	*/

	verFS := fs.NewFlagSet("version", flag.ExitOnError)
	verFS.CmdFunc = cmdVersion
	verFS.Summery = "Displays the app name and version"
	// verFS.SubcommandAlias("version", "ver")
	verFS.SubcommandAlias("ver")

	helpFS := fs.NewFlagSet("help", flag.ExitOnError)
	helpFS.CmdFunc = cmdHelp
	helpFS.HelpOutput = helpOutput
	helpFS.Summery = "Displays this help information"
	// helpFS.SubcommandAlias("help", "hlp")
	helpFS.SubcommandAlias("hlp")

	return fs
}

type Option struct {
	Type  string
	value any
}

// GetBool returns boolean true or false for a given value. If the value is
// a string it will return false if the the value is a zero length string or
// true otherwise. If the value is a number it is false if the values is zero.
func (o Option) GetBool() (result bool, err error) {
	switch o.Type {
	case OptionBool:
		b := o.value.(*bool)
		result = *b
	case OptionFloat:
		result = true
		if o.value.(float64) == 0.0 {
			result = false
		}
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
	default:
		err = fmt.Errorf("unhandled option type: %T", o.Type)
	}
	return result, err
}

func (o Option) GetFloat() (result float64, err error) {

	log.Printf("krait.Option.GetFloat() | o.Type: %s | *o.value.(*int): %v (%T)\n", o.Type, o.value.(float64), o.value)
	switch o.Type {
	case OptionBool:
		result = 0.0
		if o.value.(bool) {
			result = 1.0
		}
	case OptionFloat:
		result = o.value.(float64)
	case OptionInt:
		result = float64(o.value.(int))
	case OptionString:
		result, err = strconv.ParseFloat(o.value.(string), 64)
	case OptionUint:
		result = float64(o.value.(uint))
	default:
		err = fmt.Errorf("unhandled option type: %T", o.Type)
	}
	return result, err
}

func (o Option) GetInt() (result int, err error) {

	// log.Printf("krait.Option.GetInt() | o.Type: %s | *o.value.(*int): %v (%T)\n", o.Type, *o.value.(*int), o.value)
	switch o.Type {
	case OptionBool:
		result = 0
		if *o.value.(*bool) {
			result = 1
		}
	case OptionFloat:
		// result = strconv.FormatFloat(o.value.(float64), 'f', 0, 64)
		result = int(math.Round(o.value.(float64)))
	case OptionInt:
		result = *o.value.(*int)
	case OptionString:
		result, err = strconv.Atoi(o.value.(string))
	case OptionUint:
		result = int(o.value.(uint))
	default:
		err = fmt.Errorf("unhandled option type: %T", o.Type)
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
	case OptionFloat:
		// result = fmt.Sprintf("%f", o.value.(float64))
		result = strconv.FormatFloat(o.value.(float64), 'f', -1, 64)
	case OptionInt:
		result = strconv.Itoa(o.value.(int))
	case OptionString:
		result = o.value.(string)
	case OptionUint:
		result = fmt.Sprintf("%d", o.value.(uint))
		// result = strconv.FormatUint(uint64(o.value.(uint)), 10) // Possibly faster but seriously?
	default:
		err = fmt.Errorf("unhandled option type: %T", o.Type)
	}
	return result, err
}

func (o Option) GetUint() (result uint, err error) {

	// log.Printf("krait.Option.GetInt() | o.Type: %s | *o.value.(*int): %v (%T)\n", o.Type, *o.value.(*int), o.value)
	var tmp uint64

	switch o.Type {
	case OptionBool:
		result = 0
		if o.value.(bool) {
			result = 1
		}
	case OptionFloat:
		// result = strconv.FormatFloat(o.value.(float64), 'f', 0, 64)
		tmp := o.value.(float64)
		result = 0
		if tmp > 0.0 {
			result = uint(math.Round(tmp))
		}
	case OptionInt:
		result = uint(*o.value.(*int))
	case OptionString:
		tmp, err = strconv.ParseUint(o.value.(string), 10, 64)
		result = uint(tmp)
	case OptionUint:
		tmp := o.value.(*uint)
		result = *tmp
	default:
		err = fmt.Errorf("unhandled option type: %T", o.Type)
	}
	return result, err
}

func aliasSort(s []string) {
	// log.Printf("krait.aliasSort() | s: %q\n", s)
	posix := []string{}
	multics := []string{}
	gnu := []string{}

	for _, v := range s {
		if v[0:2] == "--" {
			// GNU Options
			gnu = append(gnu, v)
		} else if len(v) == 2 {
			// POSIX Options
			posix = append(posix, v)
		} else {
			// Multics Options
			multics = append(multics, v)
		}
	}

	i := -1

	if len(posix) > 0 {
		if len(posix) > 1 {
			sort.Strings(posix)
		}

		for _, v := range posix {
			i++
			// log.Printf("krait.aliasSort() | i: %d (POSIX)\n", i)
			s[i] = v
		}
	}

	if len(multics) > 0 {
		if len(multics) > 1 {
			sort.Strings(multics)
		}

		for _, v := range multics {
			i++
			// log.Printf("krait.aliasSort() | i: %d (Multics)\n", i)
			s[i] = v
		}
	}

	if len(gnu) > 0 {
		if len(gnu) > 1 {
			sort.Strings(gnu)
		}

		for _, v := range gnu {
			i++
			// log.Printf("krait.aliasSort() | i: %d (GNU)\n", i)
			s[i] = v
		}
	}
}

func cmdHelp(fs *FlagSet, args ...string) {
	// log.Printf("krait.cmdHelp() | args: %q | fs: %s\n", args, fs)
	// log.Printf("krait.cmdHelp() | fs.cmd: %q | fs.HelpOutput: %v | args: %q\n", fs.cmd, fs.HelpOutput, args)
	// log.Printf("krait.cmdHelp() | fs.cmd: %q | args: %q\n", fs.cmd, args)

	if fs != nil && fs.HelpOutput != nil {
		if len(args) > 0 {
			fs.HelpOutput(fs, args...)
		} else {
			fs.HelpOutput(fs)
		}
	}
	os.Exit(0)
}

func cmdVersion(fs *FlagSet, args ...string) {
	// log.Printf("krait.cmdVersion() | args: %q | fs: %s\n", args, fs)

	fmt.Fprintln(flag.CommandLine.Output(), fs.getRoot().AppLabel)

	os.Exit(0)
}

func helpOutput(fs *FlagSet, args ...string) {
	// log.Printf("krait.helpOutput() | args: %q | fs: %s\n", args, fs)
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
		// log.Printf("krait.helpOutput() | rfs: %s\n", rfs)
		fmt.Println()
		fmt.Fprintln(flag.CommandLine.Output(), rfs.AppLabel) // Initial help output is the Application Name and Version AKA Label

		// Output for all commands
		var (
			aliases       = make(map[string][]string)
			widestCommand int
		)

		cmdList := make([][]string, 1)
		for i := range cmdList {
			// log.Printf("krait.helpOutput() | len(cmdList): %d\n", len(cmdList))
			cmdList[i] = make([]string, 1)
		}
		// log.Printf("krait.helpOutput() | cmdList: %#v\n", cmdList)

		for lvl := range rfs.subcommands {
			// log.Printf("krait.helpOutput() | lvl: %d\n", lvl)
			for subCmdName := range fs.parent.subcommands[lvl] {
				// log.Printf("krait.helpOutput() | %d | subCmdName: %q | subCmdFS.flagSet.Name(): %q, (%d)\n", lvl, subCmdName, subCmdFS.flagSet.Name(), len(subCmdFS.flagSet.Name()))
				// log.Printf("krait.helpOutput() | %d | subCmdName: %q, (%d)\n", lvl, subCmdName, len(subCmdName))
				aliases[subCmdName] = fs.getSubCommandAliases(subCmdName)
				// log.Printf("krait.helpOutput() | %d | aliases[%q]: %q\n", lvl, subCmdName, aliases[subCmdName])
				if len(subCmdName) > widestCommand {
					widestCommand = len(subCmdName)

					// Check if there is a single alias and if there is it will
					// be added to the subcommand width + 2 for the comma space
					if len(aliases[subCmdName]) == 1 {
						widestCommand += 2 + len(aliases[subCmdName])
					}
				}
				for lvl >= len(cmdList) {
					cmdList = append(cmdList, []string{})
				}
				cmdList[lvl] = append(cmdList[lvl], subCmdName)
				// log.Printf("krait.helpOutput() | lvl: %d | len(cmdList): %d\n", lvl, len(cmdList))
			}
			sort.Strings(cmdList[lvl])
			// for i, subCmdName := range cmdList[lvl] {
			// 	log.Printf("krait.helpOutput() | %d:%d | subCmdName: %q\n", lvl, i, subCmdName)
			// }
			// for subcmd, aliases := range aliases {
			// 	log.Printf("krait.helpOutput() | %d | subcmd: %q | aliases: %q\n", lvl, subcmd, aliases)
			// }
		}

		format := fmt.Sprintf("  %%-%ds  %%s\n", widestCommand+3)
		// format := fmt.Sprintf("  %%-%ds  %%s", widestCommand)
		// formatOptionDefault := fmt.Sprintf("  %%-%ds  %%s (default: %%v)\n", widestCommand)
		// formatOptionNodefault := fmt.Sprintf("  %%-%ds  %%s (no default)\n", widestCommand)

		// Commands
		// log.Printf("krait.helpOutput() | format: %q\n", format)
		// log.Printf("krait.helpOutput() | cmdList: %q\n", cmdList)
		fmt.Fprintln(flag.CommandLine.Output())

		fmt.Fprintln(flag.CommandLine.Output(), summeryTitle)
		for lvl := range cmdList {
			for _, cmdName := range cmdList[lvl] {
				if len(cmdName) > 0 {
					// log.Printf("krait.helpOutput() | cmdName: %q\n", cmdName)
					subCmdFS := rfs.subcommands[lvl][cmdName]
					// helpLine := fmt.Sprintf(format, subCmdFS.flagSet.Name(), subCmdFS.Summery)

					if len(aliases[cmdName]) == 1 {
						cmdName += ", " + aliases[cmdName][0]
					}
					fmt.Fprintf(flag.CommandLine.Output(), format, cmdName, subCmdFS.Summery)

					// subCmdFS.flagSet.Usage()
					// aliases := rfs.subcmdAliases[lvl][cmdName]
					// aliases := fs.getSubCommandAliases(cmdName)
					// log.Printf("krait.helpOutput() | rfs.subcmdAliases[%d][%q] = aliases: %q\n", lvl, cmdName, aliases)
					if len(aliases[cmdName]) > 1 {
						aliasList := ""
						for _, alias := range aliases[cmdName] {
							aliasList += ", " + alias
						}
						aliasList = aliasList[2:]
						fmt.Fprintf(flag.CommandLine.Output(), "      aliases: %s\n", aliasList)
					}
					// fmt.Fprintf(flag.CommandLine.Output(), helpLine)
				}
			}
		}

		/*

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
			// log.Printf("krait.helpOutput() | fs.cmd: %q | fs.AppLabel: %q | fs.Epilogue: %q\n", fs.cmd, fs.AppLabel, fs.Epilogue)
			// log.Printf("krait.helpOutput() | fs.parent.cmd: %q | fs.parent.Epilogue: %q\n", fs.parent.cmd, fs.parent.Epilogue)
			if fs.Epilogue != "" {
				fmt.Fprintf(flag.CommandLine.Output(), fs.Epilogue)
				fmt.Fprintln(flag.CommandLine.Output())
			} else {
				parent := fs.parent
				for parent != nil {
					// log.Printf("krait.helpOutput() | parent.cmd: %q\n", parent.cmd)
					if parent.Epilogue != "" {
						fmt.Fprintf(flag.CommandLine.Output(), parent.Epilogue)
						fmt.Fprintln(flag.CommandLine.Output())
						break
					}
					parent = parent.parent
				}
			}
		*/
	} else {
		for i := range fs.parent.subcommands {
			for _, subcmd := range fs.parent.subcommands[i] {
				// log.Printf("krait.helpOutput() | subcmd.flagSet.Name(): %q\n", subcmd.flagSet.Name())
				if subcmd.flagSet.Name() == cmdName {
					// log.Printf("krait.helpOutput() | cmdName match: %q\n", cmdName)
					subcmd.flagSet.Usage()
					// flag.PrintDefaults()
					// subcmd.flagSet.PrintDefaults()

					// fmt.Fprintf(flag.CommandLine.Output(), subcmd.Summery)
					break
				}
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

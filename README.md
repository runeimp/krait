Krait v0.0.0
============

Full featured CLI library to take command of all subcommand line input


Features
--------

* Argument parsing
* Option parsing
	* POSIX single letter options and option grouping with a single hyphen
	* GNU word based option prefixed with a double hyphen
* Subcommand parsing


### Recursive Argument Parsing

```go
args := ["tester", "test", "-c" "3", "two", "three"]

// 1st call
func (fs *FlagSet) Parse(args []string) (subcmd string, err error) {
	// ["tester", "test", "-c" "3", "two", "three"]
	subcmd := args[0] // "tester" the command AKA the command line name of the application

	if len(args) > 1 {
		if isASubCmd(args[1]) == true {
			fs.Parse(args[1:])
		} else {
			// handle options and command argument if any
		}
	}


	return subcmd
}

// 2nd call
func Parse(args []string) {
	// ["test", "-c" "3", "two", "three"]
	subcmd := args[0]         // "test" the name of the subcommand
	optionName := args[1]  // "-c"
	optionValue := args[2] // 3
	arg1 := args[3]        // "two"
	arg2 := args[4]        // "three"

	return subcmd
}
```



Limitations
-----------

* There can be no options or bare arguments between successive subcommands


Is It an Argument to Have or Own?
---------------------------------

* All bare arguments (no prefix or suffix) for the root FlagSet (`FlagSet.ParentName() == "nil"`) must be sub-commands only
* The root FlagSet (application command) should not take any bare arguments
* A bare arguments immediately following a subcommand may be a subcommand or an arguments
* A subcommand that accepts bare arguments should not also accept subcommands as it is far to easy to mess up unless the bare arguments must be part of a specific set of allowed arguments that do not overlap with subcommands, which is still dodgy at best
* Everything that follows `--` is an argument even it if looks like a valid subcommand

1. Check if an argument is an option or "bare" argument
	1. Bare
		1. Check if the current command (application command or subcommand) has subcommands available
		2. If subcommands are a valid type and the argument is "bare", check if the argument is a valid subcommand
		3. If the bare argument is not a valid subcommand add to the argument list
	2. Option
		1. If an option is a binary flag bare arguments after it may be subcommands or arguments
		2. If an option requires an option-argument then the argument after it or attached to it via an equal sign or as a suffix to it is be the value for the option-argument and may then be followed by an option, argument, or subcommand
		3. If an option accepts an optional option-argument then if the argument is to follow it must be connected via an equal sign or as a suffix of the option and may then be followed by an option, argument, or subcommand



Logic
=====

Is It an Argument to Have or Own?
---------------------------------

* All bare arguments (no prefix or suffix) for the root FlagSet (`FlagSet.parent == nil`) must be sub-commands only
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
		1. If an option is a binary flag bare arguments after it may be subcommands or arguments -- not yet implemented
		2. If an option requires an option-argument then the argument after it or attached to it via an equal sign or as a suffix to it is be the value for the option-argument and may then be followed by an option, argument, or subcommand -- not yet implemented
		3. If an option accepts an optional option-argument then if the argument is to follow it must be connected via an equal sign or as a suffix of the option and may then be followed by an option, argument, or subcommand -- not yet implemented


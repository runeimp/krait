Krait v0.2.0
============

Go flag package wrapper for simplified subcommand interface building. Sort of like a Cobra lite. I have no idea if this produces lighter binaries. But is possibly lighter in setup?


Features
--------

* Argument parsing
* Options
	* Prefix Support
		* POSIX single letter options and option grouping with a single hyphen
		* POSIX option grouping not supported
		* GNU long options prefixed with a double hyphen
		* Multics style long options with a single hyphen prefix
	* Data Types
		* `bool`
		* `float64`
		* `int`
		* `string`
		* `uint`
* Subcommand parsing
	* Almost infinite levels of subcommands
	* Subcommand of a subcommand can have the same name


### Basic Example with Recursive Argument Parsing


```go
package main

import (
	"fmt"

	"github.com/runeimp/krait"
)

// testFunc is a basic example Subcommand Function handler
func testFunc(fs *krait.FlagSet, args ...string) {
	fmt.Printf("testFunc was called with %d parameters %q\n", len(args), args)
}

func main() {

	// Root command / FlagSet
	cli := krait.NewFlagSet("myapp")
	cli.AppLabel = "MyApp v0.1.0"

	testFS := cli.NewFlagSet("test") // 1st level subcommand for the application
	testFS.CmdFunc = testFunc
	testFS.Summery = "Tests the basic usage of Krait"

	aliases := []string{"c", "count"}
	testCountOption := testFS.OptionInt(aliases, 0, "What number will invoke 'The Count'")

	subcmd, err := cli.Parse()
	// The default help and version systems will exit the app at this point if called

	fmt.Println()
	fmt.Printf("cli.SubCommand(): %q | subcmd: %q | error: %v\n", cli.SubCommand(), subcmd, err)
	fmt.Printf("*testCountOption: %#v\n", *testCountOption)
	fmt.Printf("Argument Count: %d | All arguments: %q\n", len(cli.Args()), cli.Args())
	fmt.Println()
}
```

If the example command above is run with no arguments or options the output would be the default help...

```text
MyApp v0.1.0

COMMAND SUMMERY
---------------
  help        Displays this help information
  test        Tests the basic usage of Krait
  version     Displays the app name and version

```


Limitations
-----------

* There can be no options or bare arguments between successive subcommands


ToDo
----

* Add more option types to better match the `flag` package and possibly go beyond
	* Duration: planned for v1.0.0
	* Int64: planned for v1.0.0
	* Uint64: planned for v1.0.0
	* Func: planned for v1.0.0
	* Date: ISO-8601 and possibly others
	* DateTime: ISO-8601 and possibly others
	* Time: ISO-8601
	* UNIX Timestamp: an int with attitude?



Advanced Example with Recursive Argument Parsing
------------------------------------------------

```go
package main

import "github.com/runeimp/krait"

// jabberwocky is an example Subcommand Function handler that is triggered by
// its subcommand on the command line
func jabberwocky(fs *krait.FlagSet, args ...string) {
	if len(args) > 0 {
		// Do something with the args
	}

	// Access the FlagSet in some way, such as...
	fmt.Println(fs.Summery)
}

// testFunc is a basic example Subcommand Function handler
func testFunc(fs *krait.FlagSet, args ...string) {
	fmt.Printf("testFunc was called with %d arguments\n", len(args))
}

func main() {

	// Root command / FlagSet
	cli := krait.NewFlagSet("myapp")
	cli.AppLabel = "MyApp v0.1.0"

	// The subcommand to use if one is not provided on the command line.
	// The normal default is the internal "help" subcommand. So this is
	// really just for special setups.
	cli.DefaultSubCommand = "jabberwock"

	cli.Epilogue = `The main help epilogue...

This is not required!
`

	jFS := cli.NewFlagSet("jabberwock") // 1st level subcommand for the application
	jFS.CmdFunc = jabberwocky
	jFS.Summery = "The jaws that bite, the claws that catch!"

	testFS := cli.NewFlagSet("test") // 1st level subcommand for the application
	testFS.CmdFunc = testFunc
	testFS.Summery = "Tests the basic usage of Krait"

	// The longest name becomes the key in testFS.Options map
	aliases := []string{"c", "count", "rep"}

	// Capturing "testCountOption" is not necessary
	testCountOption := testFS.OptionInt(aliases, 0, "What number will invoke 'The Count'")

	testOne := testFS.NewFlagSet("one") // 2nd level subcommand for the test subcommand
	testOne.CmdFunc = testFunc // Reusing subcommand handlers is reasonable
	testFS.Summery = "The first named test"

	fmt.Println()

	// commandLine := ["myapp", "test", "-c" "3", "two", "three"]
	subcmd, err := cli.Parse()    // The "commandLine" variable above could be passed for testing purposes
	subcmd, err := testFS.Parse() // This would panic as you should only call Parse on the root command/FlagSet

	fmt.Println()

	fmt.Printf("cli.SubCommand(): %q | subcmd: %q | error: %v\n", cli.SubCommand(), subcmd, err)
	fmt.Printf("*cliCountOption: %d (should be zero if the subcommand test was called as --count is an option for test and not root on the command line)\n", *cliCountOption)

	// This shows accessing the option value in a traditional way but is unnecessary
	fmt.Printf("*testCountOption: %#v\n", *testCountOption)

	// This shows accessing the option in a more dynamic (as needed) way
	countInt, err := testFS.Options["count"].GetInt()
	fmt.Printf("countInt: %d | err: %v\n", countInt, err)

	// Accessing the bare arguments of the command line
	fmt.Printf("Argument Count: %d | All arguments: %q\n", len(cli.Args()), cli.Args())

	fmt.Println()

}
```
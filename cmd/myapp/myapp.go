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

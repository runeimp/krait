package krait

import (
	"flag"
	"testing"
)

func kraitTestFunction(fs *FlagSet, args ...string) {
	//
}

// TestNewRootKrait ensures Krait creation works
func TestNewRootKrait(t *testing.T) {
	want := flag.NewFlagSet("root", flag.ExitOnError)
	root := NewFlagSet("root")
	got := root.flagSet

	if got.Name() != want.Name() {
		t.Fatalf("got: %#v\n             want: %#v", got, want)
	}
}

// TestNewCreateKrait ensures Krait creation works
func TestNewCreateKrait(t *testing.T) {
	want := flag.NewFlagSet("create", flag.ExitOnError)
	got := NewFlagSet("create").flagSet

	if got.Name() != want.Name() {
		t.Fatalf("got: %q | want: %q", got.Name(), want.Name())
	}
}

// TestRootParse ensures Krait.Parse() works
func TestRootParse(t *testing.T) {
	want := "test"

	args := []string{"krait", want, "two", "three"}

	root := NewFlagSet("krait")
	// root.NewFlagSet(want, kraitTestFunction) // ExitOnError is the default
	fs := root.NewFlagSet(want) // ExitOnError is the default
	fs.CmdFunc = kraitTestFunction
	got, _ := root.Parse(args)

	if got != want {
		t.Fatalf("got: %q | want: %q", got, want)
	}
}

// TestOptionBool ensure boolean option parsing works
func TestOptionBool(t *testing.T) {
	want := true
	args := []string{"krait", "test", "--confirm", "one", "two", "three"}
	optionAliases := []string{"c", "confirm"}

	root := NewFlagSet("root")
	// root.NewFlagSet("test", kraitTestFunction) // ExitOnError is the default
	fs := root.NewFlagSet("test") // ExitOnError is the default
	fs.CmdFunc = kraitTestFunction
	confirm := fs.OptionBool(optionAliases, false, "Should we confirm the processing?")
	root.Parse(args)

	got := *confirm

	if got != want {
		t.Fatalf("got: %t | want: %t", got, want)
	}
}

// TestOptionFloat ensure int option parsing works
func TestOptionFloat(t *testing.T) {
	want := 1.1
	args := []string{"krait", "test", "--count", "1.1", "two", "three"}
	optionAliases := []string{"c", "count"}

	root := NewFlagSet("root")
	// root.NewFlagSet("test", kraitTestFunction) // ExitOnError is the default
	testFS := root.NewFlagSet("test")
	testFS.CmdFunc = kraitTestFunction
	count := testFS.OptionFloat(optionAliases, 0.0, "What number will invoke 'The Count'")
	root.Parse(args)

	got := *count

	if got != want {
		t.Fatalf("got: %f | want: %f", got, want)
	}
}

// TestOptionInt ensure int option parsing works
func TestOptionInt(t *testing.T) {
	want := 1
	args := []string{"krait", "test", "--count", "1", "two", "three"}
	optionAliases := []string{"c", "count"}

	root := NewFlagSet("root")
	// root.NewFlagSet("test", kraitTestFunction) // ExitOnError is the default
	testFS := root.NewFlagSet("test")
	testFS.CmdFunc = kraitTestFunction
	count := testFS.OptionInt(optionAliases, 0, "What number will invoke 'The Count'")
	root.Parse(args)

	got := *count

	if got != want {
		t.Fatalf("got: %d | want: %d", got, want)
	}
}

// TestOptionUint1 ensure uint option parsing works
func TestOptionUint1(t *testing.T) {
	var want uint = 1
	args := []string{"krait", "test", "--count", "1", "two", "three"}
	optionAliases := []string{"c", "count"}

	root := NewFlagSet("root")
	// root.NewFlagSet("test", kraitTestFunction) // ExitOnError is the default
	testFS := root.NewFlagSet("test")
	testFS.CmdFunc = kraitTestFunction
	count := testFS.OptionUint(optionAliases, 0, "What number will invoke 'The Count'")
	root.Parse(args)

	got := *count

	if got != want {
		t.Fatalf("got: %d | want: %d", got, want)
	}
}

/*
// TestOptionUint2 ensure float to uint option parsing works
func TestOptionUint2(t *testing.T) {
	var want uint = 1
	args := []string{"krait", "test", "--count", "1.0", "two", "three"}
	optionAliases := []string{"c", "count"}

	root := NewFlagSet("root")
	// root.NewFlagSet("test", kraitTestFunction) // ExitOnError is the default
	testFS := root.NewFlagSet("test")
	testFS.CmdFunc = kraitTestFunction
	count := testFS.OptionUint(optionAliases, 0, "What number will invoke 'The Count'")
	root.Parse(args)

	got := *count

	if got != want {
		t.Fatalf("got: %d | want: %d", got, want)
	}
}
*/

func TestOptionUintFromInt(t *testing.T) {
	var want uint = 1
	args := []string{"krait", "test", "--count", "1", "two", "three"}
	optionAliases := []string{"c", "count"}

	root := NewFlagSet("root")
	// root.NewFlagSet("test", kraitTestFunction) // ExitOnError is the default
	testFS := root.NewFlagSet("test")
	testFS.CmdFunc = kraitTestFunction
	testFS.OptionInt(optionAliases, 0, "What number will invoke 'The Count'")
	root.Parse(args)

	count, err := testFS.Options["count"].GetUint()
	if err != nil {
		t.Fatalf("error %d", err)
	}

	got := count

	if got != want {
		t.Fatalf("got: %d | want: %d", got, want)
	}
}

// TestOptionAliases1 ensure int option parsing works with aliases
func TestOptionAliases1(t *testing.T) {
	want := 2
	args := []string{"krait", "test", "-c", "2"}
	optionAliases := []string{"c", "count"}

	root := NewFlagSet("root")
	testFS := root.NewFlagSet("test") // ExitOnError is the default
	testFS.CmdFunc = kraitTestFunction
	count := testFS.OptionInt(optionAliases, 0, "What number will invoke 'The Count'")
	// count := root.OptionInt(optionAliases, 0, "What number will invoke 'The Count'")
	root.Parse(args)

	got := *count

	if got != want {
		t.Fatalf("got: %d | want: %d", got, want)
	}
}

// TestOptionAliases2 ensure int option parsing works with aliases
func TestOptionAliases2(t *testing.T) {
	want := 3
	args := []string{"krait", "test", "-count", "3"}
	optionAliases := []string{"c", "count"}

	root := NewFlagSet("root")
	testFS := root.NewFlagSet("test") // ExitOnError is the default
	testFS.CmdFunc = kraitTestFunction
	count := testFS.OptionInt(optionAliases, 0, "What number will invoke 'The Count'")
	// count := root.OptionInt(optionAliases, 0, "What number will invoke 'The Count'")
	root.Parse(args)

	got := *count

	if got != want {
		t.Fatalf("got: %d | want: %d", got, want)
	}
}

package ff_test

import (
	"fmt"
	"os"
	"time"

	"github.com/peterbourgon/ff"
)

func ExampleFlagSet_Parse() {
	os.Setenv("EXAMPLE_FOO", "foo_from_env")
	os.Setenv("EXAMPLE_BAR", "2s")
	os.Setenv("EXAMPLE_BAZ", "baz_from_env")
	args := []string{"--foo=xyz", "-bar", "3s"}

	fs := ff.NewFlagSet("example [flags]")
	var (
		foo = fs.String("foo", "default_value", "a string variable", ff.Env("EXAMPLE_FOO"))
		bar = fs.Duration("bar", time.Second, "a duration variable", ff.Env("EXAMPLE_BAR"))
		baz = fs.String("baz", "", "another string variable", ff.Env("EXAMPLE_BAZ"))
	)
	if err := fs.Parse(args, ff.FromEnvironment("EXAMPLE_")); err != nil {
		panic(err)
	}

	fmt.Println("foo", *foo)
	fmt.Println("bar", bar.String())
	fmt.Println("baz", *baz)
	// Output:
	// foo xyz
	// bar 3s
	// baz baz_from_env
}

package fftest_test

import (
	"fmt"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/fftest"
)

func ExampleUnindentString() {
	fs := ff.NewFlagSet("testcommand")
	fs.String('f', "foo", "", "the foo parameter")
	fs.IntLong("bar", 3, "the bar parameter")

	want := fftest.UnindentString(`
		NAME
		  testcommand

		FLAGS
		  -f, --foo STRING   the foo parameter
		      --bar INT      the bar parameter (default: 3)
	`)

	have := fftest.UnindentString(ffhelp.Flags(fs).String())

	if want == have {
		fmt.Println("strings are identical")
	} else {
		fmt.Println(fftest.DiffString(want, have))
	}

	// Output:
	// strings are identical
}

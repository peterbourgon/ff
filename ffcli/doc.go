/*

Package ffcli is for building declarative commandline applications.

Rationale

This package is intended to be a lightweight alternative to common commandline
application helper packages like https://github.com/spf13/cobra and
https://github.com/alecthomas/kingpin. In contrast to those packages, the API
surface area of package ffcli is very small, with the immediate goal of being
intuitive and productive, and the long-term goal of supporting commandline
applications that are substantially easier to understand and maintain.

To support these goals, the package is concerned only with the core mechanics of
defining a command tree, parsing flags, and selecting a command to run. It does
not intend to be a one-stop-shop for everything your commandline application
needs. Features like tab completion or colorized output are orthogonal to
command tree parsing, and are easy to provide on top of ffcli.

Finally, this package follows in the philosophy of its parent package ff, or
"flags-first". Flags, and more specifically the Go stdlib flag.FlagSet, should
be the primary mechanism of getting configuration from the execution environment
into your program. The affordances provided by package ff, including environment
variable and config file parsing, are also available in package ffcli. Support
for other flag packages is a non-goal.

Dependency relationships

One important design principle of this package is that commandline applications
aren't fundamentally different from any other type of application. Their
component graphs, and the dependency relationships between their parts, don't
require special affordances beyond what is already commonly available in the
language.

Concretely, if an Exec function depends on a flag, in the simple case, it can
reference that flag directly from the outer scope.

    func main() {
        var (
            rootFlagSet = flag.NewFlagSet("root", flag.ExitOnError)
            token       = rootFlagSet.String("token", "", "API token")
            verbose     = rootFlagSet.Bool("verbose", false, "increase log verbosity")
        )

        root := &ffcli.Command{
            FlagSet: rootFlagSet,
            Exec: func(ctx context.Context, args []string) error {
                println("token", *token, "verbose", *verbose)
                return nil
            },
        }

Another option is to capture dependencies with a function closure.

    func rootExec(token string, verbose bool) func(context.Context, []string) error {
        return func(ctx context.Context, args []string) error {
            println("token", token, "verbose", verbose)
            return nil
        }
    }

    func main() {
            var (
                rootFlagSet = flag.NewFlagSet("root", flag.ExitOnError)
                token       = rootFlagSet.String("token", "", "API token")
                verbose     = rootFlagSet.Bool("verbose", false, "increase log verbosity")
            )

            root := &ffcli.Command{
                FlagSet: rootFlagSet,
                Exec:    rootExec(*token, *verbose),
            }

Still another option is to capture dependencies with types.

    package rootcmd

    type Config struct {
        token   string
        verbose bool
    }

    func NewConfig(target *flag.FlagSet) *Config {
        var c Config
        target.StringVar(&c.token, "token", "", "API token")
        target.BoolVar(&c.verbose, "verbose", false, "increase log verbosity")
        return &c
    }

    func (c *Config) Exec(ctx context.Context, args []string) error {
        println("token", c.token, "verbose", c.verbose)
        return nil
    }

    package main

    func main() {
        var (
            rootFlagSet = flag.NewFlagSet("root", flag.ExitOnError)
            rootConfig  = rootcmd.NewConfig(rootFlagSet)
        )

        root := &ffcli.Command{
            FlagSet: rootFlagSet,
            Exec:    rootConfig.Exec,
        }

There is no one-size-fits-all solution. The right answer depends on the
requirements of your application. Package ffcli is designed to be flexible
enough to express sophisticated application architectures, while remaining
simple enough to be immediately intuitive at even a very small scale. See the
examples directory for several applications of differing complexity.

Testing

Just as CLIs aren't fundamentally different from other applications regarding
dependencies and component graphs, neither are they fundamentally different
regarding testing.

Commands can be made more testable by wrapping their construction in a function
which takes all of the dependencies, like flag.FlagSets or specific values, as
parameters. Providing an input io.Reader and output io.Writer instead of using
os.Stdin and os.Stdout directly allows you to effectively test Exec functiopns.

See objectctl in the examples directory for tips.

*/
package ffcli

// Package textcli implements the stdin/stdout text conversion commands.
package textcli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"nimbleterminal.dev/ntemoji"
)

// version is what --version prints. Release builds set it with
// -ldflags "-X nimbleterminal.dev/ntemoji/internal/textcli.version=v1.2.3".
var version = "dev"

// Run returns a process exit code: 0 for success, 2 for usage, 1 for I/O errors.
func Run(reverse bool, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	name := "demoji"
	before, after := "", ""
	if reverse {
		name, before, after = "remoji", ":", ":"
	}
	flags := pflag.NewFlagSet(name, pflag.ContinueOnError)
	flags.SetOutput(stderr)
	locale := flags.String("locale", "en", "CLDR metadata language; unsupported locales fall back to English")
	flags.StringVar(&before, "open", before, "opening name delimiter")
	flags.StringVar(&after, "close", after, "closing name delimiter")
	help := flags.BoolP("help", "h", false, "show usage and exit")
	list := flags.Bool("list-locales", false, "list bundled language codes and exit")
	showVersion := flags.BoolP("version", "V", false, "print the version and exit")
	flags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s [options] [file|-]\nReads stdin when no file is given; writes converted text to stdout.\n", name)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		flags.Usage()
		return 2
	}
	if *help {
		flags.Usage()
		return 0
	}
	if *showVersion {
		fmt.Fprintln(stdout, name+" "+version)
		return 0
	}
	if flags.NArg() > 1 {
		flags.Usage()
		return 2
	}
	if *list {
		if _, err := fmt.Fprintln(stdout, strings.Join(ntemoji.Locales(), "\n")); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	input := stdin
	if flags.NArg() == 1 && flags.Arg(0) != "-" {
		file, err := os.Open(flags.Arg(0))
		if err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", name, err)
			return 1
		}
		defer file.Close()
		input = file
	}
	opts := []ntemoji.TextOption{ntemoji.WithTextLocale(*locale), ntemoji.WithDelimiters(before, after)}
	var err error
	if reverse {
		err = ntemoji.Remoji(stdout, input, opts...)
	} else {
		err = ntemoji.Demoji(stdout, input, opts...)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}
	return 0
}

// Demoji replaces emoji in stdin or a file with descriptive names.
package main

import (
	"os"

	"nimbleterminal.dev/ntemoji/internal/textcli"
)

func main() { os.Exit(textcli.Run(false, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

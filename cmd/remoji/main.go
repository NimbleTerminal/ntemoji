// Remoji replaces delimited descriptive names in stdin or a file with emoji.
package main

import (
	"os"

	"nimbleterminal.dev/ntemoji/internal/textcli"
)

func main() { os.Exit(textcli.Run(true, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

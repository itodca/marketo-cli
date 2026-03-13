package main

import (
	"os"

	"github.com/itodca/marketo-cli/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}

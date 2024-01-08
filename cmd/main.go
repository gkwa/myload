package main

import (
	"os"

	"github.com/taylormonacelli/myload"
)

func main() {
	code := myload.Execute()
	os.Exit(code)
}

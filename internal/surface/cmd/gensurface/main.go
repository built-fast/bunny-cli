package main

import (
	"fmt"

	"github.com/built-fast/bunny-cli/cmd"
	"github.com/built-fast/bunny-cli/internal/surface"
)

func main() {
	root := cmd.NewRootCmd()
	fmt.Print(surface.Generate(root))
}

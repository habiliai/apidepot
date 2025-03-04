package main

import (
	"context"
	"fmt"
	"github.com/habiliai/apidepot/pkg/cli/apidepotctl"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := apidepotctl.Execute(ctx); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error: %+v\n", err))
		os.Exit(1)
	}
}

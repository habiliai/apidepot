package main

import (
	"context"
	"fmt"
	"github.com/habiliai/apidepot/pkg/cli/apidepot"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := apidepot.Execute(ctx); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error: %+v\n", err))
		os.Exit(1)
	}
}

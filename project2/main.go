package main

import (
	"fmt"
	cobra2 "github.com/spf13/cobra"
	"os"

	"computer_vision/project2/cmd"
)

func main() {
	var root cobra2.Command

	root.AddCommand(
		cmd.EnlargeImage(),
		cmd.AddTextureToImage(),
		)

	if err := root.Execute(); err != nil {
		fmt.Printf("error %v", err)
		os.Exit(1)
	}
}


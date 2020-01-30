package main

import (
	"fmt"
	"os"

	"github.com/reservoird/reservoird/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

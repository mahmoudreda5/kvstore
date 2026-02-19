package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: kvstore <data-dir>")
		return
	}

	dataDir := os.Args[1]
	fmt.Println("opening store at:", dataDir)

	_ = dataDir // placeholder until we implement Store
}

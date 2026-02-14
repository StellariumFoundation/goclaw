package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	fmt.Printf("GoClaw — AI Digital Worker v%s\n", version)
	fmt.Println("A Golang rewrite of Openclaw")
	fmt.Println()

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("goclaw version %s\n", version)
		return
	}

	fmt.Println("Starting GoClaw AI Digital Worker...")
	fmt.Println("Ready to process tasks.")
}

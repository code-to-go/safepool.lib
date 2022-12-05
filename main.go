package main

import (
	_ "embed"
	"fmt"

	"github.com/code-to-go/safepool.lib/api"
)

func main() {
	api.Start()
	fmt.Print("This is just a library! ")

}

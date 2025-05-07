package main

import (
	"os"
	"fmt"

	"github.com/ipfs/kubo/cmd/ipfs/kubo"
)

func main() {
	fmt.Println("YW: Running my custom IPFS build")
	os.Exit(kubo.Start(kubo.BuildDefaultEnv))
}

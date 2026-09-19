package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/The127/miso/internal/cachedisk"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: cachedisk PATH SIZE")
		os.Exit(2)
	}

	size, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := cachedisk.Make(os.Args[1], size); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

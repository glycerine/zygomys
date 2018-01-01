/*
The zygomys command line REPL is known as `zygo`.
*/
package main

import (
	"flag"
	"fmt"
	"github.com/glycerine/zygomys/zygo"
	"os"
)

func usage(myflags *flag.FlagSet) {
	fmt.Printf("zygo command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	cfg := zygo.NewZlispConfig("zygo")
	cfg.DefineFlags()
	err := cfg.Flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		usage(cfg.Flags)
	}

	if err != nil {
		panic(err)
	}
	err = cfg.ValidateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "zygo command line error: '%v'\n", err)
		usage(cfg.Flags)
	}

	// the library does all the heavy lifting.
	zygo.ReplMain(cfg)
}

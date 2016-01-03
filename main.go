package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glycerine/glisp/interpreter"
)

func usage(myflags *flag.FlagSet) {
		fmt.Printf("glisp command line help:\n")
		myflags.PrintDefaults()
		os.Exit(1)
}

func main() {
	myflags := flag.NewFlagSet("glisp", flag.ExitOnError)
	var cfg glisp.GlispConfig
	cfg.DefineFlags(myflags)
	err := myflags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		usage(myflags)
	}
	if err != nil {
		panic(err)
	}
	err = cfg.ValidateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "glisp command line error: '%v'\n", err)
		usage(myflags)
	}
	glisp.ReplMain(&cfg)
}

package main

import (
	"flag"
	"fmt"
	"github.com/glycerine/zygomys/repl"
	"os"
)

func usage(myflags *flag.FlagSet) {
	fmt.Printf("zygo command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	cfg := zygo.NewGlispConfig("zygo")
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

	registerExts := func(env *zygo.Glisp) {
		// this mechanism not used at the moment, but the
		// syntax would be: zygoext.ImportRandom(env)
	}
	//cfg.ExtensionsVersion = zygoext.Version()
	zygo.ReplMain(cfg, registerExts)
}

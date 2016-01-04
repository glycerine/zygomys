package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glycerine/glisp/extensions"
	"github.com/glycerine/glisp/interpreter"
)

func usage(myflags *flag.FlagSet) {
	fmt.Printf("glisp command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	cfg := glisp.NewGlispConfig("glisp")
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
		fmt.Fprintf(os.Stderr, "glisp command line error: '%v'\n", err)
		usage(cfg.Flags)
	}

	registerExts := func(env *glisp.Glisp) {
		glispext.ImportRandom(env)
		glispext.ImportChannels(env)
		glispext.ImportCoroutines(env)
		glispext.ImportRegex(env)
	}
	cfg.ExtensionsVersion = glispext.Version()
	glisp.ReplMain(cfg, registerExts)
}

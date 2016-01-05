package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glycerine/godiesel/extensions"
	"github.com/glycerine/godiesel/interpreter"
)

func usage(myflags *flag.FlagSet) {
	fmt.Printf("gdsl command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	cfg := gdsl.NewGlispConfig("gdsl")
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
		fmt.Fprintf(os.Stderr, "gdsl command line error: '%v'\n", err)
		usage(cfg.Flags)
	}

	registerExts := func(env *gdsl.Glisp) {
		gdslext.ImportRandom(env)
		gdslext.ImportChannels(env)
		gdslext.ImportCoroutines(env)
		gdslext.ImportRegex(env)
	}
	cfg.ExtensionsVersion = gdslext.Version()
	gdsl.ReplMain(cfg, registerExts)
}

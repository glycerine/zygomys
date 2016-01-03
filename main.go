package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glycerine/glisp/interpreter"
	"github.com/glycerine/glisp/extensions"
)

func usage(myflags *flag.FlagSet) {
	fmt.Printf("glisp command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	myflags := flag.NewFlagSet("glisp", flag.ExitOnError)
	cfg := &glisp.GlispConfig{}
	cfg.DefineFlags(myflags)
	//fmt.Printf("Args = %#v\n", os.Args)
	err := myflags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		//fmt.Printf("\n ErrHelp returned from Parse()\n")
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

	registerExts := func(env *glisp.Glisp) {
		glispext.ImportRandom(env)
		glispext.ImportTime(env)
		glispext.ImportChannels(env)
		glispext.ImportCoroutines(env)
		glispext.ImportRegex(env)
	}
	glisp.ReplMain(cfg, myflags, registerExts)
}

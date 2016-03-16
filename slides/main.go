/*
The zygomys command line REPL is known as `zygo`.
*/
package main

// START OMIT

import (
...
	zygo "github.com/glycerine/zygomys/repl"
)

func main() {
	// (1) configure it
	// See library configuration convention: https://github.com/glycerine/configs-in-golang
	cfg := zygo.NewGlispConfig("zygo")

	// (2) register your Go data types
	// here we register snoopy as a handle to Go struct &Snoopy{}
	zygo.GoStructRegistry.RegisterUserdef("snoopy",
		&zygo.RegisteredType{GenDefMap: true, Factory: func(env *zygo.Glisp) (interface{}, error) {
			return &Snoopy{}, nil
		}}, true)
	
	// (3) run the zygo repl
	// -- the library does all the heavy lifting.
	zygo.ReplMain(cfg)
}

// END OMIT

func usage(myflags *flag.FlagSet) {
	fmt.Printf("zygo command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

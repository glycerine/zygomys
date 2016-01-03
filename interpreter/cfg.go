package glisp

import (
	"flag"
)

type GlispConfig struct {
	CpuProfile string
	MemProfile    string
	ExitOnFailure bool
	CountFuncCalls bool
}

// call DefineFlags before myflags.Parse()
func (c *GlispConfig) DefineFlags(fs *flag.FlagSet) {	
	fs.StringVar(&c.CpuProfile, "cpuprofile", "", "write cpu profile to file")
	fs.StringVar(&c.MemProfile, "memprofile", "", "write mem profile to file")
	fs.BoolVar(&c.ExitOnFailure, "exitonfail", false, "exit on failure instead of starting repl")
	fs.BoolVar(&c.CountFuncCalls, "countcalls", false, "count how many times each function is run")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *GlispConfig) ValidateConfig() error {
	return nil
}

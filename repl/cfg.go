package zygo

import (
	"flag"
)

// configure a glisp repl
type GlispConfig struct {
	CpuProfile        string
	MemProfile        string
	ExitOnFailure     bool
	CountFuncCalls    bool
	Flags             *flag.FlagSet
	ExtensionsVersion string
}

func NewGlispConfig(cmdname string) *GlispConfig {
	return &GlispConfig{
		Flags: flag.NewFlagSet(cmdname, flag.ExitOnError),
	}
}

// call DefineFlags before myflags.Parse()
func (c *GlispConfig) DefineFlags() {
	c.Flags.StringVar(&c.CpuProfile, "cpuprofile", "", "write cpu profile to file")
	c.Flags.StringVar(&c.MemProfile, "memprofile", "", "write mem profile to file")
	c.Flags.BoolVar(&c.ExitOnFailure, "exitonfail", false, "exit on failure instead of starting repl")
	c.Flags.BoolVar(&c.CountFuncCalls, "countcalls", false, "count how many times each function is run")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *GlispConfig) ValidateConfig() error {
	return nil
}

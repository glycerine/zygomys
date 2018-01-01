package zygo

import (
	"flag"
)

// configure a glisp repl
type ZlispConfig struct {
	CpuProfile        string
	MemProfile        string
	ExitOnFailure     bool
	CountFuncCalls    bool
	Flags             *flag.FlagSet
	ExtensionsVersion string
	Command           string
	Sandboxed         bool
	Quiet             bool
	Trace             bool
	LoadDemoStructs   bool

	// liner bombs under emacs, avoid it with this flag.
	NoLiner bool
	Prompt  string // default "zlisp> "

}

func NewZlispConfig(cmdname string) *ZlispConfig {
	return &ZlispConfig{
		Flags: flag.NewFlagSet(cmdname, flag.ExitOnError),
	}
}

// call DefineFlags before myflags.Parse()
func (c *ZlispConfig) DefineFlags() {
	c.Flags.StringVar(&c.CpuProfile, "cpuprofile", "", "write cpu profile to file")
	c.Flags.StringVar(&c.MemProfile, "memprofile", "", "write mem profile to file")
	c.Flags.BoolVar(&c.ExitOnFailure, "exitonfail", false, "exit on failure instead of starting repl")
	c.Flags.BoolVar(&c.CountFuncCalls, "countcalls", false, "count how many times each function is run")
	c.Flags.StringVar(&c.Command, "c", "", "expressions to evaluate")
	c.Flags.BoolVar(&c.Sandboxed, "sandbox", false, "run sandboxed; disallow system/external interaction functions")
	c.Flags.BoolVar(&c.Quiet, "quiet", false, "start repl without printing the version/mode/help banner")
	c.Flags.BoolVar(&c.Trace, "trace", false, "trace execution (warning: very verbose and slow)")
	c.Flags.BoolVar(&c.LoadDemoStructs, "demo", false, "load the demo structs: Event, Snoopy, Hornet, Weather and friends.")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *ZlispConfig) ValidateConfig() error {
	if c.Prompt == "" {
		c.Prompt = "zlisp> "
	}
	return nil
}

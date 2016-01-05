package gdsl

import "fmt"

func (env *Glisp) ImportRequire() {
	env.AddMacro("req", RequireMacro)
}

// (req path) avoids the need to put quotes around path you are sourcing.
func RequireMacro(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("path to source missing. use: " +
			"(req path-to-source) ;;no deep to quote the path\n")
	}

	// (source "path")
	return MakeList([]Sexp{env.MakeSymbol("source"),
		SexpStr(args[0].(SexpSymbol).name)}), nil
}

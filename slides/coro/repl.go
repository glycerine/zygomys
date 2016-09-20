package main

// reference: github.com/glycerine/zygomys/repl/repl.go
//
// much detail elided for instructional purposes...

func main() {
	// env represents one interpreter:
	// It will have one parsing and one execution goroutine.
	env = NewGlisp()

	// cfg configures the env
	cfg := &GlsipConfig{
		// some of the options:
		ExitOnFailure: false, // action to take on error
		Sandboxed:     false, // restrict scripts to sandobox?
		Quiet:         false, // display startup banner?
		Trace:         false, // for debugging, print actions step-by-step
	}

	// start a repl
	Repl(env, cfg)
}

func Repl(env *Glisp, cfg *GlispConfig) {

	// Prompter prints a prompt and returns a single line
	// of input from stdin when pr.getExpressionWithLiner(env)
	// is invoked.
	pr := NewPrompter()
	defer pr.Close()

	// the LOOP of the REPL. REPL stands for
	//
	// (1) READ, (2) EVAL, (3) PRINT, (4) LOOP
	//
	for {
		// (1) READ
		line, exprsInput, err := pr.getExpressionWithLiner(env)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				os.Exit(0)
			}
			env.Clear()
			continue
		}

		// (2) EVAL
		expr, err := env.Eval(exprsInput)
		switch err {
		case nil:
		case NoExpressionsFound:
			env.Clear()
			continue
		default:
			// display error. reset.
			fmt.Print(env.GetStackTrace(err))
			env.Clear()
			continue
		}

		if expr != SexpNull {
			// (3) PRINT expr to stdout here
			// ...
		}
	}
}

// continuationPrompt is displayed when the parser needs more input
var continuationPrompt = "... "

// reads Stdin only
func (pr *Prompter) getExpressionWithLiner(env *Glisp) (readin string, xs []Sexp, err error) {

	line, err := pr.Getline(nil)
	if err != nil {
		return "", nil, err
	}

	err = UnexpectedEnd
	var x []Sexp

	// parse and pause the parser if we need more input.
	env.parser.ResetAddNewInput(bytes.NewBuffer([]byte(line + "\n")))
	x, err = env.parser.ParseTokens()

	if len(x) > 0 {
		xs = append(xs, x...)
	}

	for err == ErrMoreInputNeeded || err == UnexpectedEnd || err == ResetRequested {
		nextline, err := pr.Getline(&continuationPrompt)
		if err != nil {
			return "", nil, err
		}
		// provide more input
		env.parser.NewInput(bytes.NewBuffer([]byte(nextline + "\n")))

		// get parsed expression tree(s) back in x
		x, err = env.parser.ParseTokens()
		if len(x) > 0 {
			for i := range x {
				if x[i] == SexpEnd {
					P("found an SexpEnd token, omitting it")
					continue
				}
				xs = append(xs, x[i])
			}
		}
		switch err {
		case nil:
			line += "\n" + nextline
			// no problem, parsing went fine.
			return line, xs, nil
		case ResetRequested:
			continue
		case ErrMoreInputNeeded:
			continue
		default:
			return "", nil, fmt.Errorf("Error on line %d: %v\n", env.parser.lexer.Linenum(), err)
		}
	}
	return line, xs, nil
}

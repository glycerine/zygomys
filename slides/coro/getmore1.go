// getMoreInput is called by the Parser routines mid-parse, if
// need be, to obtain the next line/rune of input.
//
// getMoreInput() is used by Parser.ParseList(), Parser.ParseArray(),
// Parser.ParseBlockComment(), and Parser.ParseInfix().
//
// getMoreInput() is also used by Parser.infiniteParsingLoop() which
// is the main driver behind parsing.
//
// This function should *return* when it has more input
// for the parser/lexer, which will call it when they get wedged.
//
// Listeners on p.ParsedOutput should know the Convention: sending
// a length 0 []ParserReply on p.ParsedOutput channel means: we need more
// input! They should send some in on p.AddInput channel; or request
// a reset and simultaneously give us new input with p.ReqReset channel.
func (p *Parser) getMoreInput(deliverThese []Sexp, errorToReport error) error {
    if len(deliverThese) == 0 && errorToReport == nil {
        p.FlagSendNeedInput = true
    } else {
        p.sendMe = append(p.sendMe,ParserReply{Expr: deliverThese,Err:  errorToReport})
    }
    for {
        select {
        case <-p.reqStop:
            return ParserHaltRequested
        case input := <-p.AddInput:
            p.lexer.AddNextStream(input)
            p.FlagSendNeedInput = false
            return nil
        case input := <-p.ReqReset:
            p.lexer.Reset()
            p.lexer.AddNextStream(input)
            p.FlagSendNeedInput = false
            return ResetRequested
        case p.HaveStuffToSend() <- p.sendMe:
            // that was a conditional send, because
            // HaveStuffToSend() will return us a
            // nil channel if there's nothing ready.
            p.sendMe = make([]ParserReply, 0, 1)
            p.FlagSendNeedInput = false
        }
    }
}

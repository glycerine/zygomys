// getMoreInput does I/O: it is called by the Parser routines mid-parse to get the user's next line
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
        case p.HaveStuffToSend() <- p.sendMe:  // a conditional send!
            p.sendMe = make([]ParserReply, 0, 1)
            p.FlagSendNeedInput = false
}}}

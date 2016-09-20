func (p *Parser) HaveStuffToSend() chan []ParserReply {
    if len(p.sendMe) > 0 || p.FlagSendNeedInput {
        return p.ParsedOutput
    }
    return nil
}

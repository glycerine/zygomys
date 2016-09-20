// Start() commences the background parse loop goroutine.
func (p *Parser) Start() 


// ParseTokens is the main service the Parser provides.
// Currently returns first error encountered, ignoring
// any expressions after that.
func (p *Parser) ParseTokens() ([]Sexp, error)


// NewInput is the principal API function to
// supply parser with addition textual
// input lines
func (p *Parser) NewInput(s io.RuneScanner)


// ResetAddNewInput is the principal API function to
// tell the parser to forget everything it has stored,
// reset, and take as new input the scanner s.
func (p *Parser) ResetAddNewInput(s io.RuneScanner) {

	
// Stop gracefully shutsdown the parser and its background goroutine.
func (p *Parser) Stop() error


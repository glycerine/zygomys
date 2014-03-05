package main

import (
    "io"
    "bytes"
    "regexp"
    "errors"
    "strconv"
)

type TokenType int
const (
    TokenLParen TokenType = iota
    TokenRParen
    TokenLSquare
    TokenRSquare
    TokenQuoted
    TokenUnquoted
    TokenDecimal
    TokenHex
    TokenBinary
    TokenChar
)

type Token struct {
    typ TokenType
    str string
}

func (t Token) String() string {
    switch (t.typ) {
    case TokenLParen:   return "("
    case TokenRParen:   return ")"
    case TokenLSquare:  return "["
    case TokenRSquare:  return "]"
    case TokenQuoted:   return "'" + t.str
    case TokenUnquoted: return t.str
    case TokenDecimal:  return t.str
    case TokenHex:      return "0x" + t.str
    case TokenBinary:   return "0b" + t.str
    case TokenChar:
        quoted := strconv.Quote(t.str)
        return "#" + quoted[1:len(quoted)-1]
    }
    return ""
}

var DecimalRegex *regexp.Regexp
var HexRegex *regexp.Regexp
var BinaryRegex *regexp.Regexp
var QuotedRegex *regexp.Regexp
var UnquotedRegex *regexp.Regexp
var CharRegex *regexp.Regexp
var lexInit = false

func InitLexer() {
    var err error

    DecimalRegex, err = regexp.Compile("^-?[0-9]+$")
    if err != nil { panic(err) }
    HexRegex, err = regexp.Compile("^0x[0-9a-fA-F]+$")
    if err != nil { panic(err) }
    BinaryRegex, err = regexp.Compile("^0b[01]+$")
    if err != nil { panic(err) }
    QuotedRegex, err = regexp.Compile("^'[^'#]+$")
    if err != nil { panic(err) }
    UnquotedRegex, err = regexp.Compile("^[^'#]+$")
    if err != nil { panic(err) }
    CharRegex, err = regexp.Compile("^#\\\\?.$")
    if err != nil { panic(err) }
    lexInit = true
}

func DecodeChar(atom string) (string, error) {
    if len(atom) == 3 {
        switch (atom[2]) {
        case 'n': return "\n", nil
        case 'r': return "\r", nil
        case 'a': return "\a", nil
        case '#': return "#", nil
        default: return "", errors.New("invalid escape sequence")
        }
    }

    return atom[1:2], nil
}

func DecodeAtom(atom string) (Token, error) {
    if DecimalRegex.MatchString(atom) {
        return Token{TokenDecimal, atom}, nil
    } else if HexRegex.MatchString(atom) {
        return Token{TokenHex, atom[2:]}, nil
    } else if BinaryRegex.MatchString(atom) {
        return Token{TokenBinary, atom[2:]}, nil
    } else if UnquotedRegex.MatchString(atom) {
        return Token{TokenUnquoted, atom}, nil
    } else if QuotedRegex.MatchString(atom) {
        return Token{TokenQuoted, atom[1:]}, nil
    } else if CharRegex.MatchString(atom) {
        char, err := DecodeChar(atom)
        if err != nil {
            return Token{}, err
        }
        return Token{TokenChar, char}, nil
    }

    return Token{}, errors.New("Unrecognized atom")
}

func dumpBuffer(tokens []Token, buffer *bytes.Buffer) ([]Token, error) {
    if buffer.Len() <= 0 {
        return tokens, nil
    }

    tok, err := DecodeAtom(buffer.String())
    if err != nil {
        return tokens, err
    }
    buffer.Reset()
    return append(tokens, tok), nil
}

func DecodeBrace(brace rune) Token {
    switch (brace) {
    case '(': return Token{TokenLParen, ""}
    case ')': return Token{TokenRParen, ""}
    case '[': return Token{TokenLSquare, ""}
    default : return Token{TokenRSquare, ""}
    }
}

func LexStream(stream io.RuneReader) ([]Token, error) {
    if !lexInit {
        InitLexer()
    }

    tokens := make([]Token, 0, 20)
    comment := false
    buffer := new(bytes.Buffer)

    for {
        r, _, err := stream.ReadRune()
        if err != nil {
            break
        }

        if comment {
            if r == '\n' {
                comment = false
            }
        } else if r == '(' || r == ')' || r == '[' || r == ']' {
            tokens, err = dumpBuffer(tokens, buffer)
            if err != nil {
                return tokens, err
            }
            tokens = append(tokens, DecodeBrace(r))
        } else if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
            tokens, err = dumpBuffer(tokens, buffer)
            if err != nil {
                return tokens, err
            }
        } else if r == ';' {
            comment = true
        } else {
            _, err = buffer.WriteRune(r)
            if err != nil {
                return tokens, err
            }
        }
    }

    return tokens, nil
}

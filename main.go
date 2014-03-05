package main

import (
    "fmt"
    "os"
    "bufio"
)

func main() {
    tokens, err := LexStream(bufio.NewReader(os.Stdin))

    if err != nil {
        fmt.Println(err)
        os.Exit(-1)
    }

    for _, tok := range tokens {
        fmt.Println(tok)
    }
}

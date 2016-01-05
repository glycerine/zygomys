package gdsl

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/avail/obj"
	"cmd/link/avail/amd64"
	"cmd/link/avail/arm"
	"cmd/link/avail/arm64"
	"cmd/link/avail/mips64"
	"cmd/link/avail/ppc64"
	"cmd/link/avail/s390x"
	"cmd/link/avail/x86"
	"fmt"
	"os"
)

func main() {
	switch obj.Getgoarch() {
	default:
		fmt.Fprintf(os.Stderr, "link: unknown architecture %q\n", obj.Getgoarch())
		os.Exit(2)
	case "386":
		x86.Main()
	case "amd64", "amd64p32":
		amd64.Main()
	case "arm":
		arm.Main()
	case "arm64":
		arm64.Main()
	case "mips64", "mips64le":
		mips64.Main()
	case "ppc64", "ppc64le":
		ppc64.Main()
	case "s390x":
		s390x.Main()
	}
}

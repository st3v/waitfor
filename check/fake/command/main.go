package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	rc   = flag.Int("exit", 0, "exit code to return")
	out  = flag.String("out", "", "content to print on stdout")
	err  = flag.String("err", "", "content to print on stderr")
	echo = flag.Bool("echo", false, "redirect stdin to stdout")
	env  = flag.Bool("env", false, "print environment variables")
)

func main() {
	flag.Parse()

	if *echo {
		_, err := io.Copy(os.Stdout, os.Stdin)
		if err != nil {
			panic(fmt.Sprintf("Error copying stdin to stdout: %s", err.Error()))
		}
		os.Exit(*rc)
	}

	fmt.Fprint(os.Stderr, *err)
	fmt.Fprint(os.Stdout, *out)

	if *env {
		fmt.Fprint(os.Stdout, strings.Join(os.Environ(), "\n"))
	}

	os.Exit(*rc)
}

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const correctPassword = "iTWasNotThaTHard?Right?123"

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "[-] Please provide password and command\n")
		fmt.Fprintf(os.Stderr, "Usage: chowned password command")
		os.Exit(1)
	}
	password := os.Args[1]
	if password != correctPassword {
		fmt.Fprintf(os.Stderr, "[-] Wrong password")
		os.Exit(1)
	}

	syscall.Setuid(0)
	syscall.Setgid(0)

	commandSlice := []string{"-c"}
	commandSlice = append(commandSlice, os.Args[2])

	cmd := exec.Command("/bin/bash", commandSlice...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(out.String())
}

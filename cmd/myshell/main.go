package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input: ", err)
			os.Exit(1)
		}

		command, args := getCmdAndArgs(input)

		handleCommand(command, args)
	}
}

func handleCommand(command string, args []string) {
	switch command {
	case "exit":
		if len(args) > 0 {
			exitCode, err := strconv.Atoi(args[0])
			if err != nil {
				os.Exit(1)
			}
			os.Exit(exitCode)
		} else {
			os.Exit(0)
		}
	case "echo":
		fmt.Println(strings.Join(args, " "))
	case "type":
		switch args[0] {
		case "type", "echo", "exit":
			fmt.Println(args[0] + " is a shell builtin")
		default:
			fmt.Println(args[0] + ": not found")
		}
	default:
		fmt.Println(command + ": command not found")
	}
}

func getCmdAndArgs(input string) (string, []string) {
	// Ok so we are trimming space cause it also trims the space from the beginning, which sometimes
	// causes the command to be empty (space)
	input = strings.TrimSpace(input)

	args := strings.Split(input, " ")

	return strings.TrimSpace(args[0]), args[1:]
}

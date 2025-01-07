package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	initCommands()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input: ", err)
			os.Exit(1)
		}

		command, args := getCmdAndArgs(input)

		if handler, ok := builtInCommands[command]; ok {
			err := handler(args)
			handleError(err)
		} else {
			err := handleExecutables(command, args)
			if err != nil {
				fmt.Println(command + ": command not found")
			}
		}
	}
}

func handleExecutables(command string, args []string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func getCmdAndArgs(input string) (string, []string) {
	// Ok so we are trimming space cause it also trims the space from the beginning, which sometimes
	// causes the command to be empty (space)
	input = strings.TrimSpace(input)

	args := strings.Split(input, " ")

	return strings.TrimSpace(args[0]), args[1:]
}

type CommandHandler func(args []string) error

var builtInCommands = make(map[string]CommandHandler)

func initCommands() {
	builtInCommands["type"] = typeCommand
	builtInCommands["exit"] = exitCommand
	builtInCommands["echo"] = echoCommand
}

func typeCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println("")
	}

	_, builtin := builtInCommands[args[0]]

	// First we check if the command is builtin command, if yes then return from func
	if builtin {
		fmt.Println(args[0] + " is a shell builtin")
		return nil
	}

	// If the command is not a builtin command we then check if it is an executable
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		fullPath := path + "/" + args[0]

		if _, err := os.Stat(fullPath); err == nil {
			fmt.Println(args[0] + " is " + fullPath)
			return nil
		}
	}

	fmt.Println(args[0] + " not found")

	return nil
}

func exitCommand(args []string) error {
	if len(args) > 0 {
		exitCode, err := strconv.Atoi(args[0])
		if err != nil {
			os.Exit(1)
		}
		os.Exit(exitCode)
	} else {
		os.Exit(0)
	}
	return nil
}

func echoCommand(args []string) error {
	_, err := fmt.Println(strings.Join(args, " "))

	return handleError(err)
}

func handleError(err error) error {
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	return nil
}

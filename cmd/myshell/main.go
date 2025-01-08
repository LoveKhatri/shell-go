package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
			handler(args)
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
	// Ok so we are trimming space cause it also trims the space from the beginning,
	// which sometimes causes the command to be empty (space)
	input = strings.TrimSpace(input)
	var args []string
	var currentArg strings.Builder
	var inSingleQuotes bool

	for _, char := range input {
		switch char {
		case '\'':
			inSingleQuotes = !inSingleQuotes
		case ' ':
			if inSingleQuotes {
				currentArg.WriteRune(char)
			} else {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			}
		default:
			currentArg.WriteRune(char)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	if len(args) == 0 {
		return "", nil
	}

	return strings.TrimSpace(args[0]), args[1:]
}

type CommandHandler func(args []string)

var builtInCommands = make(map[string]CommandHandler)

func initCommands() {
	builtInCommands["type"] = typeCommand
	builtInCommands["exit"] = exitCommand
	builtInCommands["echo"] = echoCommand
	builtInCommands["pwd"] = pwdCommand
	builtInCommands["cd"] = cdCommand
}

func typeCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("")
	}

	_, builtin := builtInCommands[args[0]]

	// First we check if the command is builtin command, if yes then return from func
	if builtin {
		fmt.Println(args[0] + " is a shell builtin")
		return
	}

	// If the command is not a builtin command we then check if it is an executable
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		fullPath := path + "/" + args[0]

		if _, err := os.Stat(fullPath); err == nil {
			fmt.Println(args[0] + " is " + fullPath)
			return
		}
	}

	fmt.Println(args[0] + " not found")
}

func exitCommand(args []string) {
	if len(args) > 0 {
		exitCode, err := strconv.Atoi(args[0])
		if err != nil {
			os.Exit(1)
		}
		os.Exit(exitCode)
	} else {
		os.Exit(0)
	}
}

func echoCommand(args []string) {
	_, err := fmt.Println(strings.Join(args, " "))

	if err != nil {
		fmt.Println("Error: ", err)
	}

}

func handleError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
}

func pwdCommand(args []string) {
	dir, err := os.Getwd()
	handleError(err)

	fmt.Println(dir)

}

func cdCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("cd: missing argument")
	}

	pathInput := path.Clean(args[0])

	if strings.HasPrefix(pathInput, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("cd: " + pathInput + ": No such file or directory")
			return
		}

		pathInput = filepath.Join(home, strings.TrimPrefix(pathInput, "~"))
	}

	err := os.Chdir(pathInput)

	if err != nil {
		fmt.Println("cd: " + pathInput + ": No such file or directory")
		return
	}
}

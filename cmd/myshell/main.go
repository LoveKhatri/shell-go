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
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute: %v: %v", command, err)
	}

	return nil
}

func getCmdAndArgs(input string) (string, []string) {
	input = strings.TrimSpace(input)

	var (
		args           []string
		currentArg     strings.Builder
		inSingleQuotes bool
		inDoubleQuotes bool
		escapeNext     bool
	)

	// Convert to a slice for lookahead capability
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		if escapeNext {
			// If we were escaping, just add the character literally
			currentArg.WriteRune(char)
			escapeNext = false
			continue
		}

		switch char {
		case '\\':
			if inSingleQuotes {
				// Inside single quotes, backslash is literal
				currentArg.WriteRune(char)
			} else if inDoubleQuotes {
				// Inside double quotes, check next character
				if i+1 < len(runes) {
					nextChar := runes[i+1]
					// Special meaning if nextChar is one of \, $, ", or a newline
					if nextChar == '\\' || nextChar == '$' || nextChar == '"' || nextChar == '\n' {
						escapeNext = true
					} else {
						// Otherwise it's literal
						currentArg.WriteRune(char)
					}
				} else {
					// Backslash at the end
					currentArg.WriteRune(char)
				}
			} else {
				// Outside of quotes, backslash escapes next char
				escapeNext = true
			}

		case '\'':
			if !inDoubleQuotes {
				// Toggle single quotes if not in double quotes
				inSingleQuotes = !inSingleQuotes
			} else {
				// If in double quotes, treat it as literal
				currentArg.WriteRune(char)
			}

		case '"':
			if !inSingleQuotes {
				// Toggle double quotes if not in single quotes
				inDoubleQuotes = !inDoubleQuotes
			} else {
				// If in single quotes, treat it as literal
				currentArg.WriteRune(char)
			}

		case ' ':
			// Space splits arguments only if we are not in quotes
			if inSingleQuotes || inDoubleQuotes {
				currentArg.WriteRune(char)
			} else {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			}

		default:
			// Normal character, just add to current argument
			currentArg.WriteRune(char)
		}
	}

	// Add the last argument if present
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	// If nothing was parsed, return no command
	if len(args) == 0 {
		return "", nil
	}

	return args[0], args[1:]
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

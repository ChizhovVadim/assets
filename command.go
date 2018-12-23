package main

import (
	"fmt"
	"os"
	"strings"
)

type command struct {
	name    string
	handler func(commandArgs) error
}

type commandArgs struct {
	name   string
	params map[string]string
}

func runCommands(commands []command) {
	var cmdArgs = parseCommandArg()
	var cmd, found = findCommand(cmdArgs.name, commands)
	if !found {
		fmt.Println("command not found")
		return
	}
	var err = cmd.handler(cmdArgs)
	if err != nil {
		fmt.Println("command error", err)
		return
	}
}

func parseCommandArg() commandArgs {
	var args = os.Args
	var cmdName = ""
	var flags = make(map[string]string)
	for i := 1; i < len(args); i++ {
		var arg = args[i]
		if strings.HasPrefix(arg, "-") {
			if i < len(args)-1 {
				var k = strings.TrimPrefix(arg, "-")
				var v = args[i+1]
				flags[k] = v
			}
		} else if cmdName == "" {
			cmdName = arg
		}
	}
	return commandArgs{cmdName, flags}
}

func findCommand(name string, commands []command) (command, bool) {
	for i := range commands {
		if commands[i].name == name {
			return commands[i], true
		}
	}
	return command{}, false
}

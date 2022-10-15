package cli

import (
	"weshare/engine"

	"github.com/fatih/color"
)

func processAdd(commands []string, options Options) {
	if len(commands) < 1 {
		color.Red("command add requires the target file path")
		return
	}

	err := engine.Add(commands[0], true)
	if err != nil {
		color.Red("internal error:%v", err)
		return
	}
	color.Green("%s added", commands[0])
}

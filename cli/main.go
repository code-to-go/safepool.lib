package cli

import (
	"flag"
	"os"
	"strings"
	"weshare/core"
	"weshare/engine"

	"github.com/sirupsen/logrus"
)

var shortcuts = map[byte]string{
	'j': "join",
	's': "state",
	'a': "add",
	't': "trust",
	'u': "update",
}

func setLogLevel(logLevel string) {
	logLevel = strings.ToUpper(logLevel)
	switch logLevel {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "WARNING":
		logrus.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	}
}

func replaceShortcuts(commands []string) []string {
	if len(commands) == 0 {
		return commands
	}
	switch len(commands[0]) {
	case 1:
		if a := shortcuts[commands[0][0]]; a != "" {
			return append([]string{a}, commands[1:]...)
		}
	case 2:
		a := shortcuts[commands[0][0]]
		b := shortcuts[commands[0][1]]

		if a != "" && b != "" {
			return append([]string{a, b}, commands[1:]...)
		}
	}
	return commands

}

type Options struct {
	Verbose       bool
	Verbose2      bool
	LogLevel      string
	PrintExchange bool
}

// ProcessArgs analyze the
func ProcessArgs() {
	options := Options{}

	err := engine.Start()
	if err == core.ErrNotInitialized {
		err = initIdentity()
		if core.IsErr(err, "cannot init identity: %v") {
			return
		}
	}

	_, completion := os.LookupEnv("COMP_LINE")
	if completion {
		//		complete(strings.Split(cl, " ")[1:])
		return
	}

	flag.Usage = usage
	flag.StringVar(&options.LogLevel, "d", "error",
		"Items level to display")

	flag.BoolVar(&options.Verbose, "v", false,
		"shows verbose log")

	flag.BoolVar(&options.Verbose2, "vv", false,
		"shows very verbose log")

	flag.BoolVar(&options.PrintExchange, "printExchange", false,
		"shows very verbose log")

	flag.Parse()

	switch {
	case options.Verbose2:
		setLogLevel("DEBUG")
	case options.Verbose:
		setLogLevel("INFO")
	default:
		setLogLevel("ERROR")
	}

	nArg := flag.NArg()
	if nArg == 0 {
		flag.Usage()
		return
	}

	commands := os.Args[len(os.Args)-nArg:]
	commands = replaceShortcuts(commands)

	engine.Start()

	switch commands[0] {
	case "join":
		processJoin(commands[1:], options)
	case "token":
		processToken(commands[1:], options)
	case "state":
		processState(commands[1:], options)
	case "add":
		processAdd(commands[1:], options)
	case "trust":
		processTrust(commands[1:], options)
	case "pwd":
		processUpdate(commands[1:], options)

	default:
		logrus.Debugf("Unknown command %s", commands[0])
		flag.Usage()
	}

}

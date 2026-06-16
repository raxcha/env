package main

import (
	"env/master"
	"os"
	"strings"
)

func main() {
	stage, args := parseArgs(os.Args[1:])
	arg := strings.Join(args, " ")
	if arg == "" {
		arg = defaultClientArg(stage)
	}

	master.CreateMasterWithStage(arg, stage)

	select {}
}

func parseArgs(args []string) (string, []string) {
	stage := ""
	rest := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--stage" || arg == "--upto" {
			if i+1 < len(args) {
				stage = args[i+1]
				i++
			}
			continue
		}

		if strings.HasPrefix(arg, "--stage=") {
			stage = strings.TrimPrefix(arg, "--stage=")
			continue
		}

		if strings.HasPrefix(arg, "--upto=") {
			stage = strings.TrimPrefix(arg, "--upto=")
			continue
		}

		rest = append(rest, arg)
	}

	return stage, rest
}

func defaultClientArg(stage string) string {
	switch master.NormalizeStage(stage) {
	case "master", "tabs":
		return ""
	default:
		return "editor"
	}
}

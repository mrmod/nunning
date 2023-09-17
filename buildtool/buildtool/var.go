package buildtool

import (
	"log"
	"os"
)

var (
	BuildToolDefaultArguments = map[BuildTool][]string{
		BuildTool("buildtool/go"): []string{
			"run",
			"--rm",
			"-w",
			"/build",
			"-v",
			mustGetCWD() + ":/build",
			"buildtool/go",
			"build",
		},
	}

	BuildToolEntrypoint = map[BuildTool]string{
		BuildTool("buildtool/go"): "docker",
	}
)

func mustGetCWD() string {
	d, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get cwd: %s", err)
	}
	return d
}

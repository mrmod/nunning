package main

import (
	"log"
	"mrmod/buildtool/buildtool"
	"os/exec"
	"strings"
)

func main() {
	buildFile := buildtool.LoadBuildFile("BUILD.yaml")
	log.Printf("Loaded buildFile %#v\n", buildFile)

	for _, target := range buildFile.Targets {
		log.Printf("Running buildtool %s", target.BuildTool)

		args := target.BuildToolArguments()
		entrypoint := target.BuildTool.Entrypoint()
		log.Printf("Running `%s %#v`", entrypoint, args)
		cmd := exec.Command(entrypoint, args...)
		var stdOut, stdErr strings.Builder

		cmd.Stdout = &stdOut
		cmd.Stderr = &stdErr

		if err := cmd.Run(); err != nil {
			log.Printf(stdErr.String())
			panic(err)
		}
		log.Println(stdErr.String())
		log.Println(stdOut.String())
	}
}

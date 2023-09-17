package buildtool

type BuildTarget struct {
	Build            `yaml:"Build"`
	BuildTool        `yaml:"BuildTool"`
	BuildToolOptions `yaml:"BuildToolOptions"`
	Name             string `yaml:"Name"`
}

// BuildToolArguments Returns the default tool arguments with the user arguments appended
func (b BuildTarget) BuildToolArguments() []string {
	args := ParseArguments(b, b.Arguments...)
	return args.Strings()
}

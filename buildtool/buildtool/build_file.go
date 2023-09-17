package buildtool

type BuildFile struct {
	Targets []BuildTarget `yaml:"Targets"`
}

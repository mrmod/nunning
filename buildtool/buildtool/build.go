package buildtool

type Build struct {
	Inputs  []Input  `yaml:"Inputs"`
	Outputs []Output `yaml:"Outputs"`
}

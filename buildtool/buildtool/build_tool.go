package buildtool

type BuildTool string

func (b BuildTool) Entrypoint() string {
	return BuildToolEntrypoint[b]
}

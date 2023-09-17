package buildtool

import "log"

/*
	BuildToolArgument

Implements Stringable
Can derive a String() => string for BuildToolArgument
*/
type BuildToolArgument string

func (arg BuildToolArgument) String() string {
	return string(arg)
}

func (arg BuildToolArgument) RenderWithTarget(t BuildTarget, args BuildToolArguments) BuildToolArguments {
	log.Printf("Rendering %v into %#v", arg, args)
	insAt, newArgs := AppendArgument(t, arg, args)
	log.Printf("Inserted %d args at %d", len(newArgs)-insAt, insAt)
	return newArgs
}

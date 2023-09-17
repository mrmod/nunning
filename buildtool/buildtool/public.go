package buildtool

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func Merge(a any) string {
	log.Printf("nothing merge")
	return ""
}
func InputsToString(inputs []Input) (s []string) {
	for _, input := range inputs {
		s = append(s, input.String())
	}
	return
}

/*
	ParseArguments

Create s a list of BuildArguments.
When there's a parseable template
Then templates will by applied, in order
to the BuildArguments.
When it's not a template
Then a BuildArgument is added to the arguments
*/
func ParseArguments(t BuildTarget, args ...BuildToolArgument) BuildToolArguments {
	log.Printf("Parsing %d arguments", len(args))
	for _, arg := range args {
		log.Printf("[%d] parsing %s", len(args), arg)

		args = BuildToolArguments(args).Append(t, arg)
	}
	return args
}

func (args BuildToolArguments) Append(t BuildTarget, arg BuildToolArgument) BuildToolArguments {
	log.Printf("Append %v to self: %#v", arg, args)
	return arg.RenderWithTarget(t, args)
}

type RenderWithTargetFunc func(BuildTarget, BuildToolArguments) BuildToolArguments

/* AppendArgument
*
* Adds an argument to the tail of the build tool arguments list
 */
func AppendArgument(t BuildTarget, arg BuildToolArgument, args BuildToolArguments) (int, BuildToolArguments) {
	log.Printf("appending %v to %#v", arg, args)
	insertedAt := len(args)
	renderer := template.New("render")
	renderer.Funcs(template.FuncMap{
		// Merge
		// Appends a list of arguments to the tail of another list of arguments
		"Merge": func(inputs []Input) string {
			log.Printf("merging %v", inputs)
			newArgs := args[0 : len(args)-1]
			for _, _s := range inputs {
				newArgs = append(newArgs, BuildToolArgument(_s.String()))
			}

			log.Printf("Args[%d] newArgs[%d]", len(args), len(newArgs))
			args = newArgs
			return ""
		},
	})

	// The template mutates the input
	argTemplate, err := renderer.Parse(arg.String())
	if err != nil {
		log.Printf("unable to parse template: %s", err)
		log.Printf("inserting as string %s", arg)
		return insertedAt, append(args, BuildToolArgument(arg.String()))
	}

	// Interpolated output should be kept
	output := new(bytes.Buffer)
	if err := argTemplate.Execute(output, t); err != nil {
		log.Printf("unable to apply the template: %s", err)
		return insertedAt, append(args, arg)
	}
	if output.Len() > 0 {
		log.Printf("template produced: %s", output.String())
		return insertedAt, append(args, BuildToolArgument(output.String()))
	}
	return insertedAt, args

}

func RenderInto(bt BuildTarget, arg BuildToolArgument, args []string) []string {
	renderer := template.New("render")

	renderer.Funcs(template.FuncMap{
		"Merge": func(s []Input) string {

			log.Printf("Merging %v into %v", s, args)
			args = append(args, InputsToString(s)...)
			return ""
		},
	})

	t, err := renderer.Parse(string(arg))
	if err != nil {
		return append(args, string(arg))
	}

	output := new(bytes.Buffer)
	if err := t.Execute(output, bt); err != nil {
		log.Printf("Error: %s", err)
		return append(args, string(arg))
	}
	return append(args, output.String())
}

func LoadBuildFile(f string) *BuildFile {

	fh, err := os.Open(f)
	if err != nil {
		panic(err)
	}

	buildFile := new(BuildFile)
	if err := yaml.NewDecoder(fh).Decode(buildFile); err != nil {
		panic(err)
	}

	return buildFile
}

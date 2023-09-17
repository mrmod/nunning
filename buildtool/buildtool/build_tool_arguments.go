package buildtool

import (
	"log"
	"strings"
)

/*
	BuildToolArguments

Implements Stringable
Implements Sluggable
Can derive a String() => string behavior from BuildToolArguments
Can derive a Slug() => string bevavior from BuildToolArguments
*/
type BuildToolArguments []BuildToolArgument

func (args BuildToolArguments) String() string {
	return args.Slug()
}

func (args BuildToolArguments) Strings() []string {
	slug := []string{}
	for n, arg := range args {
		log.Printf("[%d] %v;", n, slug)
		slug = append(slug, arg.String())
	}
	log.Printf("final: :%v:", slug)
	return slug
}

func (args BuildToolArguments) Slug() string {
	return strings.Join(args.Strings(), spaceSlug)
}

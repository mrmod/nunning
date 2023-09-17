package buildtool

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

func exactArgsGiven(ctx context.Context, arg string) {
	// context.WithValue(ctx, godogsCtxKey)

}

func exactArgumentsArePassed(ctx context.Context) {
	fmt.Printf("...ArePassed: %#v", 1)
}
func InitializeScenario(ctx *godog.ScenarioContext) {

	ctx.Step(`^a lteral string \"(\w)\"$`, exactArgsGiven)
	ctx.Step(`^I call BuildToolArguments$`, exactArgumentsArePassed)
}

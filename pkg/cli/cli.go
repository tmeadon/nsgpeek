package cli

import (
	"github.com/alecthomas/kong"
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

var (
	cli struct {
		Debug bool `help:"Enable debug mode"`

		Stream StreamCmd `cmd:"" help:"Stream NSG flow logs"`
	}

	cred *azure.Credential
	subs []string
)

type cliContext struct {
	Debug bool
}

func Run() {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.Name("nsgpeek"))
	getCredential(ctx)
	getUserSubscriptions(ctx)

	err := ctx.Run(&cliContext{Debug: cli.Debug})

	ctx.FatalIfErrorf(err)
}

func getCredential(ctx *kong.Context) {
	c, err := azure.GetCredential()
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	cred = c
}

func getUserSubscriptions(ctx *kong.Context) {
	s, err := azure.GetSubscriptions(cred)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	subs = s
}

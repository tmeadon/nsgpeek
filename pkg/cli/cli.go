package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/tmeadon/nsgpeek/pkg/azure"
	"github.com/tmeadon/nsgpeek/pkg/flowwriter"
)

var (
	cli struct {
		Debug bool `help:"Enable debug mode"`

		Stream StreamCmd `cmd:"" help:"Stream NSG flow logs"`
		Search SearchCmd `cmd:"" help:"Search historical NSG flow logs"`
	}

	cred *azure.Credential
	subs []string
)

type cliContext struct {
	Debug bool
}

type commonArgs struct {
	NsgName   string `required:"" short:"n" help:"Name of the NSG to stream logs from"`
	Quiet     bool   `short:"q" help:"(Optional) Don't print to console"`
	File      string `short:"f" help:"(Optional) File path to write logs to in CSV format"`
	Overwrite bool   `help:"(Optional) Overwrite file if already exists"`
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

func initWriterGroup(args commonArgs) (*flowwriter.WriterGroup, error) {
	consoleWriter := flowwriter.NewConsoleWriter(os.Stdout)
	writers := flowwriter.NewWriterGroup(consoleWriter)

	if err := addCsvWriter(args.File, args.Overwrite, writers); err != nil {
		return nil, err
	}

	return writers, nil
}

func addCsvWriter(path string, overwrite bool, wg *flowwriter.WriterGroup) error {
	if path != "" {
		if _, err := os.Stat(path); err == nil && !overwrite {
			return fmt.Errorf("file already exists at path %v - add --overwrite or specify a different filepath, see command help for details", path)
		}

		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file %v: %w", path, err)
		}

		csvWriter, err := flowwriter.NewCsvFileWriter(file)
		if err != nil {
			return fmt.Errorf("failed to create csv file writer: %w", err)
		}

		wg.AddWriter(csvWriter)
	}
	return nil
}

package main

import (
	"github.com/jinfwhuang/go-light-eth/internal/consensus/node"
	"github.com/urfave/cli/v2"
	tmplog "log"
	"os"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func main() {
	app := cli.App{}
	app.Name = "beacon-chain-light-client"
	app.Usage = "Beacon Chain Light Client"
	app.Action = start

	app.Flags = node.AppFlags

	if err := app.Run(os.Args); err != nil {
		tmplog.Println(err)
	}
}

func start(ctx *cli.Context) error {
	node, err := node.New(ctx)
	if err != nil {
		panic(err)
	}
	node.Start()
	return nil
}

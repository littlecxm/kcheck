package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/littlecxm/kcheck/configs"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var (
	listType  string
	version   = "bleeding-edge"
	buildDate = "0000-00-00 00:00:00"
	commitID  = "*******"
	json      = jsoniter.ConfigCompatibleWithStandardLibrary
)

func main() {
	configs.WorkDir, _ = os.Getwd()

	app := &cli.App{
		Name:    "kcheck",
		Usage:   "check files through list",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Usage: fmt.Sprintf(
					"input list `TYPE`, support: `%s`, `%s`, `%s`, `%s`",
					configs.KCheckType,
					configs.XMLType,
					configs.KBinType,
					configs.MetadataType,
				),
				Destination: &listType,
			},
		},
		Action: commandHandler,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

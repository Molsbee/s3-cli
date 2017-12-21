package main

import (
	"github.com/molsbee/s3-cli/command"
	"github.com/molsbee/s3-cli/model/config"
	"github.com/molsbee/s3-cli/service"
	"github.com/urfave/cli"
	"os"
)

func main() {
	s3Service := service.NewS3Service(config.S3ServiceConfig{
		AccessKey:        "",
		SecretAccessKey:  "",
		Endpoint:         "",
		Region:           "",
		SignatureVersion: config.V2,
	})

	app := cli.NewApp()
	app.Name = "s3-cli"
	app.Usage = "interact and consume s3 compatible object storage"
	app.Commands = []cli.Command{
		command.List(s3Service),
	}
	app.Run(os.Args)
}

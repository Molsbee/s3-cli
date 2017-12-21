package command

import (
	"fmt"
	"github.com/molsbee/s3-cli/service"
	"github.com/molsbee/s3-cli/util"
	"github.com/urfave/cli"
)

func List(s3Service service.S3Service) cli.Command {
	return cli.Command{
		Name:        "ls",
		Usage:       "List all objects/buckets based on request parameter.",
		Description: "If no [s3://BUCKET[/PREFIX]] provided gets a list of buckets owned by user.",
		UsageText:   "s3-cli ls [s3://BUCKET[/PREFIX]]",
		Action: func(ctx *cli.Context) error {
			args := ctx.Args()
			if len(args) > 1 {
				return fmt.Errorf("incorrect number of arguments provided")
			}

			if len(args) == 0 {
				buckets, err := s3Service.ListBuckets()
				if err != nil {
					return err
				}

				for _, bucket := range buckets {
					util.Print(bucket)
				}

				return nil
			}

			response, listErr := s3Service.ListObjects(args.Get(0))
			if listErr != nil {
				return listErr
			}

			util.Print(response)
			return nil
		},
	}
}

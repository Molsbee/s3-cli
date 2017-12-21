package model

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/molsbee/s3-cli/util"
)

type Bucket struct {
	CreationDate string `json:"creationDate"`
	Name         string `json:"fileName"`
}

func NewBucket(b *s3.Bucket) Bucket {
	return Bucket{
		CreationDate: util.FormatTime(b.CreationDate),
		Name:         *b.Name,
	}
}

func (b Bucket) String() string {
	return fmt.Sprintf("%16s    %s", b.CreationDate, b.Name)
}

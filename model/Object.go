package model

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/molsbee/s3-cli/util"
	"time"
)

type ObjectResponse struct {
	Prefixes []Prefix
	Objects  []Object
}

func (or ObjectResponse) String() string {
	var out string
	for _, prefix := range or.Prefixes {
		out += fmt.Sprintln(prefix)
	}

	for _, object := range or.Objects {
		out += fmt.Sprintln(object)
	}

	return out
}

type Prefix struct {
	Name string
}

func (p Prefix) String() string {
	return fmt.Sprintf("%16s    %-19s    %s", "", "DIR", p.Name)
}

type Object struct {
	LastModified string `json:"lastModified"`
	Size         int64  `json:"size"`
	Key          string `json:"key"`
}

func NewObject(lastModified *time.Time, size int64, key string) Object {
	return Object{
		LastModified: util.FormatTime(lastModified),
		Size:         size,
		Key:          key,
	}
}

func NewObjectFromS3Object(o *s3.Object) Object {
	return Object{
		LastModified: util.FormatTime(o.LastModified),
		Size:         *o.Size,
		Key:          *o.Key,
	}
}

func (o Object) String() string {
	return fmt.Sprintf("%16s    %-19d    %s", o.LastModified, o.Size, o.Key)
}

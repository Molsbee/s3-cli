package config

type SignatureVersion int

const (
	V2 SignatureVersion = iota
	V4
)

type S3ServiceConfig struct {
	AccessKey        string
	SecretAccessKey  string
	Endpoint         string
	Region           string
	SignatureVersion SignatureVersion
}

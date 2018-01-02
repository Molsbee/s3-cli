package namedhandler

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
)

const (
	timeFormat     = "Mon, 2 Jan 2006 15:04:05 +0000"
	logSignInfoMsg = `DEBUG: Request Signature:
---[ STRING TO SIGN ]--------------------------------
%s
---[ SIGNATURE ]-------------------------------------
%s
-----------------------------------------------------`
)

type signer struct {
	// Values that must be populated from the request
	Request     *http.Request
	Time        time.Time
	Credentials *credentials.Credentials
	Debug       aws.LogLevelType
	Logger      aws.Logger
	pathStyle   bool
	bucket      string

	Query        url.Values
	stringToSign string
	signature    string

	baseEndpoint string
}

var V2SignRequestHandler = request.NamedHandler{
	Name: "v2.SignRequestHandler", Fn: SignSDKRequest,
}

var subresources = []string{
	"acl",
	"delete",
	"lifecycle",
	"location",
	"logging",
	"notification",
	"partNumber",
	"policy",
	"requestPayment",
	"torrent",
	"uploadId",
	"uploads",
	"versionId",
	"versioning",
	"versions",
	"website",
}

// Sign requests with signature version 2.
//
// Will sign the requests with the service config's Credentials object
// Signing is skipped if the credentials is the credentials.AnonymousCredentials
// object.
func SignSDKRequest(req *request.Request) {
	// If the request does not need to be signed ignore the signing of the
	// request if the AnonymousCredentials object is used.
	if req.Config.Credentials == credentials.AnonymousCredentials {
		return
	}

	v2 := signer{
		Request:      req.HTTPRequest,
		Time:         req.Time,
		Credentials:  req.Config.Credentials,
		Debug:        req.Config.LogLevel.Value(),
		Logger:       req.Config.Logger,
		pathStyle:    aws.BoolValue(req.Config.S3ForcePathStyle),
		baseEndpoint: *req.Config.Endpoint,
	}

	req.Error = v2.Sign()
}

func (v2 *signer) Sign() error {
	credValue, err := v2.Credentials.Get()
	if err != nil {
		return err
	}

	// Current code
	v2.Query = v2.Request.URL.Query()

	v2.Request.Header.Set("x-amz-date", v2.Time.UTC().Format(timeFormat))
	if credValue.SessionToken != "" {
		v2.Request.Header.Set("x-amz-security-token", credValue.SessionToken)
	}

	// in case this is a retry, ensure no signature present
	v2.Request.Header.Del("Authorization")

	path := v2.createPath()
	query := v2.createQueryString()
	if query != "" {
		path += "?" + query
	}

	v2.stringToSign = createStringToSign(v2.Request, path)

	hash := hmac.New(sha1.New, []byte(credValue.SecretAccessKey))
	hash.Write([]byte(v2.stringToSign))
	v2.signature = base64.StdEncoding.EncodeToString(hash.Sum(nil))
	v2.Request.Header.Set("Authorization", "AWS "+credValue.AccessKeyID+":"+v2.signature)

	if v2.Debug.Matches(aws.LogDebugWithSigning) {
		v2.logSigningInfo()
	}

	return nil
}

func (v2 *signer) logSigningInfo() {
	msg := fmt.Sprintf(logSignInfoMsg, v2.stringToSign, v2.Request.Header.Get("Authorization"))
	v2.Logger.Log(msg)
}

func (v2 *signer) createPath() string {
	path := v2.Request.URL.Path
	if path == "" {
		path = "/"
	}

	if v2.baseEndpoint != v2.Request.URL.Host {
		bucketName := strings.Replace(v2.Request.URL.Host, "."+v2.baseEndpoint, "", 1)
		path = "/" + bucketName + path
	}

	return path
}

func (v2 *signer) createQueryString() string {
	queryKeys := make([]string, 0, len(v2.Query))
	for key := range v2.Query {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)

	// build URL-encoded query keys and values
	var queryKeysAndValues []string
	for _, key := range subresources {
		if _, ok := v2.Query[key]; ok {
			k := strings.Replace(url.QueryEscape(key), "+", "%20", -1)
			v := strings.Replace(url.QueryEscape(v2.Query.Get(key)), "+", "%20", -1)
			queryKeysAndValues = append(queryKeysAndValues, k+"="+v)
		}
	}

	return strings.Join(queryKeysAndValues, "&")
}

func createStringToSign(request *http.Request, path string) string {
	signatureParts := []string{
		request.Method,
		request.Header.Get("Content-MD5"),
		request.Header.Get("Content-Type"),
		"",
	}

	for k := range request.Header {
		k = strings.ToLower(k)
		if strings.HasPrefix(k, "x-amz-") {
			signatureParts = append(signatureParts, k+":"+strings.Join(request.Header[http.CanonicalHeaderKey(k)], ","))
		}
	}

	signatureParts = append(signatureParts, path)
	return strings.Join(signatureParts, "\n")
}

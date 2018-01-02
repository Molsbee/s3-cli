package util

import (
	"fmt"
	"net/url"
	"time"
	"strings"
)

func FormatTime(t *time.Time) string {
	month := normalizeTimeValue(int(t.Month()))
	day := normalizeTimeValue(t.Day())
	hour := normalizeTimeValue(t.Hour())
	minute := normalizeTimeValue(t.Minute())
	return fmt.Sprintf("%d-%s-%s %s:%s", t.Year(), month, day, hour, minute)
}

func normalizeTimeValue(v int) string {
	value := fmt.Sprintf("%d", v)
	if len(value) == 1 {
		value = "0" + value
	}

	return value
}

func ParseAndValidateBucketURL(bucket string) (*url.URL, error) {
	url, err := url.Parse(bucket)
	if err != nil {
		return nil, fmt.Errorf("unable to parse provided bucket - correct format s3://[BUCKET][/PREFIX]")
	}

	if strings.ToLower(url.Scheme) != "s3" {
		return nil, fmt.Errorf("unsupported schema provided - correct format s3://[[BUCKET][/PREFIX]")
	}

	if len(url.Host) == 0 {
		return nil, fmt.Errorf("missing bucket name - correct format s3://[BUCKET][/PREFIX]")
	}

	return url, nil
}

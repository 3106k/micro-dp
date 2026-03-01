package edition

import (
	"os"
	"strings"
)

const (
	OSS = "oss"
	Web = "web"
)

func Current() string {
	e := strings.ToLower(strings.TrimSpace(os.Getenv("EDITION")))
	if e == Web {
		return Web
	}
	return OSS
}

func IsOSS() bool { return Current() == OSS }
func IsWeb() bool { return Current() == Web }

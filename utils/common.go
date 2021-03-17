package utils

import (
	"errors"
	"strconv"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3355.0 Safari/537.36"
const configKeyVerbose = "VERBOSE"

var DelayHandlers = make(map[string]DelayHandler)

func ConfigGetInt(u Utils, key string) (int64, error) {
	if env := u.ConfigGet(key); env != "" {
		if int, err := strconv.ParseInt(env, 10, 64); err != nil {
			return 0, err
		} else {
			return int, nil
		}
	}

	return 0, errors.New("Not yet configured")
}

func ConfigGetIntWithDefault(u Utils, key string, valueDefault int64) int64 {
	if value, err := ConfigGetInt(u, key); err == nil {
		return value
	}

	return valueDefault
}

func ConfigGetTTLDefault(u Utils) int64 {
	return ConfigGetIntWithDefault(u, "TTL_DEFAULT", 300)
}

func Verbosef(u Utils, format string, args ...interface{}) {
	if ConfigGetIntWithDefault(u, configKeyVerbose, 0) < 1 {
		return
	}

	u.Debugf(format, args...)
}

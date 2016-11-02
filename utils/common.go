package utils

import (
	"errors"
	"strconv"
)

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
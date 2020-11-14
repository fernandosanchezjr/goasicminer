package url

import (
	"net/url"
	"strconv"
)

func ParseInt(name string, values url.Values, result *int) (err error) {
	intStr := values.Get(name)
	if intStr != "" {
		*result, err = strconv.Atoi(intStr)
		return err
	}
	return
}

func ParseBool(name string, values url.Values, result *bool) (err error) {
	boolStr := values.Get(name)
	if boolStr != "" {
		*result, err = strconv.ParseBool(boolStr)
		return err
	}
	return
}

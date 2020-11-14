package charting

import (
	url2 "github.com/fernandosanchezjr/goasicminer/backend/url"
	"net/url"
)

type ServiceParams struct {
	Days    int
	Hours   int
	Minutes int
	Refresh bool
}

func ParseServiceParams(values url.Values) (params *ServiceParams, err error) {
	params = &ServiceParams{
		Days:    0,
		Hours:   0,
		Minutes: 30,
		Refresh: true,
	}
	if err = url2.ParseInt("days", values, &params.Days); err != nil {
		return
	}
	if err = url2.ParseInt("hours", values, &params.Hours); err != nil {
		return
	}
	if err = url2.ParseInt("minutes", values, &params.Minutes); err != nil {
		return
	}
	if err = url2.ParseBool("refresh", values, &params.Refresh); err != nil {
		return
	}
	return
}

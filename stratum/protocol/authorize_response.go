package protocol

import "github.com/epiclabs-io/elastic"

type AuthorizeResponse struct {
	Result bool
}

func NewAuthorizeResponse(reply *Reply) (*AuthorizeResponse, error) {
	ar := &AuthorizeResponse{}
	if err := reply.HasError(); err != nil {
		return nil, err
	}
	if err := elastic.Set(&ar.Result, reply.Result); err != nil {
		return nil, err
	}
	return ar, nil
}

package client

import (
	"github.com/fission/fission"
	"github.com/fission/fission/environments/envsidecar"
)

type (
	Client struct {
		url string
	}
)

func MakeClient(fetcherUrl string) *Client {
	return &Client{
		url: fetcherUrl,
	}
}

func (c *Client) getSpecializeUrl() string {
	return c.url + "/specialize"
}

func (c *Client) Specialize(req *fission.FunctionSpecializeRequest) error {
	_, err := envsidecar.SendRequest(req, c.getSpecializeUrl())
	return err
}

package client

import (
	"strings"

	"github.com/fission/fission"
	"github.com/fission/fission/environments/envsidecar"
)

type (
	Client struct {
		url string
	}
)

func MakeClient(specializerUrl string) *Client {
	return &Client{
		url: strings.TrimSuffix(specializerUrl, "/"),
	}
}

func (c *Client) getSpecializeUrl() string {
	return c.url + "/specialize"
}

func (c *Client) Specialize(req *fission.FunctionSpecializeRequest) error {
	_, err := envsidecar.SendRequest(req, c.getSpecializeUrl())
	return err
}

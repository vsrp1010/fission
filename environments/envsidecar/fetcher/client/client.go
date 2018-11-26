package client

import (
	"encoding/json"
	"strings"

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
		url: strings.TrimSuffix(fetcherUrl, "/"),
	}
}

func (c *Client) getFetchUrl() string {
	return c.url + "/fetch"
}

func (c *Client) getUploadUrl() string {
	return c.url + "/upload"
}

func (c *Client) Fetch(fr *fission.FunctionFetchRequest) error {
	_, err := envsidecar.SendRequest(fr, c.getFetchUrl())
	return err
}

func (c *Client) Upload(fr *fission.ArchiveUploadRequest) (*fission.ArchiveUploadResponse, error) {
	body, err := envsidecar.SendRequest(fr, c.getUploadUrl())

	uploadResp := fission.ArchiveUploadResponse{}
	err = json.Unmarshal(body, &uploadResp)
	if err != nil {
		return nil, err
	}

	return &uploadResp, nil
}

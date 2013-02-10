package xmlrpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	client *http.Client
	url    *url.URL
}

func NewClient(url *url.URL) (*Client, error) {
	return &Client{client: new(http.Client), url: url}, nil
}

func (this *Client) Call(method string, args ...interface{}) (interface{}, error) {

	var (
		err  error
		resp *http.Response
		res  interface{}
	)

	buf := new(bytes.Buffer)
	err = Marshal(buf, method, args...)
	if err != nil {
		return nil, err
	}

	// keep-alive is handled by the transport layer
	resp, err = http.Post(this.url.String(), "text/xml", buf)
	var b []byte
	b, err = ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))
	// res, err = Unmarshal(resp.Body)

	if resp.Close {
		// ignore error
		resp.Body.Close()
	}

	return res, err
}

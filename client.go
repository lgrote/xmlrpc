package xmlrpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client interface {
	Call(method string, args ...interface{}) (interface{}, error)
}

type clientImpl struct {
	client *http.Client
	url    *url.URL
}

func NewClient(url *url.URL) (Client, error) {
	return &clientImpl{client: new(http.Client), url: url}, nil
}

func (this *clientImpl) Call(method string, args ...interface{}) (interface{}, error) {

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
	if err != nil {
		return nil, fmt.Errorf("error calling rpc endpoint: %v", err)
	}
	var b []byte
	b, err = ioutil.ReadAll(resp.Body)
	fmt.Printf("Resp: %s\n", string(b))
	if err != nil {
		return nil, fmt.Errorf("error reading rpc result body: %v", err)
	}
	err = Unmarshal(resp.Body, &res)

	if resp.Close {
		// ignore error
		resp.Body.Close()
	}

	return res, err
}

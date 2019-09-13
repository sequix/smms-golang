package smms

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
)

const (
	smmsHost = "https://sm.ms"
)

type client struct {
	token  string
	client *http.Client
}

func New(username, password string) (*client, error) {
	token, err := Token(username, password)
	if err != nil {
		return nil, err
	}
	return &client{
		token:  token,
		client: http.DefaultClient,
	}, nil
}

func NewFromToken(token string) *client {
	return &client{
		token:  token,
		client: http.DefaultClient,
	}
}

func (c *client) SetHTTPClient(client *http.Client) {
	c.client = client
}

func Token(username, password string, clients ...*http.Client) (string, error) {
	client := http.DefaultClient
	if len(clients) > 0 {
		client = clients[0]
	}

	bodyStr := url.Values{"username": {username}, "password": {password}}.Encode()

	rsp, err := do(
		client,
		http.MethodPost,
		smmsHost+"/api/v2/token",
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Accept":       "application/json",
		},
		strings.NewReader(bodyStr),
	)
	if err != nil {
		return "", err
	}

	data := tokenData{}
	if err := mapstructure.Decode(rsp.Data, &data); err != nil {
		return "", fmt.Errorf("smms client: not proper data field for token %v: %w", rsp.Data, err)
	}
	return data.Token, nil
}

// 池化上传buffer，减少gc开销
var uploadBufferPool byteBufferPool

func (c *client) Upload(filename string, img io.Reader) (*ImageRsp, error) {
	buf := uploadBufferPool.Get()
	defer uploadBufferPool.Put(buf)

	w := multipart.NewWriter(buf)
	boundary := w.Boundary()

	ww, err := w.CreateFormFile("smfile", filename)
	if err != nil {
		return nil, fmt.Errorf("smms client: creating multipart form file: %w", err)
	}
	written, err := io.Copy(ww, img)
	if err != nil {
		return nil, fmt.Errorf("smms client: generating req for %s with %d bytes written: %w",
			filename, written, err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("smms client: closing multipart writer: %w", err)
	}

	rsp, err := do(
		c.client,
		http.MethodPost,
		smmsHost+"/api/v2/upload",
		map[string]string{
			"Content-Type":  "multipart/form-data; boundary=" + boundary,
			"Accept":        "application/json",
			"Authorization": c.token,
		},
		buf.NewReader(),
	)
	if err != nil {
		return nil, err
	}

	data := ImageRsp{}
	if err := mapstructure.Decode(rsp.Data, &data); err != nil {
		return nil, fmt.Errorf("smms client: not proper data field for upload %v: %w", rsp.Data, err)
	}
	return &data, nil
}

func (c *client) History() ([]ImageRsp, error) {
	rsp, err := do(
		c.client,
		http.MethodGet,
		smmsHost+"/api/v2/history",
		map[string]string{
			"Accept":        "application/json",
			"Authorization": c.token,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	data := []ImageRsp{}
	if err := mapstructure.Decode(rsp.Data, &data); err != nil {
		return nil, fmt.Errorf("smms client: not proper data field for history %v: %w", rsp.Data, err)
	}
	return data, nil
}

func (c *client) UploadHistory() ([]ImageRsp, error) {
	rsp, err := do(
		c.client,
		http.MethodGet,
		smmsHost+"/api/v2/upload_history",
		map[string]string{
			"Accept":        "application/json",
			"Authorization": c.token,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	data := []ImageRsp{}
	if err := mapstructure.Decode(rsp.Data, &data); err != nil {
		return nil, fmt.Errorf("smms client: not proper data field for upload_history %v: %w", rsp.Data, err)
	}
	return data, nil
}

func (c *client) Profile() (*ProfileRsp, error) {
	rsp, err := do(
		c.client,
		http.MethodPost,
		smmsHost+"/api/v2/profile",
		map[string]string{
			"Accept":        "application/json",
			"Authorization": c.token,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	data := ProfileRsp{}
	if err := mapstructure.Decode(rsp.Data, &data); err != nil {
		return nil, fmt.Errorf("smms client: not proper data field for profile %v: %w", rsp.Data, err)
	}
	return &data, nil
}

func (c *client) Delete(hash string) error {
	_, err := do(
		c.client,
		http.MethodGet,
		smmsHost+"/api/v2/delete/"+hash,
		map[string]string{
			"Accept":        "application/json",
			"Authorization": c.token,
		},
		nil,
	)
	return err
}

func (c *client) Clear() error {
	_, err := do(
		c.client,
		http.MethodGet,
		smmsHost+"/api/v2/clear",
		map[string]string{
			"Accept":        "application/json",
			"Authorization": c.token,
		},
		nil,
	)
	return err
}

func do(
	client *http.Client,
	method string,
	url string,
	headers map[string]string,
	body io.Reader,
) (*smmsResponse, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("smms client: building req: %w", err)
	}

	rh := req.Header
	for k, v := range headers {
		rh.Set(k, v)
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("smms client: doing req %v: %w", req, err)
	}

	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("smms client: reading rsp body of req %v: %w", req, err)
	}

	smmsRsp := &smmsResponse{}
	if err := json.Unmarshal(buf, smmsRsp); err != nil {
		return nil, fmt.Errorf("smms client: unmarshaling rsp body of req %v: %w", req, err)
	}

	if !smmsRsp.Success || rsp.StatusCode > 399 {
		return smmsRsp, fmt.Errorf("smms server: statusCode: %d method: %s url: %s requestID: %s code: %s message: %s",
			rsp.StatusCode, method, url, smmsRsp.RequestID, smmsRsp.Code, smmsRsp.Message)
	}

	if err := rsp.Body.Close(); err != nil {
		return nil, fmt.Errorf("smms client: closing rsp.Body of req %v: %w", req, err)
	}

	return smmsRsp, nil
}

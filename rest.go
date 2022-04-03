package transco

import (
	"github.com/go-resty/resty/v2"
)

type Rest struct {
	*resty.Client
}

type RestRequest struct {
	*resty.Request
}

func NewRest() *Rest {
	r := &Rest{
		Client: resty.New(),
	}

	return r
}

type ErrResponse struct {
	Err string `json:"err"`
	Msg string `json:"msg"`
}

type OkResponse struct {
	Data interface{} `json:"data"`
}

// result must be a pointer that will be decoded to
func (r *Rest) requester() *RestRequest {
	// respBody := &OkResponse{Data: result}

	req := r.R().
		SetHeader("Content-Type", "application/json").
		// SetResult(respBody).
		SetError(&ErrResponse{}).
		SetBody(`{}`) // default value for POST

	return &RestRequest{req}
}

func (req *RestRequest) SetResult(result interface{}) *RestRequest {
	okRes := &OkResponse{Data: result}
	req.Request.SetResult(okRes)

	return req
}

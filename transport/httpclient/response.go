package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Tsukikage7/servex/errors"
)

type Response struct {
	*http.Response
}

func (r *Response) JSON(v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func (r *Response) Text() (string, error) {
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *Response) Bytes() ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func (r *Response) CheckStatus() error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}
	return errors.New(
		r.StatusCode,
		fmt.Sprintf("http.%d", r.StatusCode),
		fmt.Sprintf("HTTP %d: %s", r.StatusCode, http.StatusText(r.StatusCode)),
	).WithHTTP(r.StatusCode)
}

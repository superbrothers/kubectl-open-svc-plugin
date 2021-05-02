package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func StripModifierFunc(s string) func(r *http.Response) error {
	return func(r *http.Response) error {
		for key, vals := range r.Header {
			for i, val := range vals {
				vals[i] = strings.ReplaceAll(val, s, "")
			}
			r.Header[key] = vals
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		b = bytes.ReplaceAll(b, []byte(s), []byte(""))
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		r.Header.Set("Content-Length", strconv.Itoa(len(b)))

		return nil
	}
}

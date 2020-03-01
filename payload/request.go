// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func NewEmptyHttpRequest() *HttpRequest {
	return &HttpRequest{Payload: NewNamespaced("request")}
}

func NewHttpRequestFromMap(m map[string]interface{}) *HttpRequest {
	r := NewEmptyHttpRequest()
	r.Data = m
	return r
}

func NewHttpRequestFromHTTP(hr *http.Request) (*HttpRequest, error) {
	r := NewEmptyHttpRequest()
	if err := r.SetVersion(fmt.Sprintf("%v.%v", hr.ProtoMajor, hr.ProtoMinor)); err != nil {
		return nil, err
	}
	if err := r.SetMethod(hr.Method); err != nil {
		return nil, err
	}
	if err := r.SetQuery(hr.URL.Query()); err != nil {
		return nil, err
	}
	if err := r.SetHeaders(hr.Header); err != nil {
		return nil, err
	}

	scheme := "http"
	if hr.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s%s", scheme, hr.Host, hr.URL.Path)
	if err := r.SetURL(url); err != nil {
		return nil, err
	}

	// Parse submitted form data. ParseForm already ignores invalid HTTP methods.
	// NOTE: Multipart forms and files are not initialized here. Files must be
	//       assigned by calling SetFiles() after calling this function.
	if err := hr.ParseForm(); err != nil {
		return nil, err
	}
	// TODO: test that this form has the submitted values when multipart is used
	if len(hr.PostForm) != 0 {
		if err := r.SetPostData(hr.PostForm); err != nil {
			return nil, err
		}
	}

	// Read body contents
	body, err := ioutil.ReadAll(hr.Body)
	if err != nil {
		return nil, err
	}
	if err := r.SetBody(body); err != nil {
		return nil, err
	}

	return r, nil
}

type HttpRequest struct {
	*Payload
}

func (r *HttpRequest) SetVersion(version string) error {
	if err := r.Set("version", version); err != nil {
		return err
	}
	return nil
}

func (r *HttpRequest) GetVersion() string {
	return r.GetString("version")
}

func (r *HttpRequest) SetMethod(method string) error {
	if err := r.Set("method", method); err != nil {
		return err
	}
	return nil
}

func (r *HttpRequest) GetMethod() string {
	return r.GetString("method")
}

func (r *HttpRequest) SetURL(url string) error {
	if err := r.Set("url", url); err != nil {
		return err
	}
	return nil
}

func (r *HttpRequest) GetURL() string {
	return r.GetString("url")
}

func (r *HttpRequest) SetQuery(v url.Values) error {
	if err := r.Set("query", v); err != nil {
		return err
	}
	return nil
}

func (r *HttpRequest) GetQuery() url.Values {
	queryValues := url.Values{}
	for name, values := range r.GetMap("query") {
		for _, v := range values.([]interface{}) {
			queryValues.Add(name, v.(string))
		}
	}
	return queryValues
}

func (r *HttpRequest) SetHeaders(h http.Header) error {
	if err := r.Set("headers", h); err != nil {
		return err
	}
	return nil
}

func (r *HttpRequest) GetHeaders() http.Header {
	header := http.Header{}
	for name, values := range r.GetMap("header") {
		for _, v := range values.([]interface{}) {
			header.Add(name, v.(string))
		}
	}
	return header
}

func (r *HttpRequest) SetPostData(v url.Values) error {
	if err := r.Set("post_data", v); err != nil {
		return err
	}
	return nil
}

func (r *HttpRequest) GetPostData() url.Values {
	formValues := url.Values{}
	for name, values := range r.GetMap("post_data") {
		for _, v := range values.([]interface{}) {
			formValues.Add(name, v.(string))
		}
	}
	return formValues
}

func (r *HttpRequest) SetBody(b []byte) error {
	if err := r.Set("body", b); err != nil {
		return err
	}
	return nil
}

func (r *HttpRequest) GetBody() []byte {
	return r.GetDefault("body", nil).([]byte)
}

func (r *HttpRequest) SetFiles(files []*File) error {
	var fps []map[string]interface{}
	for _, f := range files {
		fps = append(fps, f.Data)
	}
	return r.Set("files", fps)
}

func (r *HttpRequest) GetFiles() (fs []*File) {
	for _, m := range r.GetSliceMap("files") {
		fs = append(fs, NewFileFromMap(m))
	}
	return nil
}

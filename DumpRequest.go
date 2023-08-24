package main

import (
	"fmt"
	"github.com/vitorsalgado/mocha/v3/params"
	"github.com/vitorsalgado/mocha/v3/reply"
	"io"
	"net/http"
)

func DumpRequest(r *http.Request, m reply.M, p params.P) (*reply.Response, error) {
	fmt.Println("request received : " + r.RequestURI)

	body, err := io.ReadAll(r.Body)

	fmt.Printf("headers: %v\n", r.Header)
	fmt.Printf("body: %s\n", string(body))
	fmt.Printf("post form: %s\n", r.PostForm)
	fmt.Printf("form: %s\n", r.Form)
	fmt.Printf("body err %v\n", err)

	response, _ := reply.OK().Build(r, m, p)
	return response, nil
}

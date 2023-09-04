package main

import (
	"fmt"
	"github.com/vitorsalgado/mocha/v3"
	"io"
)

type DumpRequestPostAction struct{}

func (action *DumpRequestPostAction) Run(args mocha.PostActionArgs) error {
	r := args.Request
	fmt.Println("request received : " + r.RequestURI)

	body, err := io.ReadAll(r.Body)

	fmt.Printf("headers: %v\n", r.Header)
	fmt.Printf("body: %s\n", string(body))
	fmt.Printf("post form: %s\n", r.PostForm)
	fmt.Printf("form: %s\n", r.Form)
	fmt.Printf("body err %v\n", err)

	return nil
}

var DumpRequest = &DumpRequestPostAction{}

package main

import (
	"fmt"
	"testing"

	"github.com/oneclickvirt/portchecker/email"
)

func TestMain(t *testing.T) {
	res := email.EmailCheck()
	fmt.Println(res)
}

package dError

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestA(t *testing.T) {
	//var chan1  chan int
	////close(chan1)
	//a, ok := <-chan1
	//spew.Dump(a, ok)
	var err interface{}
	err = NewError("123", errors.New("测呃呃"))

	newErr := err.(error)
	spew.Dump(newErr.Error())
}

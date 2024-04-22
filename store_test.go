package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T){
	key := "MyLooks"
	pathName := CASPathTransformFunc(key)
	fmt.Println(pathName)
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("Some jpg bytes"))
	if err := s.writeStream("myPicture", data); err != nil {
		t.Error(err)
	}
	
}
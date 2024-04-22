package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T){
	key := "MyLooks"
	pathName := CASPathTransformFunc(key)
	expectedPathName := "bdbdc/c7520/e3730/94b7a/93d34/a7ddc/ce322/1e025"
	fmt.Println(pathName)

	if pathName != expectedPathName{
		t.Error(t, "Have %s want %s", pathName, expectedPathName)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("Some jpg bytes"))
	if err := s.writeStream("myPicture", data); err != nil {
		t.Error(err)
	}
	
}
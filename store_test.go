package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPathTransformFunc(t *testing.T){
	key := "MyLooks"
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "bdbdcc7520e373094b7a93d34a7ddcce3221e025"
	expectedPathName := "bdbdc/c7520/e3730/94b7a/93d34/a7ddc/ce322/1e025"
	fmt.Println(pathKey.PathName)

	if pathKey.PathName != expectedPathName{
		t.Error(t, "Have %s want %s", pathKey.PathName, expectedPathName)
	}

	if pathKey.FileName != expectedOriginalKey{
		t.Error(t, "Have %s want %s", pathKey.FileName, expectedOriginalKey)
	}
}

func TestStroreDeleteKey(t *testing.T){
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myPicture"
	data := []byte("Some jpg bytes")

	if err := s.writeStream(key,bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil{
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "myPicture"
	data := []byte("Some jpg bytes")

	if err := s.writeStream(key,bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, err := ioutil.ReadAll(r)

	if string(b) != string(data){
		t.Errorf("want %s have %s", data, b)
	}

	// s.Delete(key)
}
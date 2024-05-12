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
	expectedFileName := "bdbdcc7520e373094b7a93d34a7ddcce3221e025"
	expectedPathName := "bdbdc/c7520/e3730/94b7a/93d34/a7ddc/ce322/1e025"
	fmt.Println(pathKey.PathName)

	if pathKey.PathName != expectedPathName{
		t.Error(t, "Have %s want %s", pathKey.PathName, expectedPathName)
	}

	if pathKey.FileName != expectedFileName{
		t.Error(t, "Have %s want %s", pathKey.FileName, expectedFileName)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	id := generateId()
	defer teardown(t, s)
	for i:= 0; i < 50; i++{
		key :=fmt.Sprintf("Dis_Nuts_%d", i)
		data := []byte("Some jpg bytes")

		if _,err := s.writeStream(id, key,bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(id, key); !ok{
			t.Errorf("expected to have key %s", key)
		}
		
		_, r, err := s.Read(id, key)
		if err != nil {
			t.Error(err)
		}

		b, err := ioutil.ReadAll(r)

		if string(b) != string(data){
			t.Errorf("want %s have %s", data, b)
		}

		if err := s.Delete(id, key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(id, key); ok{
			t.Errorf("expected to NOT have key %s", key)
		}

	}
}

func newStore() *Store{
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store){
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
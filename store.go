package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:])

	// blockSize describes the depth of the folder tree
	blockSize := 5
	sliceLength := len(hashString) / blockSize

	paths := make([]string, sliceLength)

	for i := 0; i < sliceLength; i++{
		from, to := i * blockSize, (i * blockSize) + blockSize
		paths[i] = hashString[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashString,
	}
}

type PathTransformFunc func (string) PathKey

type PathKey struct{
	PathName string
	FileName string
}

func (p PathKey) fullPath () string{
	return fmt.Sprintf("%s%s", p.PathName, p.FileName)
}

type StoreOpts struct{
	PathTransformFunc PathTransformFunc
}

type Store struct{
	StoreOpts
}

var DefaultPathTransformFunc = func (key string) string {
	return key
}

func NewStore(opts StoreOpts) *Store{
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Read(key string) (io.Reader, error){
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Store) readStream(key string)(io.ReadCloser, error){
	return os.Open(s.PathTransformFunc(key).fullPath())
}

func (s *Store) writeStream(key string, r io.Reader) error{
	pathKey := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	
	fullPath := pathKey.fullPath()

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil{
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, fullPath)

	return nil
}

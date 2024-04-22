package main

import (
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
		Original: hashString,
	}
}

type PathTransformFunc func (string) PathKey

type PathKey struct{
	PathName string
	Original string
}

func (p PathKey) filename () string{
	return fmt.Sprintf("%s%s", p.PathName, p.Original)
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



func (s *Store) writeStream(key string, r io.Reader) error{
	pathKey := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	
	pathAndFileName := pathKey.filename()

	f, err := os.Create(pathAndFileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil{
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, pathAndFileName)

	return nil
}
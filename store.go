package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "GGNetwork"

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

func (p PathKey) FirstPathName() string{
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p PathKey) fullPath () string{
	return fmt.Sprintf("%s%s", p.PathName, p.FileName)
}

type StoreOpts struct{
	// Root is the folder name of the root, containing all the files and folders of the system
	Root string
	PathTransformFunc PathTransformFunc
}

type Store struct{
	StoreOpts
}

var DefaultPathTransformFunc = func (key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

func NewStore(opts StoreOpts) *Store{
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}

	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(id, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.fullPath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, fs.ErrNotExist)
	
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(id, key string) error{
	pathKey := s.PathTransformFunc(key)

	defer func ()  {
		log.Printf("deleted [%s] from disk", pathKey.FileName)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id,pathKey.FirstPathName())		

	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Write(id string,key string, r io.Reader) (int64, error){
	return s.writeStream(id, key, r)
}

func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader)(int64, error){
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}

	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Store) openFileForWriting(id, key string)(*os.File, error){
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id,pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.fullPath())

	return os.Create(fullPathWithRoot)
}

func (s *Store) writeStream(id, key string, r io.Reader) (int64,error){
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

func (s *Store) Read(id, key string) (int64, io.Reader, error){
	return s.readStream(id, key)
}

func (s *Store) readStream(id, key string)(int64, io.ReadCloser, error){
	pathKey := s.PathTransformFunc(key) 
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s",s.Root,id, pathKey.fullPath())

	// fi, err := os.Stat(fullPathWithRoot)
	// if err != nil {
	// 	return 0, nil, err
	// }

	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fi ,err := file.Stat()
	if err != nil {
		return 0, nil ,err
	}
	return fi.Size(), file, err
}



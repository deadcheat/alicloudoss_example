package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/apex/log"
	"golang.org/x/sync/semaphore"
)

const (
	UploadChunkSize   int64 = 100 * 1024
	UploadConcurrency       = 30
	UploadTaskWeight        = 1
)

func main() {
	targetDir := flag.String("d", "", "directory has files to upload")
	flag.Parse()

	if targetDir == nil || *targetDir == "" {
		log.Errorf("missing args: `-d dirname`")
		os.Exit(1)
	}

	c, err := oss.New(Endpoint, AccessKeyID, AccessKeySecret)
	if err != nil {
		log.Errorf("error when connect endpoint %v \n", err)
		os.Exit(1)
	}

	b, err := c.Bucket(BucketName)
	if err != nil {
		log.Errorf("error when access bucket %v \n", err)
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(*targetDir)
	if err != nil {
		log.Errorf("error when access target directory %s, %v \n", *targetDir, err)
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)
	abPath, _ := filepath.Abs(*targetDir)
	s := semaphore.NewWeighted(UploadConcurrency)
	for i := range files {
		f := files[i]
		if f.IsDir() {
			continue
		}
		wg.Add(1)
		filePath := filepath.Join(abPath, f.Name())
		go func(objectKey, path string) {
			fmt.Println("upload:", path)
			defer wg.Done()
			s.Acquire(context.Background(), UploadTaskWeight)
			defer s.Release(UploadTaskWeight)
			err = b.UploadFile(objectKey, path, UploadChunkSize)
			if err != nil {
				errChan <- err
			}
		}(f.Name(), filePath)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	errors := make([]error, 0)
	func() {
		for {
			select {
			case err, ok := <-errChan:
				if !ok {
					fmt.Println("error chan closed")
					return
				}
				errors = append(errors, err)
			}
		}
	}()
	if len(errors) > 0 {
		fmt.Println("errors below occurred during processing")
		for i := range errors {
			e := errors[i]
			fmt.Printf("  %d : %v \n", i, e)
		}
	}
	fmt.Printf("upload succeded \n")
}

package main

import (
	"fmt"
	"strings"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Test a single read/write pair.
func TestPipe1(t *testing.T) {
	data := []byte("hello, world")
	r, w := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(2)

	var buf = make([]byte, 64)
	go func() {
		defer func() {
			r.Close()
			wg.Done()
		}()

		n, err := r.Read(buf)
		if err != nil {
			t.Errorf("read: %v", err)
		} else if n != 12 || string(buf[0:12]) != "hello, world" {
			t.Errorf("bad read: got %q", buf[0:n])
		}
	}()

	go func() {
		defer func() {
			w.Close()
			wg.Done()
		}()

		n, err := w.Write(data)
		if err != nil {
			t.Errorf("write: %v", err)
		}
		if n != len(data) {
			t.Errorf("short write: %d != %d", n, len(data))
		}
	}()

	wg.Wait()

	fmt.Println(string(buf[0:13]))
}

//func (f *File) Seek(offset int64, whence int)
//
//方法标准格式是：seek(offset,whence=0)
//
//offset：开始偏移量，也就是代表需要移动偏移的字节数。
//
//whence：给offset参数一个定义，表示要从哪个位置开始偏移；0代表从文件开头开始算起，1代表从当前位置开始算起，2代表从文件末尾算起。

func TestSeek(t *testing.T) {
	f, err := os.Open("./test.txt")
	if err != nil {
		t.Fatal(err)
	}

	fileStat, err := f.Stat() //Get info from file
	if err != nil {
		return
	}

	fileSize := fileStat.Size()
	fileHeader := make([]byte, 2)
	f.Read(fileHeader) //读入文件头部
	fileContentType := http.DetectContentType(fileHeader)

	fmt.Println(fileStat.Mode(), fileSize, fileContentType)

	r, w := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer func() {
			w.Close()
			wg.Done()
		}()

		//f.Seek(5,0)//从文件开头开始，偏移5
		//f.Seek(5,1)//从当前位置开始，偏移5
		f.Seek(-5, 2) //从文件末尾开始，向前偏移5

		if _, err = io.Copy(w, f); err != nil { //从文件中读取数据
			return
		}
	}()

	var buf = make([]byte, 64)
	go func() {
		defer func() {
			r.Close()
			wg.Done()
		}()

		_, err := r.Read(buf)
		if err != nil {
			t.Errorf("read: %v", err)
		}
	}()

	wg.Wait()

	fmt.Println(string(buf[0:6]))
}

func TestSeekError(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Seek(0, 0)
	if err == nil {
		t.Fatal("Seek on pipe should fail")
	}

	if perr, ok := err.(*os.PathError); !ok || perr.Err != syscall.ESPIPE {
		t.Errorf("Seek returned error %v, want &PathError{Err: syscall.ESPIPE}", err)
	}

	_, err = w.Seek(0, 0)
	if err == nil {
		t.Fatal("Seek on pipe should fail")
	}

	if perr, ok := err.(*os.PathError); !ok || perr.Err != syscall.ESPIPE {
		t.Errorf("Seek returned error %v, want &PathError{Err: syscall.ESPIPE}", err)
	}
}

func TestRemoveDirFile(t *testing.T) {
	root := "./"

	now := time.Now()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if !strings.Contains(info.Name(),".del"){
			return nil
		}

		if now.Sub(info.ModTime()) > time.Duration(time.Minute*1) {
			fmt.Println(info.Name()+"文件创建已超时，需要删除")

			os.Remove(info.Name())

		}
		return nil
	})
	if err != nil {
		t.Errorf("filepath.Walk error(%v)", err)
	}
}

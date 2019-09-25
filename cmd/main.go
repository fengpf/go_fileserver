package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"go_fileserver"
)

var (
	addr = flag.String("addr", ":9000", "http service address")

	fh, _ = go_fileserver.NewFileServer(
		&go_fileserver.FileInfo{},
	)

	data = [][]string{{"姓名", "电话", "公司", "职位", "加入时间"}, {"1", "2", "刘犇,刘犇,刘犇", "4", "5"}}
)

func main() {
	flag.Parse()

	schedule := go_fileserver.NewSchedule()
	go schedule.Run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/csv/export/", export)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		go_fileserver.ServeWs(schedule, w, r)
	})

	fmt.Println("file server start")

	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "home.html")
}

func export(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/csv/export/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	fmt.Println(r.URL.Path, filename)

	csv := go_fileserver.CSV{
		Data:  data,
		Title: fmt.Sprintf("%s", time.Now().Format("2006-01-02")),
	}

	csv.Render(w)
	csv.WriteContentType(w)
}

func writeCsvFile(filePath string) {
	fp, err := os.Create(filePath) // 创建文件句柄
	if err != nil {
		log.Fatalf("创建文件["+filePath+"]句柄失败,%v", err)
		return
	}
	defer fp.Close()

	csv := go_fileserver.CSV{
		Data:  data,
		Title: fmt.Sprintf("%s", time.Now().Format("2006-01-02")),
	}

	csv.Render(fp)
}

package go_fileserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type FileServer struct {
	FileInfo *FileInfo
}

func NewFileServer(fi *FileInfo) (fs *FileServer, err error) {
	fs = &FileServer{
		FileInfo: NewFile(fi.Dir, fi.FileName),
	}

	return
}

func (fs *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fs.FileInfo.OpenFile()

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	file := r.FormValue("file")
	if file == "" {
		return
	}

}

type Req struct {
	Data    []byte
	MsgType int
	Params  string
	Client  *WSClient
}

//设置活跃的客户端连接，分发客户端请求
type Schedule struct {
	// 已经注册的客户端
	wsClientsMap map[*WSClient]chan *Req

	// 分发客户端请求
	dispatch chan *Req

	// 注册请求
	register chan *WSClient

	// 删除已注册请求
	unregister chan *WSClient
}

func NewSchedule() *Schedule{
	return &Schedule{
		dispatch:  make(chan *Req),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		wsClientsMap:  make(map[*WSClient]chan *Req),
	}
}

func (s *Schedule) Run() {
	for {
		select {
		case client := <-s.register:
			s.wsClientsMap[client] = client.send

		case client := <-s.unregister:
			if _, ok := s.wsClientsMap[client]; ok {
				delete(s.wsClientsMap, client)
				close(client.send)
			}

		case msg := <-s.dispatch:
			fmt.Println(s.wsClientsMap)

			if _, ok := s.wsClientsMap[msg.Client]; !ok {
				continue
			}

			switch msg.MsgType {
			case websocket.TextMessage:
				if msg.Client.params == msg.Params {

					tmp := string(msg.Data)
					tmp += "我处理了10秒钟~~~"
					msg.Data = []byte(tmp)

					time.Sleep(time.Second * 10) //假设业务处理
					select {
					case msg.Client.send <- msg:
					default:
						close(msg.Client.send)
						delete(s.wsClientsMap, msg.Client)
					}
				}
			}
		}
	}
}

package go_fileserver

import (
	"github.com/gorilla/websocket"
	"net/http"
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
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	wsClients map[*WSClient]chan *Req

	// Inbound messages from the clients.
	broadcast chan *Req

	// Register requests from the clients.
	register chan *WSClient

	// Unregister requests from clients.
	unregister chan *WSClient
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Req),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		wsClients:  make(map[*WSClient]chan *Req),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.wsClients[client] = client.send

		case client := <-h.unregister:
			if _, ok := h.wsClients[client]; ok {
				delete(h.wsClients, client)
				close(client.send)
			}

		case msg := <-h.broadcast:


			switch msg.MsgType {
			case websocket.TextMessage:

			}


			for client := range h.wsClients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.wsClients, client)
				}
			}
		}
	}
}

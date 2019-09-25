package go_fileserver

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// websocket 客户端
type WSClient struct {
	schedule *Schedule

	conn *websocket.Conn

	// 发送消息队列
	send chan *Req

	msgType int

	params string
}

// 升级websocket请求
func ServeWs(schedule *Schedule, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("get param", r.FormValue("filename"))

	client := &WSClient{
		schedule: schedule,
		conn:     conn,
		send:     make(chan *Req, 2),
		params:   r.FormValue("filename"),
	}
	client.schedule.register <- client

	go client.read()

	go client.write()
}

//读取数据并分发
func (c *WSClient) read() {
	defer func() {
		c.schedule.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		fmt.Println("read", string(data))
		c.schedule.dispatch <- &Req{
			MsgType: msgType,
			Data:    data,
			Params:  c.params,
			Client:  c,
		}
	}
}

//消费send 带缓冲的chan队列并执行所有写数据请求
func (c *WSClient) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //写超时
			if !ok || msg == nil {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) //关闭连接
				return
			}

			w, err := c.conn.NextWriter(msg.MsgType)
			if err != nil {
				return
			}

			w.Write(msg.Data)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

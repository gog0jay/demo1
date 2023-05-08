package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.Mutex

	//广播消息的channel
	Message chan string
}

// NewServer 创建一个Server的接口
func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

}

// ListenMessage 监听Message广播消息channel的goroutine，一旦有消息，反送给全部在线user
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		s.mapLock.Lock()
		for _, u := range s.OnlineMap {
			u.C <- msg
		}

		s.mapLock.Unlock()
	}
}

// BroadCast 广播消息的方法
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s] %s:%s", user.Addr, user.Name, msg)

	s.Message <- sendMsg

}

func (s *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	//fmt.Println("链接建立成功")

	user := NewUser(conn)
	// 用户上线，将用户加入到OnlineMap中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// 广播当前用户上线的消息
	s.BroadCast(user, "already online")

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				s.BroadCast(user, "is offline")
			}
			if err != nil && err != io.EOF {
				fmt.Print("Conn Read err: ", err)
			}
			// 提取用户的消息，去除最后的\n
			msg := strings.Trim(string(buf[:]), "\r\n")
			s.BroadCast(user, msg)
		}
	}()
	// 当前handler阻塞
	select {}
}

// Start 启动服务器的接口
func (s *Server) Start() {
	// 1. socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))

	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	// 4. close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go s.ListenMessage()

	for {
		// 2. accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// 3. do handler
		go s.Handler(conn)
	}

}

package main

import (
	"fmt"
	"net"
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

// 监听Message广播消息channel的goroutine，一旦有消息，反送给全部在线user
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}

		this.mapLock.Unlock()
	}
}

// BroadCast 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s] %s:%s", user.Addr, user.Name, msg)

	this.Message <- sendMsg

}

func (this *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	fmt.Println("链接建立成功")

	user := NewUser(conn)
	// 1. 用户上线，将用户加入到OnlineMap中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 2. 广播当前用户上线的消息
	this.BroadCast(user, "已上线")

	// 3. 当前handler阻塞
	select {}
}

// Start 启动服务器的接口
func (this *Server) Start() {
	// 1. socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))

	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	// 4. close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessage()

	for {
		// 2. accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// 3. do handler
		go this.Handler(conn)
	}

}

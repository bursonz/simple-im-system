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

	// OnlineMap 锁
	mapLock sync.RWMutex

	//消息广播的Channel
	Message chan string
}

// NewServer 创建一个Server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMessager 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message

		//将msg发送给全部的在线User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.Ch <- msg
		}
		s.mapLock.Unlock()

	}
}

// BroadCast 广播消息的方法
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	//...当前连接的业务
	//fmt.Println("连接建立成功")

	user := NewUser(conn)

	//用户上线，将用户加入到OnlineMap中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	//广播当前用户上线消息
	s.BroadCast(user, "已上线")

	//当前handler阻塞
}

// Start 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	//close Listen socket
	defer listener.Close()

	//启动监听Message的goroutine
	go s.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		//do handler
		go s.Handler(conn)
	}
}

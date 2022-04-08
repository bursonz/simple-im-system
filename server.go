package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

	user := NewUser(conn, s)

	//用户上线
	user.Online()

	//监听用户是否活跃的Channel
	isLive := make(chan bool)

	//接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			//从conn中读取字节
			n, err := conn.Read(buf)

			//返回长度为0，用户下线
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])

			//用户针对msg进行消息处理
			user.DoMessage(msg)

			//用户的任意消息,代表当前用户是一个活跃的
			isLive <- true

		}
	}()

	//当前handler阻塞
	for {
		select {
		case <-isLive:
			//当前用户是活跃的,重置定时器
			//不做任何事情,为了激活select,更新下面的定时器

		case <-time.After(time.Second * 10):
			//已经超时
			//将当前的User强制关闭
			user.SendMsg("长时间不活跃,已被踢下线")

			//销毁用的资源
			close(user.Ch)

			//关闭资源
			conn.Close()

			//退出当前Handler
			return //runtime.Goexit()
		}
	}
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

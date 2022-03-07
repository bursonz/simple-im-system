package main

import "net"

type User struct {
	Name string
	Addr string
	Ch   chan string
	conn net.Conn
}

// NewUser 创建一个用户的API
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		Ch:   make(chan string),
		conn: conn,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.Ch
		u.conn.Write([]byte(msg + "\n"))
	}
}
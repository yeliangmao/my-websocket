package request

import "github.com/gorilla/websocket"

// Node 用户连接节点结构体，用于管理单个用户的WebSocket连接及相关通信通道
// Node struct for user connection, managing a single user's WebSocket connection and related communication channels
type Node struct {
	Conn *websocket.Conn // WebSocket连接实例 / WebSocket connection instance
	Data chan []byte     // 数据通道，用于接收待发送给用户的消息 / Data channel for receiving messages to be sent to the user
	Exit chan bool       // 退出通道，用于通知关闭连接 / Exit channel for notifying connection closure
}
type InitialInformation struct {
	Id int
}

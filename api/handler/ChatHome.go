package handler

import (
	"Gin/api/request"
	"Gin/global/model"
	"Gin/inits"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// ChatHome WebSocket聊天连接建立的核心业务接口
// ChatHome is the core business interface for establishing WebSocket chat connections
func ChatHome(context *gin.Context) {
	model.ActiveConnWG.Add(1) //记录在线用户
	defer model.ActiveConnWG.Done()
	var ID int
	// 1. 从Redis的LoginBucket中获取一个用户ID（实现ID复用机制）
	// 1. Get a user ID from Redis's LoginBucket (implements ID reuse mechanism)
	if id, err := model.RDB.RPop(model.Ctx, "LoginBucket").Int(); err != nil {
		model.Logger.Error("Connection failed: Failed to get user ID from LoginBucket", zap.Error(err))
		return
	} else {
		ID = id
	}
	// 延迟操作：连接关闭后将ID归还到LoginBucket，确保ID可循环使用
	// Deferred operation: Return ID to LoginBucket after connection closes to ensure ID reuse
	defer model.RDB.LPush(model.Ctx, "LoginBucket", ID)

	// 2. 初始化用户连接节点（Node），用于管理WebSocket连接、数据通道和退出通知
	// 2. Initialize user connection node (Node) to manage WebSocket connection, data channel and exit notification

	// 3. 配置WebSocket升级器（将HTTP请求升级为WebSocket连接）
	// 3. Configure WebSocket upgrader (upgrades HTTP request to WebSocket connection)
	Upgrader := websocket.Upgrader{
		WriteBufferSize: 1024, // 写缓冲区大小 / Write buffer size
		ReadBufferSize:  1024, // 读缓冲区大小 / Read buffer size
		// 注意：若需处理跨域WebSocket请求，需添加CheckOrigin配置（例：CheckOrigin: func(r *http.Request) bool { return true }）
		// Note: To handle cross-origin WebSocket requests, add CheckOrigin config (e.g., CheckOrigin: func(r *http.Request) bool { return true })
		CheckOrigin: func(r *http.Request) bool { // 新增跨域允许
			return true
		},
	}
	var Node request.Node
	// 延迟操作：确保连接关闭、数据通道和退出通道释放，避免资源泄漏
	// Deferred operations: Ensure connection closure and release of data/exit channels to avoid resource leaks
	// 执行HTTP到WebSocket的连接升级
	// Execute HTTP to WebSocket connection upgrade
	if conn, err := Upgrader.Upgrade(context.Writer, context.Request, nil); err != nil {
		model.Logger.Error("websocket upgrade failed", zap.Error(err))
		return
	} else {
		// 升级成功，初始化Node的连接和通道
		// Upgrade successful, initialize Node's connection and channels
		Node = request.Node{
			Conn:     conn,              // WebSocket连接实例 / WebSocket connection instance
			Data:     make(chan []byte), // 消息发送通道（接收待推送给用户的消息） / Message send channel (receives messages to push to user)
			Exit:     make(chan bool),   // 连接退出通道（用于通知连接关闭） / Connection exit channel (notifies connection closure)
			ExitFlag: &sync.Once{},
		}
	}
	defer func() {
		// 关闭WebSocket连接（此时conn非nil）
		if err := Node.Conn.Close(); err != nil {
			model.Logger.Warn("关闭WebSocket连接失败", zap.Error(err))
		}
		// 关闭数据通道
		close(Node.Data)
		// 关闭退出通道
		close(Node.Exit)
	}()
	// 4. 将用户节点加入全局连接池，便于后续消息分发和连接管理
	// 4. Add user node to global connection pool for subsequent message distribution and connection management
	model.ConnectionPool[ID] = Node
	// 延迟操作：连接关闭后从连接池移除该用户，避免无效连接残留
	// Deferred operation: Remove user from connection pool after connection closes to avoid residual invalid connections
	defer delete(model.ConnectionPool, ID)

	// 5. 向Redis写入用户ID与节点标识（OnlyMark）的映射，用于跨节点消息路由
	// 5. Write mapping of user ID and node identifier (OnlyMark) to Redis for cross-node message routing
	// 0表示键永不过期 / 0 means the key never expires
	model.RDB.Set(model.Ctx, fmt.Sprintf("%d", ID), model.OnlyMark, 0)
	// 延迟操作：连接关闭后删除Redis中的用户ID映射，避免脏数据
	// Deferred operation: Delete user ID mapping in Redis after connection closes to avoid dirty data
	defer model.RDB.Del(model.Ctx, fmt.Sprintf("%d", ID))

	// 6. 将用户ID加入当前节点的Redis集合（OnlyMark），用于统计节点下在线用户
	// 6. Add user ID to Redis set (OnlyMark) of current node for counting online users under the node
	model.RDB.SAdd(model.Ctx, model.OnlyMark, fmt.Sprintf("%d", ID))
	// 延迟操作：连接关闭后从Redis集合移除用户ID，确保在线状态准确
	// Deferred operation: Remove user ID from Redis set after connection closes to ensure accurate online status
	defer model.RDB.SRem(model.Ctx, model.OnlyMark, fmt.Sprintf("%d", ID))

	// 7. 记录用户上线日志（X-Forwarded-For获取客户端真实IP，需反向代理配置支持）
	// 7. Record user online log (X-Forwarded-For gets client's real IP, requires reverse proxy configuration support)
	model.Logger.Info("Users go live", zap.String("Client IP", context.Request.Header.Get("X-Forwarded-For")))
	var ll = request.InitialInformation{Id: ID}
	InitialInformation, _ := json.Marshal(ll)
	Node.Conn.WriteMessage(websocket.TextMessage, InitialInformation)
	// 8. 使用WaitGroup等待读写协程完成，确保连接关闭前读写操作正常收尾
	// 8. Use WaitGroup to wait for read/write goroutines to complete, ensuring proper cleanup of read/write operations before connection closes
	go ChatWrite(Node, context.Request.Header.Get("X-Forwarded-For"))
	// 启动消息读取协程（接收客户端发送的消息）
	// Start message read goroutine (receives messages from client)
	go CharRead(Node, context.Request.Header.Get("X-Forwarded-For"), ID)
	var WG sync.WaitGroup
	WG.Add(1) // 注册1个待等待的协程（退出监听） / Register 1 goroutines to wait for (exit goroutine)
	go CharExit(&WG, Node)
	WG.Wait() // 阻塞等待读写协程结束 / Block and wait for read/write goroutines to end
	return
}

func CharExit(s *sync.WaitGroup, node request.Node) {
	for {
		ok := <-node.Exit
		if ok {
			s.Done()
			return
		}
	}
}

// ChatWrite 消息读取协程：从客户端接收消息并推送到数据通道
// ChatWrite message read goroutine: Receives messages from client and pushes to data channel
func ChatWrite(node request.Node, clientIP string) {
	for { // 改为无限循环，同时监听两个通道
		// 正常接收消息并发送
		data, ok := <-node.Data
		// 检查 data 通道是否已关闭（避免永久阻塞）
		if !ok {
			model.Logger.Info(fmt.Sprintf("User data channel closed: %s", clientIP))
			node.ExitFlag.Do(
				func() {
					node.Exit <- true
				},
			)
			return
		}
		// 发送消息
		if err := node.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			model.Logger.Error("Failed to write message", zap.String("Client IP", clientIP), zap.Error(err))
			node.ExitFlag.Do(
				func() {
					node.Exit <- true
				},
			)
			return
		}

	}
}

// CharRead 消息写入协程：从客户端读取消息并处理
// CharRead message write goroutine: Reads messages from client and processes them
func CharRead(node request.Node, Name string, ID int) {
	for {
		// 从WebSocket连接读取消息（忽略消息类型，仅关注消息内容）
		// Read message from WebSocket connection (ignore message type, only focus on content)
		_, message, err := node.Conn.ReadMessage()
		if err != nil {
			model.Logger.Error("read message failed", zap.String("Client IP", Name), zap.Error(err))
			// 读取消息失败，发送退出信号并终止协程
			// Failed to read message, send exit signal and terminate goroutine
			node.ExitFlag.Do(
				func() {
					node.Exit <- true
				},
			)
			return
		}

		// 1. 反序列化客户端消息（将JSON字节流转为Message结构体）
		// 1. Deserialize client message (convert JSON byte stream to Message struct)
		var Message model.Message
		if err := json.Unmarshal(message, &Message); err != nil {
			model.Logger.Error("unmarshal message failed", zap.String("Client IP", Name), zap.Error(err))
			// 反序列化失败，向客户端返回错误提示
			// Deserialization failed, return error prompt to client
			node.Data <- []byte("Message resolution failed")
			continue
		}
		model.Logger.Info("user seed data ok")
		// 2. 构造消息响应体（添加发送方ID，用于接收方识别来源）
		// 2. Construct message response body (add sender ID for receiver to identify source)
		var Response = model.Response{
			Data:   Message.Data,   // 消息内容 / Message content
			Target: Message.Target, // 消息目标（群ID或单个用户ID） / Message target (group ID or single user ID)
			Type:   Message.Type,   // 消息类型（group：群聊；once：单聊） / Message type (group: group chat; once: private chat)
			FormId: ID,             // 发送方用户ID / Sender user ID
		}

		// 3. 序列化响应体（转为JSON字节流，便于RabbitMQ传输）
		// 3. Serialize response body (convert to JSON byte stream for RabbitMQ transmission)
		data, err := json.Marshal(Response)
		if err != nil {
			model.Logger.Error("Data serialization failed", zap.String("Client IP", Name), zap.Error(err))
			// 序列化失败，向客户端返回错误提示
			// Serialization failed, return error prompt to client
			node.Data <- []byte("Message push failed")
			continue
		}
		// 4. 根据消息类型分发消息（通过RabbitMQ实现跨节点消息路由）
		// 4. Distribute messages by message type (achieve cross-node message routing via RabbitMQ)
		switch Message.Type {
		case "group":
			// 群聊消息：获取所有节点标识，向每个节点的RabbitMQ队列发送消息
			// Group chat message: Get all node identifiers, send message to each node's RabbitMQ queue
			OtherOnlyMarks := model.RDB.HGetAll(model.Ctx, "Nodes").Val()
			for mark, _ := range OtherOnlyMarks {
				// 检查当前节点的RabbitMQ连接是否存在，不存在则重新同步连接
				// Check if RabbitMQ connection for current node exists; resync if not
				_, ok := model.RabbieMqPoll[mark]
				if !ok {
					inits.PullAndConnRabbieMq()
				}
				// 向该节点的RabbitMQ队列发布群聊消息
				// Publish group chat message to the node's RabbitMQ queue
				model.RabbieMqPoll[mark].PublishSimple(string(data))
			}

		case "once":
			// 单聊消息：根据目标用户ID获取其所在节点标识，向对应节点的RabbitMQ队列发送消息
			// Private chat message: Get target user's node identifier by user ID, send message to the node's RabbitMQ queue
			OtherOnlyMark, RedisOk := model.RDB.Get(model.Ctx, fmt.Sprintf("%d", Message.Target)).Result()
			if RedisOk != nil {
				var res = model.Response{
					Data:   "当前用户不在线",
					Target: ID,
					Type:   "once",
					FormId: -1,
				}
				data, _ := json.Marshal(res)
				node.Conn.WriteMessage(websocket.TextMessage, data)
				continue
			}
			// 检查目标节点的RabbitMQ连接是否存在，不存在则重新同步连接
			// Check if RabbitMQ connection for target node exists; resync if not
			_, ok := model.RabbieMqPoll[OtherOnlyMark]
			if !ok {
				inits.PullAndConnRabbieMq()
			}
			// 向目标节点的RabbitMQ队列发布单聊消息
			// Publish private chat message to the target node's RabbitMQ queue
			model.RabbieMqPoll[OtherOnlyMark].PublishSimple(string(data))
			node.Conn.WriteMessage(websocket.TextMessage, data)
		default:
			// 非法消息类型：记录错误日志并向客户端返回提示
			// Illegal message type: Record error log and return prompt to client
			model.Logger.Error("Illegal message type", zap.String("Client IP", Name), zap.String("Message Type", Message.Type))
			node.Data <- []byte("Illegal import the type")
		}
	}
}

客户端连接分配一个随机的id，用于标识客户端

函数
ws.Start()
加载WebSocket模块

发送消息给指定id的客户端
ws.Send(id string, msg []byte)

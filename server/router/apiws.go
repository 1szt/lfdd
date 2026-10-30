package router

import (
	"log"

	"hxzl/core/web"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

type ApiWs struct {
}

// apiws WebSocket
func apiws() {
	ApiWs := ApiWs{}
	Group := web.Engine.Group("/api/ws")
	Group.GET("/", ApiWs.Ws)
}

// Ws WebSocket 路由处理函数
func (ApiWs) Ws(c *gin.Context) {
	// 获取客户端真实 IP
	clientIP := c.ClientIP()
	// 获取客户端请求参数 token
	token := c.Query("token")
	// 升级ws
	Upgrade(c, token, clientIP)

}

// Upgrade 升级ws函数
// 传入token ip地址
func Upgrade(c *gin.Context, token, ip string) {
	// 升级 HTTP 为 WebSocket 连接
	// AcceptOptions 里可以配置跨域，若前后端同域可不传，开发测试建议加上 InsecureSkipVerify: true
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("❌ [APIWS] 升级失败: %v", err)
		c.JSON(400, gin.H{"error": "升级失败"})
		return
	}

	// 确保连接在退出时关闭
	defer conn.CloseNow()

}

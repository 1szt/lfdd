// Package apitest 热重载测试页：手动触发 server/db 重载并观察效果，挂载于 /api/test/*。
package apitest

import (
	"hxzl/db"
	"hxzl/server"
	"hxzl/web"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Start 注册测试页路由（须在 web.Start 之后调用）
func Start() {
	group := web.Engine.Group("/api")
	group.GET("/test/reload", GetPage)
	group.GET("/test/reload/status", GetStatus)
	group.POST("/test/reload/apply", PostApply)
	log.Print("✅ [APITEST] 热重载测试页 加载完成！")
}

// GetPage 测试页面
func GetPage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(pageHTML))
}

// GetStatus 返回当前服务与数据库状态
func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"server": server.Status(),
			"db":     db.Status(),
		},
	})
}
// PostApply 触发重载：先返回响应再异步重载（Shutdown 会等待在途请求，同步调用会阻塞到超时）
func PostApply(c *gin.Context) {
	go server.Reload()
	go db.Reload()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "已触发重载，请观察日志与状态变化"})
}

const pageHTML = `<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<title>热重载测试</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 640px; margin: 40px auto; padding: 0 16px; color: #222; }
  h1 { font-size: 20px; }
  .card { border: 1px solid #ddd; border-radius: 8px; padding: 16px; margin: 16px 0; }
  .row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px dashed #eee; }
  .row:last-child { border-bottom: none; }
  .label { color: #888; }
  .val { font-family: Consolas, monospace; }
  button { font-size: 16px; padding: 10px 24px; border: none; border-radius: 6px; background: #2563eb; color: #fff; cursor: pointer; }
  button:disabled { background: #999; }
  #log { font-family: Consolas, monospace; font-size: 13px; background: #111; color: #0f0; border-radius: 8px; padding: 12px; min-height: 80px; white-space: pre-wrap; }
</style>
</head>
<body>
<h1>🔥 热重载测试页</h1>
<div class="card">
  <div class="row"><span class="label">服务状态</span><span class="val" id="srvStatus">-</span></div>
  <div class="row"><span class="label">监听地址</span><span class="val" id="srvAddr">-</span></div>
  <div class="row"><span class="label">数据库类型</span><span class="val" id="dbType">-</span></div>
  <div class="row"><span class="label">数据库状态</span><span class="val" id="dbStatus">-</span></div>
</div>
<button id="btn" onclick="apply()">重载 server + db</button>
<h3 style="margin-top:24px">操作日志</h3>
<div id="log"></div>
<script>
const $ = id => document.getElementById(id);
function log(msg) {
  $('log').textContent += '[' + new Date().toLocaleTimeString() + '] ' + msg + '\n';
}
async function refresh() {
  try {
    const t0 = performance.now();
    const r = await fetch('/api/test/reload/status');
    const d = (await r.json()).data;
    const ms = Math.round(performance.now() - t0);
    $('srvStatus').textContent = d.server.status;
    $('srvAddr').textContent = d.server.ip + ':' + d.server.port;
    $('dbType').textContent = d.db.type || '-';
    $('dbStatus').textContent = d.db.status + ' (' + ms + 'ms)';
  } catch (e) {
    $('srvStatus').textContent = '请求失败';
    log('状态查询失败: ' + e);
  }
}
async function apply() {
  $('btn').disabled = true;
  log('触发重载...');
  try {
    const r = await fetch('/api/test/reload/apply', { method: 'POST' });
    log((await r.json()).message);
  } catch (e) {
    log('触发失败（服务正在重启属正常现象）: ' + e);
  }
  // 重载期间请求会失败，轮询直到恢复
  let tries = 0;
  const t = setInterval(async () => {
    tries++;
    await refresh();
    if ($('srvStatus').textContent === 'running' || tries > 10) {
      clearInterval(t);
      $('btn').disabled = false;
      log('服务已恢复: ' + $('srvAddr').textContent);
    }
  }, 500);
}
refresh();
setInterval(refresh, 3000);
</script>
</body>
</html>`

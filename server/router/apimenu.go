package router

import (
	"net/http"

	"hxzl/core/web"

	"github.com/gin-gonic/gin"
)

// apimenu 菜单接口
func apimenu() {
	web.Engine.Group("/api").GET("/menu", GetMenu)
}

// -------- 数据结构 --------

type GoodsItem struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Desc  string `json:"desc"`
	Price int    `json:"price"`
}

type CategoryItem struct {
	ID    int         `json:"id"`
	Name  string      `json:"name"`
	Goods []GoodsItem `json:"goods"`
}

// -------- 模拟数据 --------

var menuData = []CategoryItem{
	{
		ID: 1, Name: "热销推荐",
		Goods: []GoodsItem{
			{ID: 101, Name: "招牌炒饭", Icon: "🍚", Desc: "大火爆炒 粒粒分明", Price: 18},
			{ID: 102, Name: "红烧牛肉面", Icon: "🍜", Desc: "浓郁汤底 大块牛肉", Price: 22},
			{ID: 103, Name: "香煎鸡排", Icon: "🍗", Desc: "外酥里嫩 鲜嫩多汁", Price: 25},
			{ID: 104, Name: "番茄蛋汤", Icon: "🍲", Desc: "酸甜开胃 营养丰富", Price: 10},
		},
	},
	{
		ID: 2, Name: "经典套餐",
		Goods: []GoodsItem{
			{ID: 201, Name: "商务A套餐", Icon: "🍱", Desc: "两荤一素 含汤", Price: 35},
			{ID: 202, Name: "商务B套餐", Icon: "🍱", Desc: "三荤一素 含汤", Price: 38},
			{ID: 203, Name: "学生套餐", Icon: "🍱", Desc: "一荤两素 含饭", Price: 25},
			{ID: 204, Name: "家庭套餐", Icon: "🍱", Desc: "四荤两素 3-4人", Price: 88},
		},
	},
	{
		ID: 3, Name: "凉菜小食",
		Goods: []GoodsItem{
			{ID: 301, Name: "凉拌黄瓜", Icon: "🥒", Desc: "清脆爽口", Price: 8},
			{ID: 302, Name: "皮蛋豆腐", Icon: "🥚", Desc: "经典凉菜", Price: 12},
			{ID: 303, Name: "口水鸡", Icon: "🐔", Desc: "麻辣鲜香", Price: 18},
			{ID: 304, Name: "酸辣土豆丝", Icon: "🥔", Desc: "酸辣开胃", Price: 10},
		},
	},
	{
		ID: 4, Name: "饮品甜点",
		Goods: []GoodsItem{
			{ID: 401, Name: "柠檬水", Icon: "🍋", Desc: "冰镇解渴", Price: 6},
			{ID: 402, Name: "奶茶", Icon: "🧋", Desc: "香浓丝滑", Price: 12},
			{ID: 403, Name: "双皮奶", Icon: "🍮", Desc: "顺德风味", Price: 15},
			{ID: 404, Name: "红豆冰沙", Icon: "🍧", Desc: "冰爽甜蜜", Price: 14},
		},
	},
}

// GetMenu 获取菜单
func GetMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":       200,
		"message":    "success",
		"categories": menuData,
	})
}

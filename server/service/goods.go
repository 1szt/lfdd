package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"server/core/db"
	"server/model"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// ------------------------ GetGoodsSidebar 商品侧边菜单 ------------------------

// SidebarMenu 侧边菜单项
type SidebarMenu struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}

type GetGoodsSidebarInput struct {
}
type GetGoodsSidebarOutput struct {
	Menus []SidebarMenu `json:"menus"`
}

// GetGoodsSidebar 商品页侧边菜单：全部菜单按 sort 升序（数值越小越靠前）
func GetGoodsSidebar(i GetGoodsSidebarInput) (o GetGoodsSidebarOutput, err error) {
	var menus []model.Menu
	if err = db.DB.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return o, err
	}

	o.Menus = make([]SidebarMenu, 0, len(menus))
	for _, m := range menus {
		o.Menus = append(o.Menus, SidebarMenu{ID: m.ID, Name: m.Name})
	}
	return o, nil
}

// ------------------------ GetGoodsPage 商品页面 ------------------------

// GoodsView 商品视图（价格单位：分，1850 = 18.50 元）
type GoodsView struct {
	ID     int64        `json:"id,string"`
	MenuID int64        `json:"menu_id,string"`
	Name   string       `json:"name"`
	Alias  string       `json:"alias"` // 别名（多个用逗号分隔），AI 匹配时与名称一起识别
	Price  int          `json:"price"`
	Cost   int          `json:"cost,omitempty"` // 进价（分），仅管理员请求时返回
	Spec   string       `json:"spec"`
	Stock  int          `json:"stock"`
	Image  string       `json:"image"` // 第一张图片哈希（首页展示用）
	Media  []GoodsMedia `json:"media"` // 图片/视频列表
	Sort   int          `json:"sort"`
	Status int8         `json:"status"`
}

// GoodsMedia 商品媒体：本站上传存内容哈希，第三方资源只存直链（不下载）
type GoodsMedia struct {
	Hash string `json:"hash,omitempty"` // 上传文件的内容哈希，访问路径 /api/upload/{hash}
	URL  string `json:"url,omitempty"`  // 第三方图片/视频直链，前端直接用该地址
	Type string `json:"type"`           // image / video
}

type GetGoodsPageInput struct {
	MenuID   int64 `json:"menu_id,string"` // 菜单（分类）ID，0 = 全部
	Page     int   `json:"page"`           // 页码，从 1 开始
	PageSize int   `json:"page_size"`      // 每页条数
	All      bool  `json:"all"`            // true = 不过滤上下架（管理端用）
	WithCost bool  `json:"-"`              // true = 返回进价（仅管理员）
}

type GetGoodsPageOutput struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     []GoodsView `json:"list"`
}

// GetGoodsPage 商品页面：指定菜单下上架商品（status=1）的分页列表
func GetGoodsPage(i GetGoodsPageInput) (o GetGoodsPageOutput, err error) {
	// 分页参数：默认 1 页 20 条，单页上限 100
	if i.Page < 1 {
		i.Page = 1
	}
	if i.PageSize < 1 {
		i.PageSize = 20
	}
	if i.PageSize > 100 {
		i.PageSize = 100
	}

	// 1. 统计总数（默认只算上架商品，管理端 all=1 时含下架）
	countQ := db.DB.Model(&model.Goods{})
	if !i.All {
		countQ = countQ.Where("status = ?", 1)
	}
	if i.MenuID > 0 {
		countQ = countQ.Where("menu_id = ?", i.MenuID)
	}
	if err = countQ.Count(&o.Total).Error; err != nil {
		return o, err
	}

	// 2. 分页查询（新的查询链，避免复用条件串扰）
	listQ := db.DB.Model(&model.Goods{})
	if !i.All {
		listQ = listQ.Where("status = ?", 1)
	}
	if i.MenuID > 0 {
		listQ = listQ.Where("menu_id = ?", i.MenuID)
	}
	var goods []model.Goods
	if err = listQ.Order("sort ASC, id DESC").Limit(i.PageSize).Offset((i.Page - 1) * i.PageSize).
		Find(&goods).Error; err != nil {
		return o, err
	}

	o.List = make([]GoodsView, 0, len(goods))
	for _, g := range goods {
		o.List = append(o.List, toGoodsView(g, i.WithCost))
	}
	o.Page = i.Page
	o.PageSize = i.PageSize
	return o, nil
}

// ------------------------ GetGoodsAll 全部上架商品 ------------------------
type GetGoodsAllInput struct {
	WithCost bool `json:"-"` // true = 返回进价（仅管理员）
}
type GetGoodsAllOutput struct {
	List []GoodsView `json:"list"`
}

// GetGoodsAll 全部上架商品：供前端一次性拉取后本地模糊搜索（数据量小，避免逐页拉取）
func GetGoodsAll(i GetGoodsAllInput) (o GetGoodsAllOutput, err error) {
	var goods []model.Goods
	if err = db.DB.Where("status = ?", 1).Order("sort ASC, id DESC").Find(&goods).Error; err != nil {
		return o, err
	}
	o.List = make([]GoodsView, 0, len(goods))
	for _, g := range goods {
		o.List = append(o.List, toGoodsView(g, i.WithCost))
	}
	return o, nil
}

// toGoodsView 模型转视图，集中一处避免各接口字段遗漏；withCost=false 时不返回进价
func toGoodsView(g model.Goods, withCost bool) GoodsView {
	media := mediaOf(g)
	image := g.Image
	if image == "" {
		image = firstImage(media)
	}
	v := GoodsView{
		ID: g.ID, MenuID: g.MenuID, Name: g.Name, Alias: g.Alias, Price: g.Price,
		Spec: g.Spec, Stock: g.Stock, Image: image, Media: media, Sort: g.Sort, Status: g.Status,
	}
	if withCost {
		v.Cost = g.Cost
	}
	return v
}

// normalizeMedia 归一化媒体列表：哈希与第三方直链二选一，类型只认 image/video（其余按图片处理）
func normalizeMedia(list []GoodsMedia) []GoodsMedia {
	out := make([]GoodsMedia, 0, len(list))
	for _, m := range list {
		m.Hash = strings.TrimSpace(m.Hash)
		m.URL = strings.TrimSpace(m.URL)
		if m.URL != "" {
			// 第三方直链：只校验协议，不下载；同一项不同时保留哈希
			if !isHTTPURL(m.URL) {
				continue
			}
			m.Hash = ""
		} else if m.Hash == "" {
			continue
		}
		if m.Type != "video" {
			m.Type = "image"
		}
		out = append(out, m)
	}
	return out
}

// isHTTPURL 校验第三方媒体直链（仅允许 http/https）
func isHTTPURL(raw string) bool {
	if len(raw) > 1000 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// mediaOf 视图媒体列表：优先 media 字段，旧数据（只有 image）回退为单张图片
func mediaOf(g model.Goods) []GoodsMedia {
	if g.Media != "" {
		var list []GoodsMedia
		if err := json.Unmarshal([]byte(g.Media), &list); err == nil {
			return normalizeMedia(list)
		}
	}
	if g.Image != "" {
		return []GoodsMedia{imageMedia(g.Image)}
	}
	return []GoodsMedia{}
}

// imageMedia 单张图片媒体：兼容旧数据（存哈希）与第三方直链（存完整 URL）
func imageMedia(v string) GoodsMedia {
	if isHTTPURL(v) {
		return GoodsMedia{URL: v, Type: "image"}
	}
	return GoodsMedia{Hash: v, Type: "image"}
}

// mergeMedia 入参媒体归一化：兼容旧客户端只传 image 的情况
func mergeMedia(list []GoodsMedia, image string) []GoodsMedia {
	out := normalizeMedia(list)
	if len(out) == 0 && image != "" {
		out = []GoodsMedia{imageMedia(image)}
	}
	return out
}

// mediaJSON 媒体列表序列化（写入 goods.media）
func mediaJSON(list []GoodsMedia) string {
	b, err := json.Marshal(list)
	if err != nil {
		return ""
	}
	return string(b)
}

// firstImage 第一张图片：首页展示用（本站哈希或第三方直链）
func firstImage(list []GoodsMedia) string {
	for _, m := range list {
		if m.Type == "image" {
			if m.Hash != "" {
				return m.Hash
			}
			return m.URL
		}
	}
	return ""
}

// ------------------------ CreateGoods 新建商品 ------------------------
type CreateGoodsInput struct {
	MenuID int64        `json:"menu_id,string"`
	Name   string       `json:"name"`
	Alias  string       `json:"alias"` // 别名（多个用逗号分隔），可空
	Price  int          `json:"price"` // 单位：分
	Cost   int          `json:"cost"`  // 进价，单位：分
	Spec   string       `json:"spec"`
	Stock  int          `json:"stock"`  // -1 表示不限库存
	Image  string       `json:"image"`  // 旧字段：单图哈希（无 media 时生效）
	Media  []GoodsMedia `json:"media"`  // 图片/视频列表
	Status int8         `json:"status"` // 1 上架、0 下架
	Sort   int          `json:"sort"`
}
type CreateGoodsOutput struct {
	ID int64 `json:"id,string"`
}

// CreateGoods 新建商品（必须挂在一个已存在的菜单下）
func CreateGoods(i CreateGoodsInput) (o CreateGoodsOutput, err error) {
	if i.Name == "" {
		return o, errors.New("商品名称不能为空")
	}
	if i.Price < 0 {
		return o, errors.New("商品价格不能为负数")
	}
	if i.Cost < 0 {
		return o, errors.New("商品进价不能为负数")
	}
	if i.Status != 0 && i.Status != 1 {
		return o, errors.New("上架状态只能是 0 或 1")
	}
	if _, err = checkMenuExists(CheckMenuExistsInput{MenuID: i.MenuID}); err != nil {
		return o, err
	}

	g := model.Goods{
		MenuID: i.MenuID, Name: i.Name, Alias: normalizeAlias(i.Alias), Price: i.Price, Cost: i.Cost,
		Spec: i.Spec, Stock: i.Stock, Image: firstImage(mergeMedia(i.Media, i.Image)),
		Media: mediaJSON(mergeMedia(i.Media, i.Image)), Status: i.Status, Sort: i.Sort,
	}
	if err = db.DB.Create(&g).Error; err != nil {
		return o, err
	}

	o.ID = g.ID
	return o, nil
}

// ------------------------ UpdateGoods 修改商品 ------------------------
type UpdateGoodsInput struct {
	ID     int64        `json:"id,string"`
	MenuID int64        `json:"menu_id,string"` // 可改所属菜单（移动分类）
	Name   string       `json:"name"`
	Alias  string       `json:"alias"` // 别名（多个用逗号分隔），可空
	Price  int          `json:"price"` // 单位：分
	Cost   int          `json:"cost"`  // 进价，单位：分
	Spec   string       `json:"spec"`
	Stock  int          `json:"stock"`
	Image  string       `json:"image"`  // 旧字段：单图哈希（无 media 时生效）
	Media  []GoodsMedia `json:"media"`  // 图片/视频列表
	Status int8         `json:"status"` // 1 上架、0 下架
	Sort   int          `json:"sort"`
}
type UpdateGoodsOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

// UpdateGoods 修改商品（价格、顺序、库存、上下架等全量更新，由前端回传完整对象）
func UpdateGoods(i UpdateGoodsInput) (o UpdateGoodsOutput, err error) {
	if i.Name == "" {
		return o, errors.New("商品名称不能为空")
	}
	if i.Price < 0 {
		return o, errors.New("商品价格不能为负数")
	}
	if i.Cost < 0 {
		return o, errors.New("商品进价不能为负数")
	}
	if i.Status != 0 && i.Status != 1 {
		return o, errors.New("上架状态只能是 0 或 1")
	}
	if _, err = checkMenuExists(CheckMenuExistsInput{MenuID: i.MenuID}); err != nil {
		return o, err
	}

	media := mergeMedia(i.Media, i.Image)
	res := db.DB.Model(&model.Goods{}).Where("id = ?", i.ID).Updates(map[string]interface{}{
		"menu_id": i.MenuID,
		"name":    i.Name,
		"alias":   normalizeAlias(i.Alias),
		"price":   i.Price,
		"cost":    i.Cost,
		"spec":    i.Spec,
		"stock":   i.Stock,
		"image":   firstImage(media),
		"media":   mediaJSON(media),
		"status":  i.Status,
		"sort":    i.Sort,
	})
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("商品不存在")
	}
	return o, nil
}

// ------------------------ DeleteGoods 删除商品 ------------------------
type DeleteGoodsInput struct {
	ID int64 `json:"id,string"`
}
type DeleteGoodsOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

// DeleteGoods 删除商品
func DeleteGoods(i DeleteGoodsInput) (o DeleteGoodsOutput, err error) {
	res := db.DB.Where("id = ?", i.ID).Delete(&model.Goods{})
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("商品不存在")
	}
	return o, nil
}

// ------------------------ SortGoods 商品顺序整组重写 ------------------------
type SortGoodsInput struct {
	IDs []string `json:"ids"` // 拖拽后的最终顺序（商品 id 字符串数组，与列表返回的 id 格式一致）
}
type SortGoodsOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

// SortGoods 商品顺序整组重写：按数组下标写 sort（0,1,2...），事务保证整组一致
func SortGoods(i SortGoodsInput) (o SortGoodsOutput, err error) {
	if len(i.IDs) == 0 {
		return o, errors.New("ids 不能为空")
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for idx, s := range i.IDs {
			id, e := strconv.ParseInt(s, 10, 64)
			if e != nil {
				return fmt.Errorf("无效的商品 id: %s", s)
			}
			if e = tx.Model(&model.Goods{}).Where("id = ?", id).Update("sort", idx).Error; e != nil {
				return e
			}
		}
		return nil
	})
	return o, err
}

type CheckMenuExistsInput struct {
	MenuID int64 `json:"menu_id,string"` // 待校验的菜单 ID
}
type CheckMenuExistsOutput struct{}

// checkMenuExists 校验菜单存在（商品必须挂在有效菜单下）
func checkMenuExists(i CheckMenuExistsInput) (o CheckMenuExistsOutput, err error) {
	var count int64
	if err = db.DB.Model(&model.Menu{}).Where("id = ?", i.MenuID).Count(&count).Error; err != nil {
		return o, err
	}
	if count == 0 {
		return o, errors.New("所属菜单不存在")
	}
	return o, nil
}

// normalizeAlias 归一化别名：中文逗号转英文逗号，按逗号分段去空白、去重后重新拼接
func normalizeAlias(s string) string {
	s = strings.ReplaceAll(s, "，", ",")
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return strings.Join(out, ",")
}

package service

import (
	"errors"
	"fmt"
	"server/core/db"
	"server/model"
	"strconv"

	"gorm.io/gorm"
)

// ------------------------ CreateMenu 新建菜单 ------------------------
type CreateMenuInput struct {
	Name string `json:"name"`
	Sort int    `json:"sort"`
}
type CreateMenuOutput struct {
	ID int64 `json:"id,string"`
}

// CreateMenu 新建菜单（sort 数值越小越靠前，不传默认 0）
func CreateMenu(i CreateMenuInput) (o CreateMenuOutput, err error) {
	if i.Name == "" {
		return o, errors.New("菜单名称不能为空")
	}

	m := model.Menu{Name: i.Name, Sort: i.Sort}
	if err = db.DB.Create(&m).Error; err != nil {
		return o, err
	}

	o.ID = m.ID
	return o, nil
}

// ------------------------ UpdateMenu 修改菜单 ------------------------
type UpdateMenuInput struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}
type UpdateMenuOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

// UpdateMenu 修改菜单名称（顺序走 SortMenus 整组重写）
func UpdateMenu(i UpdateMenuInput) (o UpdateMenuOutput, err error) {
	if i.Name == "" {
		return o, errors.New("菜单名称不能为空")
	}

	res := db.DB.Model(&model.Menu{}).Where("id = ?", i.ID).Update("name", i.Name)
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("菜单不存在")
	}
	return o, nil
}

// ------------------------ DeleteMenu 删除菜单 ------------------------
type DeleteMenuInput struct {
	ID int64 `json:"id,string"`
}
type DeleteMenuOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

// DeleteMenu 删除菜单：菜单下还有商品时拒绝，避免商品失去归属
func DeleteMenu(i DeleteMenuInput) (o DeleteMenuOutput, err error) {
	// 1. 检查菜单下是否还有商品
	var count int64
	if err = db.DB.Model(&model.Goods{}).Where("menu_id = ?", i.ID).Count(&count).Error; err != nil {
		return o, err
	}
	if count > 0 {
		return o, errors.New("菜单下还有商品，请先删除或转移商品")
	}

	// 2. 删除菜单
	res := db.DB.Where("id = ?", i.ID).Delete(&model.Menu{})
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("菜单不存在")
	}
	return o, nil
}

// ------------------------ SortMenus 菜单顺序整组重写 ------------------------
type SortMenusInput struct {
	IDs []string `json:"ids"` // 拖拽后的最终顺序（菜单 id 字符串数组，与列表返回的 id 格式一致）
}
type SortMenusOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

// SortMenus 菜单顺序整组重写：按数组下标写 sort（0,1,2...），事务保证整组一致
func SortMenus(i SortMenusInput) (o SortMenusOutput, err error) {
	if len(i.IDs) == 0 {
		return o, errors.New("ids 不能为空")
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for idx, s := range i.IDs {
			id, e := strconv.ParseInt(s, 10, 64)
			if e != nil {
				return fmt.Errorf("无效的菜单 id: %s", s)
			}
			if e = tx.Model(&model.Menu{}).Where("id = ?", id).Update("sort", idx).Error; e != nil {
				return e
			}
		}
		return nil
	})
	return o, err
}

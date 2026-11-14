package service

import (
	"errors"
	"strings"
	"time"

	"server/core/db"
	"server/model"
)

// ------------------------ 用户管理（管理页） ------------------------

type GetMeInput struct {
	UserID int64 `json:"user_id,string"`
}
type GetMeOutput struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     int8   `json:"role"` // 1 普通 2 管理员
	Status   int8   `json:"status"`
}

// GetMe 当前登录用户信息（前端据此决定是否展示管理页签）
func GetMe(i GetMeInput) (o GetMeOutput, err error) {
	var u model.User
	if err = db.DB.First(&u, "id = ?", i.UserID).Error; err != nil {
		return o, err
	}
	o = GetMeOutput{
		ID: u.ID, Username: u.Username, Name: u.Name,
		Role: u.Role, Status: u.Status,
	}
	return o, nil
}

type GetUsersInput struct {
	Keyword  string `json:"keyword"` // 按 用户名/邮箱/姓名 模糊搜索
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
type UserItem struct {
	ID          int64      `json:"id,string"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Status      int8       `json:"status"` // 1 正常 0 禁用
	Role        int8       `json:"role"`   // 1 普通 2 管理员
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
type GetUsersOutput struct {
	List  []UserItem `json:"list"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
}

// GetUsers 用户分页列表（keyword 模糊搜用户名/邮箱/姓名）
func GetUsers(i GetUsersInput) (o GetUsersOutput, err error) {
	o.List = make([]UserItem, 0)
	if i.Page <= 0 {
		i.Page = 1
	}
	if i.PageSize <= 0 || i.PageSize > 100 {
		i.PageSize = 20
	}

	q := db.DB.Model(&model.User{})
	if kw := strings.TrimSpace(i.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("username LIKE ? OR email LIKE ? OR name LIKE ?", like, like, like)
	}
	if err = q.Count(&o.Total).Error; err != nil {
		return o, err
	}

	var users []model.User
	if err = q.Order("created_at DESC").
		Offset((i.Page - 1) * i.PageSize).Limit(i.PageSize).
		Find(&users).Error; err != nil {
		return o, err
	}

	for _, u := range users {
		o.List = append(o.List, UserItem{
			ID: u.ID, Username: u.Username, Email: u.Email, Name: u.Name,
			Status: u.Status, Role: u.Role,
			LastLoginAt: u.LastLoginAt, CreatedAt: u.CreatedAt,
		})
	}
	o.Page = i.Page
	return o, nil
}

type UpdateUserInput struct {
	ID     int64 `json:"id,string"`
	Role   int8  `json:"role"`   // 1 普通 2 管理员（与 status 一起全量提交）
	Status int8  `json:"status"` // 1 正常 0 禁用
}
type UpdateUserOutput struct {
}

// UpdateUser 修改用户角色/状态（前端回传完整对象，两字段都更新）
func UpdateUser(i UpdateUserInput) (o UpdateUserOutput, err error) {
	if i.ID == 0 {
		return o, errors.New("缺少用户 ID")
	}
	if i.Role != 1 && i.Role != 2 {
		return o, errors.New("角色取值非法")
	}
	if i.Status != 0 && i.Status != 1 {
		return o, errors.New("状态取值非法")
	}

	res := db.DB.Model(&model.User{}).Where("id = ?", i.ID).
		Updates(map[string]any{"role": i.Role, "status": i.Status})
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("用户不存在")
	}
	return o, nil
}

type DeleteUserInput struct {
	ID int64 `json:"id,string"`
	// 操作人 ID（禁止删除自己，handler 从上下文取）
	OperatorID int64 `json:"operator_id,string"`
}
type DeleteUserOutput struct {
}

// DeleteUser 删除用户（不能删自己；同时清掉其会话）
func DeleteUser(i DeleteUserInput) (o DeleteUserOutput, err error) {
	if i.ID == 0 {
		return o, errors.New("缺少用户 ID")
	}
	if i.ID == i.OperatorID {
		return o, errors.New("不能删除自己的账号")
	}

	res := db.DB.Where("id = ?", i.ID).Delete(&model.User{})
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("用户不存在")
	}
	// 顺带清掉该用户的登录会话（token 立即失效）
	db.DB.Where("user_id = ?", i.ID).Delete(&model.Session{})
	return o, nil
}

type ResetUserPasswordInput struct {
	ID       int64  `json:"id,string"`
	Password string `json:"password"`
	// 当前操作人 token：重置的是自己时保留该会话，避免改完密码立刻被踢下线
	KeepToken string `json:"-"`
}
type ResetUserPasswordOutput struct {
}

// ResetUserPassword 管理员重置用户密码（≥8 位，argon2id 哈希）
func ResetUserPassword(i ResetUserPasswordInput) (o ResetUserPasswordOutput, err error) {
	if len(i.Password) < 8 {
		return o, errors.New("密码长度至少 8 位")
	}
	h, err := HashPassword(HashPasswordInput{Password: i.Password})
	if err != nil {
		return o, err
	}
	res := db.DB.Model(&model.User{}).Where("id = ?", i.ID).
		Update("password", h.Hash)
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("用户不存在")
	}
	// 改密后旧会话作废（保留当前操作人的会话）
	q := db.DB.Where("user_id = ?", i.ID)
	if i.KeepToken != "" {
		q = q.Where("token <> ?", i.KeepToken)
	}
	q.Delete(&model.Session{})
	return o, nil
}

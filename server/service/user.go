package service

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"server/core/db"
	"server/model"

	"gorm.io/gorm"
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

// ------------------------ 添加 / 编辑用户 ------------------------

type CreateUserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     int8   `json:"role"` // 1 普通 2 管理员
}
type CreateUserOutput struct {
	ID int64 `json:"id,string"`
}

// CreateUser 管理员创建用户（用户名/邮箱唯一，argon2id 哈希）
func CreateUser(i CreateUserInput) (o CreateUserOutput, err error) {
	username := strings.TrimSpace(i.Username)
	email := strings.ToLower(strings.TrimSpace(i.Email))
	if username == "" {
		return o, errors.New("用户名不能为空")
	}
	if utf8.RuneCountInString(username) > 32 {
		return o, errors.New("用户名不能超过 32 个字")
	}
	if email == "" {
		return o, errors.New("邮箱不能为空")
	}
	if len(i.Password) < 8 {
		return o, errors.New("密码长度至少 8 位")
	}
	if i.Role != 1 && i.Role != 2 {
		i.Role = 1
	}

	// 唯一性校验
	var count int64
	if err = db.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return o, err
	}
	if count > 0 {
		return o, errors.New("用户名已存在")
	}
	if err = db.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return o, err
	}
	if count > 0 {
		return o, errors.New("邮箱已存在")
	}

	h, err := HashPassword(HashPasswordInput{Password: i.Password})
	if err != nil {
		return o, err
	}
	id, err := GenUserID(GenUserIDInput{})
	if err != nil {
		return o, err
	}
	user := &model.User{
		ID:       id.ID,
		Username: username,
		Email:    email,
		Password: h.Hash,
		Role:     i.Role,
	}
	if err = db.DB.Create(user).Error; err != nil {
		return o, err
	}
	o.ID = user.ID
	return o, nil
}

type UpdateUserInfoInput struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`        // 修改后的用户名
	Password string `json:"password"`        // 新密码，留空则不修改
	// 当前操作人 token：改的是自己时保留该会话
	KeepToken string `json:"-"`
}
type UpdateUserInfoOutput struct {
}

// UpdateUserInfo 修改用户名/密码（改密码后旧会话作废，保留当前操作人会话）
func UpdateUserInfo(i UpdateUserInfoInput) (o UpdateUserInfoOutput, err error) {
	if i.ID == 0 {
		return o, errors.New("缺少用户 ID")
	}
	// 先确认用户存在（避免 Updates 值未变化时 RowsAffected=0 误判）
	var u model.User
	if err = db.DB.First(&u, "id = ?", i.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, errors.New("用户不存在")
		}
		return o, err
	}

	updates := map[string]any{}

	if username := strings.TrimSpace(i.Username); username != "" {
		if utf8.RuneCountInString(username) > 32 {
			return o, errors.New("用户名不能超过 32 个字")
		}
		var count int64
		if err = db.DB.Model(&model.User{}).
			Where("username = ? AND id <> ?", username, i.ID).Count(&count).Error; err != nil {
			return o, err
		}
		if count > 0 {
			return o, errors.New("用户名已存在")
		}
		updates["username"] = username
	}

	if i.Password != "" {
		if len(i.Password) < 8 {
			return o, errors.New("密码长度至少 8 位")
		}
		h, e := HashPassword(HashPasswordInput{Password: i.Password})
		if e != nil {
			return o, e
		}
		updates["password"] = h.Hash
	}

	if len(updates) == 0 {
		return o, errors.New("没有要修改的内容")
	}
	if err = db.DB.Model(&model.User{}).Where("id = ?", i.ID).Updates(updates).Error; err != nil {
		return o, err
	}

	// 改了密码：旧会话作废（保留当前操作人的会话）
	if _, changed := updates["password"]; changed {
		q := db.DB.Where("user_id = ?", i.ID)
		if i.KeepToken != "" {
			q = q.Where("token <> ?", i.KeepToken)
		}
		q.Delete(&model.Session{})
	}
	return o, nil
}

// ------------------------ 设备（登录会话）管理 ------------------------

type GetUserSessionsInput struct {
	UserID int64 `json:"user_id,string"`
	// 当前操作人 token：标记哪条是自己的会话
	CurrentToken string `json:"-"`
}
type UserSessionItem struct {
	Token     string    `json:"token"`
	IP        string    `json:"ip"`
	Device    string    `json:"device"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Expired   bool      `json:"expired"`
	Current   bool      `json:"current"` // 是否当前操作人的会话
}
type GetUserSessionsOutput struct {
	List []UserSessionItem `json:"list"`
}

// GetUserSessions 某用户的全部登录设备/会话（含已过期，管理页设备管理用）
func GetUserSessions(i GetUserSessionsInput) (o GetUserSessionsOutput, err error) {
	o.List = make([]UserSessionItem, 0)
	if i.UserID == 0 {
		return o, errors.New("缺少用户 ID")
	}
	var ss []model.Session
	if err = db.DB.Where("user_id = ?", i.UserID).
		Order("created_at DESC").Find(&ss).Error; err != nil {
		return o, err
	}
	now := time.Now()
	for _, s := range ss {
		o.List = append(o.List, UserSessionItem{
			Token:     s.Token,
			IP:        s.IP,
			Device:    s.Device,
			ExpiresAt: s.ExpiresAt,
			CreatedAt: s.CreatedAt,
			Expired:   !s.ExpiresAt.After(now),
			Current:   i.CurrentToken != "" && s.Token == i.CurrentToken,
		})
	}
	return o, nil
}

type KickSessionInput struct {
	UserID int64  `json:"user_id,string"`
	Token  string `json:"token"`
}
type KickSessionOutput struct {
}

// KickSession 踢指定设备下线（删除其会话，token 立即失效）
func KickSession(i KickSessionInput) (o KickSessionOutput, err error) {
	if i.UserID == 0 || i.Token == "" {
		return o, errors.New("参数不完整")
	}
	res := db.DB.Where("user_id = ? AND token = ?", i.UserID, i.Token).
		Delete(&model.Session{})
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("会话不存在或已失效")
	}
	return o, nil
}

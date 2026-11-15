package service

import (
	"crypto/rand"
	"encoding/hex"
	"server/core/db"
	"server/model"
)

// ------------------------ CreateGuest 申请游客身份 ------------------------

type CreateGuestInput struct {
	Device string `json:"device"` // 设备信息（UA 等），可空
}
type CreateGuestOutput struct {
	GuestID string `json:"guest_id"`
}

// CreateGuest 申请游客身份：生成随机 token 并入库，之后通过 Authorization: Bearer 携带即可访问业务接口
func CreateGuest(i CreateGuestInput) (o CreateGuestOutput, err error) {
	id, err := newGuestID()
	if err != nil {
		return o, err
	}
	if err = db.DB.Create(&model.Guest{ID: id, Device: i.Device}).Error; err != nil {
		return o, err
	}
	o.GuestID = id
	return o, nil
}

// newGuestID 生成游客 token：guest_ 前缀 + 29 字节随机数（16 进制 58 字符），总长 64
// 前缀用于与登录 token 共用 Authorization 头时区分身份类型
func newGuestID() (string, error) {
	bytes := make([]byte, 29)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return model.GuestTokenPrefix + hex.EncodeToString(bytes), nil
}

// ------------------------ GetGuest 校验游客身份 ------------------------

type GetGuestInput struct {
	GuestID string `json:"guest_id"`
}
type GetGuestOutput struct {
	GuestID string `json:"guest_id"`
}

// GetGuest 按身份查游客记录：查不到（未申请或已注销）返回 gorm.ErrRecordNotFound
func GetGuest(i GetGuestInput) (o GetGuestOutput, err error) {
	var g model.Guest
	if err = db.DB.Where("id = ?", i.GuestID).First(&g).Error; err != nil {
		return o, err
	}
	o.GuestID = g.ID
	return o, nil
}

// ------------------------ DeleteGuest 注销游客身份 ------------------------

type DeleteGuestInput struct {
	GuestID string `json:"guest_id"`
}
type DeleteGuestOutput struct{}

// DeleteGuest 注销游客身份：删除后该身份立即失效，需重新申请
func DeleteGuest(i DeleteGuestInput) (o DeleteGuestOutput, err error) {
	if err = db.DB.Where("id = ?", i.GuestID).Delete(&model.Guest{}).Error; err != nil {
		return o, err
	}
	return o, nil
}

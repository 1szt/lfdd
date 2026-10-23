// Package db 数据库模块：持有全局 DB 实例，负责数据库连接的建立/关闭/重连（ORM 底层为 gorm）。
package db

import (
	"fmt"
	"hxzl/config"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var lastConfig string // 上次连接时的配置指纹

type DBStatus struct {
	Type   string `json:"type"`
	Host   string `json:"host"`
	Port   string `json:"port"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func Start() {
	if config.Get("server_install").Value != "true" {
		log.Println("⚠️ [DB] 服务器未初始化，跳过数据库连接，等待初始化...")
		return
	}
	connect()
}

func Stop() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("❌ [DB] 获取底层连接失败: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("❌ [DB] 关闭数据库连接失败: %v", err)
		return
	}
	DB = nil
	log.Println("✅ [DB] 数据库连接已关闭")
}

func Restart() {
	Stop()
	connect()
}

func Reload() {
	// 构造当前配置指纹
	currentConfig := configFingerprint()

	// 检查数据库是否已连接
	connected := false
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			if err = sqlDB.Ping(); err == nil {
				connected = true
			}
		}
	}

	// 配置没变且已连接 → 跳过
	if connected && currentConfig == lastConfig {
		log.Println("🔄 [DB] 配置无变化且已连接，跳过重载")
		return
	}

	// 有变化或未连接 → 重连
	if !connected {
		log.Println("🔄 [DB] 数据库未连接，重新连接...")
	} else {
		log.Println("🔄 [DB] 配置已变化，重新连接...")
	}
	Stop()
	connect()
}

// configFingerprint 根据数据库相关配置生成指纹，用于判断配置是否变化
func configFingerprint() string {
	dbType := config.Get("db_type").Value
	return dbType + "|" + config.Get("db_host").Value + "|" + config.Get("db_port").Value + "|" + config.Get("db_name").Value + "|" + config.Get("db_path").Value
}

func Status() DBStatus {
	dbType := config.Get("db_type").Value
	status := "disconnected"
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			if err = sqlDB.Ping(); err == nil {
				status = "connected"
			}
		}
	}
	return DBStatus{
		Type:   dbType,
		Host:   config.Get("db_host").Value,
		Port:   config.Get("db_port").Value,
		Name:   config.Get("db_name").Value,
		Status: status,
	}
}

func connect() {
	dbType := config.Get("db_type").Value
	if dbType == "" {
		log.Fatal("❌ [DB] 未找到 db_type 配置")
	}

	var dialector gorm.Dialector

	switch dbType {
	case "sqlite", "sqlite3":
		path := config.Get("db_path").Value
		dialector = sqlite.Open(path)

	case "pgsql", "postgres", "postgresql":
		host := config.Get("db_host").Value
		port := config.Get("db_port").Value
		user := config.Get("db_user").Value
		password := config.Get("db_password").Value
		name := config.Get("db_name").Value
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
		dialector = postgres.Open(dsn)

	case "mysql", "mariadb":
		host := config.Get("db_host").Value
		port := config.Get("db_port").Value
		user := config.Get("db_user").Value
		password := config.Get("db_password").Value
		name := config.Get("db_name").Value
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)
		dialector = mysql.Open(dsn)

	default:
		log.Fatalf("⚠️ [DB] 数据库模块 配置错误！无法识别的数据库类型: %s", dbType)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		log.Printf("❌ [DB] 初始化数据库 %s 失败: %v", dbType, err)
		return
	}

	log.Printf("✅ [DB] 已成功初始化数据库 %s", dbType)
	lastConfig = configFingerprint()
}

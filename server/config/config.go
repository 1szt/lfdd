package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// cache 配置内存缓存，避免每次 Get 都查库；Set 写库成功后同步更新
var (
	cacheMu sync.RWMutex
	cache   = map[string]string{}
)

type Config struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	User  string `json:"modifyUser"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func Start() {

	// 确保 data 目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		panic("failed to create data directory: " + err.Error())
	}

	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("data/config.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect config database: " + err.Error())
	}

	// 自动迁移数据库表
	if err = db.AutoMigrate(&Config{}); err != nil {
		panic("failed to migrate config table: " + err.Error())
	}

	// 首次启动为缺失的配置项写入默认值（只补缺不覆盖）：
	// 之后所有 Get/Set 都能查到行，不会出现 record not found 刷屏
	seedDefaults()

	// 从环境变量中读取配置
	envPrefix := "LFDD_"
	configFields := []string{
		"server_ip", "server_port", "server_https", "server_https_auto",
		"server_cert", "server_key", "server_install",
		"db_type", "db_path", "db_host", "db_port",
		"db_name", "db_user", "db_password",
	}

	for _, field := range configFields {
		envKey := envPrefix + strings.ToUpper(field)
		if val, ok := os.LookupEnv(envKey); ok && len(val) > 0 {
			if err := Set(field, val, "env"); err != nil {
				log.Printf("⚠️ [Config] 环境变量覆盖配置 %s 失败: %v", field, err)
			} else {
				log.Printf("✅ [Config] 环境变量覆盖配置 %s = %s", field, val)
			}
		}
	}

}

// TableName 设置表名
func (Config) TableName() string {
	return "config"
}

// defaultValues 各配置项的首启默认值（与代码里的空值兜底行为保持一致）
var defaultValues = map[string]string{
	"server_ip":         "",     // 空 = 监听所有网卡
	"server_port":       "9081", // 与 server 模块 addrFromConfig 的缺省一致
	"server_https":      "false",
	"server_https_auto": "false",
	"server_cert":       "", // 安装时填写
	"server_key":        "",
	"server_install":    "false", // false/空 都算未初始化
	"db_type":           "",      // 初始化时由安装页填写
	"db_path":           "",
	"db_host":           "",
	"db_port":           "",
	"db_name":           "",
	"db_user":           "",
	"db_password":       "",
}

// seedDefaults 为缺失的配置项补写默认值（一次 Find 拿全部已有项，只补缺不覆盖）
func seedDefaults() {
	var existing []Config
	if err := db.Find(&existing).Error; err != nil {
		log.Printf("⚠️ [Config] 读取现有配置失败: %v", err)
		return
	}
	exist := make(map[string]bool, len(existing))
	for _, c := range existing {
		exist[c.Name] = true
	}
	for name, value := range defaultValues {
		if exist[name] {
			continue
		}
		if err := db.Create(&Config{Name: name, Value: value, User: "system"}).Error; err != nil {
			log.Printf("⚠️ [Config] 写入默认配置 %s 失败: %v", name, err)
		}
	}
}

// Get 读取配置（优先走内存缓存），不存在时返回零值（Value 为空字符串），不视为错误
func Get(name string) Config {
	cacheMu.RLock()
	val, ok := cache[name]
	cacheMu.RUnlock()
	if ok {
		return Config{Name: name, Value: val}
	}

	var cfg Config
	if err := db.Where("name = ?", name).First(&cfg).Error; err != nil {
		// 配置不存在是正常情况（如首次安装前），只有真正的查询错误才记录日志
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("❌ [Config] 查询配置 %s 失败: %v", name, err)
		}
		return cfg
	}

	cacheMu.Lock()
	cache[name] = cfg.Value
	cacheMu.Unlock()
	return cfg
}

func Set(name, value, user string) error {
	if len(user) == 0 {
		return errors.New("用户名不能为空")
	}
	var cfg Config
	err := db.Where("name = ?", name).First(&cfg).Error
	if err != nil {
		// 不存在则创建
		cfg = Config{
			Name:  name,
			Value: value,
			User:  user,
		}
		if err := db.Create(&cfg).Error; err != nil {
			return fmt.Errorf("创建配置 %s 失败: %v", name, err)
		}
		putCache(name, value)
		return nil
	}
	// 存在则修改
	cfg.Value = value
	cfg.User = user
	if err := db.Save(&cfg).Error; err != nil {
		return fmt.Errorf("保存配置 %s 失败: %v", name, err)
	}
	putCache(name, value)
	return nil
}

// putCache 写入缓存（Set 成功后调用，保证缓存与库一致）
func putCache(name, value string) {
	cacheMu.Lock()
	cache[name] = value
	cacheMu.Unlock()
}

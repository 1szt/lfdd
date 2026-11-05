package service

import (
	crand "crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// HashPassword 哈希密码（argon2id，PHC 密码哈希竞赛冠军，当前业界首选）。
// 返回 PHC 自描述编码串（含算法、参数、盐），原样存库即可。
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := crand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		algoArgon2id, argon2.Version, argon2Memory, argon2Time, argon2Threads,
		phcB64.EncodeToString(salt), phcB64.EncodeToString(key)), nil
}

// VerifyPassword 校验密码，格式不合法一律返回 false（恒定时间比较）。
func VerifyPassword(password, encoded string) bool {
	// 格式: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != algoArgon2id {
		return false
	}
	var v, memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &v); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false
	}
	salt, err := phcB64.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := phcB64.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// 参数取 RFC 9106 推荐第二档（64MB），适合交互式登录场景
const (
	algoArgon2id = "argon2id"

	argon2Time    = 1
	argon2Memory  = 64 * 1024 // KiB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// phcB64 PHC 规范要求无填充 base64
var phcB64 = base64.RawStdEncoding

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/1szt/go-pkg/env"
)

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var videoExts = map[string]bool{
	".mp4": true, ".mov": true, ".webm": true, ".m4v": true,
}

// 图片文件名：64 位十六进制内容哈希（无后缀，类型由 GET 时嗅探内容得出）
var imageNameRe = regexp.MustCompile(`^[0-9a-f]{64}$`)

const maxImageSize = 5 << 20  // 5MB
const maxVideoSize = 50 << 20 // 50MB

type UploadDirInput struct{}
type UploadDirOutput struct {
	Dir string `json:"dir"` // 图片存储目录
}

// uploadDir 图片存储目录（LFDD_UPLOAD_PATH，空则 data/upload）
func uploadDir(_ UploadDirInput) (o UploadDirOutput, err error) {
	dir := env.Get("LFDD_UPLOAD_PATH")
	if dir == "" {
		dir = "data/upload"
	}
	o.Dir = dir
	return o, nil
}

type UploadImageInput struct {
	File *multipart.FileHeader
}
type UploadImageOutput struct {
	Hash string `json:"hash"` // 内容 SHA-256 哈希，数据库 goods.image 只存它，访问路径为 /api/upload/{hash}
}

// UploadImage 媒体上传（图片/视频）：文件按内容 SHA-256 哈希命名存储（无后缀，同内容自动去重）
func UploadImage(i UploadImageInput) (o UploadImageOutput, err error) {
	fh := i.File
	if fh == nil {
		return o, errors.New("请选择图片或视频文件")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	isVideo := videoExts[ext]
	if !imageExts[ext] && !isVideo {
		return o, errors.New("仅支持 jpg/png/gif/webp 图片或 mp4/mov/webm 视频")
	}
	limit := int64(maxImageSize)
	if isVideo {
		limit = maxVideoSize
	}
	if fh.Size > limit {
		if isVideo {
			return o, errors.New("视频不能超过 50MB")
		}
		return o, errors.New("图片不能超过 5MB")
	}

	f, err := fh.Open()
	if err != nil {
		return o, err
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return o, err
	}
	if int64(len(data)) > limit {
		if isVideo {
			return o, errors.New("视频不能超过 50MB")
		}
		return o, errors.New("图片不能超过 5MB")
	}

	sum := sha256.Sum256(data)
	o.Hash = hex.EncodeToString(sum[:])

	dirOut, e := uploadDir(UploadDirInput{})
	if e != nil {
		return o, e
	}
	if err = os.MkdirAll(dirOut.Dir, 0o755); err != nil {
		return o, err
	}

	// 同内容文件已存在则跳过写入（哈希去重）
	dst := filepath.Join(dirOut.Dir, o.Hash)
	if _, err = os.Stat(dst); os.IsNotExist(err) {
		if err = os.WriteFile(dst, data, 0o644); err != nil {
			return o, err
		}
	}

	return o, nil
}

type GetUploadImageInput struct {
	Name string // 文件名：64 位内容哈希（无后缀）
}
type GetUploadImageOutput struct {
	Path        string // 服务器本地文件路径
	ContentType string // 按文件内容嗅探出的 MIME 类型（文件无后缀）
}

// GetUploadImage 校验文件名并返回本地路径与内容类型（供 GET /api/upload/{hash} 直出图片/视频）
func GetUploadImage(i GetUploadImageInput) (o GetUploadImageOutput, err error) {
	if !imageNameRe.MatchString(i.Name) {
		return o, errors.New("非法文件名")
	}
	dirOut, e := uploadDir(UploadDirInput{})
	if e != nil {
		return o, e
	}
	path := filepath.Join(dirOut.Dir, i.Name)
	f, err := os.Open(path)
	if err != nil {
		return o, errors.New("文件不存在")
	}
	defer f.Close()

	// 文件名无后缀：读头部字节嗅探真实类型（jpg/png/gif/webp/mp4 等）
	head := make([]byte, 512)
	n, _ := f.Read(head)
	o.ContentType = http.DetectContentType(head[:n])
	o.Path = path
	return o, nil
}

package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/1szt/go-pkg/env"
	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var videoExts = map[string]bool{
	".mp4": true, ".mov": true, ".webm": true, ".m4v": true,
}

// 图片文件名：64 位十六进制内容哈希（无后缀，类型由 GET 时嗅探内容得出）
var imageNameRe = regexp.MustCompile(`^[0-9a-f]{64}$`)

const (
	maxImageSize  = 20 << 20 // 图片原图上限 20MB（压缩后保存）
	maxVideoSize  = 50 << 20 // 视频上限 50MB（原样保存，不压缩）
	maxImageBytes = 1 << 20  // 图片压缩目标：≤1MB
	maxImageDim   = 1600     // 图片最长边像素
)

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

// UploadImage 媒体上传：图片用 Go 压缩为 WebP（≤1MB），视频原样保存；
// 按最终内容的 SHA-256 哈希命名存储（无后缀，同内容自动去重）
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
		return o, errors.New("图片不能超过 20MB")
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
		return o, errors.New("图片不能超过 20MB")
	}

	// 图片：纯 Go 压缩成 WebP；视频：不压缩，原样保存
	if !isVideo {
		compressed, err := compressImage(CompressImageInput{Data: data})
		if err != nil {
			return o, err
		}
		data = compressed.Data
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

type CompressImageInput struct {
	Data []byte // 原始图片字节
}
type CompressImageOutput struct {
	Data []byte // 压缩后的 WebP 字节
}

// compressImage 纯 Go 图片压缩：等比缩到最长边 ≤1600，统一转 WebP（保留透明），
// 逐步降低质量/尺寸直到 ≤1MB，保证体积可控
func compressImage(i CompressImageInput) (o CompressImageOutput, err error) {
	data := i.Data
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return o, errors.New("无法识别的图片格式")
	}
	// 防解压炸弹：超大尺寸直接拒绝
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 12000 || cfg.Height > 12000 {
		return o, errors.New("图片尺寸过大")
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return o, errors.New("图片解析失败")
	}

	for dim := maxImageDim; ; dim = dim * 3 / 4 {
		img := resizeTo(ResizeToInput{Src: src, Dim: dim}).Img
		for _, q := range []int{80, 75, 70, 65, 60, 55, 50, 45} {
			out, err := encodeWebP(EncodeWebPInput{Img: img, Quality: q})
			if err != nil {
				return o, err
			}
			if len(out.Data) <= maxImageBytes {
				o.Data = out.Data
				return o, nil
			}
		}
		// 已经缩到很小仍超限：返回最低质量结果兜底
		if dim <= 640 {
			out, err := encodeWebP(EncodeWebPInput{Img: img, Quality: 40})
			if err != nil {
				return o, err
			}
			o.Data = out.Data
			return o, nil
		}
	}
}

type ResizeToInput struct {
	Src image.Image // 原始图片
	Dim int         // 最长边像素上限
}
type ResizeToOutput struct {
	Img image.Image // 缩放后的图片（保留透明通道）
}

// resizeTo 等比缩放至最长边不超过 dim
func resizeTo(i ResizeToInput) (o ResizeToOutput) {
	b := i.Src.Bounds()
	w, h := b.Dx(), b.Dy()
	dim := i.Dim
	if w > dim || h > dim {
		if w >= h {
			h = h * dim / w
			w = dim
		} else {
			w = w * dim / h
			h = dim
		}
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), i.Src, b, draw.Over, nil)
	o.Img = dst
	return o
}

type EncodeWebPInput struct {
	Img     image.Image // 待编码图片
	Quality int         // 质量 [0,100]
}
type EncodeWebPOutput struct {
	Data []byte // WebP 字节
}

// encodeWebP 编码为 WebP（Method 4 为质量/速度折中）
func encodeWebP(i EncodeWebPInput) (o EncodeWebPOutput, err error) {
	var buf bytes.Buffer
	if err = webp.Encode(&buf, i.Img, webp.Options{Quality: i.Quality, Method: 4}); err != nil {
		return o, err
	}
	o.Data = buf.Bytes()
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

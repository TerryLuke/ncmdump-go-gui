package convert

import (
	"fmt"
	"strings"

	"git.taurusxin.com/taurusxin/ncmdump-go/ncmcrypt"
)

// OutputFormat 为最终产出容器/编码（除 auto 外，转码依赖系统已安装 ffmpeg）。
type OutputFormat string

const (
	FormatAuto OutputFormat = "auto" // 保持解密结果：mp3 或 flac
	FormatMP3  OutputFormat = "mp3"
	FormatFLAC OutputFormat = "flac"
	FormatWAV  OutputFormat = "wav"
	FormatAAC  OutputFormat = "aac" // 实际文件扩展名为 .m4a（AAC in MP4）
)

// ParseOutputFormat 解析 CLI / 配置字符串（大小写不敏感）。
func ParseOutputFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return FormatAuto, nil
	case "mp3":
		return FormatMP3, nil
	case "flac":
		return FormatFLAC, nil
	case "wav":
		return FormatWAV, nil
	case "aac", "m4a":
		return FormatAAC, nil
	default:
		return FormatAuto, fmt.Errorf("unknown output format %q (supported: auto, mp3, flac, wav, aac)", s)
	}
}

func (f OutputFormat) fileExtension() string {
	switch f {
	case FormatMP3:
		return ".mp3"
	case FormatFLAC:
		return ".flac"
	case FormatWAV:
		return ".wav"
	case FormatAAC:
		return ".m4a"
	default:
		return ""
	}
}

func needsTranscode(want OutputFormat, native ncmcrypt.NcmFormat) bool {
	if want == FormatAuto {
		return false
	}
	switch want {
	case FormatMP3:
		return native != ncmcrypt.Mp3
	case FormatFLAC:
		return native != ncmcrypt.Flac
	case FormatWAV, FormatAAC:
		return true
	default:
		return false
	}
}

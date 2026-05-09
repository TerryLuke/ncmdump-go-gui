package convert

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// 环境变量 NCMDUMP_FFMPEG 可显式指定 ffmpeg 可执行路径（macOS .app 从访达启动时 PATH 常不含 Homebrew）。

var ffmpegResolveOnce sync.Once
var ffmpegResolvedPath string
var ffmpegResolveErr error

func ffmpegFallbackPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/opt/homebrew/bin/ffmpeg",
			"/usr/local/bin/ffmpeg",
			"/opt/homebrew/opt/ffmpeg/bin/ffmpeg",
		}
	case "windows":
		return []string{
			filepath.Join(os.Getenv("ProgramFiles"), "ffmpeg", "bin", "ffmpeg.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "ffmpeg", "bin", "ffmpeg.exe"),
		}
	default:
		out := []string{"/usr/bin/ffmpeg", "/usr/local/bin/ffmpeg", "/snap/bin/ffmpeg"}
		if home, err := os.UserHomeDir(); err == nil {
			out = append(out,
				filepath.Join(home, ".local", "bin", "ffmpeg"),
				filepath.Join(home, "bin", "ffmpeg"),
			)
		}
		return out
	}
}

func verifyFFmpeg(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("is a directory")
	}
	cmd := exec.Command(path, "-version")
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	cmd.Env = os.Environ()
	return cmd.Run()
}

// ffmpegExecutable 返回可用的 ffmpeg 绝对路径；从桌面 .app 启动时依赖常见安装目录或 NCMDUMP_FFMPEG。
func ffmpegExecutable() (string, error) {
	ffmpegResolveOnce.Do(func() {
		if p := strings.TrimSpace(os.Getenv("NCMDUMP_FFMPEG")); p != "" {
			if err := verifyFFmpeg(p); err != nil {
				ffmpegResolveErr = fmt.Errorf("环境变量 NCMDUMP_FFMPEG=%q 无效: %w", p, err)
				return
			}
			ffmpegResolvedPath = p
			return
		}
		if p, err := exec.LookPath("ffmpeg"); err == nil {
			if err := verifyFFmpeg(p); err == nil {
				ffmpegResolvedPath = p
				return
			}
		}
		for _, p := range ffmpegFallbackPaths() {
			if err := verifyFFmpeg(p); err == nil {
				ffmpegResolvedPath = p
				return
			}
		}
		ffmpegResolveErr = fmt.Errorf("未找到 ffmpeg。macOS：从应用程序打开 GUI 时常无 Homebrew PATH，已尝试 /opt/homebrew/bin 等；" +
			"仍可安装 ffmpeg 后设置 NCMDUMP_FFMPEG=/绝对路径/ffmpeg；终端 dev 模式下通常 PATH 完整故正常")
	})
	return ffmpegResolvedPath, ffmpegResolveErr
}

// transcodeTo remux/transcode the tagged dump file to another format using ffmpeg (-map_metadata 0 尽量保留标签)。
func transcodeTo(src, dst string, want OutputFormat) error {
	fbin, err := ffmpegExecutable()
	if err != nil {
		return err
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", src, "-map_metadata", "0"}
	switch want {
	case FormatMP3:
		args = append(args, "-c:a", "libmp3lame", "-q:a", "0")
	case FormatFLAC:
		args = append(args, "-c:a", "flac", "-compression_level", "8")
	case FormatWAV:
		args = append(args, "-c:a", "pcm_s16le")
	case FormatAAC:
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	default:
		return fmt.Errorf("unsupported transcode format %q", want)
	}
	args = append(args, dst)
	cmd := exec.Command(fbin, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func lookPathFFmpeg() error {
	_, err := ffmpegExecutable()
	return err
}

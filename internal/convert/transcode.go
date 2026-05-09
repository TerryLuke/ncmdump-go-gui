package convert

import (
	"fmt"
	"os/exec"
	"strings"
)

// transcodeTo remux/transcode the tagged dump file to another format using ffmpeg (-map_metadata 0 尽量保留标签).
func transcodeTo(src, dst string, want OutputFormat) error {
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
	out, err := exec.Command("ffmpeg", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func lookPathFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("未找到 ffmpeg：选择非「自动」且需转码时请安装 ffmpeg 并加入 PATH（如 brew install ffmpeg）")
	}
	return nil
}

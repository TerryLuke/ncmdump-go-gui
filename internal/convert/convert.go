package convert

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"git.taurusxin.com/taurusxin/ncmdump-go/ncmcrypt"
	"git.taurusxin.com/taurusxin/ncmdump-go/utils"
)

// ErrSkipped is returned when the path is not a .ncm file and is intentionally ignored.
var ErrSkipped = errors.New("skipped: not a .ncm file")

// ProcessFile converts one NetEase Cloud Music .ncm file to mp3/flac and fixes metadata.
// If outputDir is non-empty, the output file is written there; otherwise it stays beside the source file.
// When metadata fix fails, the returned path is still the dump path (file was created).
func ProcessFile(filePath, outputDir string) (outPath string, err error) {
	if len(filePath) < 4 || !strings.EqualFold(filePath[len(filePath)-4:], ".ncm") {
		return "", ErrSkipped
	}

	currentFile, err := ncmcrypt.NewNeteaseCloudMusic(filePath)
	if err != nil {
		return "", err
	}
	dump, err := currentFile.Dump(outputDir)
	if err != nil {
		return "", err
	}
	if !dump {
		return "", nil
	}
	metadata, err := currentFile.FixMetadata(true)
	if !metadata {
		return currentFile.GetDumpFilePath(), err
	}
	return currentFile.GetDumpFilePath(), nil
}

// CollectNCMInDir lists .ncm files under dir. If recursive is false, only the top level is scanned.
func CollectNCMInDir(dir string, recursive bool) ([]string, error) {
	if !utils.IsDir(dir) {
		return nil, errors.New("not a directory")
	}
	var out []string
	if recursive {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(path), ".ncm") {
				out = append(out, path)
			}
			return nil
		})
		return out, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range entries {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.EqualFold(filepath.Ext(name), ".ncm") {
			out = append(out, filepath.Join(dir, name))
		}
	}
	return out, nil
}

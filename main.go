package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"

	"git.taurusxin.com/taurusxin/ncmdump-go/internal/convert"
	"git.taurusxin.com/taurusxin/ncmdump-go/internal/gui"
	"git.taurusxin.com/taurusxin/ncmdump-go/internal/version"
	"git.taurusxin.com/taurusxin/ncmdump-go/utils"
	flag "github.com/spf13/pflag"
)

//go:embed Icon.png
var embeddedAppIcon []byte

func appWindowIcon() fyne.Resource {
	if len(embeddedAppIcon) == 0 {
		return nil
	}
	return fyne.NewStaticResource("Icon.png", embeddedAppIcon)
}

// macOS：从「应用程序」里双击 .app 时，多数情况下只有可执行文件路径这一条 argv，不会带 -psn_。
// 若在解析 CLI 之前无法识别为 GUI 启动，会误走「无参数打印帮助并退出」，看起来像闪退。
func runningInMacOSAppBundle() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(exe, ".app/Contents/MacOS/")
}

// 在 .app 内且除 Finder 遗留的 -psn_* 外没有任何参数时，应直接打开图形界面
// （访达「将文件拖到程序坞图标上」等会附带真实路径，此时须走 CLI 分支）。
func shouldLaunchBundledGUI() bool {
	if !runningInMacOSAppBundle() {
		return false
	}
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "-psn_") {
			continue
		}
		return false
	}
	return true
}

func processOneCLI(filePath, outDir string, format convert.OutputFormat) error {
	out, err := convert.ProcessFile(filePath, outDir, format)
	if errors.Is(err, convert.ErrSkipped) {
		return nil
	}
	if err != nil {
		if out != "" {
			utils.WarningPrintfln("'%s' failed: %s", filePath, err.Error())
		} else {
			utils.ErrorPrintfln("Processing '%s' failed: %s", filePath, err.Error())
		}
		return err
	}
	if out != "" {
		utils.DonePrintfln("'%s' -> '%s'", filePath, out)
	}
	return nil
}

func main() {
	if shouldLaunchBundledGUI() {
		gui.Run(appWindowIcon())
		return
	}

	var sourceDir string
	var outputDir string
	showHelp := flag.BoolP("help", "h", false, "Display help message")
	showVersion := flag.BoolP("version", "v", false, "Display version information")
	showGUI := flag.Bool("gui", false, "Launch graphical interface")
	processRecursive := flag.BoolP("recursive", "r", false, "Process all files in the directory recursively")
	outputFormat := flag.String("format", "auto", "Output format: auto (keep mp3/flac), mp3, flac, wav, aac (m4a); transcoding needs ffmpeg in PATH")
	flag.StringVarP(&outputDir, "output", "o", "", "Output directory for the dump files")
	flag.StringVarP(&sourceDir, "dir", "d", "", "Process all files in the directory")
	flag.Parse()

	outFmt, err := convert.ParseOutputFormat(*outputFormat)
	if err != nil {
		utils.ErrorPrintfln("%s", err.Error())
		os.Exit(1)
	}

	if *showGUI {
		gui.Run(appWindowIcon())
		return
	}

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println("ncmdump version " + version.String)
		os.Exit(0)
	}

	if !flag.Lookup("dir").Changed && sourceDir == "" && len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if flag.Lookup("recursive").Changed && !flag.Lookup("dir").Changed {
		utils.ErrorPrintfln("The -r option can only be used with the -d option")
		os.Exit(1)
	}

	outputDirSpecified := flag.Lookup("output").Changed

	if outputDirSpecified {
		if utils.PathExists(outputDir) {
			if !utils.IsDir(outputDir) {
				utils.ErrorPrintfln("Output directory '%s' is not valid.", outputDir)
				os.Exit(1)
			}
		} else {
			_ = os.MkdirAll(outputDir, os.ModePerm)
		}
	}

	if sourceDir != "" {
		if !utils.IsDir(sourceDir) {
			utils.ErrorPrintfln("The source directory '%s' is not valid.", sourceDir)
			os.Exit(1)
		}

		if *processRecursive {
			_ = filepath.WalkDir(sourceDir, func(p string, d os.DirEntry, err_ error) error {
				if !outputDirSpecified {
					outputDir = sourceDir
				}
				relativePath := utils.GetRelativePath(sourceDir, p)
				destinationPath := filepath.Join(outputDir, relativePath)

				if utils.IsRegularFile(p) {
					parentDir := filepath.Dir(destinationPath)
					_ = os.MkdirAll(parentDir, os.ModePerm)
					_ = processOneCLI(p, parentDir, outFmt)
				}
				return nil
			})
		} else {
			files, err := os.ReadDir(sourceDir)
			if err != nil {
				utils.ErrorPrintfln("Unable to read directory: '%s'", sourceDir)
				os.Exit(1)
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				filePath := filepath.Join(sourceDir, file.Name())
				if outputDirSpecified {
					_ = processOneCLI(filePath, outputDir, outFmt)
				} else {
					_ = processOneCLI(filePath, sourceDir, outFmt)
				}
			}
		}
	} else {
		for _, filePath := range flag.Args() {
			if !strings.EqualFold(filepath.Ext(filePath), ".ncm") {
				continue
			}
			if outputDirSpecified {
				_ = processOneCLI(filePath, outputDir, outFmt)
			} else {
				_ = processOneCLI(filePath, "", outFmt)
			}
		}
	}

}

package gui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"git.taurusxin.com/taurusxin/ncmdump-go/internal/convert"
	"git.taurusxin.com/taurusxin/ncmdump-go/internal/version"
	"git.taurusxin.com/taurusxin/ncmdump-go/utils"
)

// Run opens the graphical interface and blocks until the window is closed.
// icon 建议传入嵌入的 PNG，否则程序坞/窗口仍会显示 Fyne 默认图标（与 .app 外层图标不一致）。
func Run(icon fyne.Resource) {
	a := app.NewWithID("com.my.ncmdump-go")
	if icon != nil {
		a.SetIcon(icon)
	}
	w := a.NewWindow("ncmdump-go-gui")
	if icon != nil {
		w.SetIcon(icon)
	}
	w.Resize(fyne.NewSize(720, 560))
	w.SetFixedSize(false)

	var taskPaths []string

	pathList := widget.NewList(
		func() int { return len(taskPaths) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(taskPaths[id])
		},
	)

	outputEntry := widget.NewEntry()
	outputEntry.SetPlaceHolder("留空则输出到各 .ncm 文件所在目录")

	formatOpts := []string{
		"自动（与解密一致：MP3 或 FLAC）",
		"MP3",
		"FLAC",
		"WAV",
		"AAC（M4A）",
	}
	formatSel := widget.NewSelect(formatOpts, nil)
	formatSel.SetSelected(formatOpts[0])

	recursive := widget.NewCheck("递归扫描子文件夹（仅“添加文件夹”时有效）", nil)

	var logText string
	logLabel := widget.NewLabel("")
	logLabel.Wrapping = fyne.TextWrapWord

	appendLog := func(line string) {
		if logText != "" {
			logText += "\n" + line
		} else {
			logText = line
		}
		logLabel.SetText(logText)
	}

	addFilesBtn := widget.NewButton("添加 .ncm 文件…", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			p := reader.URI().Path()
			taskPaths = appendUnique(taskPaths, p)
			pathList.Refresh()
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".ncm"}))
		fd.Show()
	})

	addDirBtn := widget.NewButton("添加文件夹…", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			dir := uri.Path()
			ncms, err := convert.CollectNCMInDir(dir, recursive.Checked)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if len(ncms) == 0 {
				dialog.ShowInformation("未找到文件", "该文件夹下没有 .ncm 文件。", w)
				return
			}
			for _, p := range ncms {
				taskPaths = appendUnique(taskPaths, p)
			}
			pathList.Refresh()
			appendLog(fmt.Sprintf("[信息] 从目录添加 %d 个 .ncm：%s", len(ncms), dir))
		}, w)
	})

	selectedFormat := func() convert.OutputFormat {
		switch formatSel.Selected {
		case "MP3":
			return convert.FormatMP3
		case "FLAC":
			return convert.FormatFLAC
		case "WAV":
			return convert.FormatWAV
		case "AAC（M4A）":
			return convert.FormatAAC
		default:
			return convert.FormatAuto
		}
	}

	clearBtn := widget.NewButton("清空列表", func() {
		taskPaths = nil
		pathList.Refresh()
	})

	outBrowse := widget.NewButton("选择…", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			outputEntry.SetText(uri.Path())
		}, w)
	})

	doConvert := func() {
		if len(taskPaths) == 0 {
			dialog.ShowInformation("提示", "请先添加要转换的 .ncm 文件或文件夹。", w)
			return
		}
		outDir := strings.TrimSpace(outputEntry.Text)
		if outDir != "" {
			if utils.PathExists(outDir) {
				if !utils.IsDir(outDir) {
					dialog.ShowError(errors.New("输出路径不是目录"), w)
					return
				}
			} else if err := os.MkdirAll(outDir, 0o755); err != nil {
				dialog.ShowError(fmt.Errorf("创建输出目录失败: %w", err), w)
				return
			}
		}

		go func() {
			total := len(taskPaths)
			for i, p := range taskPaths {
				idx := i + 1
				fyne.Do(func() {
					appendLog(fmt.Sprintf("[进度] %d/%d — %s", idx, total, p))
				})
				out, err := convert.ProcessFile(p, outDir, selectedFormat())
				if errors.Is(err, convert.ErrSkipped) {
					fyne.Do(func() {
						appendLog(fmt.Sprintf("[跳过] 非 .ncm 文件：%s", p))
					})
					continue
				}
				if err != nil {
					if out != "" {
						fyne.Do(func() {
							appendLog(fmt.Sprintf("[警告] 元数据写入失败（文件已生成）：%s — %v", p, err))
						})
					} else {
						fyne.Do(func() {
							appendLog(fmt.Sprintf("[错误] %s — %v", p, err))
						})
					}
					continue
				}
				if out != "" {
					fyne.Do(func() {
						appendLog(fmt.Sprintf("[完成] %s -> %s", p, out))
					})
				}
			}
			fyne.Do(func() {
				appendLog("[信息] 全部任务结束。")
			})
		}()
	}

	convertBtn := widget.NewButton("开始转换", func() {
		doConvert()
	})
	convertBtn.Importance = widget.HighImportance

	aboutBtn := widget.NewButton("关于", func() {
		dialog.ShowInformation("关于 ncmdump-go",
			"版本 "+version.String+"\n\n"+
				"将网易云音乐 .ncm 解密为 MP3/FLAC，并写入元数据；"+
				"网易云 3.x 若文件中无封面，会尝试从网络获取（需联网）。\n\n"+
				"选择 WAV / AAC 等非原始格式时需本机已安装 ffmpeg 并完成转码。",
			w)
	})

	toolbar := container.NewHBox(addFilesBtn, addDirBtn, clearBtn, aboutBtn)

	pathScroll := container.NewScroll(pathList)
	pathScroll.SetMinSize(fyne.NewSize(0, 160))

	outRow := container.NewBorder(nil, nil, nil, outBrowse, outputEntry)

	formIntro := widget.NewLabel("待转换列表（可多次添加文件；添加文件夹时受「递归扫描」选项影响）")
	formIntro.Wrapping = fyne.TextWrapWord

	logScroll := container.NewScroll(logLabel)
	logScroll.SetMinSize(fyne.NewSize(0, 120))

	content := container.NewVBox(
		widget.NewLabel("选择输出目录（可选）"),
		outRow,
		container.NewHBox(widget.NewLabel("输出格式"), formatSel),
		widget.NewLabel("非「自动」且需转码时需 ffmpeg；打包后若无 PATH 可设置 NCMDUMP_FFMPEG 或安装在 /opt/homebrew/bin"),
		recursive,
		formIntro,
		toolbar,
		pathScroll,
		convertBtn,
		widget.NewSeparator(),
		widget.NewLabel("日志"),
		logScroll,
	)
	w.SetContent(container.NewPadded(content))
	w.ShowAndRun()
}

func appendUnique(slice []string, item string) []string {
	for _, x := range slice {
		if x == item {
			return slice
		}
	}
	return append(slice, item)
}

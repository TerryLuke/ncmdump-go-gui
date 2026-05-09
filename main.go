package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.taurusxin.com/taurusxin/ncmdump-go/internal/convert"
	"git.taurusxin.com/taurusxin/ncmdump-go/internal/gui"
	"git.taurusxin.com/taurusxin/ncmdump-go/internal/version"
	"git.taurusxin.com/taurusxin/ncmdump-go/utils"
	flag "github.com/spf13/pflag"
)

func processOneCLI(filePath, outDir string) error {
	out, err := convert.ProcessFile(filePath, outDir)
	if errors.Is(err, convert.ErrSkipped) {
		return nil
	}
	if err != nil {
		if out != "" {
			utils.WarningPrintfln("Fix metadata for '%s' failed: %s", filePath, err.Error())
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
	var sourceDir string
	var outputDir string
	showHelp := flag.BoolP("help", "h", false, "Display help message")
	showVersion := flag.BoolP("version", "v", false, "Display version information")
	showGUI := flag.Bool("gui", false, "Launch graphical interface")
	processRecursive := flag.BoolP("recursive", "r", false, "Process all files in the directory recursively")
	flag.StringVarP(&outputDir, "output", "o", "", "Output directory for the dump files")
	flag.StringVarP(&sourceDir, "dir", "d", "", "Process all files in the directory")
	flag.Parse()

	if *showGUI {
		gui.Run()
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
					_ = processOneCLI(p, parentDir)
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
					_ = processOneCLI(filePath, outputDir)
				} else {
					_ = processOneCLI(filePath, sourceDir)
				}
			}
		}
	} else {
		for _, filePath := range flag.Args() {
			if !strings.EqualFold(filepath.Ext(filePath), ".ncm") {
				continue
			}
			if outputDirSpecified {
				_ = processOneCLI(filePath, outputDir)
			} else {
				_ = processOneCLI(filePath, "")
			}
		}
	}

}

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

var (
	buildTime, commitId, versionData, author string
	inputFolder, outputFolder, passWord      string
	help, version                            bool
)

var (
	fileList []string
)

func Flag() {
	flag.BoolVar(&help, "h", false, "Display help information")
	flag.BoolVar(&help, "help", false, "Display help information")
	flag.BoolVar(&version, "v", false, "PDF_Decrypt version")
	flag.BoolVar(&version, "version", false, "PDF_Decrypt version")
	flag.StringVar(&inputFolder, "in", "PDF", "in explorer")
	flag.StringVar(&outputFolder, "out", "out", "out explorer")
	flag.StringVar(&passWord, "pass", "123456", "password ('abc' | 'abc\\def\\ghi')")
}

func passWordLists(passWord string) []string {
	return strings.Split(passWord, "\\")
}

func adsPath(folder string) string {
	var (
		adspath string
		err     error
	)
	if adspath, err = filepath.Abs(folder); err != nil {
		_, _ = color.New(color.FgYellow).Println(err)
		os.Exit(1)
	}
	return adspath
}

func inputFolders(filePath string) []string {
	if err := filepath.Walk(adsPath(filePath), func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".pdf" {
			fileList = append(fileList, path)
		}
		if info.IsDir() {
			var outFileName, outFilePath string
			if outFileName, err = filepath.Rel(adsPath(inputFolder), path); err != nil {
				_, _ = color.New(color.FgYellow).Println(err)
				os.Exit(1)
			}
			outFilePath = filepath.Join(outputFolder, outFileName)
			if _, err := os.Stat(outFilePath); err != nil {
				if err := os.Mkdir(outFilePath, 0755); err != nil {
					_, _ = color.New(color.FgYellow).Println(err)
					os.Exit(1)
				}
			}
		}
		return nil
	}); err != nil {
		_, _ = color.New(color.FgYellow).Println(err)
		os.Exit(1)
	}
	return fileList
}

func outPutFolder(filePath string) {
	filePath = adsPath(filePath)
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsExist(err) {
			err := os.Mkdir(filePath, 0755)
			if err != nil {
				_, _ = color.New(color.FgYellow).Println(err)
				os.Exit(1)
			}
		}
	}
}

func outPutFolderClean(filePath string) {
	var (
		files []fs.FileInfo
		err   error
	)
	filePath = adsPath(filePath)
	if files, err = ioutil.ReadDir(filePath); err != nil {
		_, _ = color.New(color.FgYellow).Println(err)
		os.Exit(1)
	}
	if len(files) == 0 {
		if err := os.Remove(filePath); err != nil {
			_, _ = color.New(color.FgYellow).Println(err)
			os.Exit(1)
		}
		_, _ = color.New(color.FgRed).Printf(" Remove Folder: ")
		_, _ = color.New(color.FgWhite).Println(filePath)
		return
	}
	for _, v := range files {
		if v.IsDir() {
			outPutFolderClean(path.Join(filePath, v.Name()))
		}
	}
}

func Decrypt(fileList []string) {
	for _, v := range fileList {
		for _, p := range passWordLists(passWord) {
			if DecryptFile(v, p) {
				break
			}
		}
	}

}

func DecryptFile(filePath, p string) bool {
	//conf.UserPW = i
	filePath = adsPath(filePath)
	if !pdfcpu.MemberOf(pdfcpu.ConfigPath, []string{"default", "disable"}) {
		if err := pdfcpu.EnsureDefaultConfigAt(pdfcpu.ConfigPath); err != nil {
			_, _ = color.New(color.FgYellow).Println(err)
			os.Exit(1)
		}
	}
	conf := pdfcpu.NewDefaultConfiguration()
	conf.OwnerPW = p
	_, _ = color.New(color.FgCyan).Printf(" Decrypting: ")
	_, _ = color.New(color.FgWhite).Printf(path.Base(filePath))
	outFileRouter, _ := filepath.Rel(adsPath(inputFolder), filePath)
	outFilePath := path.Join(adsPath(outputFolder), outFileRouter)
	if err := api.DecryptFile(filePath, outFilePath, conf); err != nil {
		if strings.Contains(err.Error(), "not encrypted") {
			_, _ = color.New(color.FgGreen).Println("Not encrypted")
			var l3 []byte
			if l3, err = ioutil.ReadFile(filePath); err != nil {
				_, _ = color.New(color.FgYellow).Println(err)
				os.Exit(1)
			}
			if err = ioutil.WriteFile(outFilePath, l3, 0644); err != nil {
				_, _ = color.New(color.FgYellow).Println(err)
				os.Exit(1)
			}
			return true
		}
		_, _ = color.New(color.FgYellow).Println(err)
		return false
	}
	_, _ = color.New(color.FgGreen).Printf(" Decrypted!")
	_, _ = color.New(color.FgWhite).Println(" PassWord:", p)
	return true
}

func main() {
	_, _ = color.New(color.FgMagenta).Println("\n --------------------------------------")
	Flag()
	flag.Parse()
	if help || len(os.Args) == 1 {
		fmt.Println(` Usage of PDF_Decrypt:
   -help -h
         Display help information
   -in string
         in explorer
   -out string
         out explorer
   -pass string
         password (abc | abc\\def\\ghi)
   -version -v
         PDF_Decrypt version`, "\n")
		return
	}
	if version {
		_, _ = color.New(color.FgMagenta).Println(" |  Version:", versionData)
		_, _ = color.New(color.FgMagenta).Println(" |  BuildTime:", buildTime)
		_, _ = color.New(color.FgMagenta).Println(" |  Author:", author)
		_, _ = color.New(color.FgMagenta).Println(" |  CommitId:", commitId)
		_, _ = color.New(color.FgMagenta).Println(" --------------------------------------")
		return
	}
	_, _ = color.New(color.FgMagenta).Println(" Start PDF_Decrypt ...")
	outPutFolder(outputFolder)
	Decrypt(inputFolders(inputFolder))
	for i := 0; i < 7; i++ {
		outPutFolderClean(outputFolder)
	}
	_, _ = color.New(color.FgGreen).Printf("\n Decrypted %v File\n", len(fileList))
}

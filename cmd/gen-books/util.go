package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kjk/u"
)

var (
	softErrorMode bool
	errors        []string
)

func maybePanicIfErr(err error) {
	if err == nil {
		return
	}
	if !softErrorMode {
		u.PanicIfErr(err)
	}
	errors = append(errors, err.Error())
}

func clearErrors() {
	errors = nil
}

func printAndClearErrors() {
	if len(errors) == 0 {
		return
	}
	errStr := strings.Join(errors, "\n")
	fmt.Printf("\n%d errors:\n%s\n\n", len(errors), errStr)
	clearErrors()
}

func createDirForFileMaybeMust(path string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	maybePanicIfErr(err)
}

func copyFileMaybeMust(dst, src string) {
	createDirForFileMaybeMust(dst)
	err := copyFile(dst, src)
	maybePanicIfErr(err)
}

// "foo.js" => "foo-${sha1}.js"
func nameToSha1Name(name, sha1Hex string) string {
	ext := filepath.Ext(name)
	n := len(name)
	s := name[:n-len(ext)]
	return s + "-" + sha1Hex[:8] + ext
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDirectory(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func createDirMust(dir string) {
	err := os.MkdirAll(dir, 0755)
	u.PanicIfErr(err)
}

func copyFile(dst, src string) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()
	fout, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fout.Close()
	_, err = io.Copy(fout, fin)
	return err
}

func copyFileMust(dst, src string) {
	err := copyFile(dst, src)
	u.PanicIfErr(err)
}

func getDirsRecur(dir string) ([]string, error) {
	toVisit := []string{dir}
	idx := 0
	for idx < len(toVisit) {
		dir = toVisit[idx]
		idx++
		fileInfos, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, fi := range fileInfos {
			if !fi.IsDir() {
				continue
			}
			path := filepath.Join(dir, fi.Name())
			toVisit = append(toVisit, path)
		}
	}
	return toVisit, nil
}

// "foo" + "bar" = "foo/bar", only one "/"
func urlJoin(s1, s2 string) string {
	if strings.HasSuffix(s1, "/") {
		if strings.HasPrefix(s2, "/") {
			return s1 + s2[1:]
		}
		return s1 + s2
	}

	if strings.HasPrefix(s2, "/") {
		return s1 + s2
	}
	return s1 + "/" + s2
}

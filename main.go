package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func main() {
	color.NoColor = false

	ds := []delete{}
	us := []upload{}
	ss := []string{}

	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		line := s.Text()

		switch c := classifyLine(line).(type) {
		case delete:
			ds = append(ds, c)
		case upload:
			us = append(us, c)
		case string:
			ss = append(ss, c)
		default:
			panic(fmt.Sprintf("unhandled type: %T", c))
		}
	}

	ds, us, ms := extractMoves(ds, us)

	printDropboxCache(ms, ds, us, ss)
	printOthers(ms, ds, us, ss)
}

type delete struct {
	dropboxCache bool
	s3Path       string
}

type upload struct {
	cache     bool
	localPath string
	s3Path    string
}

func classifyLine(s string) interface{} {
	d := parseDelete(s)
	if d != nil {
		return *d
	}

	u := parseUpload(s)
	if u != nil {
		return *u
	}

	return s
}

func parseDelete(s string) *delete {
	prefix := "(dryrun) delete: s3://"
	if !strings.HasPrefix(s, prefix) {
		return nil
	}
	return &delete{
		dropboxCache: isDropboxCache(s),
		s3Path:       strings.TrimPrefix(s, prefix),
	}
}

func parseUpload(s string) *upload {
	prefix := "(dryrun) upload: "
	if !strings.HasPrefix(s, prefix) {
		return nil
	}
	s = strings.TrimPrefix(s, prefix)

	midfix := " to s3://"
	i := strings.Index(s, midfix)
	if i < 0 {
		return nil
	}

	localPath := s[0:i]
	s3Path := strings.TrimPrefix(s[i:], midfix)

	return &upload{
		cache:     isDropboxCache(s),
		localPath: localPath,
		s3Path:    s3Path,
	}
}

func isDropboxCache(s string) bool {
	return strings.Contains(s, "/.dropbox.cache/")
}

type move struct {
	fileName  string
	s3FromDir string
	s3ToDir   string
}

func extractMoves(ds []delete, us []upload) ([]delete, []upload, []move) {
	resDs := []delete{}
	resMs := []move{}

	for _, d := range ds {
		uIndex := -1
		for j, u := range us {
			if isMove(d, u) {
				resMs = append(resMs, move{
					fileName:  fileName(d.s3Path),
					s3FromDir: dirName(d.s3Path),
					s3ToDir:   dirName(u.s3Path),
				})
				uIndex = j
				break
			}
		}
		if uIndex < 0 {
			resDs = append(resDs, d)
		} else {
			us = append(us[0:uIndex], us[uIndex+1:]...)
		}
	}

	return resDs, us, resMs
}

func isMove(d delete, u upload) bool {
	return fileName(d.s3Path) == fileName(u.s3Path)
}

func fileName(s string) string {
	i := strings.LastIndex(s, "/")
	return s[i+1:]
}

func dirName(s string) string {
	i := strings.LastIndex(s, "/")
	return s[0:i]
}

func printDropboxCache(ms []move, ds []delete, us []upload, ss []string) {
	printAll(ms, ds, us, ss, true)
}

func printOthers(ms []move, ds []delete, us []upload, ss []string) {
	printAll(ms, ds, us, ss, false)
}

func printAll(ms []move, ds []delete, us []upload, ss []string, cache bool) {
	if !cache {
		for _, x := range ms {
			printMove(x)
		}
	}
	for _, x := range ds {
		if x.dropboxCache == cache {
			printDelete(x)
		}
	}
	for _, x := range us {
		if x.cache == cache {
			printUpload(x)
		}
	}
	printSeparator()
	if !cache {
		for _, x := range ss {
			printUnrecognized(x)
		}
	}
}

func printDelete(x delete) {
	c := red
	if x.dropboxCache {
		c = gray
	}
	fmt.Printf("DELETE: %s\n", c(stripBucket(x.s3Path)))
}

func printUpload(x upload) {
	c := green
	if x.cache {
		c = gray
	}
	fmt.Printf("NEW: %s\n", c(stripBucket(x.s3Path)))
}

func printMove(x move) {
	fmt.Printf(
		"{%s -> %s}/%s\n",
		yellow(stripBucket(x.s3FromDir)), yellow(stripBucket(x.s3ToDir)), x.fileName,
	)
}

func printSeparator() {
	fmt.Printf("---\n")
}

func printUnrecognized(x string) {
	fmt.Printf("%s\n", x)
}

var gray = color.New(color.FgBlue).SprintFunc()
var red = color.New(color.FgHiRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

func stripBucket(s3Path string) string {
	i := strings.Index(s3Path, "/")
	if i < 0 {
		return ""
	}
	return s3Path[i+1:]
}

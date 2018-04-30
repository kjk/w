package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/siser"
	"github.com/kjk/u"
	"golang.org/x/net/html"
)

// URLInfo describes a crawled URL
type URLInfo struct {
	URL  string
	Code int    // 200, 404 etc.
	Body string // body of the url

	// data calculated, not saved
	Offset int64 // offset of the record so that we can read the data
}

const (
	startURL = "https://msdn.microsoft.com/en-us/library/windows/desktop/aa964920(v=vs.85).aspx"
	keyURL   = "url"
	keyCode  = "code"
	//keySize  = "size"
	keyBody = "body"
	//keyWhenDownloaded = "whenDownloaded"
	//keyDownloadTime = "dl_time"

	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"
)

var (
	dataDir        string // directory where we store data files
	dataFilePath   string
	urls           = make(map[string]*URLInfo)
	urlsDownloaded = make(map[string]bool)
)

func setDataDirMust() {
	dataDir = u.ExpandTildeInPath("~/data/winapigen")
	err := os.MkdirAll(dataDir, 0755)
	u.PanicIfErr(err)
	dataFilePath = filepath.Join(dataDir, "crawled_pages.txt")
}

func recordGetInt(rec *siser.Record, key string) (int, bool) {
	s, ok := rec.Get(key)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}

func recordGetIntMust(rec *siser.Record, key string) int {
	n, ok := recordGetInt(rec, key)
	u.PanicIf(!ok)
	return n
}

func recordGetStringMust(rec *siser.Record, key string) string {
	s, ok := rec.Get(key)
	u.PanicIf(!ok, "missing value for key '%s'", key)
	return s
}

func recordToURLInfoMust(rec *siser.Record) URLInfo {
	i := URLInfo{}
	i.URL = recordGetStringMust(rec, keyURL)
	i.Code = recordGetIntMust(rec, keyCode)
	return i
}

func appendToFile(path string, d []byte) error {
	fmt.Printf("Appending %d bytes to %s\n", len(d), path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(d)
	return err
}

func writeDownloadData(i *URLInfo) error {
	rec := siser.Record{}
	rec.Append(keyURL, i.URL)
	rec.Append(keyCode, strconv.Itoa(i.Code))
	rec.Append(keyBody, i.Body)
	d := rec.Marshal()
	return appendToFile(dataFilePath, d)
}

func openFileAtOffset(path string, off int64) (*os.File, error) {
	f, err := os.Open(dataFilePath)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(off, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func readDataMust() {
	f, err := os.Open(dataFilePath)
	if err != nil {
		return
	}
	defer f.Close()
	r := siser.NewReader(f)
	for r.ReadNext() {
		off, rec := r.Record()
		i := recordToURLInfoMust(rec)
		i.Offset = off
		urls[i.URL] = &i
	}
	u.PanicIfErr(r.Err())
}

func loadRecordBody(i *URLInfo) error {
	f, err := openFileAtOffset(dataFilePath, i.Offset)
	if err != nil {
		return err
	}
	defer f.Close()
	r := siser.NewReader(f)
	ok := r.ReadNext()
	if !ok {
		return fmt.Errorf("no record in %s at offset %d", dataFilePath, i.Offset)
	}
	_, rec := r.Record()
	i.Body = recordGetStringMust(rec, keyBody)
	return nil
}

var (
	httpClient *http.Client
)

func httpGet(url string) (*http.Response, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
		httpClient.Timeout = time.Second * 15
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func downloadURLCached(url string) (*URLInfo, error) {
	i, ok := urls[url]
	timeStart := time.Now()
	if ok {
		if i.Body == "" {
			err := loadRecordBody(i)
			if err != nil {
				return nil, err
			}
		}
		fmt.Printf("loaded '%s' from cache in %s\n", url, time.Since(timeStart))
		return i, nil
	}
	rsp, err := httpGet(url)
	if err != nil {
		return i, err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return i, err
	}
	fmt.Printf("downloaded '%s' in %s\n", url, time.Since(timeStart))
	i = &URLInfo{}
	i.URL = url
	i.Body = string(body)
	i.Code = rsp.StatusCode

	err = writeDownloadData(i)
	if err != nil {
		return i, nil
	}
	urlsDownloaded[url] = true
	return i, nil
}

func didDownload(url string) bool {
	_, ok := urlsDownloaded[url]
	return ok
}

const (
	strSpaces = "                                 "
)

// returns strings that consists of n spaces
func getSpaces(n int) string {
	s := strSpaces
	for n >= len(s) {
		s = s + s
	}
	return s[:n]
}

func supressTextForTag(tag string) bool {
	if tag == "script" {
		return true
	}
	return false
}

func prettyPrintHTML(s string) {
	var attrName, attrVal []byte
	r := bytes.NewBufferString(s)
	z := html.NewTokenizer(r)
	var tagsStack []string
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			//fmt.Printf("token: %d ErrorToken\n", tt)
			if z.Err() == io.EOF {
				fmt.Printf("Finsihed parsing\n")
				return
			}
			fmt.Printf("Finished parsing with error: %s\n", z.Err())
			return
		}

		if tt == html.StartTagToken {
			//fmt.Printf("token: %d StartTagToken\n", tt)
			tn, hasAttr := z.TagName()
			level := len(tagsStack)
			tns := string(tn)
			tagsStack = append(tagsStack, tns)
			var attrID []byte
			var attrClass []byte
			for hasAttr {
				attrName, attrVal, hasAttr = z.TagAttr()
				attrNameStr := string(attrName)
				if attrNameStr == "id" {
					attrID = attrVal
				} else if attrNameStr == "class" {
					attrClass = attrVal
				}
			}
			if attrID != nil {
				tns += "#" + string(attrID)
			}
			if attrClass != nil {
				classes := strings.Split(string(attrClass), " ")
				for _, cls := range classes {
					tns += "." + cls
				}
			}

			fmt.Printf("%s%s\n", getSpaces(level), tns)
		}

		if tt == html.EndTagToken {
			//fmt.Printf("token: %d EndTagToken\n", tt)
			n := len(tagsStack)
			tagsStack = tagsStack[:n-1]
		}

		if tt == html.TextToken {
			var currTag string
			level := len(tagsStack) + 1
			if level >= 2 {
				currTag = tagsStack[level-2]
			}
			if supressTextForTag(currTag) {
				continue
			}
			b := z.Text()
			s := strings.TrimSpace(string(b))
			if len(s) > 0 {
				if len(s) > 128 {
					s = s[:128] + "..."
				}
				fmt.Printf("%s%s\n", getSpaces(level), s)
			}
		}
		//fmt.Printf("token: %d\n", tt)
	}
}

func crawlStartingFrom(startURL string) {
	fmt.Printf("crawStartingFrom\n")
	urlsToVisit := []string{startURL}
	for len(urlsToVisit) > 0 {
		urlStr := urlsToVisit[0]
		urlsToVisit = urlsToVisit[1:]
		_, err := url.Parse(urlStr)
		if err != nil {
			fmt.Printf("Couldn't parse url '%s'\n", urlStr)
			continue
		}
		if didDownload(urlStr) {
			fmt.Printf("skipping, already downloaded '%s'\n", urlStr)
			continue
		}
		i, err := downloadURLCached(urlStr)
		u.PanicIfErr(err)
		prettyPrintHTML(i.Body)
	}
}

func parseMSDN() {
	setDataDirMust()
	readDataMust()
	crawlStartingFrom(startURL)
}

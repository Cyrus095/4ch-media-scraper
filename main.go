package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Thread struct {
	Posts []Post `json:"posts"`
}

type Post struct {
	No          int64  `json:"no"`
	Now         string `json:"now"`
	Name        string `json:"name"`
	Sub         string `json:"sub"`
	Com         string `json:"com"`
	Filename    string `json:"filename"`
	Ext         string `json:"ext"`
	W           int    `json:"w"`
	H           int    `json:"h"`
	TnW         int    `json:"tn_w"`
	TnH         int    `json:"tn_h"`
	Tim         int64  `json:"tim"`
	Time        int64  `json:"time"`
	Md5         string `json:"md5"`
	Fsize       int    `json:"fsize"`
	Resto       int    `json:"resto"`
	Bumplimit   int    `json:"bumplimit"`
	Imagelimit  int    `json:"imagelimit"`
	SemanticURL string `json:"semantic_url"`
	Replies     int    `json:"replies"`
	Images      int    `json:"images"`
}

var mediaQueue []int64

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: 4chan-scraper <thread-url> [dest]")
		os.Exit(1)
	}

	rawurl := os.Args[1]

	dest := ""
	if len(os.Args) >= 3 {
		dest = os.Args[2]
	} else {
		var err error
		dest, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	if dest == "" {
		dest, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	sects := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(sects) < 3 {
		panic("url not complete")
	}
	board := sects[0]
	threadid := sects[2]

	apiUrl := "https://a.4cdn.org/" + board + "/thread/" + threadid + ".json"
	fmt.Println(apiUrl)
	resp, err := http.Get(apiUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var thread Thread
	err = json.NewDecoder(resp.Body).Decode(&thread)
	if err != nil {
		panic(err)
	}
	fmt.Println("posts:", len(thread.Posts))
	var threadTitle = sanitizeTitle(thread.Posts[0].Sub)
	if threadTitle == "" {
		threadTitle = strconv.FormatInt(thread.Posts[0].Tim, 10)
	}
	dirPath := dest + "/" + threadTitle
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		panic(err)
	}

	postCount := 0
	for _, post := range thread.Posts {
		if post.Ext == ".jpg" ||
			post.Ext == ".jpeg" ||
			post.Ext == ".png" ||
			post.Ext == ".gif" ||
			post.Ext == ".webm" ||
			post.Ext == ".mp4" {
			mediaUrl := "https://i.4cdn.org/" + board + "/" + strconv.FormatInt(post.Tim, 10) + post.Ext
			resp, err := http.Get(mediaUrl)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			fullDest := dirPath + "/" + strconv.FormatInt(post.No, 10) + post.Ext
			out, err := os.Create(fullDest)
			if err != nil {
				panic(err)
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			postCount++
			fmt.Printf("[ %d ] downloaded\n", postCount)
		}
		duration := time.Duration(222) * time.Millisecond
		time.Sleep(duration)
	}

}

func sanitizeTitle(title string) string {
	title = strings.ToLower(title)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	title = re.ReplaceAllString(title, "-")
	title = strings.Trim(title, "-")
	return title
}

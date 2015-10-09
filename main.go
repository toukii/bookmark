package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	. "github.com/everfore/oauth/oauth2"
	"github.com/shaalx/leetcode/lfu2"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Bookmark struct {
	Title    string `json:"title"`
	Official string `json:"official"`
	Bgpic    string `json:"bgpic"`
	Site     string `json:"site"`
	Remark   string `json:"remark"`
	N        int    `json:"n"`
}

var (
	t      *template.Template
	hacker *template.Template
	v      []*Bookmark
	update chan bool
	cache  *lfu2.LFUCache
	OA     *OAGithub
)

func init() {
	update = make(chan bool, 10)
	hacker, _ = template.New("hacker.html").ParseFiles("hacker.html")
	t, _ = template.New("bookmark.html").ParseFiles("bookmark.html")
	// b := readFile("bookmark.md")
	b := get("http://7xku3c.com1.z0.glb.clouddn.com/bookmark.md")
	v = unmarshal(b)
	cache = lfu2.NewLFUCache(len(v))
	for i := len(v) - 1; i >= 0; i-- {
		cache.Set(v[i].Title, v[i])
	}
	update <- true
	OA = NewOAGithub("8ba2991113e68b4805c1", "b551e8a640d53904d82f95ae0d84915ba4dc0571", "user", "http://bookmark.daoapp.io/callback")
}

func main() {
	go updateBookmarks(time.Second)
	go flushBookmarks(time.Hour * 24 * 30)
	http.HandleFunc("/", bookmark)
	http.HandleFunc("/hacker", hackerHandler)
	http.HandleFunc("/lfu", lfu)
	http.HandleFunc("/signin", signin)
	http.HandleFunc("/callback", callback)
	http.ListenAndServe(":80", nil)
}

func readFile(name string) []byte {
	f, _ := os.OpenFile(name, os.O_RDONLY, 0064)
	defer f.Close()
	rd := bufio.NewReader(f)
	b, _ := ioutil.ReadAll(rd)
	fmt.Println(string(b))
	return b
}

func unmarshal(b []byte) []*Bookmark {
	var v []*Bookmark
	err := json.Unmarshal(b, &v)
	if nil != err {
		fmt.Println(err)
		return nil
	}
	return v
}

func get(_url string) []byte {
	resp, _ := http.Get(_url)
	b, _ := ioutil.ReadAll(resp.Body)
	return b
}

func flushBookmarks(d time.Duration) {
	ticker := time.NewTicker(d)
	for {
		<-ticker.C
		cache.Flush()
		update <- true
	}
}

func updateBookmarks(d time.Duration) {
	ticker := time.NewTicker(d)
	for {
		<-update
		<-ticker.C
		vals := cache.Vals()
		ret := make([]*Bookmark, len(vals))
		for i, it := range vals {
			bok := it.V.(*Bookmark)
			bok.N = it.N
			ret[i] = bok
		}
		v = ret
	}
}

func bookmark(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s  ", req.RemoteAddr)
	t.Execute(rw, v)
}

func hackerHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s  ", req.RemoteAddr)
	hacker.Execute(rw, nil)
}

func lfu(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}
	req.ParseForm()
	title := req.Form.Get("title")
	fmt.Printf("[ %s %s ]", req.RemoteAddr, title)
	cache.Get(title)
	update <- true
}

func signin(rw http.ResponseWriter, req *http.Request) {
	http.Redirect(rw, req, OA.AuthURL(), 302)
}

func callback(rw http.ResponseWriter, req *http.Request) {
	log.Printf("%s\n", req.RemoteAddr)
	b, err := OA.NextStep(req)
	if nil != err {
		rw.Write([]byte(err.Error()))
		return
	}
	var ret map[string]interface{}
	err = json.Unmarshal(b, &ret)
	if nil == err {
		t := template.New("oauth.html")
		t, err := t.ParseFiles("oauth.html")
		if nil != err {
			return
		}
		now := time.Now().String()
		ret["now"] = now
		t.Execute(rw, ret)
	} else {
		rw.Write([]byte(err.Error()))
	}
}

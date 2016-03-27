package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/everfore/exc"
	. "github.com/everfore/oauth/oauth2"
	"github.com/everfore/rpcsv"
	"github.com/shaalx/goutils"
	"github.com/shaalx/leetcode/lfu2"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strings"
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
	t                 *template.Template
	hacker            *template.Template
	markdown_edit_tpl *template.Template
	v                 []*Bookmark
	rmv               []*Bookmark
	update            chan bool
	cache             *lfu2.LFUCache
	OA                *OAGithub
	command           *exc.CMD
)

func init() {
	update = make(chan bool, 10)
	hacker, _ = template.New("hacker.html").ParseFiles("hacker.html")
	t, _ = template.New("bookmark.html").ParseFiles("bookmark.html")
	markdown_edit_tpl, _ = template.New("markdown_edit.html").ParseFiles("markdown_edit.html")
	b := readFile("bookmark.md")
	// b := get("http://7xku3c.com1.z0.glb.clouddn.com/bookmark.md")
	v = unmarshal(b)
	cache = lfu2.NewLFUCache(len(v) / 2)
	for i := len(v) - 1; i >= 0; i-- {
		cache.Set(v[i].Title, v[i])
	}
	update <- true
	OA = NewOAGithub("8ba2991113e68b4805c1", "b551e8a640d53904d82f95ae0d84915ba4dc0571", "user", "http://bookmark.daoapp.io/callback")
	command = exc.NewCMD("go version").Debug()
}

func main() {
	go updateBookmarks(time.Second)
	go flushBookmarks(time.Hour * 24 * 30)
	http.HandleFunc("/", bookmark)
	http.HandleFunc("/update", updateMD)
	http.HandleFunc("/hacker", hackerHandler)
	http.HandleFunc("/lfu", lfu)
	http.HandleFunc("/signin", signin)
	http.HandleFunc("/callback", callback)
	http.HandleFunc("/webhook", webhook)
	http.HandleFunc("/up", up)
	http.HandleFunc("/down", down)

	http.HandleFunc("/markdown_edit", markdown_edit)
	http.HandleFunc("/markdown", markdown)
	http.HandleFunc("/markdownCB", markdownCB)

	http.ListenAndServe(":80", nil)
}

func readFile(name string) []byte {
	f, _ := os.OpenFile(name, os.O_RDONLY, 0064)
	defer f.Close()
	rd := bufio.NewReader(f)
	b, _ := ioutil.ReadAll(rd)
	// fmt.Println(string(b))
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
		rmvals := cache.RmVals()
		ret := make([]*Bookmark, len(vals))
		retrm := make([]*Bookmark, len(rmvals))
		for i, it := range vals {
			bok := it.V.(*Bookmark)
			bok.N = it.N
			ret[i] = bok
		}
		for i, it := range rmvals {
			bok := it.V.(*Bookmark)
			bok.N = it.N
			retrm[i] = bok
		}
		v = ret
		rmv = retrm
	}
}

func bookmark(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s  ", req.RemoteAddr)
	// t.Execute(rw, v)
	data := make(map[string]interface{})
	data["size"] = len(v)
	data["rmsize"] = len(rmv)
	data["v"] = v
	data["rmv"] = rmv
	t.Execute(rw, data)
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

func up(rw http.ResponseWriter, req *http.Request) {
	cache.Insize(1)
	updateMD(rw, req)
}

func down(rw http.ResponseWriter, req *http.Request) {
	cache.Desize(1)
	update <- true
	http.Redirect(rw, req, "/", 302)
}

func updateMD(rw http.ResponseWriter, req *http.Request) {
	// b := get("http://7xku3c.com1.z0.glb.clouddn.com/bookmark.md")
	// b := get("https://raw.githubusercontent.com/shaalx/bookmark/master/bookmark.md")
	b := readFile("bookmark.md")
	v = unmarshal(b)
	allVals := make(map[string]int)
	for _, it := range cache.Vals() {
		allVals[it.Key] = 0
	}
	for _, it := range cache.RmVals() {
		allVals[it.Key] = 0
	}
	for i := len(v) - 1; i >= 0; i-- {
		// cur := cache.Attach(v[i].Title)
		// cur.N -= 1
		// cache.Set(v[i].Title, v[i])
		if _, exist := allVals[v[i].Title]; exist {
			cache.WhistPut(v[i].Title, v[i])
		} else {
			cache.Set(v[i].Title, v[i])
			fmt.Println("*****************IN******************", v[i].Title, v[i])
		}
		allVals[v[i].Title] = 1
	}
	for k, v := range allVals {
		if v <= 0 {
			cache.Remove(k)
			fmt.Println("rm ", k)
		}
	}
	update <- true
	http.Redirect(rw, req, "/", 302)
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

func webhook(rw http.ResponseWriter, req *http.Request) {
	usa := req.UserAgent()
	fmt.Println(usa)
	if !strings.Contains(usa, "GitHub-Hookshot/") {
		fmt.Println("CSRF Attack!")
		http.Redirect(rw, req, "/", 302)
		return
	}
	command.Reset("git pull origin master:master").Execute()
	updateMD(rw, req)
}

/*  markdown --start */

var (
	RPC_Client     *rpc.Client
	rpc_tcp_server = "tcphub.t0.daoapp.io:61142"
)

func connect() {
	RPC_Client = rpcsv.RPCClient(rpc_tcp_server)
	go func() {
		time.Sleep(2e9)
		RPC_Client.Close()
	}()
}

func markdown(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	rawContent := req.Form.Get("rawContent")
	fmt.Println(req.RemoteAddr, req.Referer())
	// fmt.Println(rawContent)
	out := make([]byte, 0, 100)
	in := goutils.ToByte(rawContent)
	RPC_Client = rpcsv.RPCClient(rpc_tcp_server)
	err := rpcsv.Markdown(RPC_Client, &in, &out)
	if goutils.CheckErr(err) {
		rw.Write(goutils.ToByte(err.Error()))
		return
	}
	if len(out) <= 0 {
		rw.Write(goutils.ToByte("{response:nil}"))
		return
	}
	writeCrossDomainHeaders(rw, req)
	rw.Write(out)
}

func writeCrossDomainHeaders(w http.ResponseWriter, req *http.Request) {
	// Cross domain headers
	if acrh, ok := req.Header["Access-Control-Request-Headers"]; ok {
		w.Header().Set("Access-Control-Allow-Headers", acrh[0])
	}
	w.Header().Set("Access-Control-Allow-Credentials", "True")
	if acao, ok := req.Header["Access-Control-Allow-Origin"]; ok {
		w.Header().Set("Access-Control-Allow-Origin", acao[0])
	} else {
		if _, oko := req.Header["Origin"]; oko {
			w.Header().Set("Access-Control-Allow-Origin", req.Header["Origin"][0])
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	// w.Header().Set("Connection", "Close")
}

func markdownCB(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	rawContent := req.Form.Get("rawContent")
	fmt.Println(req.RemoteAddr, req.Referer())
	// fmt.Println(rawContent)
	out := make([]byte, 0, 100)
	in := goutils.ToByte(rawContent)
	RPC_Client = rpcsv.RPCClient(rpc_tcp_server)
	err := rpcsv.Markdown(RPC_Client, &in, &out)
	if goutils.CheckErr(err) {
		rw.Write(goutils.ToByte(err.Error()))
		return
	}
	if len(out) <= 0 {
		rw.Write(goutils.ToByte("{response:nil}"))
		return
	}
	writeCrossDomainHeaders(rw, req)
	fmt.Println(req.RemoteAddr)
	CallbackFunc := fmt.Sprintf("CallbackFunc(%v);", string(Json(goutils.ToString(out))))
	fmt.Fprint(rw, CallbackFunc)
}

type CallbackData struct {
	Mddata interface{} `json:"mddata"`
}

func Json(data interface{}) []byte {
	bs, err := json.Marshal(CallbackData{Mddata: data})
	if goutils.CheckErr(err) {
		return nil
	}
	return bs
}

func markdown_edit(rw http.ResponseWriter, req *http.Request) {
	markdown_edit_tpl.Execute(rw, nil)
}

/*  markdown --end */

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/jamieabc/ig-check-new-post/config_parser"
	"golang.org/x/net/html"
)

const (
	instagramURL    = "https://www.instagram.com"
	sharedDataStr   = "_sharedData"
	graphImageRegex = `"GraphImage","id":"(\d+)"`
)

var (
	cache map[string]string
)

func main() {
	if 2 > len(os.Args) {
		fmt.Println("not enought parameter, usage: ig-check-new-post [config file]\n")
		return
	}

	cache = make(map[string]string)

	fileName := os.Args[1]
	config, err := config_parser.Parse(fileName)
	if nil != err {
		fmt.Printf("parser %s with error: %s\n", err)
	}

	usersWithNewPost := check(config, cache)

	notify(usersWithNewPost)
}

func check(config config_parser.Config, cache map[string]string) []string {
	usersWithNewPost := make([]string, 0)

	for _, user := range config.Accounts {
		resp, err := response(user)
		if nil != err {
			fmt.Printf("get response with error: %s\n", err)
			continue
		}
		defer resp.Body.Close()

		page, err := html.Parse(resp.Body)
		if nil != err {
			fmt.Printf("parse html with error: %s\n", err)
			continue
		}

		b, err := body(page)
		if nil != err {
			fmt.Printf("retrieve body with error: %s\n", err)
			continue
		}

		s, err := scripts(b)
		if nil != err {
			fmt.Printf("get scripts with error: %s\n", err)
		}

		id, err := latestGraphImageID(s)
		if nil != err {
			fmt.Printf("get latest graph image id with error: %s\n", err)
			continue
		}

		if newPost := newer(cache, user); newPost {
			cache[user] = id
			usersWithNewPost = append(usersWithNewPost, user)
		}
	}

	return usersWithNewPost
}

func response(user string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", instagramURL, user)
	resp, err := http.Get(url)
	if nil != err {
		fmt.Printf("get %s with error: %s\n", user, err)
	}
	return resp, err
}

func body(page *html.Node) (*html.Node, error) {
	var body *html.Node
	var crawler func(*html.Node)

	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "body" {
			body = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}

	crawler(page)
	if body != nil {
		return body, nil
	}
	return nil, fmt.Errorf("missing body in html")
}

func scripts(body *html.Node) ([]*html.Node, error) {
	scripts := make([]*html.Node, 0)
	var crawler func(*html.Node)
	crawler = func(n *html.Node) {
		if html.ElementNode == n.Type && "script" == n.Data {
			scripts = append(scripts, n)
			return
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(body)

	if 0 == len(scripts) {
		return nil, fmt.Errorf("missing script in body")
	}

	return scripts, nil
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

func latestGraphImageID(ns []*html.Node) (string, error) {
	msg := fmt.Errorf("cannot find GraphImage id")
	for _, v := range ns {
		str := renderNode(v)
		if strings.Contains(str, sharedDataStr) {
			re := regexp.MustCompile(graphImageRegex)
			match := re.FindStringSubmatch(str)
			if 0 < len(match) {
				id := match[len(match)-1]
				return id, nil
			}
			return "", msg
		}
	}
	return "", msg
}

func newer(cache map[string]string, key string) bool {
	if _, ok := cache[key]; !ok {
		return true
	}
	return false
}

func notify(user []string) {
	if 0 == len(user) {
		return
	}

	var err error
	var cmds []*exec.Cmd

	switch runtime.GOOS {
	case "linux":
		for _, u := range user {
			cmds = append(cmds, exec.Command("notify-send", u))
		}
	case "darwin":
		for _, u := range user {
			notificationCmd := fmt.Sprintf(`display notification "%s"`, u)
			cmds = append(cmds, exec.Command("osascript", "-e", notificationCmd))
		}
	}

	var stderr bytes.Buffer
	for _, c := range cmds {
		c.Stderr = &stderr
		err = c.Run()
		if nil != err {
			fmt.Printf("notify %s: %s\n", err, stderr.String())
		}
	}
}

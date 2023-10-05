package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

//go:embed templates/*.html
var templates embed.FS

var LastReverseProxyPort = uint32(8090 - 1)

type DekanatReverseProxy struct {
	Offline bool

	RemoteOriginBytes []byte
	RemoteUrl         *url.URL

	ProxyOrigin      string
	ProxyOriginBytes []byte

	BlockedRequests []*http.Request
}

func NewReverseProxy(remoteOrigin string) *DekanatReverseProxy {
	gin.SetMode(gin.ReleaseMode)

	remote, _ := url.Parse(remoteOrigin)

	port := atomic.AddUint32(&LastReverseProxyPort, 1)
	stringPort := strconv.FormatUint(uint64(port), 10)

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	proxyOrigin := "http://" + hostname + ":" + stringPort

	reverseProxy := &DekanatReverseProxy{
		Offline:           false,
		RemoteUrl:         remote,
		RemoteOriginBytes: []byte(remoteOrigin),

		ProxyOrigin:      proxyOrigin,
		ProxyOriginBytes: []byte(proxyOrigin),

		BlockedRequests: make([]*http.Request, 0, 20),
	}

	r := gin.Default()
	r.SetHTMLTemplate(
		template.Must(template.New("").ParseFS(templates, "templates/*.html")),
	)
	r.Any("/*proxyPath", reverseProxy.proxyAction)

	go r.Run(":" + stringPort)

	return reverseProxy
}

func (reverseProxy *DekanatReverseProxy) proxyAction(c *gin.Context) {
	if reverseProxy.Offline {
		reverseProxy.blockAction(c, "offline mode")
		return
	}

	if c.Request.URL.Query().Get("action") == "delete" {
		fmt.Println("[proxy] Blocked delete request by url param", c.Request.RequestURI)
		reverseProxy.blockAction(c, "delete action in url param")
		return
	}

	if strings.Contains(c.Request.RequestURI, "&n=11&action=delete") {
		fmt.Println("[proxy] Blocked delete request by substr", c.Request.RequestURI)
		reverseProxy.blockAction(c, "delete action in url substr")
		return
	}

	if c.Request.Method == "POST" {
		if c.Request.URL.Query().Get("n") == "1" && strings.HasSuffix(c.Request.URL.Path, "kaf.cgi") {
			fmt.Println("[proxy] Allowed post login request", c.Request.RequestURI)
			// allow login action
		} else {
			fmt.Println("[proxy] Blocked post request", c.Request.RequestURI)
			reverseProxy.blockAction(c, "post request")
			return
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(reverseProxy.RemoteUrl)
	proxy.ModifyResponse = reverseProxy.rewriteBody
	proxy.Director = func(req *http.Request) {
		req.URL.Path = c.Request.URL.Path
		req.Header = c.Request.Header
		req.Host = reverseProxy.RemoteUrl.Host
		req.URL.Scheme = reverseProxy.RemoteUrl.Scheme
		req.URL.Host = reverseProxy.RemoteUrl.Host
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (reverseProxy *DekanatReverseProxy) blockAction(c *gin.Context, reason string) {
	c.HTML(403, "blocked.html", gin.H{
		"reason": reason,
	})

	fmt.Println("Blocked request", c.Request.RequestURI)
	reverseProxy.BlockedRequests = append(reverseProxy.BlockedRequests, c.Request)
}

func (reverseProxy *DekanatReverseProxy) rewriteBody(resp *http.Response) (err error) {
	b, err := io.ReadAll(resp.Body) //Read html
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	// rewrite host
	b = bytes.Replace(b, reverseProxy.RemoteOriginBytes, reverseProxy.ProxyOriginBytes, -1)
	body := io.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return nil
}

func (reverseProxy *DekanatReverseProxy) SwitchOffline() {
	reverseProxy.Offline = true
}
func (reverseProxy *DekanatReverseProxy) SwitchOnline() {
	reverseProxy.Offline = false
}

func (reverseProxy *DekanatReverseProxy) ClearBlockedRequests() {
	reverseProxy.BlockedRequests = make([]*http.Request, 0, 20)
}

func (reverseProxy *DekanatReverseProxy) GetBlockedRequests() []*http.Request {
	return reverseProxy.BlockedRequests
}

func (reverseProxy *DekanatReverseProxy) GetBlockedRequestsCount() int {
	return len(reverseProxy.BlockedRequests)
}

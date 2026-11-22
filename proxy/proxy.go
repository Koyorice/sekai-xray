package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"log"

	"github.com/elazarl/goproxy"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	crypt = NewCrypt(
		[]byte{
			0x72, 0x4a, 0x6e, 0x36,
			0x56, 0x43, 0x36, 0x76,
			0x4a, 0x68, 0x45, 0x58,
			0x32, 0x78, 0x4c, 0x4f,
		},
		[]byte{
			0x50, 0x46, 0x53, 0x65,
			0x55, 0x47, 0x39, 0x76,
			0x36, 0x4d, 0x70, 0x6f,
			0x68, 0x5a, 0x4b, 0x4f,
		},
	)
	suite = regexp.MustCompile(
		`^/api/suite/user/(\d+)$`,
	)
	mysekai = regexp.MustCompile(
		`^/api/user/(\d+)/mysekai$`,
	)
)

func Start(addr string) {
	cert := ensureCA()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = log.New(io.Discard, "", 0)

	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/proxy.pac" {
			data := []byte(fmt.Sprintf(
				"function FindProxyForURL(a,b){return\"mitm.it\"===b||\"mk-zian-obt-cdn.bytedgame.com\"===b?\"PROXY %s\":\"DIRECT\"}",
				r.Host,
			))

			w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")

			_, err := w.Write(data)
			if err != nil {
				panic(err)
			}

			return
		}

		http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", http.StatusInternalServerError)
	})

	action := &goproxy.ConnectAction{
		Action:    goproxy.ConnectMitm,
		TLSConfig: goproxy.TLSConfigFromCA(cert),
	}

	proxy.OnRequest(
		goproxy.DstHostIs("mk-zian-obt-cdn.bytedgame.com"),
	).HandleConnect(
		goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
			return action, host
		}),
	)

	proxy.OnRequest(
		goproxy.ReqHostIs("mitm.it"),
	).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		buf := bytes.NewReader(cert.Certificate[0])

		res := &http.Response{
			Request:       req,
			StatusCode:    http.StatusOK,
			ContentLength: int64(buf.Len()),
			Body:          io.NopCloser(buf),
			Header: http.Header{
				"Content-Type":        []string{"application/x-x509-ca-cert"},
				"Content-Disposition": []string{"attachment; filename=ca.crt"},
			},
		}

		return req, res
	})

	proxy.OnResponse(
		goproxy.RespConditionFunc(func(res *http.Response, ctx *goproxy.ProxyCtx) bool {
			if res == nil {
				return false
			}

			if suite.MatchString(res.Request.URL.Path) {
				ctx.UserData = "suite"
				return true
			}

			if mysekai.MatchString(res.Request.URL.Path) {
				if res.Request.URL.Query().Get("isForceAllReloadOnlyMysekai") == "True" {
					ctx.UserData = "mysekai"
					return true
				}
			}

			return false
		}),
	).DoFunc(
		func(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			dataType := ctx.UserData.(string)
			log.Println("Capture data of type", dataType)

			body, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			decrypted, err := crypt.Decrypt(body)
			if err != nil {
				panic(err)
			}

			var data any
			err = msgpack.Unmarshal(decrypted, &data)
			if err != nil {
				panic(err)
			}

			indent, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				panic(err)
			}

			file := fmt.Sprintf("%s-%d.json", dataType, time.Now().Unix())

			err = os.WriteFile(file, indent, 0600)
			if err != nil {
				panic(err)
			}

			log.Println("Save data to file", file)

			res.Body = io.NopCloser(bytes.NewReader(body))

			return res
		},
	)

	log.Printf("HTTP proxy listening on %s", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           proxy,
		ReadHeaderTimeout: time.Minute,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

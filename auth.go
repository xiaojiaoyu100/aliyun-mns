package alimns

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xiaojiaoyu100/cast"
)

func withAuth(accessKeyID, accessKeySecret string) cast.RequestHook {
	return func(_ *cast.Cast, request *cast.Request) error {
		body, err := request.ReqBody()
		if err != nil {
			return err
		}
		request.RawRequest().Header.Set(contentType, textXML)
		request.RawRequest().Header.Set(date, time.Now().UTC().Format(http.TimeFormat))
		request.RawRequest().Header.Set(xMnsVersion, "2015-06-06")
		request.RawRequest().Header.Set(contentLength, strconv.FormatInt(int64(len(body)), 10))
		request.RawRequest().Header.Set(host, request.RawRequest().Host)
		h, err := Base64Md5(string(body))
		if err != nil {
			return err
		}
		request.RawRequest().Header.Set(contentMd5, h)
		toSign := []byte(
			request.RawRequest().Method + "\n" +
				request.RawRequest().Header.Get(contentMd5) + "\n" +
				textXML + "\n" +
				request.RawRequest().Header.Get(date) + "\n" +
				canonicalizedMNSHeaders(request.RawRequest().Header) + "\n" +
				canonicalizedResource(request.RawRequest()),
		)
		sig, err := signature(accessKeySecret, toSign)
		if err != nil {
			return err
		}
		auth := fmt.Sprintf("MNS %s:%s", accessKeyID, sig)
		request.RawRequest().Header.Set(authorization, auth)
		return nil
	}
}

func signature(accessKeySecret string, toSign []byte) (string, error) {
	h := hmac.New(sha1.New, []byte(accessKeySecret))
	_, err := h.Write(toSign)
	if err != nil {
		return "", err
	}
	sign := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(sign), nil
}

func canonicalizedMNSHeaders(header http.Header) string {
	var mnsHeaders []string
	for k := range header {
		if l := strings.ToLower(k); strings.HasPrefix(l, "x-mns-") {
			mnsHeaders = append(mnsHeaders, l)
		}
	}
	sort.Strings(mnsHeaders)
	var pairs []string
	for _, mnsHeader := range mnsHeaders {
		pairs = append(pairs, mnsHeader+":"+header.Get(mnsHeader))
	}
	return strings.Join(pairs, "\n")
}

func canonicalizedResource(request *http.Request) string {
	if request.URL.RawQuery == "" {
		return request.URL.Path
	}
	return request.URL.Path + "?" + request.URL.RawQuery
}

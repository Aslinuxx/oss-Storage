package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
)

//根据http请求的头部来获取 对象内容的size信息;
//put请求的完整格式: curl -v 10.29.2.2:12345/objects/test3 -XPUT -d "this is V3 test3" -H "Digest:SHA-256=hash值(散列值)" -H "Content-Length:长度值"
//注意这里客户端不需要指定object对象的数据内容长度;因为http协议会自动计算; 只需要调用h.Get()就行; 在http中该信息字段名是content-length
func GetSizeFromHeader(h http.Header) int64 {
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

//将客户端put请求的hash值取出来
func GetHashFromHeader(h http.Header) string {
	digest := h.Get("digest") // 这个digest是客户端自己加入头部的
	if len(digest) < 9 {      // 前八位刚好是SHA-256=
		return ""
	}
	if digest[:8] != "SHA-256=" {
		return ""
	}
	return digest[8:]

}

func CalculateHash(r io.Reader) string {
	h := sha256.New()
	io.Copy(h, r)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func GetOffsetFromHeader(h http.Header) int64 {
	byteRange := h.Get("range")
	if len(byteRange) < 7 {
		return 0
	}
	if byteRange[:6] != "bytes=" {
		return 0
	}
	bytePos := strings.Split(byteRange[6:], "-")
	offset, _ := strconv.ParseInt(bytePos[0], 0, 64)
	return offset
}

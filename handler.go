package main

import (
	"bytes"
	"net"
	"unsafe"

	"github.com/valyala/fasthttp"
)

func requestHandler(ctx *fasthttp.RequestCtx) {
	ip := GetFastHttpRemoteIP(ctx, true)
	ctx.Response.SetBodyString(ip.String() + "\n")

	ctx.SetContentType("text/plain")
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// see https://en.wikipedia.org/wiki/Reserved_IP_addresses
func IsReservedIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		switch ip4[0] {
		case 10:
			return true
		case 100:
			return ip4[1] >= 64 && ip4[1] <= 127
		case 127:
			return true
		case 169:
			return ip4[1] == 254
		case 172:
			return ip4[1] >= 16 && ip4[1] <= 31
		case 192:
			switch ip4[1] {
			case 0:
				switch ip4[2] {
				case 0, 2:
					return true
				}
			case 18, 19:
				return true
			case 51:
				return ip4[2] == 100
			case 88:
				return ip4[2] == 99
			case 168:
				return true
			}
		case 203:
			return ip4[1] == 0 && ip4[2] == 113
		case 224:
			return true
		case 240:
			return true
		}
	}
	return false
}

func GetFastHttpRemoteIP(ctx *fasthttp.RequestCtx, trustXForwardedFor bool) net.IP {
	ip := ctx.RemoteIP()
	if !trustXForwardedFor && !IsReservedIP(ip) {
		return ip
	}

	// Header X-Real-Ip
	if b := ctx.Request.Header.Peek("X-Real-Ip"); len(b) != 0 {
		if ip1 := net.ParseIP(b2s(b)); ip1 != nil {
			return ip1
		}
	}

	// Header X-Forwarded-For
	for _, b := range bytes.Split(ctx.Request.Header.Peek(fasthttp.HeaderXForwardedFor), []byte{',', ' '}) {
		if ip1 := net.ParseIP(b2s(b)); ip1 != nil && !IsReservedIP(ip1) {
			return ip1
		}
	}

	return ip
}

package httpapi

import "github.com/gogf/gf/v2/net/ghttp"

func Healthz(request *ghttp.Request) {
	request.Response.WriteJson(map[string]string{"status": "up"})
}

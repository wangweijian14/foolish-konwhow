package controller

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "naivete/api/v1"
	"naivete/internal/service"
)

var (
	Hello = cHello{}
)

type cHello struct{}

func (c *cHello) Hello(ctx context.Context, req *v1.HelloReq) (res *v1.HelloRes, err error) {
	service.Page().Create(ctx, "hello world")
	err = service.Graph().BuildGraph(ctx)
	if err != nil {
		g.RequestFromCtx(ctx).Response.Writeln(err.Error())
		return nil, err
	}
	g.RequestFromCtx(ctx).Response.Writeln("build-done")
	return
}

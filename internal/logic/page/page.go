package page

import (
	"context"

	"naivete/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

type sPage struct{}

func init() {
	service.RegisterPage(New())
}
func New() *sPage {
	return &sPage{}
}

// Create 创建Page
func (s *sPage) Create(ctx context.Context, wd string) (err error) {
	g.Log().Info(ctx, wd)
	return nil
}

package params

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Paging struct {
	Limit  int `form:"limit,default=25"`
	Offset int `form:"offset,default=0"`
}

func getPaging(ctx *gin.Context) Paging {
	p := Paging{Limit: 25, Offset: 0}
	// No need to check for error.
	ctx.BindQuery(&p)
	return p
}

func (p *Paging) Compose() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Limit, p.Offset)
}

package params

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type Params struct {
	DateRange []pgtype.Date
	Days      *int
	Filters   []Filter
	Page      Paging
	Search    *string
	Sorts     []Sort
}

func GetParams(ctx *gin.Context) *Params {
	return &Params{
		DateRange: getDateRange(ctx),
		Days:      getDays(ctx),
		Filters:   getFilters(ctx),
		Page:      getPaging(ctx),
		Search:    getSearch(ctx),
		Sorts:     getSorts(ctx),
	}
}

func cleanString(s string) string {
	return regNonAlpha.ReplaceAllString(s, "")
}

package params

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type Sort struct {
	Column string
	Asc    bool
}

func getSorts(ctx *gin.Context) []Sort {
	res := []Sort{}
	ss, ok := ctx.GetQueryArray("sort")
	if !ok {
		return nil
	}
	for _, s := range ss {
		var (
			col string
			asc bool
		)
		if s != "" {
			tokens := strings.SplitN(s, ":", 2)

			if len(tokens) != 2 {
				return nil
			}

			col = tokens[0]
			asc = strings.ToLower(tokens[1]) == "asc"

			res = append(res, Sort{Column: col, Asc: asc})
		} else {
			continue
		}
	}

	return res
}

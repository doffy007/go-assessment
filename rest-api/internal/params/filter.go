package params

import (
	"fmt"
	"regexp"
	"rest-api/internal/util"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

const (
	Equal            = "eq"
	NotEqual         = "neq"
	LessThan         = "lt"
	GreaterThan      = "gt"
	LessThanEqual    = "lte"
	GreaterThanEqual = "gte"
	Prefix           = "pre"
	Contain          = "con"
	In               = "in"
	NotIn            = "nin"
	Is               = "is"
	IsNot            = "isnot"
	Null             = "null"
	True             = "true"
	False            = "false"
	Like             = "like"
)

type Filter struct {
	Column           string `json:"column"`
	ColumnOverridden bool   `json:"-"`
	Operator         string `json:"operator"`
	Value            string `json:"value"`
}

var regNonAlpha *regexp.Regexp

func init() {
	var err error
	regNonAlpha, err = regexp.Compile("[^ a-zA-Z0-9_]+")
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func (arg *Params) ComposeDbQueryFromFilters(sb *strings.Builder, args []any) []any {
	for _, f := range arg.Filters {
		if f.ColumnOverridden {
			continue
		}

		if f.Operator == In || f.Operator == NotIn {
			args = append(args, strings.Split(f.Value, ","))
		} else if f.Operator != Is && f.Operator != IsNot {
			args = append(args, f.Value)
		}

		n := len(args)
		col := cleanString(f.Column)

		switch f.Operator {
		case Equal:
			sb.WriteString(fmt.Sprintf(" AND %s = $%d ", col, n))
		case NotEqual:
			sb.WriteString(fmt.Sprintf(" AND %s <> $%d ", col, n))
		case LessThan:
			sb.WriteString(fmt.Sprintf(" AND %s < $%d ", col, n))
		case GreaterThan:
			sb.WriteString(fmt.Sprintf(" AND %s > $%d ", col, n))
		case LessThanEqual:
			sb.WriteString(fmt.Sprintf(" AND %s <= $%d ", col, n))
		case GreaterThanEqual:
			sb.WriteString(fmt.Sprintf(" AND %s >= $%d ", col, n))
		case In:
			// Special handling for known array columns.
			if strings.EqualFold(col, "tags") {
				sb.WriteString(fmt.Sprintf(" AND %s && $%d ", col, n))
			} else {
				sb.WriteString(fmt.Sprintf(" AND %s = ANY($%d)", col, n))
			}
		case NotIn:
			sb.WriteString(fmt.Sprintf(" AND NOT (%s = ANY($%d))", col, n))
		case Is:
			switch strings.ToLower(f.Value) {
			case Null:
				sb.WriteString(fmt.Sprintf(" AND %s IS NULL ", col))
			case True:
				sb.WriteString(fmt.Sprintf(" AND %s IS TRUE ", col))
			case False:
				sb.WriteString(fmt.Sprintf(" AND %s IS FALSE ", col))
			}
		case IsNot:
			switch strings.ToLower(f.Value) {
			case Null:
				sb.WriteString(fmt.Sprintf(" AND %s IS NOT NULL ", col))
			case True:
				sb.WriteString(fmt.Sprintf(" AND %s IS NOT TRUE ", col))
			case False:
				sb.WriteString(fmt.Sprintf(" AND %s IS NOT FALSE ", col))
			}
		case Like:
			args = append(args, "%"+f.Value+"%")
			sb.WriteString(fmt.Sprintf(" AND %s ILIKE $%d ", col, len(args)))

		}
	}

	return args
}

func ComposeJsonbPredicate(op, val string, i int) string {
	switch op {
	case Equal:
		return fmt.Sprintf(" = $%d ", i)
	case NotEqual:
		return fmt.Sprintf(" <> $%d ", i)
	case LessThan:
		return fmt.Sprintf(" < $%d ", i)
	case GreaterThan:
		return fmt.Sprintf(" > $%d ", i)
	case LessThanEqual:
		return fmt.Sprintf(" <= $%d ", i)
	case GreaterThanEqual:
		return fmt.Sprintf(" >= $%d ", i)
	case In:
		return fmt.Sprintf(" = ANY($%d)", i)
	case Is:
		switch strings.ToLower(val) {
		case Null:
			return " IS " + val
		case True, False:
			return " = '" + val + "'"
		}
	case IsNot:
		switch strings.ToLower(val) {
		case Null:
			return " IS NOT " + val
		case True, False:
			return " <> '" + val + "'"
		}
	}

	return ""
}

func getFilters(ctx *gin.Context) []Filter {
	res := []Filter{}
	fs, ok := ctx.GetQueryArray("filter")
	if !ok {
		return nil
	}
	if len(fs) > 0 {
		for _, f := range fs {
			s := strings.SplitN(f, ":", 3)
			if len(s) == 3 {
				col := s[0]
				op := s[1]
				val := s[2]
				res = append(res, Filter{Column: col, Operator: op, Value: val})
			}
		}
	} else {
		return nil
	}
	return res
}

func getSearch(ctx *gin.Context) *string {
	s := strings.TrimSpace(ctx.Query("search"))
	if s == "" {
		return nil
	}
	return &s
}

func getDateRange(ctx *gin.Context) []pgtype.Date {
	var start, end pgtype.Date

	if err := util.DecodePgtypeDateText(ctx.Query("start"), &start); err != nil {
		if ctx.Query("start") != "" {
			log.Warn().Err(err).Msg("filter start date invalid or not present")
		}
		return nil
	}

	if err := util.DecodePgtypeDateText(ctx.Query("end"), &end); err != nil {
		log.Warn().Err(err).Msg("filter end date invalid or not present")
		end = start
	}

	return []pgtype.Date{start, end}
}

func getDays(ctx *gin.Context) *int {
	i, err := strconv.Atoi(strings.TrimSpace(ctx.Query("days")))
	if err != nil || i <= 0 {
		return nil
	}
	return &i
}

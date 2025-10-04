package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"rest-api/internal/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
	"github.com/rs/zerolog/log"
)

var (
	ErrPayload           = errors.New("something wrong with payload format")
	ErrQuery             = errors.New("something wrong with query format")
	ErrUnauthorized      = errors.New("please signin")
	ErrUnverified        = errors.New("please verify your account")
	ErrInsufficientRoles = errors.New("insufficient roles")
	ErrNotFound          = errors.New("not found")
	ErrNoAccess          = errors.New("has no access")
	ErrUpstream          = errors.New("upstream error")
	ErrNotImplemented    = errors.New("not implemented")
)

type Response struct {
	Results any   `json:"results"`
	Total   int64 `json:"total"`
}

func (resp *Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	if resp.Total == 0 && resp.Results == nil {
		resp.Results = []string{}
	}
	return json.Marshal((*Alias)(resp))
}

func Abort(ctx *gin.Context, code int, err, respErr error) {
	log.Error().
		Err(err).
		Str("method", ctx.Request.Method).
		Str("path", ctx.FullPath()).
		Msg("")

	if respErr == nil {
		respErr = err
	}

	ctx.AbortWithStatusJSON(code, gin.H{"error": respErr.Error()})
}

func HasActiveSession(ctx *gin.Context) bool {
	return ctx.GetString(UserIDKey) != "0"
}

func GetUserID(ctx *gin.Context, abort bool) (uint64, error) {
	uidVal, exists := ctx.Get(UserIDKey)
	if !exists {
		err := errors.New("missing user id in context")
		if abort {
			Abort(ctx, http.StatusBadRequest, err, nil)
		}
		return 0, err
	}

	userID, ok := uidVal.(uint64)
	if !ok {
		err := errors.New("invalid user id type")
		if abort {
			Abort(ctx, http.StatusBadRequest, err, nil)
		}
		return 0, err
	}

	return userID, nil
}

func GetStringUserID(ctx *gin.Context, abort bool) (string, error) {
	userID := ctx.GetString(UserIDKey)
	if userID == "" {
		err := errors.New("invalid user id")
		log.Debug().Err(err).Msg("invalid user id")
		if abort {
			Abort(ctx, http.StatusBadRequest, err, nil)
		}
		return "", err
	}
	return userID, nil
}

func GetInt64Param(ctx *gin.Context, k string, abort bool) (int64, error) {
	v, err := strconv.ParseInt(ctx.Param(k), 10, 64)
	if err != nil && abort {
		Abort(ctx, http.StatusBadRequest, errors.New("invalid numeric param"), nil)
	}
	return v, err
}

func GetIntParam(ctx *gin.Context, k string, abort bool) (int, error) {
	v, err := strconv.Atoi(ctx.Param(k))
	if err != nil && abort {
		Abort(ctx, http.StatusBadRequest, errors.New("invalid numeric param"), nil)
	}
	return v, err
}

func GetUint64(ctx *gin.Context, k string) (uint64, error) {
	s := ctx.GetString(k)
	if s == "" {
		return 0, errors.New("missing uint64 value in context")
	}
	return strconv.ParseUint(s, 10, 64)
}

func GetUint64Param(ctx *gin.Context, k string, abort bool) (uint64, error) {
	v, err := strconv.ParseUint(ctx.Param(k), 10, 64)
	if err != nil && abort {
		Abort(ctx, http.StatusBadRequest, errors.New("invalid numeric param"), nil)
	}
	return v, err
}

func GetUint64Query(ctx *gin.Context, k string, abort bool) (uint64, error) {
	if ctx.Query(k) == "" {
		return 0, nil
	}
	v, err := strconv.ParseUint(ctx.Query(k), 10, 64)
	if err != nil && abort {
		Abort(ctx, http.StatusBadRequest, errors.New("invalid numeric query"), nil)
	}
	return v, err
}

func GetBoolQuery(ctx *gin.Context, k string, abort bool) (bool, error) {
	if ctx.Query(k) == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(ctx.Query(k))
	if err != nil && abort {
		Abort(ctx, http.StatusBadRequest, errors.New("invalid bool query"), nil)
	}
	return v, err
}

func GetStringParam(ctx *gin.Context, k string, abort bool) (string, error) {
	v := ctx.Param(k)
	if v == "" {
		err := errors.New("invalid string param")
		if abort {
			Abort(ctx, http.StatusBadRequest, err, nil)
		}
		return "", err
	}
	return v, nil
}

func trackUser(userID, ip, ua string) error {
	return db.Service.Commit(nil, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			context.Background(),
			`UPDATE users
			SET 
				last_seen = $2,
				ip_address = $3,
				user_agent = $4
			WHERE id = $1`,
			userID,
			time.Now(),
			zeronull.Text(ip),
			zeronull.Text(ua),
		)
		return err
	})
}

func SanitizeXFFHeader(ctx *gin.Context) {
	ips := strings.Split(ctx.Request.Header.Get("X-Forwarded-For"), ",")
	if len(ips) > 3 {
		ips = ips[len(ips)-3:]
		ctx.Request.Header.Set("X-Forwarded-For", strings.Join(ips, ","))
	}
}

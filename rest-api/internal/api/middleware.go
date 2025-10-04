package api

import (
	"errors"
	"net/http"
	"rest-api/internal/db"
	"rest-api/internal/jwt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	UserIDKey = "userId"
	RoleKey   = "role"
	TokenKey  = "token"
)

func SetupParamsFactory(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		Abort(ctx, http.StatusUnauthorized, errors.New("missing Authorization token"), nil)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		Abort(ctx, http.StatusUnauthorized, errors.New("invalid authorization header format"), nil)
		return
	}
	tokenString := parts[1]

	claims, err := jwt.ParseToken(tokenString)
	if err != nil {
		Abort(ctx, http.StatusUnauthorized, errors.New("invalid token"), nil)
		return
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		Abort(ctx, http.StatusUnauthorized, errors.New("token expired"), nil)
		return
	}

	if claims.UserID == 0 {
		Abort(ctx, http.StatusUnauthorized, errors.New("invalid token claims"), nil)
		return
	}

	uid := claims.UserID
	var validUser bool
	err = db.Service.QueryRow(
		`SELECT EXISTS (
			SELECT 1 
			FROM users 
			WHERE id = $1
		)`,
		uid,
	).Scan(&validUser)
	if err != nil {
		Abort(ctx, http.StatusInternalServerError, errors.New("database query error"), nil)
		return
	}

	if !validUser {
		Abort(ctx, http.StatusUnauthorized, errors.New("invalid user"), nil)
		return
	}

	go func(userID, ip, ua string) {
		if err := trackUser(userID, ip, ua); err != nil {
			log.Error().
				Err(err).
				Str("user-id", userID).
				Str("ip", ip).
				Str("ua", ua).
				Msg("failed to track user")
		}
	}(strconv.FormatUint(uid, 10), ctx.ClientIP(), ctx.Request.UserAgent())

	ctx.Set(UserIDKey, uid)
	ctx.Set(TokenKey, tokenString)

	ctx.Next()
}

func MustHaveRoles(roles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetUint64(UserIDKey)
		if userID == 0 {
			Abort(ctx, http.StatusForbidden, ErrInsufficientRoles, nil)
			return
		}

		var ok bool
		if err := db.Service.QueryRow(
			`SELECT EXISTS(
				SELECT 1
				FROM users
				WHERE id = $1 AND role = ANY($2)
			)`,
			userID,
			roles,
		).Scan(&ok); err != nil || !ok {
			if err != nil {
				log.Error().Err(err).Msg("roles check failed")
			}
			Abort(ctx, http.StatusForbidden, ErrInsufficientRoles, nil)
			return
		}

		ctx.Next()
	}
}

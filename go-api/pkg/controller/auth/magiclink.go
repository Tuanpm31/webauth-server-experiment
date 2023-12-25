package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/dwarvesf/go-api/pkg/model"
	"github.com/dwarvesf/go-api/pkg/repository/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

func (c impl) CreateMagicLink(ctx context.Context, email string) error {
	const spanName = "CreateMagicLinkController"
	ctx, span := c.monitor.Start(ctx, spanName)

	defer span.End()

	if email == "" {
		return model.ErrMissingEmail
	}

	dbCtx := db.FromContext(ctx)
	user, err := c.repo.User.GetByEmail(dbCtx, email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.ErrInvalidCredentials
		}
		return errors.WithStack(err)
	}

	now := time.Now()
	// Generate magic link
	token, err := c.jwtHelper.GenerateJWTToken(map[string]interface{}{
		"sub":  user.ID,
		"iss":  c.cfg.App,
		"role": "user",
		"exp":  jwt.NewNumericDate(now.AddDate(1, 0, 0)),
		"nbf":  jwt.NewNumericDate(now),
		"iat":  jwt.NewNumericDate(now),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	// Send magic link to user
	fmt.Println(token)

	return nil
}

func (c impl) VerifyMagicLink(ctx context.Context, secret string) (string, error) {
	const spanName = "VerifyMagicLinkController"
	ctx, span := c.monitor.Start(ctx, spanName)

	defer span.End()

	if secret == "" {
		return "", model.ErrNotFound
	}

	// Verify magic link
	_, err := c.jwtHelper.ValidateToken(secret)
	if err != nil {
		return "", model.ErrInvalidToken
	}

	return secret, nil
}
package server

import (
	"context"
	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

const (
	CapAdmin = "admin"
	CapRead  = "read"
	CapWrite = "write"
)

type JwtClaims struct {
	UserID       string   `json:"user_id,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	jwt.StandardClaims
}

func AuthCheck(ctx context.Context, userId string, capabilities ...string) error {
	switch p := ctx.Value("jwt").(type) {
	case error:
		return p
	case JwtClaims:
		if p.ExpiresAt < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "Token expired.")
		}
		if p.Capabilities != nil {
			for pc := range p.Capabilities {
				for c := range capabilities {
					if pc == c {
						return nil
					}
				}
			}
		}
		if p.UserID == userId || userId == "*" {
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "Cannot access this api because of token do not have permission.")
	default:
		return status.Errorf(codes.Internal, "Internal server error.")
	}
}

func GenerateJWT(userId string, capabilities []string) (string, error) {
	// Create custom claims
	claims := JwtClaims{
		UserID:       userId,
		Capabilities: capabilities,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(SConfig.Auth.Expire)).Unix(), // Token will expire in 1 hour
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(SConfig.Auth.SecretKey))
	if err != nil {
		return "", err
	}

	return "Bearer " + tokenString, nil
}

func ParseJWTMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		ctx = context.WithValue(ctx, "jwt", status.Errorf(codes.Unauthenticated, "Metadata is not provided"))
		return handler(ctx, req)
	}

	token := md.Get("Authorization")
	if len(token) == 0 {
		ctx = context.WithValue(ctx, "jwt", status.Errorf(codes.Unauthenticated, "Authorization token is not provided"))
		return handler(ctx, req)
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token[0], "Bearer ") {
		token[0] = strings.TrimPrefix(token[0], "Bearer ")
	}

	claims := JwtClaims{}

	// Parse JWT
	parsedToken, err := jwt.ParseWithClaims(token[0], &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SConfig.Auth.SecretKey), nil
	})

	if err != nil {
		ctx = context.WithValue(ctx, "jwt", status.Errorf(codes.Unauthenticated, "JWT parsing error: %v", err))
		return handler(ctx, req)
	}

	// Type-assert to *jwt.Token
	if parsedToken.Valid {
		ctx = context.WithValue(ctx, "jwt", claims)
	} else {
		ctx = context.WithValue(ctx, "jwt", status.Errorf(codes.Unauthenticated, "Failed to assert token claims"))
	}

	return handler(ctx, req)
}

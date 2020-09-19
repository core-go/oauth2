package oauth2

import (
	"context"
	"github.com/common-go/auth"
)

type OAuth2IntegrationService interface {
	GetIntegrationConfiguration(ctx context.Context, sourceType string) (*IntegrationConfiguration, error)
	Authenticate(ctx context.Context, auth OAuth2Info, authorization string) (auth.AuthResult, error)
}

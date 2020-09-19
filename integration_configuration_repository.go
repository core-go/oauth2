package oauth2

import "context"

type IntegrationConfigurationRepository interface {
	GetIntegrationConfiguration(ctx context.Context, sourceType string) (*IntegrationConfiguration, string, error)
}

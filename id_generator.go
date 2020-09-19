package oauth2

import "context"

type IdGenerator interface {
	Generate(ctx context.Context) (string, error)
}

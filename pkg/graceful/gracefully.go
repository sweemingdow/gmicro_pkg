package graceful

import "context"

type Gracefully interface {
	GracefulStop(ctx context.Context) error
}

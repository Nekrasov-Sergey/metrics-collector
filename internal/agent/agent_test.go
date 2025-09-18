package agent

import (
	"context"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 11*time.Second)
	defer cancel()
	Run(ctx)
}

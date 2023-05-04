package limiter

import (
	"context"
	"testing"
)

func TestContextValue(t *testing.T) {
	priority, _ := context.Background().Value("key").(int)
	t.Fatalf("%d", priority)
}

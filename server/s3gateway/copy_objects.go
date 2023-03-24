package s3gateway

import (
	"context"
	"fmt"
)

func (svr *S3Server) CopyObjects(ctx context.Context, from, to string, recursive bool) error {
	return fmt.Errorf("not implemented")
}

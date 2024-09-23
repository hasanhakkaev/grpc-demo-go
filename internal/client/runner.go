package client

import (
	"context"
)

// Run runs the given server.
func Run(s *Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.logger.Debug("Running consumer...")
	if err := s.Run(ctx); err != nil {
		return err
	}
	return s.Shutdown(ctx)
}

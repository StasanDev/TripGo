package container

import "context"

func (a *App) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		err := a.server.Shutdown(ctx)
		a.pool.Close()
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

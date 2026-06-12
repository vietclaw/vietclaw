package agent

import "context"

type spawnNotifyKey struct{}

func withSpawnNotifier(ctx context.Context, notify spawnNotifier) context.Context {
	if notify == nil {
		return ctx
	}
	return context.WithValue(ctx, spawnNotifyKey{}, notify)
}

func spawnNotifierFromContext(ctx context.Context) spawnNotifier {
	if ctx == nil {
		return nil
	}
	if notify, ok := ctx.Value(spawnNotifyKey{}).(spawnNotifier); ok {
		return notify
	}
	return nil
}

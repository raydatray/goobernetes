package utils

type Option[T any] func(*T) error

func WithOptions[T any](obj *T, opts ...Option[T]) error {
	for _, opt := range opts {
		if err := opt(obj); err != nil {
			return err
		}
	}
	return nil
}

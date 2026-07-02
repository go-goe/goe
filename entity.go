package goe

import "context"

type Entity[T any] struct {
	entity *T
}

func (e Entity[T]) List() stateSelect[T] {
	return e.ListContext(context.Background())
}

func (e Entity[T]) ListContext(ctx context.Context) stateSelect[T] {
	return ListContext(ctx, e.entity)
}

func (e Entity[T]) Find() find[T] {
	return e.FindContext(context.Background())
}

func (e Entity[T]) FindContext(ctx context.Context) find[T] {
	return FindContext(ctx, e.entity)
}

func (e Entity[T]) Save() save[T] {
	return e.SaveContext(context.Background())
}

func (e Entity[T]) SaveContext(ctx context.Context) save[T] {
	return SaveContext(ctx, e.entity)
}

func (e Entity[T]) Create() stateInsert[T] {
	return e.CreateContext(context.Background())
}

func (e Entity[T]) CreateContext(ctx context.Context) stateInsert[T] {
	return InsertContext(ctx, e.entity)
}

func (e Entity[T]) Remove() remove[T] {
	return e.RemoveContext(context.Background())
}

func (e Entity[T]) RemoveContext(ctx context.Context) remove[T] {
	return RemoveContext(ctx, e.entity)
}

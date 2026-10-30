package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

type task[T any] struct {
	index int
	value T
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var zero_val_R_slice []Result[R]
	var zero_val_R R

	if len(in) == 0 {
		return zero_val_R_slice, nil
	}

	workers_count := min(workers, len(in))

	//массив под все задачи результата
	res := make([]Result[R], len(in))
	//для постепенной отдачи воркером тасок
	tasks := make(chan task[T], len(in))

	//контроль горутин
	var wg sync.WaitGroup

	//работа воркеров
	for i := 0; i < workers_count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range tasks {
				//проверка что контекст отменен
				if ctx.Err() != nil {
					break
				}
				res_fn, err := fn(ctx, task.value)
				var struct_res Result[R]

				if err == nil {
					struct_res = Result[R]{
						Value: res_fn,
						Err:   nil,
					}

				} else {
					struct_res = Result[R]{
						Value: zero_val_R,
						Err:   err,
					}
				}
				res[task.index] = struct_res
			}
		}()
	}
Loop:
	for i, value := range in {

		select {
		case <-ctx.Done():
			break Loop
		//добавляем задачи
		case tasks <- task[T]{
			index: i,
			value: value,
		}:

		}
	}

	close(tasks)
	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return res, nil

}

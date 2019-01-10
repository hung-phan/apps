package parallel

import "sync"

type Val interface{}

func Parallel(fns ...func() Val) []Val {
	var (
		wg  = sync.WaitGroup{}
		m   = sync.Mutex{}
		res = make([]Val, len(fns))
	)

	for index, fn := range fns {
		wg.Add(1)

		go func(index int, fn func() Val) {
			defer wg.Done()

			result := fn()

			m.Lock()
			res[index] = result
			m.Unlock()
		}(index, fn)
	}

	wg.Wait()

	return res
}

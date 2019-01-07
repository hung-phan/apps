package parallel

import "sync"

type Result interface{}

func Parallel(fns ...func() Result) []Result {
	var (
		wg  = sync.WaitGroup{}
		m   = sync.Mutex{}
		res = make([]Result, len(fns))
	)

	for index, fn := range fns {
		wg.Add(1)

		go func(index int, fn func() Result) {
			defer wg.Done()
			defer m.Unlock()

			m.Lock()

			res[index] = fn()
		}(index, fn)
	}

	wg.Wait()

	return res
}

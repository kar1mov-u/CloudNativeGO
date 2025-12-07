package circuitbreaker

// // this is small implementation from the book
// type Circuit func(context.Context) (string, error)

// // breaker is the parent fucntion that accepts the circuit, and returns the clousre, because of the closure multiple functions can share the state
// func Breaker(circuit Circuit, treshhold int) Circuit {
// 	var failures int
// 	var last = time.Now()
// 	var m sync.RWMutex

// 	return func(ctx context.Context) (string, error) {
// 		//checking the state
// 		m.RLock()
// 		d := failures - treshhold
// 		if d >= 0 {
// 			shouldRetryAt := last.Add((2 << d) * time.Second)

// 			if !time.Now().After(shouldRetryAt) {
// 				m.RUnlock()
// 				return "", errors.New("service unavaible")
// 			}
// 		}

// 		m.RUnlock()

// 		response, err := circuit(ctx)

// 		m.Lock()
// 		defer m.Unlock()

// 		last = time.Now()
// 		if err != nil {
// 			failures++
// 			return response, err
// 		}
// 		failures = 0

// 		return response, nil
// 	}
// }

func Counter() func() int {
	x := 0
	return func() int {
		x++
		return x
	}
}

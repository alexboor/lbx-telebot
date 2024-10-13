package scheduler

import (
	"fmt"
	"time"
)

// PingExample just writes string to console
//
//	This is an example
//	The function receives one single parameter from outside, see job creation line
func (s *Schedule) PingExample(now time.Time) {
	fmt.Println("ping", now.String())
}

// InMemoryStorageExample is an example of using in-memory storage
//
//	This is an example of the function which uses in-memory storage package
//	to store and retrieve a counter value. No external passed parameters.
//	If active it will increment the counter value each time it is called and print it.
func (s *Schedule) InMemoryStorageExample() {
	var counter int

	v, exist := s.Memory.Get("shc_example_counter")
	if exist {
		counter = v.(int)
	} else {
		counter = 0
	}

	fmt.Println("counter: ", counter)

	s.Memory.Set("shc_example_counter", counter+1)
}

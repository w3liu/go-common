package test

import (
	"fmt"
	"github.com/w3liu/go-common/number/ordernum"
	"sync"
	"testing"
	"time"
)

func TestGenerateOrderNum(t *testing.T) {
	mu := sync.Mutex{}
	dic := make(map[string]int32)
	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			num := ordernum.Generate(time.Now())
			mu.Lock()
			if _, ok := dic[num]; ok {
				fmt.Println("error", num)
			}
			dic[num] = int32(i)
			fmt.Println(num)
			mu.Unlock()
			wg.Done()

		}(i)
	}
	wg.Wait()
	fmt.Println("len:", len(dic))
}

package main

func Generator(nums []int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func Square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func Filter(in <-chan int, threshold int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			if n > threshold {
				out <- n
			}
		}
		close(out)
	}()
	return out
}

func Pipeline(nums []int) []int {
	stage1 := Generator(nums)
	stage2 := Square(stage1)
	stage3 := Filter(stage2, 10)
	
	var results []int
	for n := range stage3 {
		results = append(results, n)
	}
	return results
}

func main() {}

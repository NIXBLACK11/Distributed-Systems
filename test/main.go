package main

import (
	"fmt"
	"sync"
	"time"
)

const NUMJOBS = 100
const NUMWORKERS = 5

func progress() {
	x := 0

	for i := range(10000000) {
		x += i
	}

	_ = x
}

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	for job := range(jobs) {
		fmt.Printf("Worker %d working on job %d\n", id, job)
		time.Sleep(500 * time.Millisecond)
		progress()
		wg.Done()
	}
}

func main() {
	jobs := make(chan int, NUMJOBS)

	var wg sync.WaitGroup

	wg.Add(NUMJOBS)

	for workerId := range(NUMWORKERS) {
		go worker(workerId, jobs, &wg)
	}

	for jobId := range(NUMJOBS) {
		jobs <- jobId
	}

	close(jobs)

	wg.Wait()
	fmt.Println("All jobs done!")
}

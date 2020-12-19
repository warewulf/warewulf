package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/batch"
)

func exampleJob(name string, id int) string {
	fmt.Printf("[%d] %s started...\n", id, name)
	time.Sleep(time.Second * time.Duration(id))
	fmt.Printf("[%d] %s ending...\n", id, name)
	return name
}

func main() {

	fmt.Printf("GOMAXPROCS=%d\n", runtime.GOMAXPROCS(0))

	jobstorun := 100

	// create batch pool to run 100 jobs at a time every 10 seconds
	batchpool := batch.New(10)
	retvalues := make(chan string, jobstorun)

	// submit a bunch of jobs
	for i := 0; i < jobstorun; i++ {

		name := fmt.Sprintf("job-%04d", i)
		id := i
		batchpool.Submit(func() {
			retvalues <- exampleJob(name, id)
		})

	}

	// Run all batch jobs to completion
	batchpool.Run()

	close(retvalues)

	for s := range retvalues {
		fmt.Printf("Return value is %s\n", s)
	}

}

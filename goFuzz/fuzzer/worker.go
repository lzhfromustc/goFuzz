package fuzzer

import (
	"log"
	"sync"
	"time"
)

// InitWorkers starts maxParallel workers working on inputCh from fuzzer context.
func InitWorkers(maxParallel int, fuzzCtx *FuzzContext) {
	go func() {
		var wg sync.WaitGroup

		for i := 0; i < maxParallel; i++ {
			wg.Add(1)

			// Start worker
			go func(i int) {
				log.Printf("[Worker %d] Started", i)
				defer wg.Done()
				for {
					select {
					// Receive input
					case task := <-fuzzCtx.runTaskCh:
						log.Printf("[Worker %d] Working on %s\n", i, task.id)
						if ShouldSkipRunTask(fuzzCtx, task) {
							log.Printf("[Worker %d][Task %s] skipped\n", i, task.id)
							continue
						}
						result, err := Run(fuzzCtx, task)
						if err != nil {
							log.Printf("[Worker %d] [Task %s] Error: %s\n", i, task.id, err)
							continue
						}
						err = HandleRunResult(task, result, fuzzCtx)
						if err != nil {
							log.Printf("[Worker %d] [Task %s] Error: %s\n", i, task.id, err)
							continue
						}
					case <-time.After(60 * time.Second):
						log.Printf("[Worker %d] Timeout. Exiting...", i)
						return
					}
				}

			}(i)
		}

		wg.Wait()
	}()

}

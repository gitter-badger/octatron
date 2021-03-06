/************************************************************************/
/* Octatron                                                             */
/* Copyright (c) 2015 Andreas T Jonsson <mail@andreasjonsson.se>        */
/*                                                                      */
/* Octatron is free software: you can redistribute it and/or modify     */
/* it under the terms of the GNU General Public License as published by */
/* the Free Software Foundation, either version 3 of the License, or    */
/* (at your option) any later version.                                  */
/*                                                                      */
/* Octatron is distributed in the hope that it will be useful,          */
/* but WITHOUT ANY WARRANTY; without even the implied warranty of       */
/* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the        */
/* GNU General Public License for more details.                         */
/*                                                                      */
/* You should have received a copy of the GNU General Public License    */
/* along with Octatron.  If not, see <http://www.gnu.org/licenses/>.    */
/************************************************************************/

package octatron

import (
	"io"
	"sync"
	"sync/atomic"
)

type TreeResult struct {
}

type TreeConfig struct {
	Writer        io.WriteSeeker
	Bounds        Box
	VoxelsPerAxis int
}

type workerPrivateData struct {
	err error
	worker Worker
}

func processSample(data *workerPrivateData, sample *Sample) {

}

func processData(data *workerPrivateData, sampleChan <-chan Sample) error {
	for {
		sample, more := <-sampleChan
		processSample(data, &sample)
		if !more {
			err := data.err
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func collectData(workerData *workerPrivateData, node *treeNode, sampleChan chan<- Sample) {
	err := workerData.worker.Run(node.bounds, sampleChan)
	if err != nil {
		workerData.err = err
	}
	close(sampleChan)
}

func BuildTree(workers []Worker, cfg *TreeConfig) (*TreeResult, error) {
	var volumeTraversed uint64
	vpa := uint64(cfg.VoxelsPerAxis)
	totalVolume := vpa * vpa * vpa

	numWorkers := len(workers)
	workerData := make([]workerPrivateData, numWorkers)

	var wgWorkersIdle sync.WaitGroup
	wgWorkersIdle.Add(numWorkers)

	nodeMapShutdownChan, nodeMapInChan, nodeMapOutChan := startNodeCache(numWorkers)
	nodeMapInChan <- newRootNode(cfg.Bounds, cfg.VoxelsPerAxis)

	defer func() {
		nodeMapShutdownChan <- struct{}{}
	}()

	var wgWorkers sync.WaitGroup
	wgWorkers.Add(numWorkers)

	for idx, worker := range workers {
		// Spawn worker
		go func() {
			defer wgWorkers.Done()

			data := &workerData[idx]
			data.worker = worker

			// Process jobs
			for {
				node, more := <-nodeMapOutChan
				if !more {
					return
				}

				sampleChan := make(chan Sample, 10)
				go collectData(data, node, sampleChan)
				if processData(data, sampleChan) != nil {
					return
				}

				// This is a leaf
				if node.numSamples == 0 {
					parent := node.parent
					if parent != nil {
						parent.children[node.childIndex] = nil
					}

					vpa := uint64(node.voxelsPerAxis)
					volume := vpa * vpa * vpa
					newVolume := atomic.AddUint64(&volumeTraversed, volume)

					// Are we done with the octree
					if newVolume == totalVolume {
						//nodeMapShutdownChan <- struct{}{}
					}
				}
			}
		}()
	}

	wgWorkers.Wait()
	for _, data := range workerData {
		if data.err != nil {
			return nil, data.err
		}
	}
	return &TreeResult{}, nil
}

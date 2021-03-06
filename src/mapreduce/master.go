package mapreduce

import "container/list"
import "fmt"

type WorkerInfo struct {
	address string
	// You can add definitions here.
}

// Clean up all workers by sending a Shutdown RPC to each one of them Collect
// the number of jobs each work has performed.
func (mr *MapReduce) KillWorkers() *list.List {
	l := list.New()
	for _, w := range mr.Workers {
		DPrintf("DoWork: shutdown %s\n", w.address)
		args := &ShutdownArgs{}
		var reply ShutdownReply
		ok := call(w.address, "Worker.Shutdown", args, &reply)
		if ok == false {
			fmt.Printf("DoWork: RPC %s shutdown error\n", w.address)
		} else {
			l.PushBack(reply.Njobs)
		}
	}
	return l
}

func (mr *MapReduce) RunMaster() *list.List {
	// Your code here
	go mr.registerWorkers()

	for i := 0; i < mr.nMap; i++ {
		go mr.delegateJobs(Map, i)
	}

	for i := 0; i < mr.nMap; i++ {
		<-mr.donechannel
	}

	for i := 0; i < mr.nReduce; i++ {
		go mr.delegateJobs(Reduce, i)
	}

	for i := 0; i < mr.nReduce; i++ {
		<-mr.donechannel
	}

	close(mr.cancelchannel)

	return mr.KillWorkers()
}

func (mr *MapReduce) registerWorkers() {
	for {
		address := <-mr.registerChannel
		mr.Workers[address] = &WorkerInfo{address}
		mr.readychannel <- address
	}
}

func (mr *MapReduce) delegateJobs(jobType JobType, jobNo int){
	for {
		worker := <-mr.readychannel
		if ok := mr.assignJob(worker, jobType, jobNo); ok {
			mr.donechannel <- true
			select {
				case mr.readychannel <- worker:
				case <-mr.cancelchannel:
			}
			return
		}
	}
}

func (mr *MapReduce) assignJob(worker string, jobType JobType, jobNo int) bool{
	var args DoJobArgs
	switch jobType {
	case Map:
		args = DoJobArgs{mr.file, Map, jobNo, mr.nReduce}
	case Reduce:
		args = DoJobArgs{mr.file, Reduce, jobNo, mr.nMap}
	}
	var reply DoJobReply
	return call(worker, "Worker.DoJob", args, &reply)
}
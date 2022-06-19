package task

type Task struct {
	Status     int   // task status
	Priority   int   // task priority
	UpdateTime int64 // last update time stamp
}

func (t Task) Len() int {
	//TODO implement me
	panic("implement me")
}

func (t Task) Less(i, j int) bool {
	//TODO implement me
	panic("implement me")
}

func (t Task) Swap(i, j int) {
	//TODO implement me
	panic("implement me")
}

func (t Task) Push(x interface{}) {
	//TODO implement me
	panic("implement me")
}

func (t Task) Pop() interface{} {
	//TODO implement me
	panic("implement me")
}

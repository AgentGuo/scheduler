package task

const (
	TaskNameLogKey = "task"

	// task type
	KubeResourceTaskType string = "KubeResourceTask"
	ResourceTaskType     string = "ResourceTask"
	NormalTaskType       string = "NormalTask"

	// task status
	PENDING string = "pending"
	RUNNING string = "running"
	DELETED string = "deleted"
)

type Task struct {
	Name           string            `json:"Name"` // task name
	Labels         map[string]string // selector labels
	Status         string            `json:"Status"`         // task status
	Priority       int               `json:"Priority"`       // task priority
	UpdateTime     int64             `json:"UpdateTime"`     // last update time stamp
	TaskType       string            `json:"TaskType"`       // task type
	NodeName       string            `json:"NodeName"`       // run on
	Detail         interface{}       `json:"Detail"`         // detailed information
	ResourceDetail interface{}       `json:"ResourceDetail"` // resource information
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

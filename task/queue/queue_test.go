package queue

import (
	"container/heap"
	"github.com/AgentGuo/scheduler/task"
	"reflect"
	"testing"
)

func TestTaskQueue_Len(t *testing.T) {
	tests := []struct {
		name string
		t    TaskQueue
		want int
	}{
		{"test#1", TaskQueue{}, 0},
		{"test#2", TaskQueue{task.Task{}, task.Task{}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskQueue_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		t    TaskQueue
		args args
		want bool
	}{
		{"test#1", TaskQueue{task.Task{Priority: 1}, task.Task{Priority: 2}}, args{0, 1}, true},
		{"test#2", TaskQueue{task.Task{Priority: 1}, task.Task{Priority: 2}}, args{1, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskQueue_Pop(t *testing.T) {
	tests := []struct {
		name string
		t    TaskQueue
		want interface{}
	}{
		{"test#1", TaskQueue{}, nil},
		{"test#2", TaskQueue{task.Task{}}, &task.Task{}},
		{"test#3", TaskQueue{task.Task{Priority: 2}}, &task.Task{Priority: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskQueue_Push(t *testing.T) {
	type args struct {
		x interface{}
	}
	tests := []struct {
		name    string
		t       TaskQueue
		args    args
		wantLen int
	}{
		{"test#1", TaskQueue{}, args{task.Task{}}, 1},
		{"test#2", TaskQueue{task.Task{}}, args{task.Task{}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.t.Push(tt.args.x); tt.t.Len() != tt.wantLen {
				t.Errorf("After push, Len() = %v, wantLen %v", tt.t.Len(), tt.wantLen)
			}
		})
	}
}

func TestTaskQueue_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		t    TaskQueue
		args args
		want TaskQueue
	}{
		{"test#1", TaskQueue{task.Task{Priority: 1}, task.Task{Priority: 2}}, args{0, 1},
			TaskQueue{task.Task{Priority: 2}, task.Task{Priority: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.t.Swap(tt.args.i, tt.args.j); !reflect.DeepEqual(tt.t, tt.want) {
				t.Errorf("Pop() = %v, want %v", tt.args, tt.want)
			}
		})
	}
}

func TestTaskQueue(t *testing.T) {
	tests := []struct {
		name string
		t    TaskQueue
		want *task.Task
	}{
		{"test#1", TaskQueue{task.Task{Priority: 1}, task.Task{Priority: 2}},
			&task.Task{Priority: 2}},
		{"test#2", TaskQueue{task.Task{Priority: 2}, task.Task{Priority: 1}},
			&task.Task{Priority: 2}},
		{"test#3", TaskQueue{task.Task{Priority: 1}, task.Task{Priority: 2}, task.Task{Priority: 3}},
			&task.Task{Priority: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heap.Init(&tt.t)
			if got := tt.t.Pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

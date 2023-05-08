package mycontext

import "time"

type Context interface {
	// Deadline 返回绑定当前context的任务被取消的截止时间；如果没有设定期限，将返回ok == false。
	Deadline() (deadline time.Time, ok bool)
	//Done 当绑定当前context的任务被取消时，将返回一个关闭的channel；如果当前context不会被取消，将返回nil。
	Done() <-chan struct{}
	//Err 如果Done返回的channel没有关闭，将返回nil;如果Done返回的channel已经关闭，将返回非空的值表示任务结束的原因。
	//如果是context被取消，Err将返回Canceled；如果是context超时，Err将返回DeadlineExceeded。
	Err() error
	//Value 返回context存储的键值对中当前key对应的值，如果没有对应的key,则返回nil。
	Value(key interface{}) interface{}
}

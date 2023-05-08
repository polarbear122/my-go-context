package mycontext

import (
	"errors"
	"fmt"
	"sync"
)

// CanceledError 如果context已经被cancel了，就返回该错误
var CanceledError = errors.New("context canceled")

// CancelFunc 告诉operation放弃工作，不会等待工作停止
// 在第一次调用之后，对CancelFunc的后续调用将不起任何作用。
type CancelFunc func()

// WithCancel 返回带有新的Done通道的parent的副本。
// 调用返回的cancel函数或关闭parent context的Done通道时（以先发生的为准），返回的上下文的Done通道将关闭。
// 取消此context会释放与其关联的资源，因此代码应该在该context中运行的操作完成后立即调用cancel。
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := newCancelCtx(parent)
	propagateCancel(parent, &c)
	return &c, func() { c.cancel(true, CanceledError) }
}

// newCancelCtx returns an initialized cancelCtx.
func newCancelCtx(parent Context) cancelCtx {
	return cancelCtx{
		Context: parent,
		done:    make(chan struct{}),
	}
}

// 一个canceler是一个可以直接取消的context类型。其实现是*cancelCtx和*timerCtx。
type canceler interface {
	cancel(removeFromParent bool, err error)
	Done() <-chan struct{}
}

// 一个cancelCtx可以被取消。当被取消时，它也会取消任何实现了canceller的child
type cancelCtx struct {
	Context

	done chan struct{} // closed by the first cancel call.

	mu       sync.Mutex
	children map[canceler]bool // 被第一个cancel call设置为nil
	err      error             // 被第一次cancel call设置为not-nil
}

func (c *cancelCtx) Done() <-chan struct{} {
	return c.done
}

func (c *cancelCtx) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *cancelCtx) String() string {
	return fmt.Sprintf("%v.WithCancel", c.Context)
}

// cancel 关闭c.done，取消c的每一个子节点，并且如果removeFromParent为真，则将c从其父的children中移除。
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err
	close(c.done)
	for child := range c.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		child.cancel(false, err)
	}
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		removeChild(c.Context, c)
	}
}

// propagateCancel 当parent需要取消的时候，取消child
func propagateCancel(parent Context, child canceler) {
	if parent.Done() == nil {
		return // parent is never canceled
	}
	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		if p.err != nil {
			// parent has already been canceled
			child.cancel(false, p.err)
		} else {
			if p.children == nil {
				p.children = make(map[canceler]bool)
			}
			p.children[child] = true
		}
		p.mu.Unlock()
	} else {
		go func() {
			select {
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}

// parentCancelCtx 跟随一个父引用链，直到找到一个*cancelCtx
func parentCancelCtx(parent Context) (*cancelCtx, bool) {
	for {
		switch c := parent.(type) {
		case *cancelCtx:
			return c, true
		default:
			return nil, false
		}
	}
}

// removeChild 将一个context从其父代中移除。
func removeChild(parent Context, child canceler) {
	p, ok := parentCancelCtx(parent)
	if !ok {
		return
	}
	p.mu.Lock()
	if p.children != nil {
		delete(p.children, child)
	}
	p.mu.Unlock()
}

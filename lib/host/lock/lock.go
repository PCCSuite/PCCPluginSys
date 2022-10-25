package lock

import (
	"errors"
	"sort"
	"sync"

	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

const DefaultName = "windows_installer"

var ErrCancelled = errors.New("unlocked before got lock")

var globalMutex = sync.RWMutex{}

var locks = map[string]*Lock{}

type Lock struct {
	mutex    *sync.RWMutex
	have     *requester
	requests []*requester
}

// first, wait for Ch closing
// second, check Err
// if Err == nil, you have lock
// if Err != nil, you failed to get lock
type requester struct {
	action *data.RunningAction
	ch     chan struct{}
	Ch     <-chan struct{}
	// only available after Ch closed
	Err error
}

// first, wait for Ch closing
// second, check Err
// if Err == nil, you have lock
// if Err != nil, you failed to get lock
func RequestLock(name string, action *data.RunningAction) *requester {
	globalMutex.RLock()
	lock, ok := locks[name]
	globalMutex.RUnlock()
	if !ok {
		globalMutex.Lock()
		lock = &Lock{
			mutex:    &sync.RWMutex{},
			have:     nil,
			requests: []*requester{},
		}
		locks[name] = lock
		globalMutex.Unlock()
	}
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	for _, v := range lock.requests {
		if v.action == action {
			return v
		}
	}
	ch := make(chan struct{})
	requester := requester{
		action: action,
		ch:     ch,
		Ch:     ch,
		Err:    nil,
	}
	lock.requests = append(lock.requests, &requester)
	go CheckLock(name)
	return &requester
}

func Unlock(name string, action *data.RunningAction) {
	globalMutex.RLock()
	lock, ok := locks[name]
	globalMutex.RUnlock()
	if !ok {
		return
	}
	for i, v := range lock.requests {
		if v.action == action {
			lock.requests = append(lock.requests[:i], lock.requests[i+1:]...)
			select {
			case <-v.ch:
			default:
				v.Err = ErrCancelled
				close(v.ch)
			}
			go CheckLock(name)
			return
		}
	}
}

func UnlockAll(action *data.RunningAction) {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	for _, lock := range locks {
		lock.mutex.RLock()
		for i, v := range lock.requests {
			if v.action == action {
				lock.mutex.RUnlock()
				lock.mutex.Lock()
				lock.requests = append(lock.requests[:i], lock.requests[i+1:]...)
				lock.mutex.Unlock()
				lock.mutex.RLock()
				select {
				case <-v.ch:
				default:
					v.Err = ErrCancelled
					close(v.ch)
				}
				break
			}
		}
		lock.mutex.RUnlock()
	}
	go CheckLockAll()
}

func CheckLockAll() {
	for k := range locks {
		CheckLock(k)
	}
}

func CheckLock(name string) {
	globalMutex.RLock()
	lock, ok := locks[name]
	globalMutex.RUnlock()
	if !ok {
		return
	}
	lock.mutex.RLock()
	if lock.have != nil {
		found := false
		for _, v := range lock.requests {
			if v == lock.have {
				found = true
				break
			}
		}
		lock.mutex.RUnlock()
		if found && !lock.have.action.IsEnded() {
			return
		} else {
			lock.mutex.Lock()
			lock.have = nil
			lock.mutex.Unlock()
		}
	} else {
		lock.mutex.RUnlock()
	}
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	if len(lock.requests) == 0 {
		return
	}
	sort.SliceStable(lock.requests, func(i, j int) bool {
		return lock.requests[i].action.Priority < lock.requests[j].action.Priority
	})
	lock.have = lock.requests[0]
	close(lock.have.ch)
}

func IsLocking(name string, action *data.RunningAction) bool {
	globalMutex.RLock()
	lock, ok := locks[name]
	globalMutex.RUnlock()
	if !ok {
		return false
	}
	lock.mutex.RLock()
	defer lock.mutex.RUnlock()
	if lock.have == nil {
		return false
	}
	return lock.have.action == action
}

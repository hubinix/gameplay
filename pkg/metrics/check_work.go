package metrics

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sync/atomic"
)

func GetHandleFunctionName(message protoreflect.ProtoMessage) string {
	return string(message.ProtoReflect().Descriptor().FullName())
}

type CheckWork struct {
	s    atomic.Value
	Name string
}

func (th *CheckWork) Check() bool {
	if i := th.s.Load(); i != nil {
		return i.(bool)
	}
	return false
}
func (th *CheckWork) Work() {
	if !th.Check() {
		logrus.Infof("[%s]Work", th.Name)
		th.s.Store(true)
	}
}
func (th *CheckWork) Idle() {
	if th.Check() {
		logrus.Infof("[%s]Idle", th.Name)
		th.s.Store(false)
	}
}

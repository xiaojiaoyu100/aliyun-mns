package aliyun_mns

type alert interface {
	NotifyAsync(content ...interface{})
}

var inst alert

func SetAlert(a alert) {
	inst = a
}

func notifyAsync(content ...interface{}) {
	if inst != nil {
		inst.NotifyAsync(content...)
	}
}

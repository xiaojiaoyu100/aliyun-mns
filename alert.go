package aliyun_mns

type alert interface {
	NotifyAsync(content ...interface{})
}

var a alert

func SetAlert(a alert) {
	a = a
}

func notifyAsync(content ...interface{}) {
	if a != nil {
		a.NotifyAsync(content...)
	}
}

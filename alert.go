package alimns

type alert interface {
	NotifyAsync(content ...interface{})
}

var inst alert

// SetAlert sets alert instance.
func SetAlert(a alert) {
	inst = a
}

func notifyAsync(content ...interface{}) {
	if inst != nil {
		inst.NotifyAsync(content...)
	}
}

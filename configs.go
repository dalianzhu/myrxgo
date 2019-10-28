package myrxgo

type serialConfig struct {
}

var SerialConfig = new(serialConfig)

type concurrentConfig struct {
	value uint
}

func ConcurrentConfig(v uint) *concurrentConfig {
	c := new(concurrentConfig)
	c.value = v
	return c
}

type timeoutConfig struct {
	value uint
}

func TimeoutConfig(v uint) *timeoutConfig {
	t := new(timeoutConfig)
	t.value = v
	return t
}

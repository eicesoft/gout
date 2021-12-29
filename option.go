package gout

type Options struct {
	IsEnablePProf bool
}

type Option func(*Options)

func newOptions(opts ...Option) Options {
	opt := Options{}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

func WrapOptionPProf(enable bool) Option {
	return func(option *Options) {
		option.IsEnablePProf = enable
	}
}

package cache

import "time"

const (
	defaultMaxSize                   = 1024
	defaultCleanDuration             = 10 * time.Minute
	defaultCleanSize                 = 32
	defaultCleanFullThresholdPercent = 0.8
	defaultLoadTimeOut               = 3 * time.Second
)

type Options struct {
	cleanSize     int
	cleanDuration time.Duration

	maxSize                   int
	cleanFullThresholdPercent float64

	defaultEntryOpts EntryOptions
}

type Option func(options *Options)

func NewOptions() Options {
	return Options{
		cleanSize:                 defaultCleanSize,
		cleanDuration:             defaultCleanDuration,
		maxSize:                   defaultMaxSize,
		cleanFullThresholdPercent: defaultCleanFullThresholdPercent,
		defaultEntryOpts:          NewEntryOptions(),
	}
}

func MaxSize(v int) Option {
	return func(options *Options) {
		options.maxSize = v
	}
}

func CleanSize(v int) Option {
	return func(options *Options) {
		options.cleanSize = v
	}
}

func CleanDuration(t time.Duration) Option {
	return func(options *Options) {
		options.cleanDuration = t
	}
}

func DefaultEntryOpts(opts ...EntryOption) Option {
	return func(options *Options) {
		eOpts := options.defaultEntryOpts
		for _, eOpt := range opts {
			eOpt(&eOpts)
		}
	}
}

type EntryOptions struct {
	loader      Loader
	syncLoad    bool
	loadTimeout time.Duration

	expireAfterWrite  time.Duration
	refreshAfterWrite time.Duration
}

type EntryOption func(options *EntryOptions)

func NewEntryOptions() EntryOptions {
	return EntryOptions{
		loader:      nil,
		syncLoad:    false,
		loadTimeout: defaultLoadTimeOut,

		expireAfterWrite:  0,
		refreshAfterWrite: 0,
	}
}

func WithLoader(loader Loader) EntryOption {
	return func(options *EntryOptions) {
		options.loader = loader
	}
}

func WithLoaderTimeout(timeout time.Duration) EntryOption {
	return func(options *EntryOptions) {
		options.loadTimeout = timeout
	}
}

func SyncLoad(v bool) EntryOption {
	return func(options *EntryOptions) {
		options.syncLoad = v
	}
}

func ExpirationOption(d time.Duration) EntryOption {
	return func(options *EntryOptions) {
		options.expireAfterWrite = d
	}
}

func RefreshOption(d time.Duration) EntryOption {
	return func(options *EntryOptions) {
		options.refreshAfterWrite = d
	}
}

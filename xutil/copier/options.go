package copier

type CopyOption func(*copyOptions)

type copyOptions struct {
	ignoreFields map[string]bool
	fieldMapping map[string]string
}

func defaultCopyOptions() *copyOptions {
	return &copyOptions{
		ignoreFields: make(map[string]bool),
		fieldMapping: make(map[string]string),
	}
}

func applyCopyOptions(opts []CopyOption) *copyOptions {
	o := defaultCopyOptions()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func IgnoreFields(fields ...string) CopyOption {
	return func(o *copyOptions) {
		for _, f := range fields {
			o.ignoreFields[f] = true
		}
	}
}

func FieldMapping(srcField, dstField string) CopyOption {
	return func(o *copyOptions) {
		o.fieldMapping[srcField] = dstField
	}
}

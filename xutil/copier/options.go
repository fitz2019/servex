package copier

// CopyOption 复制操作的配置选项.
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

// IgnoreFields 忽略指定字段不参与复制.
func IgnoreFields(fields ...string) CopyOption {
	return func(o *copyOptions) {
		for _, f := range fields {
			o.ignoreFields[f] = true
		}
	}
}

// FieldMapping 设置源字段到目标字段的名称映射.
func FieldMapping(srcField, dstField string) CopyOption {
	return func(o *copyOptions) {
		o.fieldMapping[srcField] = dstField
	}
}

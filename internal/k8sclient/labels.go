package k8sclient

import (
	"fmt"

	assets "github.com/danielmichaels/tawny"
)

type LabelOpts struct {
	Extra     map[string]string
	Name      string
	Component string
	Core      bool
}

type LabelOpt func(*LabelOpts)

func WithName(name string) LabelOpt {
	return func(args *LabelOpts) {
		args.Name = name
	}
}

func WithComponent(component string) LabelOpt {
	return func(args *LabelOpts) {
		args.Component = component
	}
}

func WithCustomLabel(key, value string) LabelOpt {
	return func(args *LabelOpts) {
		if args.Extra == nil {
			args.Extra = make(map[string]string)
		}
		args.Extra[key] = value
	}
}
func WithCoreLabel(core bool) LabelOpt {
	return func(l *LabelOpts) {
		l.Core = core
	}
}

func CreateLabels(opts ...LabelOpt) map[string]string {
	args := &LabelOpts{}

	for _, opt := range opts {
		opt(args)
	}

	labels := map[string]string{
		"tawny.sh/name":       fmt.Sprintf("%s-%s", assets.AppName, args.Name),
		"tawny.sh/component":  fmt.Sprintf("%s-%s", assets.AppName, args.Component),
		"tawny.sh/part-of":    assets.AppName,
		"tawny.sh/managed-by": assets.AppName,
	}
	if args.Core {
		for k, v := range labels {
			labels[k] = fmt.Sprintf("%s-core", v)
		}
	}

	for k, v := range args.Extra {
		labels[k] = v
	}

	return labels
}

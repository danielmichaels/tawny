package k8sclient

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVR is a kubernetes resource schema
// Fields are group/version/resources:subresource
type GVR struct {
	raw, g, v, r, sr string
}

// NewGVR creates a GVR from a group, version, resource
func NewGVR(gvr string) *GVR {
	var g, v, r, sr string

	t := strings.Split(gvr, ":")
	raw := gvr
	if len(t) == 2 {
		raw, sr = t[0], t[1]
	}
	t = strings.Split(raw, "/")
	switch len(t) {
	case 3:
		g, v, r = t[0], t[1], t[2]
	case 2:
		v, r = t[0], t[1]
	case 1:
		r = t[0]
	default:
		log.Error().Err(fmt.Errorf("invalid gvr: %s", gvr)).Msg("invalid gvr")
	}
	return &GVR{raw: raw, g: g, v: v, r: r, sr: sr}
}

// GVK returns a full schema representation.
func (g GVR) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   g.G(),
		Version: g.V(),
		Kind:    g.R(),
	}
}

// V returns the resource version.
func (g GVR) V() string {
	return g.v
}

// RG returns the resource and group.
func (g GVR) RG() (string, string) {
	return g.r, g.g
}

// R returns the resource name.
func (g GVR) R() string {
	return g.r
}

// G returns the resource group name.
func (g GVR) G() string {
	return g.g
}

// GV returns the group version scheme representation.
func (g GVR) GV() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   g.g,
		Version: g.v,
	}
}

// GVR returns a full schema representation.
func (g GVR) GVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.G(),
		Version:  g.V(),
		Resource: g.R(),
	}
}

// GR returns a full schema representation.
func (g GVR) GR() *schema.GroupResource {
	return &schema.GroupResource{
		Group:    g.G(),
		Resource: g.R(),
	}
}

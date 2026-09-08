// Code generated from the Rust SDK exports. DO NOT EDIT.
package core

type Plugins = struct {
	CommonMark struct {
		Core *Plugin
		Html *Plugin
	}
	Generic struct {
		Linkify       *Plugin
		Mark          *Plugin
		Math          *Plugin
		Strikethrough *Plugin
		Table         *Plugin
	}
	Trap struct {
		Compat     *Plugin
		References *Plugin
		Spoiler    *Plugin
		Stamp      *Plugin
	}
}
type Presets = struct {
	CommonMark *Grammar
	TraQ       struct {
		V1 *Grammar
	}
}

func (r *Runtime) exportCatalog(plugins []*Plugin, presets []*Grammar) {
	r.Plugins.CommonMark.Core = plugins[0]
	r.Plugins.CommonMark.Html = plugins[1]
	r.Plugins.Generic.Linkify = plugins[6]
	r.Plugins.Generic.Mark = plugins[3]
	r.Plugins.Generic.Math = plugins[2]
	r.Plugins.Generic.Strikethrough = plugins[4]
	r.Plugins.Generic.Table = plugins[5]
	r.Plugins.Trap.Compat = plugins[10]
	r.Plugins.Trap.References = plugins[8]
	r.Plugins.Trap.Spoiler = plugins[7]
	r.Plugins.Trap.Stamp = plugins[9]
	r.Presets.CommonMark = presets[0]
	r.Presets.TraQ.V1 = presets[1]
}

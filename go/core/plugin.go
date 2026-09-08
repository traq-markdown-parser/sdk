package core

import "slices"

// Names are diagnostics. Symbols, rather than names, identify definitions.
type symbol struct{ reserved byte }
type PluginGroup struct {
	identity *symbol
	parent   *PluginGroup
	name     string
}

func NewPluginGroup(name string) *PluginGroup { return &PluginGroup{identity: new(symbol), name: name} }

func (g *PluginGroup) Name() string         { return g.name }
func (g *PluginGroup) Parent() *PluginGroup { return g.parent }
func (g *PluginGroup) Group(name string) *PluginGroup {
	child := NewPluginGroup(name)
	child.parent = g
	return child
}
func (g *PluginGroup) New(name string) *Plugin { p := NewPlugin(name); p.group = g; return p }

type Rule struct {
	runtime *Runtime
	index   int
	name    string
	phase   string
}

func (r *Rule) Name() string  { return r.name }
func (r *Rule) Phase() string { return r.phase }

type Plugin struct {
	text            []int
	providerRuntime *Runtime

	identity *symbol
	name     string
	group    *PluginGroup
	rules    []*Rule
	frozen   bool
}

func NewPlugin(name string) *Plugin { return &Plugin{identity: new(symbol), name: name} }

func (p *Plugin) Name() string            { return p.name }
func (p *Plugin) Namespace() *PluginGroup { return p.group }
func (p *Plugin) Add(rule *Rule) error {
	if p.frozen {
		return &BuildError{Code: "invalid_definition", Reason: "bundled plugins are immutable"}
	}
	if rule == nil {
		return &BuildError{Code: "invalid_definition", Reason: "nil rule"}
	}
	p.rules = append(slices.Clone(p.rules), rule)
	p.identity = new(symbol)
	return nil
}
func (p *Plugin) phaseRules(phase string) []*Rule {
	var result []*Rule
	for _, rule := range p.rules {
		if rule.phase == phase {
			result = append(result, rule)
		}
	}
	return result
}
func (p *Plugin) InlineRules() []*Rule { return p.phaseRules("inline") }
func (p *Plugin) BlockRules() []*Rule  { return p.phaseRules("block") }
func (p *Plugin) TextRules() []*Rule   { return p.phaseRules("text") }

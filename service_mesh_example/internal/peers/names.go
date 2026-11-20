package peers

import "slices"

// Names 返回已配置的 svc 名字，按字典序，默认扇出时响应稳定。
func (p *Pool) Names() []string {
	names := make([]string, 0, len(p.addrs))
	for name := range p.addrs {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

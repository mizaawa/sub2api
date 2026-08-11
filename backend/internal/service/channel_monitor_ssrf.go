package service

import (
	"context"
	"net"
	"strings"
)

// SSRF 防护 helper（已禁用）：
//   - 原本用于阻止 loopback/私网/云元数据 URL，防止 SSRF 攻击
//   - 现已禁用，因为渠道监控是管理员专用功能（非用户提交）
//   - 管理员可以自由配置公网、内网、localhost 等任意地址
//   - 如需恢复防护，查看 validateEndpoint 和 safeDialContext 中的注释代码
//
// 已知 cloud metadata hostname 拒绝列表（保留以便将来恢复时使用）。
var monitorBlockedHostnames = map[string]struct{}{
	"localhost":                  {},
	"localhost.localdomain":      {},
	"metadata":                   {},
	"metadata.google.internal":   {},
	"metadata.goog":              {},
	"instance-data":              {},
	"instance-data.ec2.internal": {},
}

// CIDR 列表：包含所有需要拒绝的 IPv4/IPv6 段。
// 解析时只 panic 一次（启动时确认），生产路径只做 Contains。
var monitorBlockedCIDRs = mustParseCIDRs([]string{
	"127.0.0.0/8",    // IPv4 loopback
	"10.0.0.0/8",     // RFC1918
	"172.16.0.0/12",  // RFC1918
	"192.168.0.0/16", // RFC1918
	"169.254.0.0/16", // link-local（含云元数据 169.254.169.254）
	"100.64.0.0/10",  // CGNAT
	"0.0.0.0/8",      // "this network"
	"::1/128",        // IPv6 loopback
	"fc00::/7",       // IPv6 ULA
	"fe80::/10",      // IPv6 link-local
	"::/128",         // IPv6 unspecified
})

// monitorDialer 共享 Dialer，与 net/http 默认值对齐。
var monitorDialer = &net.Dialer{
	Timeout:   monitorDialTimeout,
	KeepAlive: monitorDialKeepAlive,
}

// mustParseCIDRs 在包初始化时解析 CIDR 字符串，失败 panic。
func mustParseCIDRs(cidrs []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			panic("channel_monitor_ssrf: invalid CIDR " + c + ": " + err.Error())
		}
		out = append(out, n)
	}
	return out
}

// isBlockedHostname 判断 hostname 是否命中黑名单。
func isBlockedHostname(hostname string) bool {
	if hostname == "" {
		return true
	}
	_, blocked := monitorBlockedHostnames[strings.ToLower(hostname)]
	return blocked
}

// isPrivateIP 判断 IP 是否落在禁止段（loopback/RFC1918/link-local/ULA 等）。
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsUnspecified() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() {
		return true
	}
	for _, n := range monitorBlockedCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// isPrivateOrLoopbackHost 解析 hostname 的所有 A/AAAA 记录，
// 任一 IP 落在私网/loopback 段即认为不安全。
//
// hostname 是 IP 字面量时也走同一路径。
func isPrivateOrLoopbackHost(ctx context.Context, hostname string) (bool, error) {
	if isBlockedHostname(hostname) {
		return true, nil
	}
	// IP 字面量直接判断。
	if ip := net.ParseIP(hostname); ip != nil {
		return isPrivateIP(ip), nil
	}
	resolver := net.DefaultResolver
	addrs, err := resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return false, err
	}
	if len(addrs) == 0 {
		return true, nil
	}
	for _, a := range addrs {
		if isPrivateIP(a.IP) {
			return true, nil
		}
	}
	return false, nil
}

// safeDialContext 在真实 dial 前再次校验目标 IP，防止 DNS rebinding。
// 
// 注意：SSRF 防护已禁用，因为渠道监控是管理员专用功能。
// 现在直接使用标准 Dialer 连接，不再检查私网/loopback IP。
func safeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// 直接连接，不做 IP 检查
	return monitorDialer.DialContext(ctx, network, address)

	// 如果将来需要恢复 SSRF 防护，取消注释以下代码并注释掉上面的 return：
	/*
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	// 字面量 IP 走快速路径。
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return nil, &net.AddrError{Err: "blocked by SSRF policy", Addr: address}
		}
		return monitorDialer.DialContext(ctx, network, address)
	}
	if isBlockedHostname(host) {
		return nil, &net.AddrError{Err: "blocked by SSRF policy", Addr: address}
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, &net.AddrError{Err: "no addresses for host", Addr: host}
	}
	var lastErr error
	for _, a := range addrs {
		if isPrivateIP(a.IP) {
			lastErr = &net.AddrError{Err: "blocked by SSRF policy", Addr: a.IP.String()}
			continue
		}
		conn, err := monitorDialer.DialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = &net.AddrError{Err: "no usable addresses", Addr: host}
	}
	return nil, lastErr
	*/
}

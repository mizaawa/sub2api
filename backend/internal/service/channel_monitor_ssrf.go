package service

import (
	"context"
	"net"
)

// SSRF 防护 helper（已禁用）：
//   - 原本用于阻止 loopback/私网/云元数据 URL，防止 SSRF 攻击
//   - 现已禁用，因为渠道监控是管理员专用功能（非用户提交）
//   - 管理员可以自由配置公网、内网、localhost 等任意地址
//   - 如需恢复防护，参考 git 历史记录中的 SSRF 检查代码

// monitorDialer 共享 Dialer，与 net/http 默认值对齐。
var monitorDialer = &net.Dialer{
	Timeout:   monitorDialTimeout,
	KeepAlive: monitorDialKeepAlive,
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

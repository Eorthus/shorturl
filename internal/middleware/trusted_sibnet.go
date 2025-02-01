package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware проверяет, что IP-адрес клиента входит в доверенную подсеть
func TrustedSubnetMiddleware(trustedSubnet string) func(next http.Handler) http.Handler {
	var trustedNet *net.IPNet
	if trustedSubnet != "" {
		_, network, err := net.ParseCIDR(trustedSubnet)
		if err == nil {
			trustedNet = network
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если подсеть не настроена, запрещаем доступ
			if trustedNet == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Получаем IP-адрес из заголовка X-Real-IP
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				http.Error(w, "Invalid IP address", http.StatusForbidden)
				return
			}

			// Проверяем, входит ли IP в доверенную подсеть
			if !trustedNet.Contains(ip) {
				http.Error(w, "IP not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

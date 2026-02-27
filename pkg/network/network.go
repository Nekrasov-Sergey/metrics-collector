package network

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/peer"
)

func ParseCIDR(cidr string) (*net.IPNet, error) {
	if cidr == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка парсинга CIDR")
	}

	return ipNet, nil
}

func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.Wrap(err, "не удалось определить локальный IP")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msg("ошибка закрытия UDP-соединения")
		}
	}()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", errors.New("не удалось привести адрес к *net.UDPAddr")
	}

	return addr.IP.String(), nil
}

func GetRemoteIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}

	addr, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		return ""
	}

	return addr.IP.String()
}

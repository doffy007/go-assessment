package uid

import (
	"errors"
	"net"
	"rest-api/internal/config"
)

var ErrAssignNewID = errors.New("failed to assign new id")

func New() (uint64, error) {
	id, err := sf.NextID()
	if err != nil {
		return 0, ErrAssignNewID
	}
	return id, nil
}

func machineID() (uint16, error) {
	if len(config.AppConfig.Sonyflake.IP) == 0 {
		return 0, errors.New("IP not set")
	}
	ip := net.ParseIP(config.AppConfig.Sonyflake.IP)
	if len(ip) < 16 {
		return 0, errors.New("invalid IP")
	}
	return uint16(ip[8])<<7 + uint16(ip[9])<<6 +
			uint16(ip[10])<<5 + uint16(ip[11])<<4 +
			uint16(ip[12])<<3 + uint16(ip[13])<<2 +
			uint16(ip[14])<<1 + uint16(ip[15]),
		nil
}

package cuid2

import (
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	DefaultLength = 25
	MaxLength     = 32
	MinLength     = 2
	base36Chars   = "0123456789abcdefghijklmnopqrstuvwxyz"
	letterChars   = "abcdefghijklmnopqrstuvwxyz"
)

var (
	counter          atomic.Int64
	fingerprintBytes []byte
)

func init() {
	var b [4]byte
	_, _ = rand.Read(b[:])
	counter.Store(int64(b[0])<<24 | int64(b[1])<<16 | int64(b[2])<<8 | int64(b[3]))
	fingerprintBytes = createFingerprint()
}

func createFingerprint() []byte {
	h := sha256.New()

	// Hardware address (MAC) — globally unique per NIC
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if len(iface.HardwareAddr) > 0 && iface.Flags&net.FlagLoopback == 0 {
				h.Write(iface.HardwareAddr)
			}
		}
	}

	// First non-loopback IP address
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				h.Write(ipNet.IP)
			}
		}
	}

	// Hostname + PID as additional differentiation
	hostname, _ := os.Hostname()
	h.Write([]byte(hostname))
	h.Write(strconv.AppendInt(nil, int64(os.Getpid()), 10))

	var entropy [32]byte
	_, _ = rand.Read(entropy[:])
	h.Write(entropy[:])

	return bigIntToBase36(h.Sum(nil))
}

func bigIntToBase36(data []byte) []byte {
	n := new(big.Int).SetBytes(data)
	if n.Sign() == 0 {
		return []byte{'0'}
	}
	base := big.NewInt(36)
	mod := new(big.Int)
	var sb strings.Builder
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		sb.WriteByte(base36Chars[mod.Int64()])
	}
	buf := []byte(sb.String())
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return buf
}

// Generate produces a new CUID2 with DefaultLength (25 characters).
func Generate() string {
	return GenerateWithLength(DefaultLength)
}

// GenerateWithLength produces a new CUID2 with the specified length (clamped to [MinLength, MaxLength]).
func GenerateWithLength(length int) string {
	if length < MinLength {
		length = MinLength
	}
	if length > MaxLength {
		length = MaxLength
	}

	// Single crypto/rand read: 1 byte for leading letter + length bytes for salt
	randBuf := make([]byte, 1+length)
	_, _ = rand.Read(randBuf)

	// Build entropy: timestamp || salt || counter || fingerprint
	ts := time.Now().UnixMilli()
	count := counter.Add(1)

	entropyBuf := make([]byte, 0, 20+length+20+len(fingerprintBytes))
	entropyBuf = strconv.AppendInt(entropyBuf, ts, 10)
	for _, b := range randBuf[1:] {
		entropyBuf = append(entropyBuf, base36Chars[b%36])
	}
	entropyBuf = strconv.AppendInt(entropyBuf, count, 10)
	entropyBuf = append(entropyBuf, fingerprintBytes...)

	// SHA-256 hash — returns [32]byte on stack, no heap allocation
	h := sha256.Sum256(entropyBuf)

	// Build result: random letter + hash-derived base36 characters
	result := make([]byte, length)
	result[0] = letterChars[randBuf[0]%26]
	for i := 1; i < length; i++ {
		result[i] = base36Chars[h[i-1]%36]
	}

	return string(result)
}

// IsCuid checks whether the given string is a valid CUID2 format.
func IsCuid(id string) bool {
	if len(id) < MinLength || len(id) > MaxLength {
		return false
	}
	if id[0] < 'a' || id[0] > 'z' {
		return false
	}
	for i := 1; i < len(id); i++ {
		c := id[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

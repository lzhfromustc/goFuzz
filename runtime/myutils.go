package runtime

import (
	"internal/bytealg"
)

func SleepMS(numMs int) {
	durationMS := 1000
	for i := 0; i < numMs; i++ {
		usleep(uint32(durationMS)) // seems usleep can at most sleep for 10ms
	}
}

func Byte_to_Uint16(b []byte) uint16 {
	return uint16(b[1]) | uint16(b[0])<<8
}

func Uint16_to_Byte(b []byte, v uint16) {
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func Byte_to_Uint32(b []byte) uint32 {
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

func Uint32_to_Byte(b []byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func XorUint16(a, b uint16) uint16 {
	byteA := []byte{}
	byteB := []byte{}
	Uint16_to_Byte(byteA, a)
	Uint16_to_Byte(byteB, a)

	byteA = XorByte(byteA, byteB)
	return Byte_to_Uint16(byteA)
}

func XorUint32(a, b uint32) uint32 {
	byteA := []byte{}
	byteB := []byte{}
	Uint32_to_Byte(byteA, a)
	Uint32_to_Byte(byteB, a)

	byteA = XorByte(byteA, byteB)
	return Byte_to_Uint32(byteA)
}


func XorByte(a, b []byte) []byte {
	for _, i := range a {
		a[i] ^= b[i]
	}
	return a
}

// Parse the string returned by Stack(*, false)
// e.g.:
//goroutine 181 [running]:
//runtime.BeforeBlock()
///usr/local/go/src/runtime/mycode.go:30 +0x90
//google.golang.org/grpc.(*addrConn).resetTransport(0xc0001a2840)
///data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc/clientconn.go:1213 +0x676
//created by google.golang.org/grpc.(*addrConn).connect
///data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc/clientconn.go:843 +0x12a

type StackSingleGo struct {
	GoID string
	GoStatus string // this may not help
	VecFuncName, VecFuncFile, VecFuncLine []string
	CreaterName, CreaterFile, CreaterLine string
	OnOtherThread bool // Sometimes the stack is unavailable, if the goroutine is on another thread
}

func ParseStackStr(stackStr string) StackSingleGo {
	stackSingleGo := StackSingleGo{}
	// first line
	indexGoroutine := Index(stackStr, "goroutine ")
	indexLeftParen := Index(stackStr, " [")
	indexRightParen := Index(stackStr, "]")
	if Index(stackStr, "goroutine ") == -1 || Index(stackStr, " [") == -1 || Index(stackStr, "]") == -1 {
		print("detected1\n")
		print(stackStr)
	}
	stackSingleGo.GoID = stackStr[indexGoroutine + 10: indexLeftParen]
	stackSingleGo.GoStatus = stackStr[indexLeftParen + 2: indexRightParen]

	if Index(stackStr, "goroutine running on other thread; stack unavailable") > -1 {
		stackSingleGo.OnOtherThread = true
		return stackSingleGo
	}

	str := stackStr
	for  {
		indexEnter := Index(str, "\n")
		if indexEnter == -1 {
			break
		}
		str = str[indexEnter + 1:] // remove the last line
		if str == "" {
			break
		}
		indexCreatedBy := Index(str, "created by ")
		boolCreatedBy := indexCreatedBy > -1 && Index(str, "\n") > indexCreatedBy // this line indicates which function creates this goroutine
		if boolCreatedBy {
			stackSingleGo.CreaterName = str[indexCreatedBy + 11 : Index(str, "\n") ]
			if Index(str, "\n") == -1 {
				print("detected2\n")
				print(str)
			}
		} else {
			indexLastLeftParen := LastIndex(str, "(")
			if LastIndex(str, "(") == -1 {
				print("detected3\n")
				print(str)
			}
			stackSingleGo.VecFuncName = append(stackSingleGo.VecFuncName, str[ : indexLastLeftParen])
		}

		str = str[Index(str, "\n") + 1:]
		indexColon := Index(str, ":")
		indexSpace := Index(str, " ")
		if -1 == Index(str, ":") || -1 == Index(str, " ") {
			print("detected4\n")
			print(str)
		}
		if boolCreatedBy {
			strFuncFile := str[ : indexColon]
			if indexTab := Index(strFuncFile, "\t"); indexTab > -1 { // remove "\t"
				strFuncFile = strFuncFile[indexTab + 1:]
			}
			stackSingleGo.CreaterFile = strFuncFile
			stackSingleGo.CreaterLine = str[indexColon + 1: indexSpace]
		} else {
			strFuncFile := str[ : indexColon]
			if indexTab := Index(strFuncFile, "\t"); indexTab > -1 { // remove "\t"
				strFuncFile = strFuncFile[indexTab + 1:]
			}
			stackSingleGo.VecFuncFile = append(stackSingleGo.VecFuncFile, strFuncFile)
			stackSingleGo.VecFuncLine = append(stackSingleGo.VecFuncLine, str[indexColon + 1: indexSpace])
		}

	}
	return stackSingleGo
}

func Index(s, substr string) int {
	n := len(substr)
	switch {
	case n == 0:
		return 0
	case n == 1:
		return IndexByte(s, substr[0])
	case n == len(s):
		if substr == s {
			return 0
		}
		return -1
	case n > len(s):
		return -1
	case n <= bytealg.MaxLen:
		// Use brute force when s and substr both are small
		if len(s) <= bytealg.MaxBruteForce {
			return bytealg.IndexString(s, substr)
		}
		c0 := substr[0]
		c1 := substr[1]
		i := 0
		t := len(s) - n + 1
		fails := 0
		for i < t {
			if s[i] != c0 {
				// IndexByte is faster than bytealg.IndexString, so use it as long as
				// we're not getting lots of false positives.
				o := IndexByte(s[i:t], c0)
				if o < 0 {
					return -1
				}
				i += o
			}
			if s[i+1] == c1 && s[i:i+n] == substr {
				return i
			}
			fails++
			i++
			// Switch to bytealg.IndexString when IndexByte produces too many false positives.
			if fails > bytealg.Cutover(i) {
				r := bytealg.IndexString(s[i:], substr)
				if r >= 0 {
					return r + i
				}
				return -1
			}
		}
		return -1
	}
	c0 := substr[0]
	c1 := substr[1]
	i := 0
	t := len(s) - n + 1
	fails := 0
	for i < t {
		if s[i] != c0 {
			o := IndexByte(s[i:t], c0)
			if o < 0 {
				return -1
			}
			i += o
		}
		if s[i+1] == c1 && s[i:i+n] == substr {
			return i
		}
		i++
		fails++
		if fails >= 4+i>>4 && i < t {
			// See comment in ../bytes/bytes.go.
			j := indexRabinKarp(s[i:], substr)
			if j < 0 {
				return -1
			}
			return i + j
		}
	}
	return -1
}

// LastIndex returns the index of the last instance of substr in s, or -1 if substr is not present in s.
func LastIndex(s, substr string) int {
	n := len(substr)
	switch {
	case n == 0:
		return len(s)
	case n == 1:
		return LastIndexByte(s, substr[0])
	case n == len(s):
		if substr == s {
			return 0
		}
		return -1
	case n > len(s):
		return -1
	}
	// Rabin-Karp search from the end of the string
	hashss, pow := hashStrRev(substr)
	last := len(s) - n
	var h uint32
	for i := len(s) - 1; i >= last; i-- {
		h = h*primeRK + uint32(s[i])
	}
	if h == hashss && s[last:] == substr {
		return last
	}
	for i := last - 1; i >= 0; i-- {
		h *= primeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i+n])
		if h == hashss && s[i:i+n] == substr {
			return i
		}
	}
	return -1
}

// hashStrRev returns the hash of the reverse of sep and the
// appropriate multiplicative factor for use in Rabin-Karp algorithm.
func hashStrRev(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := len(sep) - 1; i >= 0; i-- {
		hash = hash*primeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// LastIndexByte returns the index of the last instance of c in s, or -1 if c is not present in s.
func LastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func IndexByte(s string, c byte) int {
	return bytealg.IndexByteString(s, c)
}

func indexRabinKarp(s, substr string) int {
	// Rabin-Karp search
	hashss, pow := hashStr(substr)
	n := len(substr)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*primeRK + uint32(s[i])
	}
	if h == hashss && s[:n] == substr {
		return 0
	}
	for i := n; i < len(s); {
		h *= primeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashss && s[i-n:i] == substr {
			return i - n
		}
	}
	return -1
}

func hashStr(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}
// primeRK is the prime base used in Rabin-Karp algorithm.
const primeRK = 16777619

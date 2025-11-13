package core

// 西电深澜系统使用了魔改的XXTEA加密
func srunXXTEAEncrypt(data string, key string) []byte {
	v := s(data, true)
	k := s(key, false)
	if len(k) < 4 {
		k = append(k, make([]uint32, 4-len(k))...)
	}
	n := uint32(len(v) - 1)
	if n < 1 {
		return []byte(l(v))
	}
	z, y, delta, q := v[n], v[0], uint32(0x9E3779B9), 6+52/(n+1)
	var sum uint32 = 0
	for q > 0 {
		sum += delta
		e := (sum >> 2) & 3
		var p uint32
		for p = 0; p < n; p++ {
			y = v[p+1]
			m := (z>>5 ^ y<<2) + (y>>3 ^ z<<4 ^ (sum ^ y)) + (k[(p&3)^e] ^ z)
			v[p] += m
			z = v[p]
		}
		y = v[0]
		m := (z>>5 ^ y<<2) + (y>>3 ^ z<<4 ^ (sum ^ y)) + (k[(p&3)^e] ^ z)
		v[n] += m
		z = v[n]
		q--
	}
	return []byte(l(v))
}
func s(data string, includeLength bool) []uint32 {
	n := len(data)
	paddedData := []byte(data)
	if n%4 != 0 {
		paddedData = append(paddedData, make([]byte, 4-n%4)...)
	}
	v := make([]uint32, len(paddedData)/4)
	for i := 0; i < len(paddedData); i += 4 {
		v[i>>2] = uint32(paddedData[i]) | uint32(paddedData[i+1])<<8 | uint32(paddedData[i+2])<<16 | uint32(paddedData[i+3])<<24
	}
	if includeLength {
		v = append(v, uint32(n))
	}
	return v
}
func l(data []uint32) string {
	byteData := make([]byte, len(data)*4)
	for i, val := range data {
		byteData[i*4+0] = byte(val & 0xff)
		byteData[i*4+1] = byte(val >> 8 & 0xff)
		byteData[i*4+2] = byte(val >> 16 & 0xff)
		byteData[i*4+3] = byte(val >> 24 & 0xff)
	}
	return string(byteData)
}
package bpre

var unmarshal [][]byte = nil

func MFUnmarshal(l int) {
	unmarshal = make([][]byte, l)
}

func GetMFUnmarshal(s int) [][]byte {
	if unmarshal == nil {
		return make([][]byte, s)
	}
	return unmarshal[:s]
}

var marshal []byte = nil

func Marshal(l int) {
	marshal = make([]byte, l)
}

func GetMarshal(s int) []byte {
	if marshal == nil {
		return make([]byte, s)
	}
	return marshal[:s]
}

func Reset() {
	marshal = nil
	unmarshal = nil
}

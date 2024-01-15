package bpre

var unmarshal [][]byte

func UnmarshalMF(l int) {
	unmarshal = make([][]byte, l)
}

func GetUnmarshalMF(s int) [][]byte {
	if unmarshal == nil {
		return nil
	}
	return unmarshal[:s]
}

var marshal []byte

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

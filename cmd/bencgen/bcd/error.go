package bcd

import (
	"fmt"
	"os"
)

func (b *Bcd) error(m string) {
	errorMessage := "\n\033[1;31m[bencgen] Error:\033[0m\n"
	errorMessage += fmt.Sprintf("    \033[1;37mFile:\033[0m %s\n", b.file)
	errorMessage += fmt.Sprintf("    \033[1;37mMessage:\033[0m %s\n", m)
	fmt.Println(errorMessage)
	os.Exit(-1)
}

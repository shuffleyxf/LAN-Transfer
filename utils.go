package transfer

import (
	"bufio"
	"fmt"
)

// Input 输出提示信息并等待用户输入
func Input(scanner *bufio.Scanner, info string) string {
	fmt.Printf("%s\n>", info)
	scanner.Scan()
	return scanner.Text()
}

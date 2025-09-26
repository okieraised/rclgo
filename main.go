package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Getenv("AMENT_PREFIX_PATH"))

}

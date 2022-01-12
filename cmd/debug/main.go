package main

import (
	tmplog "log"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func main() {
	tmplog.Println("fff")

}


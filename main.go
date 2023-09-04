package main

import "fmt"


func main(){
	parser := newParser("(+ (/ 7 2) (* 2 3) )")
	fmt.Println(parser.Parse().eval())
}

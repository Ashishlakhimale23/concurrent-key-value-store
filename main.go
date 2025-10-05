package main

import (
	"fmt"
	"sync/atomic"
)

type node struct {
	childrens [10]*node
	uniname   string
}

func main() {

	Newnode := getNewNode()
	rootNode := atomic.Pointer[node]{}
	rootNode.Store(Newnode)

}

func (n *node) childrensExists() bool {
	for i := 0; i < 10; i++ {
		if n.childrens[i] != nil {
			return true
		}
	}
	return false
}

func getNewNode() *node {
	var newNode node
	for i := 0; i < 10; i++ {
		newNode.childrens[i] = nil
	}
	newNode.uniname = ""

	return &newNode
}

func (n *node) clone() *node {
	newNode := getNewNode()

	newNode.uniname = n.uniname
	for i := 0; i < 10; i++ {
		newNode.childrens[i] = n.childrens[i]
	}

	return newNode
}

func Insert(year int, uniname string, rootNode *atomic.Pointer[node]) {
	for {

		var rootnode *node = rootNode.Load()
		var newRootNode *node = rootnode.clone()
		var current *node = newRootNode
		var yearString string = fmt.Sprintf("%d", year)

		for i := 0; i < len(yearString); i++ {
			child := rune(yearString[i]) - '0'

			if current.childrens[child] != nil && i != len(yearString)-1 {
				current.childrens[child] = current.childrens[child].clone()
				current = current.childrens[child]
				continue
			}

			current.childrens[child] = getNewNode()
			current = current.childrens[child]
		}

		current.uniname = uniname
		if rootNode.CompareAndSwap(newRootNode, newRootNode) {
			return
		}
	}
}

func Delete(year int, uniname string, rootNode *atomic.Pointer[node]) {

	for {

		rootnode := rootNode.Load()
		if rootnode == nil {
			fmt.Println("The rootnode is nil")
			return
		}
		newRoot := rootnode.clone()
		current := newRoot
		yearString := fmt.Sprintf("%v", year)

		for i := 0; i < len(yearString); i++ {

			child := rune(yearString[i]) - '0'

			if current.childrens[child] != nil && i != len(yearString)-1 {
				current.childrens[child] = current.childrens[child].clone()
				current = current.childrens[child]
				continue
			}

			if current.childrens[child] != nil && current.childrens[child].uniname == uniname && i == len(yearString)-1 {
				// delete here ...
				current.childrens[child] = current.childrens[child].clone()
				current = current.childrens[child]
				current.uniname = ""
				if !current.childrensExists() {
					current = nil
				}
			}
		}

		if rootNode.CompareAndSwap(newRoot, rootnode) {
			return
		}
	}
}

func Search(year int, uniname string, rootNode *atomic.Pointer[node]) {

	rootnode := rootNode.Load()

	if rootnode == nil {
		return
	}

	var traversingNode *node = rootnode
	yearString := fmt.Sprintf("%v", year)

	for i := 0; i < len(yearString); i++ {

		child := rune(yearString[i]) - '0'

		if traversingNode.childrens[child] != nil && i != len(yearString)-1 {
			traversingNode = traversingNode.childrens[child]
			continue
		}

		if traversingNode.childrens[child] != nil && traversingNode.childrens[child].uniname == uniname && i == len(yearString)-1 {
			fmt.Printf("found : %v \n", traversingNode.childrens[child].uniname)
			return
		}
	}
	fmt.Printf("not found : %v \n", uniname)
}


package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type trie struct {
	uniName   string
	childrens [10]*trie
}

type result struct {
	uniname *string
	root    *trie
}

func main() {

	var store atomic.Pointer[trie]
	store.Store(&trie{})

	var channelWg sync.WaitGroup
	var insertingWg sync.WaitGroup
	var valueGaurd map[*string]*trie = make(map[*string]*trie)

	searchChannel := make(chan result, 5)

	insertingWg.Add(2)
	go insertWithCloningTheRoot(1270, "harvard", &store, &insertingWg)
	go insertWithCloningTheRoot(3270, "stanford", &store, &insertingWg)
	insertingWg.Wait()

	channelWg.Add(2)
	go search(&channelWg, searchChannel, &store, 1270, valueGaurd)
	go search(&channelWg, searchChannel, &store, 3270, valueGaurd)

	go func() {
		channelWg.Wait()
		close(searchChannel)
	}()

	for res := range searchChannel {
		fmt.Println(res)
	}

}

func (node *trie) clone() *trie {
	if node == nil {
		return nil
	}
	newNode := getNode()
	newNode.uniName = node.uniName
	for i := 0; i < 10; i++ {
		newNode.childrens[i] = node.childrens[i]
	}
	return newNode
}

func (node *trie) nonNilChildren() bool {
	if node == nil {
		return true
	}
	for i := 0; i < 10; i++ {
		if node.childrens[i] != nil {
			return false
		}
	}
	return true
}
func getNode() *trie {
	var newNode = &trie{}
	newNode.uniName = ""

	for i := 0; i < 10; i++ {
		newNode.childrens[i] = nil
	}

	return newNode
}

func insertWithCloningTheRoot(year int, uniName string, store *atomic.Pointer[trie], wg *sync.WaitGroup) {
	defer wg.Done()

	for {

		var rootNode *trie = store.Load()
		var oldRoot *trie = rootNode.clone()
		var lastDigit int
		var decreasingYear int = year
		var newRootNode *trie = rootNode.clone()
		var nodeForiterationPRoot *trie = rootNode
		var nodeForiterationNRoot *trie = newRootNode

		for decreasingYear > 0 {

			lastDigit = decreasingYear % 10
			decreasingYear = decreasingYear / 10

			if nodeForiterationPRoot.childrens[lastDigit] != nil {
				// clone the childrens and then set the children to nodeforiteration and continue
				nodeForiterationNRoot.childrens[lastDigit] = nodeForiterationPRoot.childrens[lastDigit].clone()
				nodeForiterationNRoot = nodeForiterationNRoot.childrens[lastDigit]
				nodeForiterationPRoot = nodeForiterationPRoot.childrens[lastDigit]
				fmt.Printf("Digit %d, PRoot: %p, NRoot: %p\n", lastDigit, nodeForiterationPRoot, nodeForiterationNRoot)
				continue
			}

			newNode := getNode()
			nodeForiterationNRoot.childrens[lastDigit] = newNode
			nodeForiterationNRoot = newNode

			if decreasingYear == 0 {
				nodeForiterationNRoot.uniName = uniName
			}
		}

		if store.CompareAndSwap(oldRoot, newRootNode) {
			return
		}

	}
}

func search(channelWg *sync.WaitGroup, searchChannel chan<- result, store *atomic.Pointer[trie], year int, valueGaurd map[*string]*trie) {

	defer channelWg.Done()

	rootNode := store.Load()
	var lastDigit int
	var decreasingYear int = year
	var nodeForiteration *trie = rootNode

	for decreasingYear > 0 {

		lastDigit = decreasingYear % 10
		decreasingYear = decreasingYear / 10

		if nodeForiteration.childrens[lastDigit] != nil && decreasingYear != 0 {
			nodeForiteration = nodeForiteration.childrens[lastDigit]
			continue
		} else if nodeForiteration.childrens[lastDigit] != nil && decreasingYear == 0 {
			valueGaurd[&nodeForiteration.childrens[lastDigit].uniName] = rootNode
			res := result{
				uniname: &nodeForiteration.childrens[lastDigit].uniName,
				root:    rootNode,
			}
			searchChannel <- res
			return
		}
	}

	notFound := "not found"
	res := result{
		uniname: &notFound,
		root:    &trie{},
	}
	searchChannel <- res
}

func deleteWithCloningTheRoot(year int, uniName string, store *atomic.Pointer[trie], wg *sync.WaitGroup) {
	defer wg.Done()

	for {

		var rootNode *trie = store.Load()
		var oldRoot *trie = rootNode.clone()
		var lastDigit int
		var decreasingYear int = year
		var newRootNode *trie = rootNode.clone()
		var nodeForiterationPRoot *trie = rootNode
		var nodeForiterationNRoot *trie = newRootNode

		for decreasingYear > 0 {

			lastDigit = decreasingYear % 10
			decreasingYear = decreasingYear / 10

			if nodeForiterationPRoot.childrens[lastDigit] == nil && decreasingYear != 0 {
				// no value found
				return

			} else if nodeForiterationPRoot.childrens[lastDigit] != nil && decreasingYear > 0 {
				// clone the childrens and then set the children to nodeforiteration and continue
				nodeForiterationNRoot.childrens[lastDigit] = nodeForiterationPRoot.childrens[lastDigit].clone()
				nodeForiterationNRoot = nodeForiterationNRoot.childrens[lastDigit]
				nodeForiterationPRoot = nodeForiterationPRoot.childrens[lastDigit]
				fmt.Printf("Digit %d, PRoot: %p, NRoot: %p\n", lastDigit, nodeForiterationPRoot, nodeForiterationNRoot)
				continue
			} else if nodeForiterationPRoot.childrens[lastDigit] != nil && decreasingYear == 0 && nodeForiterationPRoot.childrens[lastDigit].uniName == uniName {
				nodeForiterationNRoot.childrens[lastDigit] = nodeForiterationPRoot.childrens[lastDigit].clone()
				// if there are any non nil childrens then keep or just delete it
				nodeForiterationNRoot.childrens[lastDigit].uniName = ""

				check := nodeForiterationNRoot.childrens[lastDigit].nonNilChildren()
				if check {
					nodeForiterationNRoot.childrens[lastDigit] = nil
				}
			}
		}

		// add the new root here
		if store.CompareAndSwap(oldRoot, newRootNode) {
			return
		}

	}
}

func insert(rootNode *trie, year int, uniName string) {

	// check if the number exists int the childrens of the rootNode
	// if not create one && if yes iterate the children now with the number we check

	var lastDigit int
	var decreasingYear int = year
	var nodeForiteration *trie = rootNode

	for decreasingYear > 0 {

		lastDigit = decreasingYear % 10
		decreasingYear = decreasingYear / 10

		if nodeForiteration.childrens[lastDigit] != nil {
			nodeForiteration = nodeForiteration.childrens[lastDigit]
			continue
		}

		newNode := getNode()
		nodeForiteration.childrens[lastDigit] = newNode
		nodeForiteration = newNode

		if decreasingYear == 0 {
			nodeForiteration.uniName = uniName
		}
	}

}

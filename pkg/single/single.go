package single

type Node struct {
	val  int
	next *Node
}

type LinkedList struct {
	Head *Node
}

func (list *LinkedList) InsertAtHead(val int) {
	tmp := list.Head
	node := &Node{val, tmp}
	list.Head = node
}

func (list *LinkedList) RemoveVal(val int) {
	// 1. Handle empty list
	if list.Head == nil {
		return
	}

	// 2. Handle deleting the Head node
	if list.Head.val == val {
		list.Head = list.Head.next
		return
	}

	// 3. Search for the value in the rest of the list
	tmp := list.Head
	for tmp.next != nil {
		if tmp.next.val == val {
			tmp.next = tmp.next.next
			return
		}
		tmp = tmp.next
	}
}

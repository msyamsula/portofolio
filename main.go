package main

type node struct {
	val         int
	left, right *node
}

var height = make(map[*node]int)

func dfs(u *node, h int) {
	if u == nil {
		return
	}

	height[u] = h
	dfs(u.left, h+1)
	dfs(u.right, h+1)
}

func main() {
	root := &node{
		val:  3,
		left: &node{},
		right: &node{
			val:  2,
			left: &node{},
			right: &node{
				val:   1,
				left:  &node{},
				right: &node{},
			},
		},
	}

	dfs(root, 0)

}

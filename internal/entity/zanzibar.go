package entity

type Instance struct {
	Ns string `validate:"required"`
	Id string `validate:"required"`
}

// type User struct {
// 	Ns string `validate:"required"`
// 	Id string `validate:"required"`
// }

// type TreeNode struct {
// 	Root        *Instance `validate:"required"`
// 	RelChildren map[string]*TreeNode
// }

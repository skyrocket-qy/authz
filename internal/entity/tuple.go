package entity

type Tuple struct {
	Sbj *Sbj
	Rel string
	Obj *Obj
}

type Sbj struct {
	Ns   string
	Name string
	Rel  *string
}

type Obj struct {
	Ns   string
	Name string
}


type TreeNode struct {
	Root     *Sbj
	RelChildren map[string]*TreeNode
}

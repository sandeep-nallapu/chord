package main

type NodeAction string

const (
	initRing  NodeAction = "init-ring-fingers"
	join      NodeAction = "join-ring"
	stabilize NodeAction = "stabilize-ring"
	fixRing   NodeAction = "fix-ring-fingers"
	leave     NodeAction = "leave-ring"
)

type NodeQuery string

const (
	notify    NodeQuery = "ring-notify"
	getFinger NodeQuery = "get-ring-fingers"
	findSuc   NodeQuery = "find-ring-successor"
	findPre   NodeQuery = "find-ring-predecessor"
)

type HashQuery string

const (
	get    HashQuery = "get"
	put    HashQuery = "put"
	remove HashQuery = "remove"
)

type ActionMode string

const (
	immediate ActionMode = "immediate"
	orderly   ActionMode = "orderly"
)

type State int

const (
	none              State = -1
	initialization          = 0
	stabilizeResp     State = 1
	fixFingerResp     State = 2
	populateSuccResp  State = 3
	populateTableResp State = 4
	dataResponse      State = 5
)

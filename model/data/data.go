package data

import (
	forumDB "forum/model/functions"
)

var (
	CombinedData AllData
)

type ToDisplay struct {
	Posts []forumDB.Post
}

type AllData struct {
	ToDisplay      ToDisplay
	Username       string
	Categories     []forumDB.Category
	PostCategories map[int][]string
	Liked          map[int]bool
	Disliked       map[int]bool
}

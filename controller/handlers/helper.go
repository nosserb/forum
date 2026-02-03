package handlers

import (
	"database/sql"
	"strconv"

	forumDB "forum/model/functions"
)

// Récupère et formate les catégories pour chaque post
// Retourne la map des catégories par post et la slice des catégories disponibles
func PostsWithCategories(db *sql.DB, posts []forumDB.Post) (map[int][]string, []forumDB.Category, error) {
	postCats := make(map[int][]string)
	catMap := make(map[int]string)
	var categories []forumDB.Category

	// Construire la map de id à nom des catégories
	if cats, err := forumDB.FetchCategories(db); err == nil {
		categories = cats
		for _, c := range cats {
			catMap[c.ID] = c.Name
		}
	}

	// Associe les catégories à chaque post
	for _, p := range posts {
		if pcs, err := forumDB.FetchPostCategoriesBy(db, "post_id", p.ID); err == nil {
			for _, rel := range pcs {
				if name, ok := catMap[rel.CategoryID]; ok {
					postCats[p.ID] = append(postCats[p.ID], name)
				} else {
					postCats[p.ID] = append(postCats[p.ID], strconv.Itoa(rel.CategoryID))
				}
			}
		}
	}

	return postCats, categories, nil
}

// FormatPostDates formate les dates de tous les posts en temps local
func FormatPostDates(posts []forumDB.Post) {
	for i := range posts {
		posts[i].CreatedAt = posts[i].CreatedAt.Local()
	}
}

// Récupère les likes et dislikes d'un utilisateur
// Retourne la map des posts likés et map des posts dislikés
func GetUserReactions(db *sql.DB, user forumDB.User) (map[int]bool, map[int]bool) {
	liked := map[int]bool{}
	disliked := map[int]bool{}

	if user.Username != "" {
		if reacts, err := forumDB.FetchReactionsBy(db, "user_id", user.ID); err == nil {
			for _, rct := range reacts {
				if rct.PostID != nil {
					pid := *rct.PostID
					if rct.Type == "like" {
						liked[pid] = true
					} else if rct.Type == "dislike" {
						disliked[pid] = true
					}
				}
			}
		}
	}

	return liked, disliked
}

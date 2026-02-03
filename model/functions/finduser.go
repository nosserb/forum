package db

import (
	"database/sql"
)

// La fonction cherche l'utilisateur dans la Db a partir de son email
func FindUser(db *sql.DB, email string) (User, error) {
	var u User

	// Lance la requête SQL qui sélectionne les informations de l'utilisateur correspondant à l'email
	row := db.QueryRow("SELECT id, email, username, password, created_at FROM users WHERE email = ?", email)

	// On lis les données de la ligne retournée par la requête
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.CreatedAt)

	if err != nil {
		return u, err
	}

	// On retourne l'utilisateur trouvé
	return u, nil
}

/*
	cmd :-> go get golang.org/x/crypto/bcrypt
	password := []byte("monSuperMotDePasse123")

    hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }

	err := bcrypt.CompareHashAndPassword(hash, []byte("monSuperMotDePasse123"))
*/

Forum

Ce projet contient un binaire Go main, sa base de données SQLite et ses templates HTML. Le Dockerfile permet de construire et de lancer le conteneur facilement. Prérequis : Docker installé et Linux (Ubuntu 24.04 recommandé). Structure du projet :

docker/
├── main
├── main.go
├── model/
│ └── forum.db
├── view/
│ └── assets/templates/\*.html
├── controller/
├── go.mod
├── go.sum
├── Dockerfile
└── README.md

Pour construire l’image Docker depuis le dossier docker :

docker build -t forumdocker .

-t forumdocker est le nom de l’image. Pour lancer le conteneur :

docker run -it -p 8080:8080 forumdocker

-it active le mode interactif pour voir les logs

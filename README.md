### Forum

Ce projet contient le code du backend, sa base de données SQLite et ses templates HTML. 

Le Dockerfile permet de construire et de lancer le conteneur facilement
<br>
Prérequis : Docker installé et Linux (Ubuntu 24.04 recommandé).

<br>

Pour construire l’image Docker depuis le dossier docker :

`docker build -t forumdocker .` 
<br>
( forumdocker est le nom de l’image )

<br>

Pour lancer le conteneur :

`docker run -it -p 8080:8080 forumdocker`
<br>
( -it active le mode interactif pour voir les logs )
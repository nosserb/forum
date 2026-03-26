### Forum

Ce projet contient le code du backend, sa base de données SQLite et ses templates HTML. 

Le Dockerfile permet de construire et de lancer le conteneur facilement
<br>
Prérequis : Docker installé et Linux (Ubuntu 24.04 recommandé).

<br>

> [!WARNING]    
Le dockerfile est actuellement dysfonctionnel. Nous travaillons à le rendre utilisable à nouveau.

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

<br>

### Crédit

Ce repo est la version open-source d'un projet du cursus commun de Zone01 nommé `forum`, renforcé par un projet `real time forum` plus avancé. 

Le travail sur ce projet a été réalisé en collaboration avec [LeRacoune](https://github.com/LeRacoune) et [rmaillard](https://github.com/rmaillard2101)
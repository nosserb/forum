
![img](https://i.ibb.co/chxV93fN/fa7b7532dde3.png)

> Open-source project Zone01 — Go backend, SQLite database, HTML templates

---

## Overview

<b>
This project is a web forum developed in Go, using a SQLite database and HTML templates. It follows an MVC (Model-View-Controller) architecture and is easily managed via Docker.
</b>

---

## Requirements
- Docker installed
- Linux (Ubuntu 24.04 recommended)
---
## Installation & Startup

### 1. Build the Docker image
```bash
docker build -t forumdocker .
```
> forumdocker is the image name

### 2. Run the container
```bash
docker run -it -p 8080:8080 forumdocker
```
> -it enables interactive mode to view logs

<br>

> [!WARNING] WARNING
 The Dockerfile is currently not functional. We are working to make it usable again.

<!-----

## Project Structure

```text
forum/
├── controller/
│   ├── cookies/
│   ├── handlers/
│   ├── logging/
│   └── server/
├── model/
│   ├── data/
│   └── functions/
├── view/
│   └── assets/
│       ├── static/
│       ├── statics/
│       └── templates/
├── dockerfile
├── go.mod
├── go.sum
├── main.go
└── README.md
``` -->
---
## Main Features
- User authentication and management
- Post creation, editing, and deletion
- Like and comment system
- Error and access management
- Responsive interface (HTML/CSS/JS)

<br>

> [!INFO] Development Notes
> Commits are grouped by major addition or specific modification, as the full history could not be imported (study project).
<br>

---

> [!INFO] CREDITS
This repo is the open-source version of a common curriculum project from **Zone01** called `forum`, reinforced by a more advanced `real time forum` project.
<br>
**Project carried out in collaboration with:**
:octocat: [LeRacoune](https://github.com/LeRacoune)
:octocat: [rmaillard](https://github.com/rmaillard2101)
<br>
**Docs by**
:octocat: [nosserb](https://github.com/nosserb)

<br>

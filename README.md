
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

> [!WARNING]
 The Dockerfile is currently not functional. We are working to make it usable again.

<br>

---

## Main Features
- User authentication and management
- Post creation, editing, and deletion
- Like and comment system
- Error and access management
- Responsive interface (HTML/CSS/JS)

<br>

> [!NOTE]
> Commits are grouped by major addition or specific modification, as the full history could not be imported (study project).
<br>

---

## CREDIT
This repo is the open-source version of a common curriculum project from **Zone01** called `forum`, reinforced by a more advanced `real time forum` project.

**Project carried out in collaboration with:**
<table>
  <tr>
    <td align="center">
      <a href="https://github.com/LeRacoune">
        <img src="https://github.com/LeRacoune.png" width="100px;" alt="LeRacoune"/><br />
        <sub><b>LeRacoune</b></sub>
      </a>
    </td>
    <td align="center">
      <a href="https://github.com/rmaillard">
        <img src="https://github.com/rmaillard.png" width="100px;" alt="rmaillard"/><br />
        <sub><b>rmaillard</b></sub>
      </a>
    </td>
    <td align="center">
      <a href="https://github.com/IronBeagle404">
        <img src="https://github.com/IronBeagle404.png" width="100px;" alt="rmaillard"/><br />
        <sub><b>IronBeagle404</b></sub>
      </a>
    </td>
  </tr>
</table>

**Docs by**

<table>
  <tr>
    <td align="center">
      <a href="https://github.com/nosserb">
        <img src="https://github.com/nosserb.png" width="100px;" alt="Nom"/><br />
        <sub><b>nosserb</b></sub>
      </a>
    </td>
  </tr>
</table>

# Helper v3 API

Helper v3 est une API backend écrite en Go pour interagir avec un outil pédagogique existant codé en PHP. L'API permet de gérer l'authentification des utilisateurs, de récupérer les IDs des cours, de vérifier le statut de présence et de marquer la présence pour les cours de la journée.

## Prérequis

- Go 1.16 ou supérieur
- Accès à l'outil pédagogique codé en PHP
- Docker

## Installation

1. Clonez le dépôt :
    ```sh
    git clone https://github.com/votre-utilisateur/helper-v3.git
    cd helper-v3
    ```

2. Installez les dépendances :
    ```sh
    go mod tidy
    ```

## Utilisation avec Go

1. Compilez et lancez le serveur :
    ```sh
    go run main.go
    ```

2. L'API sera disponible à l'adresse suivante :
    ```
    http://localhost:8888
    ```

## Utilisation avec Docker

1. Construisez l'image Docker :
    ```sh
    docker build -t helper-api .
    ```

2. Lancez un conteneur à partir de l'image :
    ```sh
    docker run -p 8888:8888 helper-api
    ```

3. L'API sera disponible à l'adresse suivante :
    ```
    http://localhost:8888
    ```

## Dockerfile

Le Dockerfile utilisé pour créer l'image est le suivant :

```Dockerfile
FROM golang:latest as build
WORKDIR /app
COPY . .
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o helper-api .
FROM alpine:latest
COPY --from=build /app/helper-api .
EXPOSE 8888
RUN chmod +x helper-api
ENTRYPOINT [ "/helper-api" ]
```

## Endpoints

### Login

- **Endpoint**: `/login`
- **Méthode**: POST
- **Description**: Authentifie l'utilisateur et récupère le cookie de session.
- **Corps de la requête**:
    ```json
    {
        "username": "votre_nom_utilisateur",
        "password": "votre_mot_de_passe"
    }
    ```
- **Réponse**:
    ```json
    {
        "body": {
            "cookie": "cookie_de_session"
        }
    }
    ```

### Get Course IDs

- **Endpoint**: `/getCourseIDs`
- **Méthode**: POST
- **Description**: Récupère les IDs des cours de la journée.
- **Corps de la requête**:
    ```json
    {
        "cookie": "votre_cookie"
    }
    ```
- **Réponse**:
    ```json
    {
        "body": {
            "courses": [
                {
                    "id": "12345",
                    "name": "Nom du cours",
                    "period": "Matin"
                },
                {
                    "id": "67890",
                    "name": "Nom du cours",
                    "period": "Après-midi"
                }
            ]
        }
    }
    ```

### Get Attendance Status

- **Endpoint**: `/getAttendanceStatus`
- **Méthode**: POST
- **Description**: Récupère le statut de présence pour un cours spécifique.
- **Corps de la requête**:
    ```json
    {
        "cookie": "votre_cookie",
        "courseID": "id_du_cours"
    }
    ```
- **Réponse**:
    ```json
    {
        "body": {
            "status": "Fermé et a déjà été ouvert"
        }
    }
    ```

### Set Presence

- **Endpoint**: `/setPresence`
- **Méthode**: POST
- **Description**: Marque la présence pour un cours spécifique.
- **Corps de la requête**:
    ```json
    {
        "cookie": "votre_cookie",
        "courseID": "id_du_cours"
    }
    ```
- **Réponse**:
    ```json
    {
        "body": {
            "message": "Presence marked successfully"
        }
    }
    ```

### Fetch Calendar

- **Endpoint**: `/fetchCalendar`
- **Méthode**: POST
- **Description**: Télécharge et analyse un fichier iCalendar pour récupérer le programme de la semaine.
- **Corps de la requête**:
    ```json
    {
        "calUUID": "49caac7c643b4be6817db60be4374ee7"
    }
    ```
- **Réponse**:
    ```json
    {
        "body": {
            "schedule": [
                {
                    "day": "2024-06-12",
                    "full_day": true,
                    "morning": false,
                    "afternoon": false,
                    "remote": false,
                    "location": "",
                    "professor": "",
                    "subject": "entreprise"
                },
                {
                    "day": "2024-06-13",
                    "full_day": false,
                    "morning": true,
                    "afternoon": false,
                    "remote": false,
                    "location": "E 561",
                    "professor": "John DOE",
                    "subject": "GOLANG"
                }
            ]
        }
    }
    ```
    > Pour récupérer l'UUID, il faudra tout d'abord trouver le lien de téléchargement du calendrier sur Pepal. Il suffit de se diriger vers l'emploi du temps, puis il sera tout simplement en haut à droite.

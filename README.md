# Sujet de l'application
Application web qui agrège les concerts en Europe et centralise leurs données (date, salle, plan, setlist).
Elle permet de suivre ses favorites avec alert de vente et de recevoir des alerts de nouveauté.
Elle facilite la mise en relation via les SNS pour faire du WTB/WTS et pour se retrouver avant un concert.

## Fonctionnalités
- Consulter la liste des concerts en Europe
- Rechercher un concert par artiste ou salle
- Ouvrir la fiche d'un concert (date, salle, localisation, plan si dispo, setlist potentielle)
- Ajouter un concert en favorites (active l'alert de vente)
- Se mettre en WTB / WTS sur un concert
- Voir les autres personnes en WTB / WTS pour un concert
- Activer une alert de nouveaux concerts pour un artiste ou une salle
- Renseigner ses SNS sur son profil
- Voir les SNS des gens qui vont au même concert

### Cas d'Utilisation
- Scénario 1 : Alice met un concert en WTS, Bob se met en WTB ; ils consultent la fiche du concert, récupèrent les SNS et se contactent pour l'échange de place.
- Scénario 2 : Eve active une alert nouveauté sur NMIXX, reçoit l'alert d'un nouveau concert, le met en favorite (alert de vente activée) ; à l'ouverture des ventes, elle reçoit l'alert et voit les SNS des autres personnes qui vont au même concert.
- Scénario 3 : Dave cherche un concert par salle, ouvre la fiche, consulte la setlist potentielle et ajoute le concert en favorite pour être alertée au lancement des ventes.

### Notes
- Les WTB et WTS sont à prix gratuit (comme un don) pour éviter d'avoir des problèmes légaux et pour simplifier l'application.
- Un favorite active l'alert de vente associée.
- Les alerts de nouveauté sont sur des artists ou des venues au choix.

## Liste de données
- Concert
  - id / name / date / venueID / artistID / url / photoURL / seatmapURL / saleStartDateTime / createdAt
- Venue
  - id / name / city / country
- Artist
  - id / name
- User
  - id / email / passwordHash / sns
- Session
  - id / userId / tokenHash / expiresAt
- WebAuthnCredential
  - id / userId / credentialId / publicKey / signCount
- WebAuthnChallenge
  - id / userId / tokenHash / kind / sessionData / expiresAt
- WT
  - userId / concertId / wtType (wtb / wts)
- Favorite
  - userId / concertId
- Alert
  - id / userId / targetType / targetId / createdAt
- SyncTicketmaster
  - max_visibility (ex: 2026-03-24T21:59:47Z)

## API Web
- https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/
Permet de récupérer périodiquement les nouveaux concerts, les plans, les dates, les venues.

- https://api.setlist.fm/docs/1.0/index.html
Setlist potentielle / par artiste (via attractions name)

## Description des requêtes
### Récupérer de nouveaux concerts
- `GET https://app.ticketmaster.com/discovery/v2/events.json?apikey=$TICKETMASTER_API_KEY&countryCode=DE&classificationName=music&size=200&page=0`
- `GET https://app.ticketmaster.com/discovery/v2/events.json?apikey=$TICKETMASTER_API_KEY&countryCode=FR&classificationName=music&size=200&page=0`
  - Filtre nouveauté : ajouter `publicVisibilityStartDateTime=YYYY-MM-DDTHH:MM:SSZ` (ex: `2026-03-24T21:59:47Z`)
  - Champs utiles :
    - event.id, event.name, event.url
    - event.images[].url (une seule image retenue, la meilleure image 16:9 si possible)
    - event.seatmap.staticUrl
    - event.dates.start.localDate / localTime / dateTime
    - event.sales.public.startDateTime
    - venue: _embedded.venues[0].id / name / city.name / country.countryCode
    - artist: _embedded.attractions[0].id / name

### Récupérer setlist d'un artiste
- `GET https://api.setlist.fm/rest/1.0/search/setlists?artistName={nomArtiste}`
  - Parser `eventDate` (dd-mm-yyyy), `tour`, `sets`, `id`
  - Filtrer : `tour` si possible présent, `eventDate` < aujourd'hui, `sets` non vide si possible
  - Prendre le concert le plus récent (max `eventDate`)
- `GET https://api.setlist.fm/rest/1.0/setlist/{setlistId}`
  - Récupérer `sets.song.name` pour la liste des titres

## Mise à jour
- publicVisibilityStartDateTime : filtre de nouveauté (SyncTicketmaster.max_visibility)
- Récupération + insertion des concerts via Ticketmaster
- Pays synchronisés pour l'instant : Allemagne (`DE`), France (`FR`), Italie (`IT`), Espagne (`ES`) et Autriche (`AT`)
- Job automatique toutes les 15 minutes, avec clé Ticketmaster fournie par le shell (`TICKETMASTER_API_KEY`)
- Le job Ticketmaster tourne en goroutine background.
- Nettoyage sync actuel : ignore les events sans venue/artiste nommé, garde une seule photo 16:9 de meilleure qualité, ignore les dates de vente aberrantes avant 2000
- Vérification des alerts de nouveauté, des alerts de vente et des matches WTB/WTS via le radar d'alertes
- Déduplication des notifications déjà envoyées via `notifications.dedupe_key`
- Récup setlist du concert si artiste dispo dans setlist.fm

## Requêtes implémentées actuellement
- GET /healthz
  healthcheck serveur

- GET /api/concerts?artistID=...&venueID=...&country=...&status=future|all&page=1
  liste des concerts, avec filtre optionnel par artiste, salle, pays et pagination

- GET /api/concerts/{concertId}
  détail d'un concert

- GET /api/artists
  liste des artistes -> id

- GET /api/venues
  liste des salles -> id

- POST /api/auth/register
  créer un compte email/password, ouvrir une session cookie

- POST /api/auth/login
  se connecter avec email/password

- POST /api/auth/logout
  révoquer la session courante

- DELETE /api/auth/unregister
  supprimer son compte, avec confirmation password

- GET /api/auth/me
  récupérer l'utilisateur connecté

- GET /api/auth/email-exists?email=...
  vérifier si un email existe

- POST /api/auth/passkeys/register/begin
  commencer l'ajout d'une passkey pour l'utilisateur connecté

- POST /api/auth/passkeys/register/finish
  valider et stocker la passkey créée

- POST /api/auth/passkeys/login/begin
  commencer une connexion passkey sans email

- POST /api/auth/passkeys/login/finish
  valider la passkey et ouvrir une session cookie

- GET /api/auth/passkeys
  lister les passkeys de l'utilisateur connecté

- DELETE /api/auth/passkeys/{credentialId}
  supprimer une passkey de l'utilisateur connecté

- GET /api/favorites/{concertId}  
  voir les SNS des gens qui vont au même concert

- GET /api/setlist/{concertId}  
  voir la setlist potentielle d'un concert

- POST ou DELETE /api/favorites/{concertId}  
  ajouter ou retirer un favorite (active/désactive l'alert de vente)

- POST ou DELETE /api/wt/{concertId}?type=wtb|wts  
  se mettre ou retirer en WTB ou WTS
  - `POST` remplace l'autre statut du même user sur le même concert
  - `POST` sur un concert expiré renvoie `409`

- GET /api/wt/{concertId}  
  voir les WTB / WTS liés à un concert

- POST /api/alerts?targetType=artist|venue&targetId=...
  créer une alert de nouveauté

- DELETE /api/alerts/{alertId}  
  retirer une alert

- GET /api/me  
  voir le profil (sns, favorites, wt, alerts)

- PATCH /api/me  
  modifier le profil (sns)


## Mail
- Le serveur mail applicatif est dans `server/job/mailserver.go`.
- Il tourne en goroutine et consomme des `job.Envelope` depuis un canal buffered.
- L'inscription pousse un mail de bienvenue dans le canal après création de session : la réponse HTTP n'attend pas l'envoi SMTP.
- Si `SMTP_HOST` est vide, le mail est désactivé sans bloquer l'application.
- Configuration SMTP :
  - `SMTP_HOST` : serveur SMTP, par exemple `10.66.66.1` depuis Docker/VPN
  - `SMTP_PORT` : port SMTP, par défaut `25`
  - `SMTP_FROM` : expéditeur, par défaut `ticketmet@jessyfal04.dev`
  - `APP_BASE_URL` : base URL used in mail buttons and `?concert=` redirections, default `https://ticketmet.jessyfal04.dev`
- The HTML template keeps a simple content area, an account card, and an action button.

## Flux du Système
- client -> serveur : requêtes HTTP
- serveur -> client : réponses JSON
- serveur -> Ticketmaster : récupération concerts
- serveur -> setlist.fm : récupération setlist
- serveur <-> BDD : lecture / écriture

## Description serveur
- Backend Go avec `net/http`, `database/sql`
- Démarrage : `server/main/main.go`.
- Schéma : `server/main/schema.sql`, appliqué au lancement. Pas de migration : quand le schéma change, la DB de déploiement est écrasée par `make docker-deploy`.
- Accès DB : `server/job/database.go` lance le serveur de requêtes en goroutine ; on peut l'utiliser grâce au lanceurs sql utilisés par les APIs et les jobs.
- API : `server/api`, handlers pour `/api/concerts`, `/api/artists`, `/api/venues`, `/api/setlist`, `/api/favorites`, `/api/wt`, `/api/me`, `/api/alerts`, `/healthz`.
- Auth : email/password avec bcrypt, sessions serveur par cookie HttpOnly `session`, passkeys WebAuthn via `github.com/go-webauthn/webauthn`.
- WebAuthn : domaine configuré dans `server/api/passkeys.go` pour `ticketmet.jessyfal04.dev`.
- Sync Ticketmaster : `server/job/ticketmaster.go`, lancé dans une goroutine au démarrage puis toutes les 15 minutes, avec `max_visibility`.
- Radar d'alertes : `server/job/alertradar.go`, lancé dans une goroutine au démarrage, regroupe les alerts de nouveauté, de vente et les matches WTB/WTS puis envoie un mail par utilisateur.
- Secrets : `.secrets/ticketmaster.mk`, chargé par le `Makefile`, ignoré par Git.
- Déploiement : image Docker `docker.io/jessyfal04/ticketmet:latest`, DB persistée dans `/app/data/ticketmet.sqlite3`
- Setlist.fm : `server/job/setlistfm.go`, lancé dans une goroutine au démarrage et utilisé par `GET /api/setlist/{concertId}` si `SETLISTFM_API_KEY` est fourni.

## Description client
- Plan : application monopage avec sections Recherche, Fiche concert, Profil, Auth.
- Recherche : liste + filtres + pagination ; appels `GET /api/concerts?artistID=...&venueID=...&country=...&status=...&page=...`, `GET /api/artists`, `GET /api/venues`.
- Fiche concert : détails + setlist + SNS + WTB/WTS ; appels `GET /api/concerts/{concertId}`, `GET /api/setlist/{concertId}`, `GET /api/favorites/{concertId}`, `GET /api/wt/{concertId}`, actions `POST/DELETE /api/favorites/{concertId}`, `POST/DELETE /api/wt/{concertId}`.
- Profil : infos utilisateur et SNS ; appels `GET /api/me`, `PATCH /api/me`, création d'alerts via `POST /api/alerts`, suppression via `DELETE /api/alerts/{alertId}`.
- Auth : écrans inscription/connexion/déconnexion ; appels `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/auth/logout`, passkeys via `/api/auth/passkeys/...`.

## Autre Idées
- Notification push via Firebase + Mail de bienvenue

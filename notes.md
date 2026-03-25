# Sujet de l'application
Application web qui agrège les concerts en Europe et centralise leurs données (date, salle, plan, setlist).
Elle permet de suivre ses favoris avec alerte de vente et de recevoir des alertes de nouveauté.
Elle facilite la mise en relation via les SNS pour faire du WTB/WTS et pour se retrouver avant un concert.

## Fonctionnalités
- Consulter la liste des concerts en Europe
- Rechercher un concert par artiste ou salle
- Ouvrir la fiche d'un concert (date, salle, localisation, plan si dispo, setlist potentielle)
- Ajouter un concert en favoris (active l'alerte de vente)
- Se mettre en WTB / WTS sur un concert
- Voir les autres personnes en WTB / WTS pour un concert
- Activer une alerte de nouveaux concerts pour un artiste ou une salle
- Renseigner ses SNS sur son profil
- Voir les SNS des gens qui vont au même concert

### Cas d'Utilisation
- Scénario 1 : Alice met un concert en WTS, Bob se met en WTB ; ils consultent la fiche du concert, récupèrent les SNS et se contactent pour l'échange de place.
- Scénario 2 : Eve active une alerte nouveauté sur NMIXX, reçoit l'alerte d'un nouveau concert, le met en favori (alerte de vente activée) ; à l'ouverture des ventes, elle reçoit l'alerte et voit les SNS des autres personnes qui vont au même concert.
- Scénario 3 : Dave cherche un concert par salle, ouvre la fiche, consulte la setlist potentielle et ajoute le concert en favori pour être alertée au lancement des ventes.

### Notes
- Les WTB et WTS sont à prix gratuit (comme un don) pour éviter d'avoir des problèmes légaux et pour simplifier l'application.
- Un favori active l'alerte de vente associée.
- Les alertes de nouveauté sont sur des artistes ou des salles au choix.

## Liste de données
- Concert
  - id / nom / date / salleId / artisteId / url / photos / saleStartDateTime
- Salle
  - id / nom / ville / pays
- Artiste
  - id / nom
- Utilisateur
  - id / pseudo / sns
- WT
  - userId / concertId / wtType (wtb / wts)
- Favori
  - userId / concertId
- Alerte
  - id / userId / cibleType / cibleId
- SyncTicketmaster
  - lastPublicVisibilityStartDateTime (ex: 2026-03-24T21:59:47Z)

## API Web
- https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/
Permet de récupérer périodiquement les nouveaux concerts, les plans, les dates, les salles.

- https://api.setlist.fm/docs/1.0/index.html
Setlist potentielle / par artiste (via attractions name)

## Description des requêtes
### Récupérer de nouveaux concerts
- `GET https://app.ticketmaster.com/discovery/v2/events.json?apikey=REDACTED&countryCode=DE&classificationName=music`
  - Filtre nouveauté : ajouter `publicVisibilityStartDateTime=YYYY-MM-DDTHH:MM:SSZ` (ex: `2026-03-24T21:59:47Z`)
  - Champs utiles :
    - event.id, event.name, event.url
    - event.images[].url
    - event.dates.start.localDate / localTime / dateTime
    - event.sales.public.startDateTime
    - venue: _embedded.venues[0].id / name / city.name / country.countryCode
    - artiste: _embedded.attractions[0].id / name

### Récupérer setlist d'un artiste
- `GET https://api.setlist.fm/rest/1.0/search/setlists?artistName={nomArtiste}`
  - Parser `eventDate` (dd-mm-yyyy), `tour`, `sets`, `id`
  - Filtrer : `tour` si possible présent, `eventDate` < aujourd'hui, `sets` non vide si possible
  - Prendre le concert le plus récent (max `eventDate`)
- `GET https://api.setlist.fm/rest/1.0/setlist/{setlistId}`
  - Récupérer `sets.song.name` pour la liste des titres

## Mise à jour
- publicVisibilityStartDateTime : filtre de nouveauté (SyncTicketmaster.lastPublicVisibilityStartDateTime)
- Récupération + insertion des concerts via Ticketmaster
- Vérification des alertes utilisateur vis à vis des salles / artistes
- Vérification des alertes de vente pour les concerts en favoris
- Récup setlist du concert si artiste dispo dans setlist.fm

## Requêtes
- POST /auth/register  
  créer un compte

- POST /auth/login  
  se connecter

- POST /auth/logout  
  se déconnecter

- GET /concerts?artisteId=...&salleId=...
  liste des concerts, avec filtre optionnel par artiste ou salle

- GET /concerts/{concertId}  
  détail d'un concert

- GET /concerts/{concertId}/sns  
  voir les SNS des gens qui vont au même concert

- GET /concerts/{concertId}/setlist  
  voir la setlist potentielle d'un concert

- POST ou DELETE /concerts/{concertId}/favoris  
  ajouter ou retirer un favori (active/désactive l'alerte de vente)

- POST ou DELETE /concerts/{concertId}/wt?type=wtb|wts  
  se mettre ou retirer en WTB ou WTS

- GET /concerts/{concertId}/wt  
  voir les WTB / WTS liés à un concert

- POST /alertes?cibleType=artiste|salle&cibleId=...
  créer une alerte de nouveauté

- DELETE /alertes/{alerteId}  
  retirer une alerte

- GET /me  
  voir le profil (sns, favoris, wt, alertes)

- PATCH /me  
  modifier le profil (sns)

- GET /artistes?search=...
  pour l'autocomplétion artistes -> id

- GET /salles?search=...
  pour l'autocomplétion salles -> id

## Flux du Système
- client -> serveur : requêtes HTTP
- serveur -> client : réponses JSON
- serveur -> Ticketmaster : récupération concerts
- serveur -> setlist.fm : récupération setlist
- serveur <-> BDD : lecture / écriture

## Description serveur
- Approche : ressources REST + jobs de synchronisation Ticketmaster et setlist.fm.
- Auth (`/auth`) : inscription, connexion, déconnexion.
- Concerts (`/concerts`) : liste/filtre, détail, setlist, SNS, WTB/WTS, favoris (alerte vente).
- Alertes (`/alertes`) : création d'alertes de nouveauté sur artistes ou salles.
- Profil & autocomplétion (`/me`, `/artistes`, `/salles`) : SNS, favoris/WT/alertes, noms pour autocomplétion.

## Description client
- Plan : application monopage avec sections Recherche, Fiche concert, Profil, Auth.
- Recherche : liste + filtres ; appels `GET /concerts?artisteId=...&salleId=...`, `GET /artistes?search=...`, `GET /salles?search=...`.
- Fiche concert : détails + setlist + SNS + WTB/WTS ; appels `GET /concerts/{concertId}`, `GET /concerts/{concertId}/setlist`, `GET /concerts/{concertId}/sns`, `GET /concerts/{concertId}/wt`, actions `POST/DELETE /concerts/{concertId}/favoris`, `POST/DELETE /concerts/{concertId}/wt`.
- Profil : infos utilisateur et SNS ; appels `GET /me`, `PATCH /me`, création d'alertes via `POST /alertes`, suppression via `DELETE /alertes/{alerteId}`.
- Auth : écrans inscription/connexion/déconnexion ; appels `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`.

## Autre Idées
- Notification push via Firebase
- Météo avant concert
- https://aviewfrommyseat.com/ pour récupèrer la liste des blocs d'un certain stade
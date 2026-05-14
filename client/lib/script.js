let user = null;
let artists = [];
let venues = [];
let concerts = [];
let selectedArtistID = "";
let selectedVenueID = "";
let selectedConcert = null;

$(function () {
	showView("searchView");
	checkSession();
	loadConcerts();
});

function api(method, url, data) {
	let options = {
		method: method,
		url: url,
		dataType: "json",
	};

	if (method == "GET") {
		options.data = data || {};
	} else if (data) {
		options.data = JSON.stringify(data);
		options.contentType = "application/json";
	}

	return $.ajax(options);
}

function setMessages(type, message) {
	if (!message) {
		$("#messages").html("");
		return;
	}

	$("#messages").html(`
		<div class="notification is-${type}">
			<button class="delete" onclick="setMessages('', '')"></button>
			${escapeHTML(message)}
		</div>
	`);
}

function escapeHTML(value) {
	return $("<div>").text(value == null ? "" : String(value)).html();
}

function field(obj, names, fallback) {
	for (let i = 0; i < names.length; i = i + 1) {
		if (obj && obj[names[i]] != null) {
			return obj[names[i]];
		}
	}
	return fallback;
}

function itemID(obj) {
	return field(obj, ["ID", "id"], "");
}

function showView(viewID) {
	$("#views").children().hide();
	$("#" + viewID).show();

	$("#viewButtons").children().addClass("is-outlined");
	$("#button-" + viewID).removeClass("is-outlined");

	if (viewID == "profileView") {
		renderProfile();
		loadPasskeys();
	}
}

// Compte
function checkSession() {
	api("GET", "/api/auth/me")
		.done(function (response) {
			user = response.user;
			toggleConnected(true);
		})
		.fail(function () {
			user = null;
			toggleConnected(false);
		});
}

function login() {
	let email = $("#email").val();
	let password = $("#password").val();

	api("POST", "/api/auth/login", { email: email, password: password })
		.done(function (response) {
			user = response.user;
			toggleConnected(true);
			setMessages("success", "Session ouverte.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Connexion impossible."));
		});
}

function register() {
	let email = $("#email").val();
	let password = $("#password").val();

	api("POST", "/api/auth/register", { email: email, password: password })
		.done(function (response) {
			user = response.user;
			toggleConnected(true);
			setMessages("success", "Compte créé.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Inscription impossible."));
		});
}

function logout() {
	$.ajax({ method: "POST", url: "/api/auth/logout" })
		.always(function () {
			user = null;
			toggleConnected(false);
			setMessages("success", "Session fermée.");
		});
}

function emailExists() {
	let email = $("#email").val();
	if (!email) return;

	api("GET", "/api/auth/email-exists", { email: email })
		.done(function (response) {
			if (response.exists) {
				$("#emailHelp").text("Email déjà enregistré.");
				$("#registerButton").prop("disabled", true);
				$("#loginButton").prop("disabled", false);
			} else {
				$("#emailHelp").text("Email disponible.");
				$("#registerButton").prop("disabled", false);
				$("#loginButton").prop("disabled", true);
			}
		});
}

function toggleConnected(connected) {
	$(".if-login").css("display", connected ? "" : "none");
	$(".if-not-login").css("display", connected ? "none" : "");
	$("#connectedEmail").text(user ? user.Email : "");
	renderProfile();
}

// Recherche
function searchArtists() {
	api("GET", "/api/artists", { search: $("#artistSearch").val() })
		.done(function (response) {
			artists = response || [];
			renderArtists();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Recherche artiste impossible."));
		});
}

function searchVenues() {
	api("GET", "/api/venues", { search: $("#venueSearch").val() })
		.done(function (response) {
			venues = response || [];
			renderVenues();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Recherche salle impossible."));
		});
}

function renderArtists() {
	$("#artistResults").html(`<option value="">Tous les artistes</option>`);
	for (let i = 0; i < artists.length; i = i + 1) {
		let artist = artists[i];
		$("#artistResults").append(`
			<option value="${escapeHTML(itemID(artist))}">
				${escapeHTML(artist.Name)} (#${escapeHTML(itemID(artist))})
			</option>
		`);
	}
	$("#artistResults").val(selectedArtistID);
}

function renderVenues() {
	$("#venueResults").html(`<option value="">Toutes les salles</option>`);
	for (let i = 0; i < venues.length; i = i + 1) {
		let venue = venues[i];
		$("#venueResults").append(`
			<option value="${escapeHTML(itemID(venue))}">
				${escapeHTML(venue.Name)} - ${escapeHTML(venue.City)} (#${escapeHTML(itemID(venue))})
			</option>
		`);
	}
	$("#venueResults").val(selectedVenueID);
}

function selectArtist() {
	selectedArtistID = $("#artistResults").val();
	loadConcerts();
}

function selectVenue() {
	selectedVenueID = $("#venueResults").val();
	loadConcerts();
}

function clearSearch() {
	selectedArtistID = "";
	selectedVenueID = "";
	$("#artistSearch").val("");
	$("#venueSearch").val("");
	artists = [];
	venues = [];
	renderArtists();
	renderVenues();
	loadConcerts();
}

function loadConcerts() {
	let params = {};
	if (selectedArtistID) params.artistID = selectedArtistID;
	if (selectedVenueID) params.venueID = selectedVenueID;

	api("GET", "/api/concerts", params)
		.done(function (response) {
			concerts = response || [];
			renderConcerts();
		})
		.fail(function (xhr) {
			$("#concertsList").html(`<tr><td colspan="5">${escapeHTML(errorText(xhr, "Impossible de charger les concerts."))}</td></tr>`);
		});
}

function renderConcerts() {
	if (concerts.length == 0) {
		$("#concertsList").html(`<tr><td colspan="5">Aucun concert.</td></tr>`);
		return;
	}

	$("#concertsList").html("");
	for (let i = 0; i < concerts.length; i = i + 1) {
		let concert = concerts[i];
		let id = itemID(concert);
		$("#concertsList").append(`
			<tr>
				<td>${escapeHTML(concert.Name)}</td>
				<td>${escapeHTML(formatDate(concert.Date))}</td>
				<td>${escapeHTML(nameFromList(venues, concert.VenueID, "salle"))}</td>
				<td>${escapeHTML(nameFromList(artists, concert.ArtistID, "artiste"))}</td>
				<td><button class="button is-info is-small" type="button" onclick="openConcert('${escapeHTML(id)}')">Ouvrir</button></td>
			</tr>
		`);
	}
}

function openConcert(id) {
	api("GET", "/api/concerts/" + id)
		.done(function (response) {
			selectedConcert = response;
			renderConcert();
			showView("detailView");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Concert introuvable."));
		});
}

function renderConcert() {
	if (!selectedConcert) {
		$("#noConcertText").show();
		$("#concertDetails").hide();
		return;
	}

	$("#noConcertText").hide();
	$("#concertDetails").show();
	$("#detailName").text(selectedConcert.Name);
	$("#detailDate").text(formatDate(selectedConcert.Date));
	$("#detailSale").text(formatDate(selectedConcert.SaleStartDateTime) || "Non renseignée");
	$("#detailVenue").text(nameFromList(venues, selectedConcert.VenueID, "salle"));
	$("#detailArtist").text(nameFromList(artists, selectedConcert.ArtistID, "artiste"));
	$("#detailURL").attr("href", selectedConcert.URL || "#");
	$("#detailSeatmap").attr("href", selectedConcert.SeatmapURL || "#");

	let photo = "lib/concert.png";
	if (selectedConcert.Photos && selectedConcert.Photos.length > 0) {
		photo = selectedConcert.Photos[0];
	}
	$("#detailPhoto").attr("src", photo);
}

// Profil et passkeys
function renderProfile() {
	if (!user) {
		$("#profileID").text("");
		$("#profileEmail").text("");
		$("#passkeysList").html("");
		return;
	}

	$("#profileID").text(user.ID);
	$("#profileEmail").text(user.Email);
}

function loadPasskeys() {
	if (!user) return;

	api("GET", "/api/auth/passkeys")
		.done(function (response) {
			let passkeys = response.passkeys || [];
			$("#passkeysList").html("");

			if (passkeys.length == 0) {
				$("#passkeysList").append(`<span class="tag is-light">Aucune passkey</span>`);
				return;
			}

			for (let i = 0; i < passkeys.length; i = i + 1) {
				let passkey = passkeys[i];
				let id = passkey.CredentialID;
				$("#passkeysList").append(`
					<span class="tag is-info is-light">
						${escapeHTML(id.substring(0, 18))}...
						<button class="delete is-small" onclick="deletePasskey('${escapeHTML(id)}')"></button>
					</span>
				`);
			}
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Impossible de lister les passkeys."));
		});
}

function deletePasskey(id) {
	api("DELETE", "/api/auth/passkeys/" + encodeURIComponent(id))
		.done(function () {
			loadPasskeys();
			setMessages("success", "Passkey supprimée.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Suppression passkey impossible."));
		});
}

function registerPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "Navigateur incompatible avec les passkeys.");
		return;
	}

	api("POST", "/api/auth/passkeys/register/begin")
		.done(function (options) {
			options.publicKey.challenge = base64ToBuffer(options.publicKey.challenge);
			options.publicKey.user.id = base64ToBuffer(options.publicKey.user.id);
			navigator.credentials.create(options)
				.then(function (credential) {
					return api("POST", "/api/auth/passkeys/register/finish", credentialToJSON(credential));
				})
				.then(function () {
					loadPasskeys();
					setMessages("success", "Passkey ajoutée.");
				})
				.catch(function () {
					setMessages("danger", "Création passkey annulée ou refusée.");
				});
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Démarrage passkey impossible."));
		});
}

function loginPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "Navigateur incompatible avec les passkeys.");
		return;
	}

	api("POST", "/api/auth/passkeys/login/begin")
		.done(function (options) {
			options.publicKey.challenge = base64ToBuffer(options.publicKey.challenge);
			if (options.publicKey.allowCredentials) {
				for (let i = 0; i < options.publicKey.allowCredentials.length; i = i + 1) {
					options.publicKey.allowCredentials[i].id = base64ToBuffer(options.publicKey.allowCredentials[i].id);
				}
			}
			navigator.credentials.get(options)
				.then(function (credential) {
					return api("POST", "/api/auth/passkeys/login/finish", credentialToJSON(credential));
				})
				.then(function (response) {
					user = response.user;
					toggleConnected(true);
					setMessages("success", "Session passkey ouverte.");
				})
				.catch(function () {
					setMessages("danger", "Connexion passkey annulée ou refusée.");
				});
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Démarrage passkey impossible."));
		});
}

function credentialToJSON(credential) {
	let response = credential.response;
	let json = {
		id: credential.id,
		rawId: bufferToBase64(credential.rawId),
		type: credential.type,
		response: {},
	};

	if (response.clientDataJSON) json.response.clientDataJSON = bufferToBase64(response.clientDataJSON);
	if (response.attestationObject) json.response.attestationObject = bufferToBase64(response.attestationObject);
	if (response.authenticatorData) json.response.authenticatorData = bufferToBase64(response.authenticatorData);
	if (response.signature) json.response.signature = bufferToBase64(response.signature);
	if (response.userHandle) json.response.userHandle = bufferToBase64(response.userHandle);

	return json;
}

function base64ToBuffer(value) {
	let base64 = value.replace(/-/g, "+").replace(/_/g, "/");
	let raw = window.atob(base64);
	let buffer = new ArrayBuffer(raw.length);
	let bytes = new Uint8Array(buffer);
	for (let i = 0; i < raw.length; i = i + 1) {
		bytes[i] = raw.charCodeAt(i);
	}
	return buffer;
}

function bufferToBase64(buffer) {
	let bytes = new Uint8Array(buffer);
	let text = "";
	for (let i = 0; i < bytes.length; i = i + 1) {
		text = text + String.fromCharCode(bytes[i]);
	}
	return window.btoa(text).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

// Petits utilitaires
function nameFromList(list, id, label) {
	for (let i = 0; i < list.length; i = i + 1) {
		if (String(itemID(list[i])) == String(id)) {
			return list[i].Name;
		}
	}
	return label + " #" + id;
}

function formatDate(value) {
	if (!value || value == "0001-01-01T00:00:00Z") return "";
	let date = new Date(value);
	if (Number.isNaN(date.getTime())) return String(value);
	return date.toLocaleString("fr-FR", {
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
	});
}

function errorText(xhr, fallback) {
	if (xhr && xhr.responseText) {
		return xhr.responseText.trim();
	}
	return fallback;
}

function todoFeature(name) {
	setMessages("warning", name + " : route prévue dans le README, pas encore implémentée côté backend.");
}

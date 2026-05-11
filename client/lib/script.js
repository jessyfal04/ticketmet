const state = {
	artists: [],
	venues: [],
	concerts: [],
	currentConcert: null,
	selectedArtistID: "",
	selectedVenueID: "",
	connected: false,
	profile: null,
	localSession: false,
};

$(function () {
	const savedProfile = readLocalProfile();
	if (savedProfile) {
		state.profile = savedProfile;
		toggleConnected(true, savedProfile.pseudo || "local");
	}

	loadLookups();
	loadConcerts();
});

function requestJSON(method, urls, data = {}) {
	const list = Array.isArray(urls) ? urls : [urls];
	const deferred = $.Deferred();
	let index = 0;

	function attempt() {
		$.ajax({
			method: method,
			url: list[index],
			data: data,
			dataType: "json",
		})
			.done(function (payload) {
				deferred.resolve(payload);
			})
			.fail(function (xhr) {
				index += 1;
				if (index < list.length) {
					attempt();
				} else {
					deferred.reject(xhr);
				}
			});
	}

	attempt();
	return deferred.promise();
}

function setMessages(type, message) {
	if (!type || !message) {
		$("#messages").empty();
		return;
	}

	$("#messages").html(`
		<div class="notification is-${type}">
			<button class="delete" onclick="setMessages(null, null)"></button>
			${escapeHTML(message)}
		</div>
	`);
}

function escapeHTML(value) {
	return $("<div>").text(value == null ? "" : String(value)).html();
}

function field(obj, names, fallback = "") {
	for (let i = 0; i < names.length; i += 1) {
		if (obj && obj[names[i]] != null) {
			return obj[names[i]];
		}
	}
	return fallback;
}

function itemID(obj) {
	return field(obj, ["ID", "id"], "");
}

function splitSNS(value) {
	return String(value || "")
		.split(",")
		.map((entry) => entry.trim())
		.filter((entry) => entry.length > 0);
}

function readLocalProfile() {
	try {
		const raw = localStorage.getItem("ticketmet.profile");
		return raw ? JSON.parse(raw) : null;
	} catch (e) {
		return null;
	}
}

function saveLocalProfile() {
	if (state.profile) {
		localStorage.setItem("ticketmet.profile", JSON.stringify(state.profile));
	}
}

function profileOrDefault(pseudo, sns) {
	return {
		id: Date.now(),
		pseudo: pseudo,
		sns: sns,
		favoris: [],
		wts: [],
		alerts: [],
	};
}

function normalizeProfile(payload, fallback) {
	const data = payload && (payload.data || payload.user || payload.User || payload);
	return {
		id: field(data, ["ID", "id"], field(fallback, ["id"], Date.now())),
		pseudo: field(data, ["Pseudo", "pseudo"], field(fallback, ["pseudo"], "")),
		sns: field(data, ["SNS", "sns"], field(fallback, ["sns"], [])),
		favoris: field(data, ["Favoris", "favoris"], field(fallback, ["favoris"], [])),
		wts: field(data, ["WTs", "wts", "WT", "wt"], field(fallback, ["wts"], [])),
		alerts: field(data, ["Alerts", "alerts", "Alertes", "alertes"], field(fallback, ["alerts"], [])),
	};
}

function toggleConnected(connected, pseudo = "") {
	state.connected = connected;
	$(".if-login").css("display", connected ? "" : "none");
	$(".if-not-login").css("display", connected ? "none" : "");
	$("#txt-connexion-pseudo").text(pseudo);

	if (connected) {
		loadProfile();
	} else {
		state.profile = null;
		renderProfile();
	}
}

function authPayload() {
	return {
		pseudo: $("#authPseudo").val().trim(),
		password: $("#authPassword").val(),
		sns: JSON.stringify(splitSNS($("#authSNS").val())),
	};
}

function register() {
	authAction("register");
}

function login() {
	authAction("login");
}

function authAction(action) {
	const payload = authPayload();
	if (!payload.pseudo) {
		setMessages("danger", "Pseudo obligatoire.");
		return;
	}

	requestJSON("POST", `/auth/${action}`, payload)
		.done(function (response) {
			state.localSession = false;
			state.profile = normalizeProfile(response, profileOrDefault(payload.pseudo, JSON.parse(payload.sns)));
			saveLocalProfile();
			toggleConnected(true, state.profile.pseudo);
			setMessages("success", "Session ouverte.");
		})
		.fail(function () {
			state.localSession = true;
			state.profile = readLocalProfile() || profileOrDefault(payload.pseudo, JSON.parse(payload.sns));
			state.profile.pseudo = payload.pseudo;
			if (splitSNS($("#authSNS").val()).length > 0) {
				state.profile.sns = splitSNS($("#authSNS").val());
			}
			saveLocalProfile();
			toggleConnected(true, state.profile.pseudo);
			setMessages("warning", "Auth serveur indisponible : session locale activée.");
		});
}

function logout() {
	requestJSON("POST", "/auth/logout")
		.always(function () {
			state.localSession = false;
			toggleConnected(false);
			setMessages("success", "Session fermée.");
		});
}

function loadLookups() {
	requestJSON("GET", ["/artistes", "/artists"], { search: "" })
		.done(function (artists) {
			state.artists = Array.isArray(artists) ? artists : [];
			renderArtistOptions(state.artists);
		});

	requestJSON("GET", ["/salles", "/venues"], { search: "" })
		.done(function (venues) {
			state.venues = Array.isArray(venues) ? venues : [];
			renderVenueOptions(state.venues);
		});
}

function mergeByID(target, items) {
	const existing = {};
	target.forEach((item) => {
		existing[itemID(item)] = item;
	});
	items.forEach((item) => {
		existing[itemID(item)] = item;
	});
	return Object.values(existing);
}

function searchArtists() {
	requestJSON("GET", ["/artistes", "/artists"], { search: $("#artistSearch").val() })
		.done(function (artists) {
			const results = Array.isArray(artists) ? artists : [];
			state.artists = mergeByID(state.artists, results);
			renderArtistOptions(results);
		})
		.fail(function () {
			setMessages("danger", "Impossible de récupérer les artistes.");
		});
}

function searchVenues() {
	requestJSON("GET", ["/salles", "/venues"], { search: $("#venueSearch").val() })
		.done(function (venues) {
			const results = Array.isArray(venues) ? venues : [];
			state.venues = mergeByID(state.venues, results);
			renderVenueOptions(results);
		})
		.fail(function () {
			setMessages("danger", "Impossible de récupérer les salles.");
		});
}

function renderArtistOptions(artists) {
	const selected = state.selectedArtistID;
	$("#artistResults").html(`<option value="">Tous les artistes</option>`);
	artists.forEach((artist) => {
		const id = itemID(artist);
		const name = field(artist, ["Name", "name"], "");
		$("#artistResults").append(`<option value="${escapeHTML(id)}">${escapeHTML(name)} (#${escapeHTML(id)})</option>`);
	});
	$("#artistResults").val(selected);
}

function renderVenueOptions(venues) {
	const selected = state.selectedVenueID;
	$("#venueResults").html(`<option value="">Toutes les salles</option>`);
	venues.forEach((venue) => {
		const id = itemID(venue);
		const name = field(venue, ["Name", "name"], "");
		const city = field(venue, ["City", "city"], "");
		$("#venueResults").append(`<option value="${escapeHTML(id)}">${escapeHTML(name)} - ${escapeHTML(city)} (#${escapeHTML(id)})</option>`);
	});
	$("#venueResults").val(selected);
}

function selectArtist() {
	state.selectedArtistID = $("#artistResults").val();
	loadConcerts();
}

function selectVenue() {
	state.selectedVenueID = $("#venueResults").val();
	loadConcerts();
}

function clearSearch() {
	state.selectedArtistID = "";
	state.selectedVenueID = "";
	$("#artistSearch").val("");
	$("#venueSearch").val("");
	loadLookups();
	loadConcerts();
}

function loadConcerts() {
	const data = {};
	if (state.selectedArtistID) {
		data.artisteId = state.selectedArtistID;
		data.artistID = state.selectedArtistID;
	}
	if (state.selectedVenueID) {
		data.salleId = state.selectedVenueID;
		data.venueID = state.selectedVenueID;
	}

	requestJSON("GET", "/concerts", data)
		.done(function (concerts) {
			state.concerts = Array.isArray(concerts) ? concerts : [];
			renderConcerts();
		})
		.fail(function () {
			$("#concertsList").html(`<tr><td colspan="5">Impossible de charger les concerts.</td></tr>`);
		});
}

function renderConcerts() {
	if (state.concerts.length === 0) {
		$("#concertsList").html(`<tr><td colspan="5">Aucun concert.</td></tr>`);
		return;
	}

	$("#concertsList").empty();
	state.concerts.forEach((concert) => {
		const id = itemID(concert);
		const name = field(concert, ["Name", "name"], "");
		const date = formatDate(field(concert, ["Date", "date"], ""));
		const venueID = field(concert, ["VenueID", "venueID", "salleId"], "");
		const artistID = field(concert, ["ArtistID", "artistID", "artisteId"], "");

		$("#concertsList").append(`
			<tr>
				<td>${escapeHTML(name)}</td>
				<td>${escapeHTML(date)}</td>
				<td>${escapeHTML(venueName(venueID))}</td>
				<td>${escapeHTML(artistName(artistID))}</td>
				<td>
					<button class="button is-info is-small" type="button" onclick="openConcert('${escapeHTML(id)}')">Ouvrir</button>
				</td>
			</tr>
		`);
	});
}

function openConcert(id) {
	requestJSON("GET", `/concerts/${id}`)
		.done(function (concert) {
			state.currentConcert = concert;
			renderConcertDetail(concert);
			loadConcertExtras(id);
		})
		.fail(function () {
			const fallback = state.concerts.find((concert) => String(itemID(concert)) === String(id));
			if (!fallback) {
				setMessages("danger", "Concert introuvable.");
				return;
			}
			state.currentConcert = fallback;
			renderConcertDetail(fallback);
			loadConcertExtras(id);
		});
}

function renderConcertDetail(concert) {
	const id = itemID(concert);
	const name = field(concert, ["Name", "name"], "");
	const date = formatDate(field(concert, ["Date", "date"], ""));
	const saleStart = formatDate(field(concert, ["SaleStartDateTime", "saleStartDateTime"], ""));
	const venueID = field(concert, ["VenueID", "venueID", "salleId"], "");
	const artistID = field(concert, ["ArtistID", "artistID", "artisteId"], "");
	const url = field(concert, ["URL", "url"], "#");
	const photos = field(concert, ["Photos", "photos"], []);

	$("#noConcertText").hide();
	$("#concertDetails").show();
	$("#detailName").text(name);
	$("#detailDate").text(`Concert : ${date}${saleStart ? " | Vente : " + saleStart : ""}`);
	$("#detailVenue").text(`Salle : ${venueName(venueID)}`);
	$("#detailArtist").text(`Artiste : ${artistName(artistID)}`);
	$("#detailURL").attr("href", url || "#");
	$("#detailPhotos").empty();

	photos.forEach((photo) => {
		$("#detailPhotos").append(`<img src="${escapeHTML(photo)}" alt="${escapeHTML(name)}">`);
	});
	if (photos.length === 0) {
		$("#detailPhotos").append(`<div class="notification is-light">Pas de photo.</div>`);
	}

	$("#detailName").attr("data-concert-id", id);
}

function loadConcertExtras(id) {
	requestJSON("GET", `/concerts/${id}/setlist`)
		.done(function (payload) {
			renderSetlist(field(payload, ["Songs", "songs"], Array.isArray(payload) ? payload : []));
		})
		.fail(function () {
			renderSetlist([]);
		});

	requestJSON("GET", `/concerts/${id}/wt`)
		.done(function (payload) {
			renderWT(Array.isArray(payload) ? payload : field(payload, ["WTs", "wts"], []));
		})
		.fail(function () {
			const localWT = state.profile ? state.profile.wts.filter((entry) => String(field(entry, ["ConcertID", "concertID", "concertId"], "")) === String(id)) : [];
			renderWT(localWT);
		});

	requestJSON("GET", `/concerts/${id}/sns`)
		.done(function (payload) {
			renderSNS(Array.isArray(payload) ? payload : field(payload, ["SNS", "sns", "users"], []));
		})
		.fail(function () {
			const sns = state.profile && isFavori(id) ? state.profile.sns : [];
			renderSNS(sns);
		});
}

function renderSetlist(songs) {
	$("#setlist").empty();
	if (!songs || songs.length === 0) {
		$("#setlist").append(`<li>Setlist indisponible.</li>`);
		return;
	}
	songs.forEach((song) => {
		$("#setlist").append(`<li>${escapeHTML(song)}</li>`);
	});
}

function renderWT(entries) {
	$("#wtList").empty();
	if (!entries || entries.length === 0) {
		$("#wtList").append(`<span class="tag is-light">Aucun WT</span>`);
		return;
	}
	entries.forEach((entry) => {
		const type = field(entry, ["Type", "type", "wtType"], "");
		const userID = field(entry, ["UserID", "userID", "userId"], "");
		$("#wtList").append(`<span class="tag is-link is-light">${escapeHTML(type)} utilisateur #${escapeHTML(userID)}</span>`);
	});
}

function renderSNS(entries) {
	$("#snsList").empty();
	if (!entries || entries.length === 0) {
		$("#snsList").append(`<li>Aucun SNS visible.</li>`);
		return;
	}
	entries.forEach((entry) => {
		if (typeof entry === "string") {
			$("#snsList").append(`<li>${escapeHTML(entry)}</li>`);
			return;
		}
		const pseudo = field(entry, ["Pseudo", "pseudo"], "utilisateur");
		const sns = field(entry, ["SNS", "sns"], []);
		$("#snsList").append(`<li>${escapeHTML(pseudo)} : ${escapeHTML(sns.join(", "))}</li>`);
	});
}

function addFavori() {
	const id = currentConcertID();
	if (!requireLogin() || !id) return;

	requestJSON("POST", `/concerts/${id}/favoris`)
		.done(function () {
			setMessages("success", "Favori ajouté.");
			loadProfile();
		})
		.fail(function () {
			if (!isFavori(id)) {
				state.profile.favoris.push({ concertId: id });
			}
			saveLocalProfile();
			renderProfile();
			loadConcertExtras(id);
			setMessages("warning", "Endpoint favoris indisponible : favori conservé localement.");
		});
}

function removeFavori() {
	const id = currentConcertID();
	if (!requireLogin() || !id) return;

	requestJSON("DELETE", `/concerts/${id}/favoris`)
		.done(function () {
			setMessages("success", "Favori retiré.");
			loadProfile();
		})
		.fail(function () {
			state.profile.favoris = state.profile.favoris.filter((entry) => String(field(entry, ["ConcertID", "concertID", "concertId"], entry)) !== String(id));
			saveLocalProfile();
			renderProfile();
			loadConcertExtras(id);
			setMessages("warning", "Endpoint favoris indisponible : favori retiré localement.");
		});
}

function setWT(type) {
	const id = currentConcertID();
	if (!requireLogin() || !id) return;

	requestJSON("POST", `/concerts/${id}/wt`, { type: type })
		.done(function () {
			setMessages("success", "WT mis à jour.");
			loadProfile();
			loadConcertExtras(id);
		})
		.fail(function () {
			state.profile.wts = state.profile.wts.filter((entry) => String(field(entry, ["ConcertID", "concertID", "concertId"], "")) !== String(id));
			state.profile.wts.push({ userId: state.profile.id, concertId: id, type: type });
			saveLocalProfile();
			renderProfile();
			loadConcertExtras(id);
			setMessages("warning", "Endpoint WT indisponible : WT conservé localement.");
		});
}

function removeWT() {
	const id = currentConcertID();
	if (!requireLogin() || !id) return;

	requestJSON("DELETE", `/concerts/${id}/wt`)
		.done(function () {
			setMessages("success", "WT retiré.");
			loadProfile();
			loadConcertExtras(id);
		})
		.fail(function () {
			state.profile.wts = state.profile.wts.filter((entry) => String(field(entry, ["ConcertID", "concertID", "concertId"], "")) !== String(id));
			saveLocalProfile();
			renderProfile();
			loadConcertExtras(id);
			setMessages("warning", "Endpoint WT indisponible : WT retiré localement.");
		});
}

function loadProfile() {
	if (!state.connected) return;

	requestJSON("GET", "/me")
		.done(function (payload) {
			state.profile = normalizeProfile(payload, state.profile);
			saveLocalProfile();
			renderProfile();
		})
		.fail(function () {
			state.profile = state.profile || readLocalProfile() || profileOrDefault($("#authPseudo").val() || "local", []);
			renderProfile();
		});
}

function patchMe() {
	if (!requireLogin()) return;

	const sns = splitSNS($("#profileSNS").val());
	requestJSON("PATCH", "/me", { sns: JSON.stringify(sns) })
		.done(function () {
			state.profile.sns = sns;
			saveLocalProfile();
			renderProfile();
			setMessages("success", "Profil mis à jour.");
		})
		.fail(function () {
			state.profile.sns = sns;
			saveLocalProfile();
			renderProfile();
			setMessages("warning", "Endpoint profil indisponible : SNS conservés localement.");
		});
}

function createAlert() {
	if (!requireLogin()) return;

	const cibleType = $("#alertTargetType").val();
	const cibleID = $("#alertTargetID").val();
	if (!cibleID) {
		setMessages("danger", "ID cible obligatoire.");
		return;
	}

	requestJSON("POST", "/alertes", { cibleType: cibleType, cibleId: cibleID })
		.done(function () {
			setMessages("success", "Alerte créée.");
			loadProfile();
		})
		.fail(function () {
			const alert = {
				alertId: Date.now(),
				userId: state.profile.id,
				cibleType: cibleType,
				cibleId: Number(cibleID),
			};
			state.profile.alerts.push(alert);
			saveLocalProfile();
			renderProfile();
			setMessages("warning", "Endpoint alertes indisponible : alerte conservée localement.");
		});
}

function deleteAlert() {
	if (!requireLogin()) return;

	const alertID = $("#deleteAlertID").val();
	if (!alertID) {
		setMessages("danger", "ID alerte obligatoire.");
		return;
	}

	requestJSON("DELETE", `/alertes/${alertID}`)
		.done(function () {
			setMessages("success", "Alerte supprimée.");
			loadProfile();
		})
		.fail(function () {
			state.profile.alerts = state.profile.alerts.filter((entry) => String(field(entry, ["AlertID", "alertID", "alertId", "id"], "")) !== String(alertID));
			saveLocalProfile();
			renderProfile();
			setMessages("warning", "Endpoint alertes indisponible : alerte supprimée localement.");
		});
}

function renderProfile() {
	if (!state.profile) {
		$("#profileSNS").val("");
		$("#profileFavoris").empty();
		$("#profileWT").empty();
		$("#profileAlerts").empty();
		return;
	}

	$("#profileSNS").val((state.profile.sns || []).join(", "));
	renderTags("#profileFavoris", state.profile.favoris, function (entry) {
		const id = field(entry, ["ConcertID", "concertID", "concertId"], entry);
		return `concert #${id}`;
	});
	renderTags("#profileWT", state.profile.wts, function (entry) {
		const id = field(entry, ["ConcertID", "concertID", "concertId"], "");
		const type = field(entry, ["Type", "type", "wtType"], "");
		return `${type} concert #${id}`;
	});
	renderTags("#profileAlerts", state.profile.alerts, function (entry) {
		const id = field(entry, ["AlertID", "alertID", "alertId", "id"], "");
		const cibleType = field(entry, ["CibleType", "cibleType"], "");
		const cibleID = field(entry, ["CibleID", "cibleID", "cibleId"], "");
		return `alerte #${id} ${cibleType} #${cibleID}`;
	});
}

function renderTags(selector, entries, labelFn) {
	$(selector).empty();
	if (!entries || entries.length === 0) {
		$(selector).append(`<span class="tag is-light">Vide</span>`);
		return;
	}
	entries.forEach((entry) => {
		$(selector).append(`<span class="tag is-info is-light">${escapeHTML(labelFn(entry))}</span>`);
	});
}

function currentConcertID() {
	if (!state.currentConcert) {
		setMessages("danger", "Aucun concert sélectionné.");
		return "";
	}
	return itemID(state.currentConcert);
}

function requireLogin() {
	if (!state.connected || !state.profile) {
		setMessages("danger", "Connexion requise.");
		return false;
	}
	return true;
}

function isFavori(concertID) {
	if (!state.profile) return false;
	return state.profile.favoris.some((entry) => String(field(entry, ["ConcertID", "concertID", "concertId"], entry)) === String(concertID));
}

function artistName(id) {
	const artist = state.artists.find((entry) => String(itemID(entry)) === String(id));
	return artist ? field(artist, ["Name", "name"], `artiste #${id}`) : `artiste #${id}`;
}

function venueName(id) {
	const venue = state.venues.find((entry) => String(itemID(entry)) === String(id));
	if (!venue) return `salle #${id}`;
	const name = field(venue, ["Name", "name"], "");
	const city = field(venue, ["City", "city"], "");
	return city ? `${name}, ${city}` : name;
}

function formatDate(value) {
	if (!value) return "";
	const parsed = new Date(value);
	if (Number.isNaN(parsed.getTime())) return String(value);
	return parsed.toLocaleString("fr-FR", {
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
	});
}

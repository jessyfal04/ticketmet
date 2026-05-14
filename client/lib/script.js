let user = null;
let artists = [];
let venues = [];
let concerts = [];
let selectedArtistID = "";
let selectedVenueID = "";
let selectedConcert = null;
let profileData = null;

$(function () {
	showView("searchView");
	checkHealth();
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

	$("#button-detailView").css("display", viewID == "detailView" ? "" : "none");

	if (viewID == "accountView") {
		renderAccount();
		if (user) {
			loadPasskeys();
			loadProfile();
		}
	}
}

// Account
function checkHealth() {
	$.ajax({ method: "GET", url: "/healthz" })
		.done(function () {
			$("#healthStatus").removeClass("is-warning is-danger").addClass("is-success").text("API online");
		})
		.fail(function () {
			$("#healthStatus").removeClass("is-warning is-success").addClass("is-danger").text("API offline");
		});
}

function checkSession() {
	api("GET", "/api/auth/me")
		.done(function (response) {
			user = response.User;
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
			user = response.User;
			toggleConnected(true);
			setMessages("success", "Signed in.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Sign in failed."));
		});
}

function register() {
	let email = $("#email").val();
	let password = $("#password").val();

	api("POST", "/api/auth/register", { email: email, password: password })
		.done(function (response) {
			user = response.User;
			toggleConnected(true);
			setMessages("success", "Account created.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Registration failed."));
		});
}

function logout() {
	$.ajax({ method: "POST", url: "/api/auth/logout" })
		.always(function () {
			user = null;
			toggleConnected(false);
			showView("accountView");
			setMessages("success", "Signed out.");
		});
}

function unregisterAccount() {
	if (!user) return;
	if (!window.confirm("Delete this account and all its data?")) return;

	api("DELETE", "/api/auth/unregister", { password: $("#unregisterPassword").val() })
		.done(function () {
			user = null;
			toggleConnected(false);
			showView("accountView");
			setMessages("success", "Account deleted.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Account deletion failed."));
		});
}

function emailExists() {
	let email = $("#email").val();
	if (!email) return;

	api("GET", "/api/auth/email-exists", { email: email })
		.done(function (response) {
			if (response.Exists) {
				$("#emailHelp").text("Email already registered.");
				$("#registerButton").prop("disabled", true);
				$("#loginButton").prop("disabled", false);
			} else {
				$("#emailHelp").text("Email available.");
				$("#registerButton").prop("disabled", false);
				$("#loginButton").prop("disabled", true);
			}
		});
}

function toggleConnected(connected) {
	$(".if-login").css("display", connected ? "" : "none");
	$(".if-not-login").css("display", connected ? "none" : "");
	$("#connectedEmail").text(user ? user.Email : "");
	if (!connected) {
		profileData = null;
	}
	renderAccount();
	if (connected) {
		loadPasskeys();
		loadProfile();
	}
}

function renderAccount() {
	if (!user) {
		$("#accountID").text("");
		$("#accountEmail").text("");
		$("#profileSNS").val("");
		$("#passkeysList").html("");
		$("#profileFavorites").html("");
		$("#profileWT").html("");
		$("#profileAlerts").html("");
		$("#unregisterPassword").val("");
		return;
	}

	$("#accountID").text(user.ID);
	$("#accountEmail").text(user.Email);
	$("#unregisterPassword").val("");
	renderProfileData();
}

function loadProfile() {
	if (!user) return;

	api("GET", "/api/me")
		.done(function (response) {
			profileData = response;
			renderProfileData();
			renderConcertActions();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Unable to load profile."));
		});
}

function saveSNS() {
	if (!user) return;

	let sns = $("#profileSNS").val().split("\n").map(function (value) {
		return value.trim();
	}).filter(function (value) {
		return value != "";
	});

	api("PATCH", "/api/me", { SNS: sns })
		.done(function (response) {
			profileData = response;
			renderProfileData();
			loadConcertFeatures();
			setMessages("success", "SNS saved.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Unable to save SNS."));
		});
}

function renderProfileData() {
	if (!profileData) {
		$("#profileFavorites").html(`<span class="has-text-grey">No data loaded.</span>`);
		$("#profileWT").html(`<span class="has-text-grey">No data loaded.</span>`);
		$("#profileAlerts").html(`<span class="has-text-grey">No data loaded.</span>`);
		return;
	}

	let sns = profileData.SNS || [];
	$("#profileSNS").val(sns.join("\n"));

	renderProfileConcertList("#profileFavorites", profileData.Favorites || [], "No favorites yet.");
	renderProfileWTList(profileData.WT || []);
	renderProfileAlerts(profileData.Alerts || []);
}

function renderProfileConcertList(target, items, emptyText) {
	if (items.length == 0) {
		$(target).html(`<span class="has-text-grey">${escapeHTML(emptyText)}</span>`);
		return;
	}

	let html = "<ul>";
	for (let i = 0; i < items.length; i = i + 1) {
		let concert = items[i];
		html += `<li><button class="button is-small is-ghost p-0" type="button" onclick="openConcert('${escapeHTML(itemID(concert))}')">${escapeHTML(concert.Name)}</button></li>`;
	}
	html += "</ul>";
	$(target).html(html);
}

function renderProfileWTList(items) {
	if (items.length == 0) {
		$("#profileWT").html(`<span class="has-text-grey">No WTB/WTS yet.</span>`);
		return;
	}

	let html = "<ul>";
	for (let i = 0; i < items.length; i = i + 1) {
		let item = items[i];
		let concert = item.Concert || {};
		html += `<li><span class="tag is-light">${escapeHTML(String(item.Type).toUpperCase())}</span> <button class="button is-small is-ghost p-0" type="button" onclick="openConcert('${escapeHTML(itemID(concert))}')">${escapeHTML(concert.Name)}</button></li>`;
	}
	html += "</ul>";
	$("#profileWT").html(html);
}

function renderProfileAlerts(items) {
	if (items.length == 0) {
		$("#profileAlerts").html(`<span class="has-text-grey">No alerts yet.</span>`);
		return;
	}

	let html = "";
	for (let i = 0; i < items.length; i = i + 1) {
		let alert = items[i];
		html += `
			<div class="tags has-addons mb-1">
				<span class="tag is-info is-light">${escapeHTML(alert.TargetType)} #${escapeHTML(alert.TargetID)}</span>
				<span class="tag">${escapeHTML(alert.TargetName || "target")}</span>
				<a class="tag is-delete" onclick="deleteAlert('${escapeHTML(alert.ID)}')"></a>
			</div>`;
	}
	$("#profileAlerts").html(html);
}


function createAlertFromSelection(targetType) {
	let targetID = targetType == "artist" ? selectedArtistID : selectedVenueID;
	if (!targetID) {
		setMessages("warning", "Select a " + targetType + " first.");
		return;
	}
	createAlert(targetType, targetID);
}

function createAlertForSelected(targetType) {
	if (!selectedConcert) return;
	let targetID = targetType == "artist" ? selectedConcert.ArtistID : selectedConcert.VenueID;
	createAlert(targetType, targetID);
}

function createAlert(targetType, targetID) {
	if (!user) {
		setMessages("warning", "Sign in first.");
		return;
	}

	api("POST", "/api/alerts?targetType=" + encodeURIComponent(targetType) + "&targetId=" + encodeURIComponent(targetID))
		.done(function () {
			loadProfile();
			setMessages("success", "Alert created.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Unable to create alert."));
		});
}

function deleteAlert(id) {
	api("DELETE", "/api/alerts/" + encodeURIComponent(id))
		.done(function () {
			loadProfile();
			setMessages("success", "Alert deleted.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Unable to delete alert."));
		});
}

// Search
function searchArtists() {
	api("GET", "/api/artists", { search: $("#artistSearch").val() })
		.done(function (response) {
			artists = response || [];
			renderArtists();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Artist search failed."));
		});
}

function searchVenues() {
	api("GET", "/api/venues", { search: $("#venueSearch").val() })
		.done(function (response) {
			venues = response || [];
			renderVenues();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Venue search failed."));
		});
}

function renderArtists() {
	$("#artistResults").html(`<option value="">All artists</option>`);
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
	$("#venueResults").html(`<option value="">All venues</option>`);
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
			$("#concertsList").html(`<tr><td colspan="5">${escapeHTML(errorText(xhr, "Unable to load concerts."))}</td></tr>`);
		});
}

function renderConcerts() {
	if (concerts.length == 0) {
		$("#concertsList").html(`<tr><td colspan="5">No concerts.</td></tr>`);
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
					<td>${escapeHTML(concert.VenueName || "Unknown venue")}</td>
					<td>${escapeHTML(concert.ArtistName || "Unknown artist")}</td>
					<td><button class="button is-info is-small" type="button" onclick="openConcert('${escapeHTML(id)}')">Open</button></td>
				</tr>
			`);
	}
}

function openConcert(id) {
	api("GET", "/api/concerts/" + id)
		.done(function (response) {
			selectedConcert = response;
			renderConcert();
			if (user) {
				loadProfile();
			}
			showView("detailView");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Concert not found."));
		});
}

function closeConcert() {
	selectedConcert = null;
	renderConcert();
	showView("searchView");
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
	$("#detailSale").text(formatDate(selectedConcert.SaleStartDateTime) || "Not provided");
	$("#detailVenue").text(selectedConcert.VenueName || "Unknown venue");
	$("#detailArtist").text(selectedConcert.ArtistName || "Unknown artist");
	$("#detailURL").attr("href", selectedConcert.URL || "#");
	$("#detailSeatmap").attr("href", selectedConcert.SeatmapURL || "#");

	let photo = "lib/concert.png";
	if (selectedConcert.Photos && selectedConcert.Photos.length > 0) {
		photo = selectedConcert.Photos[0];
	}
	$("#detailPhoto").attr("src", photo);

	renderConcertActions();
	loadConcertFeatures();
}

function loadConcertFeatures() {
	if (!selectedConcert) {
		$("#detailSetlist").html("");
		$("#detailSNS").html("");
		$("#detailWT").html("");
		return;
	}

	let id = encodeURIComponent(itemID(selectedConcert));

	api("GET", "/api/setlist/" + id)
		.done(function (response) {
			let songs = response.Songs || [];
			if (songs.length == 0) {
				$("#detailSetlist").html(`<span class="has-text-grey">No setlist available.</span>`);
				return;
			}
			let html = "<ol>";
			for (let i = 0; i < songs.length; i = i + 1) {
				html += `<li>${escapeHTML(songs[i])}</li>`;
			}
			html += "</ol>";
			$("#detailSetlist").html(html);
		})
		.fail(function (xhr) {
			$("#detailSetlist").html(escapeHTML(errorText(xhr, "Unable to load setlist.")));
		});

	api("GET", "/api/favorites/" + id)
		.done(function (response) {
			renderTags("#detailSNS", response.SNS || [], "No SNS yet.");
		})
		.fail(function (xhr) {
			$("#detailSNS").html(escapeHTML(errorText(xhr, "Unable to load SNS.")));
		});

	api("GET", "/api/wt/" + id)
		.done(function (response) {
			renderWT(response);
			renderConcertActions();
		})
		.fail(function (xhr) {
			$("#detailWT").html(escapeHTML(errorText(xhr, "Unable to load WTB/WTS.")));
		});
}

function renderWT(response) {
	let wtb = response.WTB || [];
	let wts = response.WTS || [];
	let html = `
		<p><strong>WTB:</strong> ${escapeHTML(wtb.length)}</p>
		<div class="tags mb-2">${tagsHTML(wtb, "No WTB SNS.")}</div>
		<p><strong>WTS:</strong> ${escapeHTML(wts.length)}</p>
		<div class="tags">${tagsHTML(wts, "No WTS SNS.")}</div>`;
	$("#detailWT").html(html);
}

function renderTags(target, values, emptyText) {
	$(target).html(tagsHTML(values, emptyText));
}

function tagsHTML(values, emptyText) {
	if (values.length == 0) {
		return `<span class="tag is-light">${escapeHTML(emptyText)}</span>`;
	}
	let html = "";
	for (let i = 0; i < values.length; i = i + 1) {
		html += `<span class="tag is-info is-light">${escapeHTML(values[i])}</span>`;
	}
	return html;
}

function addFavorite() {
	setFavorite("POST", "Favorite added.");
}

function deleteFavorite() {
	setFavorite("DELETE", "Favorite removed.");
}

function setFavorite(method, message) {
	if (!selectedConcert) return;
	api(method, "/api/favorites/" + encodeURIComponent(itemID(selectedConcert)))
		.done(function () {
			loadProfile();
			loadConcertFeatures();
			setMessages("success", message);
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Favorite update failed."));
		});
}

function addWT(type) {
	setWT("POST", type, String(type).toUpperCase() + " added.");
}

function deleteWT(type) {
	setWT("DELETE", type, String(type).toUpperCase() + " removed.");
}

function setWT(method, type, message) {
	if (!selectedConcert) return;
	api(method, "/api/wt/" + encodeURIComponent(itemID(selectedConcert)) + "?type=" + encodeURIComponent(type))
		.done(function () {
			loadProfile();
			loadConcertFeatures();
			setMessages("success", message);
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "WTB/WTS update failed."));
		});
}

function renderConcertActions() {
	if (!selectedConcert) return;

	let concertID = String(itemID(selectedConcert));
	let favorite = hasFavorite(concertID);
	$("#favoriteAddButton").css("display", favorite ? "none" : "");
	$("#favoriteDeleteButton").css("display", favorite ? "" : "none");

	let mineWTB = hasWT(concertID, "wtb");
	let mineWTS = hasWT(concertID, "wts");
	$("#wtbAddButton").css("display", mineWTB ? "none" : "");
	$("#wtbDeleteButton").css("display", mineWTB ? "" : "none");
	$("#wtsAddButton").css("display", mineWTS ? "none" : "");
	$("#wtsDeleteButton").css("display", mineWTS ? "" : "none");
}

function hasFavorite(concertID) {
	let favorites = profileData ? profileData.Favorites || [] : [];
	for (let i = 0; i < favorites.length; i = i + 1) {
		if (String(itemID(favorites[i])) == String(concertID)) return true;
	}
	return false;
}

function hasWT(concertID, type) {
	let items = profileData ? profileData.WT || [] : [];
	for (let i = 0; i < items.length; i = i + 1) {
		let concert = items[i].Concert || {};
		if (String(itemID(concert)) == String(concertID) && String(items[i].Type) == type) return true;
	}
	return false;
}

// Passkeys
function loadPasskeys() {
	if (!user) return;

	api("GET", "/api/auth/passkeys")
		.done(function (response) {
			let passkeys = response.Passkeys || [];
			$("#passkeysList").html("");

			if (passkeys.length == 0) {
				$("#passkeysList").append(`<span class="tag is-light">No passkeys</span>`);
				return;
			}

			for (let i = 0; i < passkeys.length; i = i + 1) {
				let passkey = passkeys[i];
				let id = passkey.CredentialID;
				let signCount = field(passkey, ["SignCount", "signCount"], 0);
				$("#passkeysList").append(`
					<div class="tag is-info is-light is-medium">
						<span title="${escapeHTML(id)}">${escapeHTML(id.substring(0, 18))}...</span>
						<span class="ml-2 has-text-grey">#${escapeHTML(signCount)}</span>
						<button class="delete is-small ml-2" type="button" onclick="deletePasskey('${escapeHTML(id)}')"></button>
					</div>
				`);
			}
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Unable to list passkeys."));
		});
}

function deletePasskey(id) {
	api("DELETE", "/api/auth/passkeys/" + encodeURIComponent(id))
		.done(function () {
			loadPasskeys();
			setMessages("success", "Passkey deleted.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Passkey deletion failed."));
		});
}

function registerPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "This browser does not support passkeys.");
		return;
	}

	api("POST", "/api/auth/passkeys/register/begin")
		.done(function (options) {
			let publicKey = options.publicKey;
			publicKey.challenge = base64ToBuffer(publicKey.challenge);
			publicKey.user.id = base64ToBuffer(publicKey.user.id);
			if (publicKey.excludeCredentials) {
				for (let i = 0; i < publicKey.excludeCredentials.length; i = i + 1) {
					publicKey.excludeCredentials[i].id = base64ToBuffer(publicKey.excludeCredentials[i].id);
				}
			}

			navigator.credentials.create(options)
				.then(function (credential) {
					return new Promise(function (resolve, reject) {
						api("POST", "/api/auth/passkeys/register/finish", credentialToJSON(credential))
							.done(resolve)
							.fail(function (xhr) {
								reject(new Error(errorText(xhr, "server validation refused")));
							});
					});
				})
				.then(function () {
					loadPasskeys();
					setMessages("success", "Passkey added.");
				})
				.catch(function (error) {
					setMessages("danger", "Passkey creation failed: " + passkeyError(error));
				});
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Unable to start passkey creation."));
		});
}

function loginPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "This browser does not support passkeys.");
		return;
	}

	api("POST", "/api/auth/passkeys/login/begin")
		.done(function (options) {
			let publicKey = options.publicKey;
			publicKey.challenge = base64ToBuffer(publicKey.challenge);
			if (publicKey.allowCredentials) {
				for (let i = 0; i < publicKey.allowCredentials.length; i = i + 1) {
					publicKey.allowCredentials[i].id = base64ToBuffer(publicKey.allowCredentials[i].id);
				}
			}

			navigator.credentials.get(options)
				.then(function (credential) {
					return new Promise(function (resolve, reject) {
						api("POST", "/api/auth/passkeys/login/finish", credentialToJSON(credential))
							.done(resolve)
							.fail(function (xhr) {
								reject(new Error(errorText(xhr, "server validation refused")));
							});
					});
					})
					.then(function (response) {
						user = response.User;
						toggleConnected(true);
						setMessages("success", "Signed in with passkey.");
				})
				.catch(function (error) {
					setMessages("danger", "Passkey sign in failed: " + passkeyError(error));
				});
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Unable to start passkey sign in."));
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
	while (base64.length % 4 != 0) {
		base64 = base64 + "=";
	}
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

// Small utilities
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
	return date.toLocaleString("en-US", {
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

function passkeyError(error) {
	if (error && error.message) {
		return error.message;
	}
	return "cancelled or refused by the browser";
}

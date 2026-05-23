let user = null;
let artists = [];
let venues = [];
let concerts = [];
let selectedArtistID = "";
let selectedVenueID = "";
let selectedCountry = "";
let selectedStatus = "future";
let selectedConcert = null;
let profileData = null;
let messageTimer = null;
let saveSNSDelay = null;
let concertPage = 1;
let concertHasMore = true;
let concertLoading = false;
let concertRequestNonce = 0;
let countryNames = { US: "United States Of America", AD: "Andorra", AI: "Anguilla", AR: "Argentina", AU: "Australia", AT: "Austria", AZ: "Azerbaijan", BS: "Bahamas", BH: "Bahrain", BB: "Barbados", BE: "Belgium", BM: "Bermuda", BR: "Brazil", BG: "Bulgaria", CA: "Canada", CL: "Chile", CN: "China", CO: "Colombia", CR: "Costa Rica", HR: "Croatia", CY: "Cyprus", CZ: "Czech Republic", DK: "Denmark", DO: "Dominican Republic", EC: "Ecuador", EE: "Estonia", FO: "Faroe Islands", FI: "Finland", FR: "France", GE: "Georgia", DE: "Germany", GH: "Ghana", GI: "Gibraltar", GB: "Great Britain", GR: "Greece", HK: "Hong Kong", HU: "Hungary", IS: "Iceland", IN: "India", IE: "Ireland", IL: "Israel", IT: "Italy", JM: "Jamaica", JP: "Japan", KR: "Korea, Republic of", LV: "Latvia", LB: "Lebanon", LT: "Lithuania", LU: "Luxembourg", MY: "Malaysia", MT: "Malta", MX: "Mexico", MC: "Monaco", ME: "Montenegro", MA: "Morocco", NL: "Netherlands", AN: "Netherlands Antilles", NZ: "New Zealand", ND: "Northern Ireland", NO: "Norway", PE: "Peru", PL: "Poland", PT: "Portugal", RO: "Romania", RU: "Russian Federation", LC: "Saint Lucia", SA: "Saudi Arabia", RS: "Serbia", SG: "Singapore", SK: "Slovakia", SI: "Slovenia", ZA: "South Africa", ES: "Spain", SE: "Sweden", CH: "Switzerland", TW: "Taiwan", TH: "Thailand", TT: "Trinidad and Tobago", TR: "Turkey", UA: "Ukraine", AE: "United Arab Emirates", UY: "Uruguay", VE: "Venezuela" };

$(function () {
	showView("searchView");
	checkHealth();
	checkSession();
	loadArtists();
	loadVenues();
	loadConcerts(true);
	$(window).on("scroll", handleConcertScroll);
	handleConcertQueryParam();
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

	let notification = $(`
		<div class="notification is-${type} mx-5" style="display:none;">
			<button class="delete" type="button"></button>
			${escapeHTML(message)}
		</div>
	`);
	notification.find(".delete").on("click", function () {
		notification.remove();
	});
	$("#messages").html("").append(notification);
	notification.fadeIn(150);
	window.clearTimeout(messageTimer);
	messageTimer = window.setTimeout(function () {
		notification.fadeOut(200, function () {
			notification.remove();
		});
	}, 5000);
}

function handleConcertQueryParam() {
	let concertID = new URLSearchParams(window.location.search).get("concert");
	if (!concertID) return;

	openConcert(concertID).done(function () {
		window.history.replaceState({}, document.title, window.location.pathname);
	});
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

function itemName(obj) {
	return field(obj, ["Name", "name"], "");
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
			$("#healthStatus").removeClass("is-warning is-danger").addClass("is-success").text("");
		})
		.fail(function () {
			$("#healthStatus").removeClass("is-warning is-success").addClass("is-danger").text("");
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
			if (handleAuthRequired(xhr)) return;
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
	renderAlertBlock();
	renderConcertActions();
	if (connected) {
		loadPasskeys();
		loadProfile();
	}
}

function handleAuthRequired(xhr) {
	if (!xhr || xhr.status != 401) {
		return false;
	}

	let wasConnected = !!user;
	user = null;
	toggleConnected(false);
	if (wasConnected) {
		setMessages("warning", "Disconnected.");
	}
	return true;
}

function renderAccount() {
	if (!user) {
		$("#accountID").text("");
		$("#accountEmail").text("");
		$("#profileSNS").val("");
		$("#profileSNS").removeClass("is-success");
		$("#profileSNS").removeClass("is-danger");
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
			renderAlertBlock();
			renderConcertActions();
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to load profile."));
		});
}

function saveSNS() {
	if (!user) return;
	window.clearTimeout(saveSNSDelay);

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
			$("#profileSNS").removeClass("is-danger");
			$("#profileSNS").addClass("is-success");
			setMessages("success", "SNS saved.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			$("#profileSNS").removeClass("is-success");
			$("#profileSNS").addClass("is-danger");
			setMessages("danger", errorText(xhr, "Unable to save SNS."));
		});
}

function profileSNSInput() {
	$("#profileSNS").removeClass("is-success");
	$("#profileSNS").addClass("is-danger");
	scheduleSaveSNS();
}

function scheduleSaveSNS() {
	if (!user) return;
	window.clearTimeout(saveSNSDelay);
	saveSNSDelay = window.setTimeout(saveSNS, 500);
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
	$("#profileSNS").removeClass("is-danger");
	$("#profileSNS").addClass("is-success");

	renderProfileConcertList("#profileFavorites", profileData.Favorites || [], "No favorites yet.");
	renderProfileWTList(profileData.WT || []);
	renderProfileAlerts(profileData.Alerts || []);
}

function renderAlertBlock() {
	let target = $("#alertBlock");
	if (!user) {
		target.html(`<span class="has-text-grey">Sign in to create artist or venue alerts.</span>`);
		return;
	}

	let artist = selectedArtistID ? selectedName(artists, selectedArtistID) : "All artists";
	let venue = selectedVenueID ? selectedName(venues, selectedVenueID) : "All venues";
	let artistAlert = getAlert("artist", selectedArtistID);
	let venueAlert = getAlert("venue", selectedVenueID);
	target.html(`
		<div class="content mb-3">
			<p class="mb-1"><strong>Artist:</strong> ${escapeHTML(artist)}</p>
			<p><strong>Venue:</strong> ${escapeHTML(venue)}</p>
		</div>
		<div class="buttons">
			<button class="button ${artistAlert ? "is-primary is-light" : "is-primary"}" type="button" ${selectedArtistID ? "" : "disabled"} onclick="createAlertFromSelection('artist')">${artistAlert ? "Remove artist alert" : "Alert this artist"}</button>
			<button class="button ${venueAlert ? "is-primary is-light" : "is-primary"}" type="button" ${selectedVenueID ? "" : "disabled"} onclick="createAlertFromSelection('venue')">${venueAlert ? "Remove venue alert" : "Alert this venue"}</button>
		</div>
	`);
}

function selectedName(items, id) {
	for (let i = 0; i < items.length; i = i + 1) {
		if (String(itemID(items[i])) == String(id)) {
			return itemName(items[i]);
		}
	}
	return "Selected #" + id;
}

function renderProfileConcertList(target, items, emptyText) {
	if (items.length == 0) {
		$(target).html(`<span class="has-text-grey">${escapeHTML(emptyText)}</span>`);
		return;
	}

	let html = "";
	for (let i = 0; i < items.length; i = i + 1) {
		let concert = items[i];
		html += `
			<div class="tags has-addons mb-1">
				<span class="tag is-primary is-light">Favorite</span>
				<button class="tag is-light" type="button" onclick="openConcert('${escapeHTML(itemID(concert))}')">${escapeHTML(concert.Name)}</button>
			</div>`;
	}
	$(target).html(html);
}

function renderProfileWTList(items) {
	if (items.length == 0) {
		$("#profileWT").html(`<span class="has-text-grey">No WTB/WTS yet.</span>`);
		return;
	}

	let html = "";
	for (let i = 0; i < items.length; i = i + 1) {
		let item = items[i];
		let concert = item.Concert || {};
		html += `
			<div class="tags has-addons mb-1">
				<span class="tag is-primary is-light">${escapeHTML(String(item.Type).toUpperCase())}</span>
				<button class="tag is-light" type="button" onclick="openConcert('${escapeHTML(itemID(concert))}')">${escapeHTML(concert.Name)}</button>
			</div>`;
	}
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
		let type = String(alert.TargetType || "target");
		type = type.charAt(0).toUpperCase() + type.slice(1);
		html += `
			<div class="tags has-addons mb-1">
				<span class="tag is-info is-light">${escapeHTML(type)}</span>
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
	toggleAlert(targetType, targetID);
}

function createAlertForSelected(targetType) {
	if (!selectedConcert) return;
	let targetID = targetType == "artist" ? selectedConcert.ArtistID : selectedConcert.VenueID;
	toggleAlert(targetType, targetID);
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
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to create alert."));
		});
}

function toggleAlert(targetType, targetID) {
	let alert = getAlert(targetType, targetID);
	if (alert) {
		deleteAlert(alert.ID);
		return;
	}
	createAlert(targetType, targetID);
}

function deleteAlert(id) {
	api("DELETE", "/api/alerts/" + encodeURIComponent(id))
		.done(function () {
			loadProfile();
			setMessages("success", "Alert deleted.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to delete alert."));
		});
}

function getAlert(targetType, targetID) {
	if (!profileData || !profileData.Alerts) return null;
	for (let i = 0; i < profileData.Alerts.length; i = i + 1) {
		let alert = profileData.Alerts[i];
		if (String(alert.TargetType) == String(targetType) && String(alert.TargetID) == String(targetID)) {
			return alert;
		}
	}
	return null;
}

// Lists
function loadArtists() {
	api("GET", "/api/artists")
		.done(function (response) {
			artists = response || [];
			renderArtists();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Artist search failed."));
		});
}

function loadVenues() {
	api("GET", "/api/venues")
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
	renderAlertBlock();
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
	renderAlertBlock();
	renderCountries();
}

function selectArtist() {
	selectedArtistID = $("#artistResults").val();
	renderAlertBlock();
	loadConcerts(true);
}

function selectVenue() {
	selectedVenueID = $("#venueResults").val();
	renderAlertBlock();
	loadConcerts(true);
}

function renderCountries() {
	let seen = {};
	let countries = [];
	for (let i = 0; i < venues.length; i = i + 1) {
		let country = field(venues[i], ["Country", "country"], "");
		if (!country || seen[country]) continue;
		seen[country] = true;
		countries.push(country);
	}
	countries.sort(function (left, right) {
		let leftName = countryNames[left] || left;
		let rightName = countryNames[right] || right;
		return leftName.localeCompare(rightName);
	});

	$("#countryResults").html(`<option value="">All countries</option>`);
	for (let i = 0; i < countries.length; i = i + 1) {
		let country = countries[i];
		$("#countryResults").append(`<option value="${escapeHTML(country)}">${escapeHTML(countryNames[country])}</option>`);
	}
	$("#countryResults").val(selectedCountry);
}

function selectCountry() {
	selectedCountry = $("#countryResults").val();
	loadConcerts(true);
}

function selectStatus() {
	selectedStatus = $("#statusResults").val() || "future";
	loadConcerts(true);
}

function clearSearch() {
	selectedArtistID = "";
	selectedVenueID = "";
	selectedCountry = "";
	selectedStatus = "future";
	renderArtists();
	renderVenues();
	renderCountries();
	$("#statusResults").val(selectedStatus);
	renderAlertBlock();
	loadConcerts(true);
}

function loadConcerts(reset) {
	if (reset) {
		concertRequestNonce = concertRequestNonce + 1;
		concertPage = 1;
		concertHasMore = true;
		concerts = [];
		concertLoading = false;
		renderConcerts();
	}
	if (concertLoading || !concertHasMore) return;

	let requestNonce = concertRequestNonce;
	let page = concertPage;
	let params = {};
	if (selectedArtistID) params.artistID = selectedArtistID;
	if (selectedVenueID) params.venueID = selectedVenueID;
	params.country = selectedCountry || "all";
	params.status = selectedStatus || "future";
	params.page = page;
	concertLoading = true;

	api("GET", "/api/concerts", params)
		.done(function (response) {
			if (requestNonce != concertRequestNonce) return;
			response = response || [];
			if (response.length == 0) {
				concertHasMore = false;
				return;
			}
			concerts = concerts.concat(response);
			concertPage = page + 1;
			renderConcerts();
		})
		.fail(function (xhr) {
			if (requestNonce != concertRequestNonce) return;
			$("#concertsGrid").html(`<div class="column is-12"><div class="notification is-danger is-light">${escapeHTML(errorText(xhr, "Unable to load concerts."))}</div></div>`);
		})
		.always(function () {
			if (requestNonce == concertRequestNonce) {
				concertLoading = false;
			}
		});
}

function renderConcerts() {
	if (concerts.length == 0) {
		$("#concertsGrid").html(`<div class="column is-12"><div class="notification is-light">No concerts.</div></div>`);
		return;
	}

	$("#concertsGrid").html("");
	for (let i = 0; i < concerts.length; i = i + 1) {
		let concert = concerts[i];
		let id = itemID(concert);
		let expired = isExpiredConcert(concert);
		let tags = renderConcertTags(concert);
		let photo = concertPhoto(concert);
		$("#concertsGrid").append(`
			<div class="column is-half-tablet is-one-third-desktop">
				<div class="card ticketmet-concert-card ${expired ? "ticketmet-concert-card--expired" : ""}">
					<div class="card-image">
						<figure class="image is-16by9">
							<img src="${escapeHTML(photo)}" alt="${escapeHTML(concert.Name || "Concert")}">
						</figure>
					</div>
					<div class="card-content">
						<p class="title is-5 mb-2">${escapeHTML(concert.Name)}</p>
						<p class="subtitle is-6 mb-3">${escapeHTML(concert.ArtistName || "Unknown artist")} · ${escapeHTML(concert.VenueName || "Unknown venue")}</p>
						<p class="is-size-7 mb-3">${escapeHTML(formatDate(concert.Date))}</p>
						<div class="tags">${tags}</div>
					</div>
					<footer class="card-footer">
						<button class="card-footer-item button is-white" type="button" onclick="openConcert('${escapeHTML(id)}')">Open</button>
					</footer>
				</div>
			</div>
		`);
	}
}

function loadMoreConcerts() {
	loadConcerts(false);
}

function handleConcertScroll() {
	if (concertLoading || !concertHasMore) return;
	let scrollBottom = $(window).scrollTop() + $(window).height();
	let pageBottom = $(document).height() - 200;
	if (scrollBottom >= pageBottom) {
		loadMoreConcerts();
	}
}

function openConcert(id) {
	return api("GET", "/api/concerts/" + id)
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

function shareConcert() {
	if (!selectedConcert) return;

	let link = new URL(window.location.href);
	link.search = "";
	link.searchParams.set("concert", itemID(selectedConcert));

	copyTextToClipboard(link.toString())
		.then(function () {
			setMessages("info", "Concert link copied.");
		})
		.catch(function () {
			setMessages("danger", "Unable to copy concert link.");
		});
}

function copyTextToClipboard(text) {
	if (navigator.clipboard && navigator.clipboard.writeText) {
		return navigator.clipboard.writeText(text);
	}

	let textarea = $("<textarea>").val(text).css({ position: "fixed", left: "-9999px", top: "0" }).appendTo("body");
	textarea[0].select();
	let ok = document.execCommand("copy");
	textarea.remove();
	return ok ? Promise.resolve() : Promise.reject(new Error("copy failed"));
}

function renderConcert() {
	if (!selectedConcert) {
		$("#noConcertText").show();
		$("#concertDetails").hide();
		$("#concertDetails").removeClass("ticketmet-concert-detail--expired");
		$("#detailDateStatus").html("");
		$("#detailSaleStatus").html("");
		$("#detailSeatmap").hide();
		return;
	}

	$("#noConcertText").hide();
	$("#concertDetails").show();
	$("#concertDetails").toggleClass("ticketmet-concert-detail--expired", isExpiredConcert(selectedConcert));
	$("#detailName").text(selectedConcert.Name);
	$("#detailDate").text(formatDate(selectedConcert.Date));
	$("#detailSale").text(formatDate(selectedConcert.SaleStartDateTime) || "Not provided");
	$("#detailVenue").text(selectedConcert.VenueName || "Unknown venue");
	$("#detailArtist").text(selectedConcert.ArtistName || "Unknown artist");
	$("#detailURL").attr("href", selectedConcert.URL || "#");
	$("#detailDateStatus").html(isExpiredConcert(selectedConcert)
		? `<span class="tag is-dark is-light">Past event</span>`
		: `<span class="tag is-success is-light">Upcoming</span>`);
	$("#detailSaleStatus").html("");
	let sale = field(selectedConcert, ["SaleStartDateTime", "saleStartDateTime"], "");
	if (sale) {
		let saleDate = new Date(sale);
		if (!Number.isNaN(saleDate.getTime()) && saleDate.getTime() > Date.now()) {
			$("#detailSaleStatus").html(`<span class="tag is-info is-light">Sale soon</span>`);
		} else if (!Number.isNaN(saleDate.getTime())) {
			$("#detailSaleStatus").html(`<span class="tag is-warning is-light">Sale open</span>`);
		}
	}
	if (selectedConcert.SeatmapURL) {
		$("#detailSeatmap").show().attr("href", selectedConcert.SeatmapURL);
	} else {
		$("#detailSeatmap").hide();
	}

	$("#detailPhoto").attr("src", concertPhoto(selectedConcert));

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
			if (handleAuthRequired(xhr)) return;
			$("#detailSNS").html(escapeHTML(errorText(xhr, "Unable to load SNS.")));
		});

	api("GET", "/api/wt/" + id)
		.done(function (response) {
			renderWT(response);
			renderConcertActions();
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			$("#detailWT").html(escapeHTML(errorText(xhr, "Unable to load WTB/WTS.")));
		});
}

function renderWT(response) {
	let wtb = response.WTB || [];
	let wts = response.WTS || [];
	let wtbCount = field(response, ["WTBCount", "wtbCount"], wtb.length);
	let wtsCount = field(response, ["WTSCount", "wtsCount"], wts.length);
	let expired = selectedConcert ? isExpiredConcert(selectedConcert) : false;
	let note = expired ? `<p class="has-text-grey mb-2">Past event, trade is disabled.</p>` : "";
	let html = `
		${note}
		<p><strong>WTB:</strong> ${escapeHTML(wtbCount)}</p>
		<div class="tags mb-2">${tagsHTML(wtb, "No WTB SNS.")}</div>
		<p><strong>WTS:</strong> ${escapeHTML(wtsCount)}</p>
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
			if (handleAuthRequired(xhr)) return;
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
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "WTB/WTS update failed."));
		});
}

function renderConcertActions() {
	if (!selectedConcert) return;

	let connected = !!user;
	$("#detailLoginText").css("display", connected ? "none" : "");
	if (!connected) {
		$("#favoriteAddButton").hide();
		$("#favoriteDeleteButton").hide();
		$("#wtbAddButton").hide();
		$("#wtbDeleteButton").hide();
		$("#wtsAddButton").hide();
		$("#wtsDeleteButton").hide();
		return;
	}

	let concertID = String(itemID(selectedConcert));
	let expired = isExpiredConcert(selectedConcert);
	let favorite = hasFavorite(concertID);
	$("#favoriteAddButton").css("display", favorite ? "none" : "");
	$("#favoriteDeleteButton").css("display", favorite ? "" : "none");
	$("#favoriteAddButton").prop("disabled", expired);
	$("#favoriteDeleteButton").prop("disabled", false);

	let mineWTB = hasWT(concertID, "wtb");
	let mineWTS = hasWT(concertID, "wts");
	$("#wtbAddButton").css("display", mineWTB ? "none" : "");
	$("#wtbDeleteButton").css("display", mineWTB ? "" : "none");
	$("#wtsAddButton").css("display", mineWTS ? "none" : "");
	$("#wtsDeleteButton").css("display", mineWTS ? "" : "none");
	$("#wtbAddButton").prop("disabled", expired);
	$("#wtbDeleteButton").prop("disabled", expired);
	$("#wtsAddButton").prop("disabled", expired);
	$("#wtsDeleteButton").prop("disabled", expired);
}

function isExpiredConcert(concert) {
	if (!concert) return false;
	let value = field(concert, ["Date", "date"], "");
	if (!value) return false;
	let date = new Date(value);
	if (Number.isNaN(date.getTime())) return false;
	return date.getTime() < Date.now();
}

function concertPhoto(concert) {
	if (concert && concert.Photos && concert.Photos.length > 0 && concert.Photos[0]) {
		return concert.Photos[0];
	}
	return "lib/concert.png";
}

function renderConcertTags(concert) {
	let html = "";
	if (isExpiredConcert(concert)) {
		html += `<span class="tag is-dark is-light">Past event</span>`;
	} else {
		html += `<span class="tag is-success is-light">Upcoming</span>`;
	}

	let sale = field(concert, ["SaleStartDateTime", "saleStartDateTime"], "");
	if (sale) {
		let saleDate = new Date(sale);
		if (!Number.isNaN(saleDate.getTime()) && saleDate.getTime() > Date.now()) {
			html += `<span class="tag is-info is-light">Sale soon</span>`;
		} else if (!Number.isNaN(saleDate.getTime())) {
			html += `<span class="tag is-warning is-light">Sale open</span>`;
		}
	}

	return html;
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
			if (handleAuthRequired(xhr)) return;
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
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Passkey deletion failed."));
		});
}

function registerPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "This browser does not support passkeys.");
		return;
	}

	setPasskeyLoading("#registerPasskeyButton", true);
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
								if (handleAuthRequired(xhr)) {
									reject(xhr); return;
								}
								reject(new Error(errorText(xhr, "server validation refused")));
							});
					});
				})
				.then(function () {
					setPasskeyLoading("#registerPasskeyButton", false);
					loadPasskeys();
					setMessages("success", "Passkey added.");
				})
				.catch(function (error) {
					setPasskeyLoading("#registerPasskeyButton", false);
					if (handleAuthRequired(error)) return;
					setMessages("danger", "Passkey creation failed: " + passkeyError(error));
				});
		})
		.fail(function (xhr) {
			setPasskeyLoading("#registerPasskeyButton", false);
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to start passkey creation."));
		});
}

function loginPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "This browser does not support passkeys.");
		return;
	}

	setPasskeyLoading("#loginPasskeyButton", true);
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
								if (handleAuthRequired(xhr)) {
									reject(xhr); return;
								}
								reject(new Error(errorText(xhr, "server validation refused")));
							});
					});
					})
					.then(function (response) {
						setPasskeyLoading("#loginPasskeyButton", false);
						user = response.User;
						toggleConnected(true);
						setMessages("success", "Signed in with passkey.");
				})
				.catch(function (error) {
					setPasskeyLoading("#loginPasskeyButton", false);
					if (handleAuthRequired(error)) return;
					setMessages("danger", "Passkey sign in failed: " + passkeyError(error));
				});
		})
		.fail(function (xhr) {
			setPasskeyLoading("#loginPasskeyButton", false);
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to start passkey sign in."));
		});
}

function setPasskeyLoading(buttonID, loading) {
	$(buttonID).toggleClass("is-loading", loading);
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

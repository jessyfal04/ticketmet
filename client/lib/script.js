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

// Send an API request with jQuery
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

// Show or clear the notification area
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

// Open the concert from the query string
function handleConcertQueryParam() {
	let concertID = new URLSearchParams(window.location.search).get("concert");
	if (!concertID) return;

	openConcert(concertID).done(function () {
		window.history.replaceState({}, document.title, window.location.pathname);
	});
}

// Escape HTML content
function escapeHTML(value) {
	return $("<div>").text(value == null ? "" : String(value)).html();
}

// Read the first matching field value
function field(obj, names, fallback) {
	for (let i = 0; i < names.length; i = i + 1) {
		if (obj && obj[names[i]] != null) {
			return obj[names[i]];
		}
	}
	return fallback;
}

// Read the item identifier
function itemID(obj) {
	return field(obj, ["ID", "id"], "");
}

// Read the item name
function itemName(obj) {
	return field(obj, ["Name", "name"], "");
}

// Show a view and update the buttons
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

// Handle auth required responses
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

// Format a date for display
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

// Read a friendly error message
function errorText(xhr, fallback) {
	if (xhr && xhr.responseText) {
		return xhr.responseText.trim();
	}
	return fallback;
}
